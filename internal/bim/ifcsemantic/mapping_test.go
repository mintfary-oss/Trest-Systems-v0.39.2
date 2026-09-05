package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestMapToBIMElements(t *testing.T) {
	es, err := ifc.ParseSTEP(strings.NewReader("#1=IFCPROJECT('p',$,'P',$,$,$,$,$,$);\n#2=IFCWALL('w',$,'Wall',$,$,$,$,$,$,.NOTDEFINED.);\n#10=IFCRELAGGREGATES('r',$,$,$,#1,(#2));\n"))
	if err != nil {
		t.Fatal(err)
	}
	m := MapToBIMElements(es)
	if len(m) != 2 {
		t.Fatalf("got %d", len(m))
	}
	if m[1].ParentExternalID != "ifc:1" || m[1].IFCGlobalID != "w" {
		t.Fatalf("bad mapping: %+v", m[1])
	}
	p := SpatialPath(es, 2)
	if len(p) != 2 || p[1].Type != "IFCPROJECT" {
		t.Fatalf("bad path: %+v", p)
	}
}
