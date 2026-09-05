package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestCoreRelationsAllChildren(t *testing.T) {
	es, err := ifc.ParseSTEP(strings.NewReader("#10=IFCRELAGGREGATES('r',$,$,$,#1,(#2,#3));\n#11=IFCRELCONTAINEDINSPATIALSTRUCTURE('r2',$,$,$,(#4,#5),#6);\n"))
	if err != nil {
		t.Fatal(err)
	}
	rs := CoreRelations(es)
	if len(rs) != 4 {
		t.Fatalf("got %d", len(rs))
	}
	if rs[0].SubjectID != 2 || rs[0].ObjectID != 1 || rs[1].SubjectID != 3 || rs[1].ObjectID != 1 {
		t.Fatalf("aggregate mapping: %+v", rs)
	}
}
