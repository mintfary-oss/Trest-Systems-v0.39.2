package ifcgeometry

import (
	"fmt"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

type GeometryInstance struct {
	ProductEntityID   int    `json:"product_entity_id"`
	ProductType       string `json:"product_type"`
	GeometryEntityID  int    `json:"geometry_entity_id"`
	GeometryType      string `json:"geometry_type"`
	PlacementEntityID int    `json:"placement_entity_id"`
	WorldTransform    Mat4   `json:"world_transform"`
	Mesh              Mesh   `json:"mesh"`
}

// ExtractProductGeometry resolves the common IFC chain Product -> ProductDefinitionShape
// -> ShapeRepresentation -> geometric item, then applies the full LocalPlacement chain.
func ExtractProductGeometry(entities []ifc.Entity, productID int) (GeometryInstance, error) {
	byID := indexEntities(entities)
	p, ok := byID[productID]
	if !ok {
		return GeometryInstance{}, fmt.Errorf("product #%d not found", productID)
	}
	if len(p.Attributes) < 6 {
		return GeometryInstance{}, fmt.Errorf("product #%d has no ProductDefinitionShape/Representation attributes", productID)
	}
	placementID := 0
	if strings.TrimSpace(p.Attributes[4]) != "$" {
		var err error
		placementID, err = ref(p.Attributes[4])
		if err != nil {
			return GeometryInstance{}, err
		}
	}
	repID, err := ref(p.Attributes[5])
	if err != nil {
		return GeometryInstance{}, fmt.Errorf("product #%d representation: %w", productID, err)
	}
	shapeReps := resolveRepresentations(byID, repID)
	for _, srID := range shapeReps {
		sr, ok := byID[srID]
		if !ok || len(sr.Attributes) < 4 {
			continue
		}
		for _, gid := range refs(sr.Attributes[3]) {
			ge, ok := byID[gid]
			if !ok {
				continue
			}
			mesh, err := Extract(entities, gid)
			if err != nil {
				continue
			}
			tr := Identity()
			if placementID != 0 {
				x, e := ResolvePlacement(entities, placementID)
				if e != nil {
					return GeometryInstance{}, e
				}
				tr = x.Matrix
			}
			for i := range mesh.Vertices {
				mesh.Vertices[i] = tr.Point(mesh.Vertices[i])
			}
			return GeometryInstance{ProductEntityID: productID, ProductType: p.Type, GeometryEntityID: gid, GeometryType: ge.Type, PlacementEntityID: placementID, WorldTransform: tr, Mesh: mesh}, nil
		}
	}
	return GeometryInstance{}, fmt.Errorf("product #%d has no supported geometry item", productID)
}

func resolveRepresentations(byID map[int]ifc.Entity, id int) []int {
	e, ok := byID[id]
	if !ok {
		return nil
	}
	if strings.EqualFold(e.Type, "IFCPRODUCTDEFINITIONSHAPE") {
		if len(e.Attributes) < 1 {
			return nil
		}
		return refs(e.Attributes[len(e.Attributes)-1])
	}
	if strings.EqualFold(e.Type, "IFCSHAPEREPRESENTATION") {
		return []int{id}
	}
	return nil
}
