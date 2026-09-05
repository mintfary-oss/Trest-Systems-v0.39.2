package diff

import "testing"

func TestCompareSemantic(t *testing.T) {
	r := CompareSemantic([]SemanticElement{{GlobalID: "A", ElementType: "IfcWall", Name: "Old"}, {GlobalID: "B", ElementType: "IfcDoor"}}, []SemanticElement{{GlobalID: "A", ElementType: "IfcWall", Name: "New"}, {GlobalID: "C", ElementType: "IfcSlab"}})
	if len(r.Changed) != 1 || len(r.Added) != 1 || len(r.Removed) != 1 {
		t.Fatalf("bad result: %+v", r)
	}
}
