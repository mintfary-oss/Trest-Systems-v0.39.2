package ifcsemantic

import (
	"fmt"
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
)

type BIMElementMapping struct {
	ExternalID       string         `json:"external_id"`
	ElementType      string         `json:"element_type"`
	Name             string         `json:"name"`
	Properties       map[string]any `json:"properties"`
	IFCEntityID      int            `json:"ifc_entity_id"`
	IFCGlobalID      string         `json:"ifc_global_id"`
	ParentExternalID string         `json:"parent_external_id"`
	RelationType     string         `json:"relation_type"`
}

// MapToBIMElements converts the schema-aware core subset into the neutral bim_elements shape.
// It intentionally keeps raw IFC attributes for traceability and does not invent geometry.
func MapToBIMElements(entities []ifc.Entity) []BIMElementMapping {
	m := Build(entities)
	rels := CoreRelations(entities)
	parent := map[int]CoreRelation{}
	for _, r := range rels {
		if _, ok := parent[r.SubjectID]; !ok {
			parent[r.SubjectID] = r
		}
	}
	out := make([]BIMElementMapping, 0)
	for _, e := range m.CoreElements() {
		c, _ := ClassOf(e.Type)
		p := map[string]any{"ifc_type": e.Type, "ifc_attributes": append([]string(nil), e.RawAttributes...)}
		x := BIMElementMapping{ExternalID: fmt.Sprintf("ifc:%d", e.ID), ElementType: string(c), Name: e.Name, Properties: p, IFCEntityID: e.ID, IFCGlobalID: e.GlobalID}
		if r, ok := parent[e.ID]; ok {
			x.ParentExternalID = fmt.Sprintf("ifc:%d", r.ObjectID)
			x.RelationType = r.Kind
		}
		out = append(out, x)
	}
	return out
}

// SpatialPath returns the chain of known aggregate/containment parents for an entity.
func SpatialPath(entities []ifc.Entity, id int) []Element {
	m := Build(entities)
	rels := CoreRelations(entities)
	cur := id
	seen := map[int]bool{}
	out := []Element{}
	for !seen[cur] {
		seen[cur] = true
		e, ok := m.Get(cur)
		if !ok {
			break
		}
		out = append(out, e)
		next := 0
		found := false
		for _, r := range rels {
			if r.SubjectID == cur {
				next = r.ObjectID
				found = true
				break
			}
		}
		if !found {
			break
		}
		cur = next
	}
	return out
}

func NormalizeType(t string) string { return strings.ToLower(strings.TrimSpace(t)) }
