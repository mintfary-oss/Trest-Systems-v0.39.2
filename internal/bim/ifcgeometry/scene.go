package ifcgeometry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

type ProductMeshRange struct {
	GlobalID        string `json:"global_id"`
	ProductEntityID int    `json:"product_entity_id"`
	TriangleStart   int    `json:"triangle_start"`
	TriangleCount   int    `json:"triangle_count"`
}

type Scene struct {
	Vertices []Vec3             `json:"vertices"`
	Indices  []uint32           `json:"indices"`
	Ranges   []ProductMeshRange `json:"ranges"`
}

// BuildSceneGeometry extracts supported geometry for products and concatenates it
// deterministically by product entity ID. Ranges provide stable GlobalId -> triangles mapping.
func BuildSceneGeometry(entities []ifc.Entity, productIDs []int) (Scene, error) {
	ids := append([]int(nil), productIDs...)
	sort.Ints(ids)
	var s Scene
	seen := map[string]bool{}
	for _, id := range ids {
		g, err := ExtractProductGeometry(entities, id)
		if err != nil {
			continue
		}
		p := findEntity(entities, id)
		gid := productGlobalID(p)
		if gid == "" {
			gid = fmt.Sprintf("#%d", id)
		}
		if seen[gid] {
			return Scene{}, fmt.Errorf("duplicate GlobalId %q", gid)
		}
		seen[gid] = true
		vertexBase := uint32(len(s.Vertices))
		triStart := len(s.Indices) / 3
		s.Vertices = append(s.Vertices, g.Mesh.Vertices...)
		for _, idx := range g.Mesh.Indices {
			s.Indices = append(s.Indices, vertexBase+idx)
		}
		s.Ranges = append(s.Ranges, ProductMeshRange{GlobalID: gid, ProductEntityID: id, TriangleStart: triStart, TriangleCount: len(g.Mesh.Indices) / 3})
	}
	if len(s.Indices) == 0 {
		return Scene{}, fmt.Errorf("no supported product geometry extracted")
	}
	return s, nil
}
func findEntity(es []ifc.Entity, id int) ifc.Entity {
	for _, e := range es {
		if e.ID == id {
			return e
		}
	}
	return ifc.Entity{}
}
func productGlobalID(e ifc.Entity) string {
	if len(e.Attributes) == 0 {
		return ""
	}
	v := strings.TrimSpace(e.Attributes[0])
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return strings.ReplaceAll(v[1:len(v)-1], "''", "'")
	}
	if v == "$" || v == "*" {
		return ""
	}
	return v
}
