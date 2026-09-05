package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
)

type CoreRelation struct {
	SubjectID int
	ObjectID  int
	Kind      string
}

// CoreRelations maps common IFC aggregation/containment relations. For IfcRelAggregates
// the first reference is the parent and all remaining references are children.
// For IfcRelContainedInSpatialStructure the final reference is the spatial container.
func CoreRelations(entities []ifc.Entity) []CoreRelation {
	out := make([]CoreRelation, 0)
	for _, e := range entities {
		refs := referencedIDs(e.Attributes)
		switch strings.ToUpper(e.Type) {
		case "IFCRELAGGREGATES":
			if len(refs) >= 2 {
				for _, child := range refs[1:] {
					out = append(out, CoreRelation{SubjectID: child, ObjectID: refs[0], Kind: "aggregates"})
				}
			}
		case "IFCRELCONTAINEDINSPATIALSTRUCTURE":
			if len(refs) >= 2 {
				container := refs[len(refs)-1]
				for _, child := range refs[:len(refs)-1] {
					out = append(out, CoreRelation{SubjectID: child, ObjectID: container, Kind: "contained_in_spatial_structure"})
				}
			}
		}
	}
	return out
}

func referencedIDs(attrs []string) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, a := range attrs {
		for i := 0; i < len(a); i++ {
			if a[i] != '#' {
				continue
			}
			j := i + 1
			for j < len(a) && a[j] >= '0' && a[j] <= '9' {
				j++
			}
			if j == i+1 {
				continue
			}
			id := 0
			for k := i + 1; k < j; k++ {
				id = id*10 + int(a[k]-'0')
			}
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
			i = j - 1
		}
	}
	return out
}
