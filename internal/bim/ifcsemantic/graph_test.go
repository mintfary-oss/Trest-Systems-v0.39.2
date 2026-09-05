package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"testing"
)

func TestBuildGraph(t *testing.T) {
	es := []ifc.Entity{{ID: 1, Type: "IFCREL", Attributes: []string{"#2", "(#3,#2)"}}, {ID: 2, Type: "IFCWALL", Attributes: nil}, {ID: 3, Type: "IFCDOOR", Attributes: nil}}
	g := BuildGraph(es)
	if len(g.References(1)) != 2 || g.References(1)[0] != 2 || g.References(1)[1] != 3 {
		t.Fatalf("refs=%v", g.References(1))
	}
	if len(g.ReferencedBy(2)) != 1 || g.ReferencedBy(2)[0] != 1 {
		t.Fatalf("in=%v", g.ReferencedBy(2))
	}
}
