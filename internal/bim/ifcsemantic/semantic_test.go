package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	es, err := ifc.ParseSTEP(strings.NewReader("#1=IFCPROJECT('3abc',$,'Project',$,$,$,$,$,$);\n#2=IFCWALL('wall1',$,'Wall A','Desc',$,$,$,$,$,.NOTDEFINED.);\n"))
	if err != nil {
		t.Fatal(err)
	}
	m := Build(es)
	w, ok := m.GetGlobalID("wall1")
	if !ok || w.Name != "Wall A" || w.Type != "IFCWALL" {
		t.Fatalf("bad semantic: %+v", w)
	}
}
