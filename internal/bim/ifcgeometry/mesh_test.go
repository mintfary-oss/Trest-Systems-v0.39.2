package ifcgeometry

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestFacetedBrep(t *testing.T) {
	src := `#1=IFCCARTESIANPOINT((0.,0.,0.));
#2=IFCCARTESIANPOINT((1.,0.,0.));
#3=IFCCARTESIANPOINT((0.,1.,0.));
#4=IFCPOLYLOOP((#1,#2,#3));
#5=IFCFACEOUTERBOUND(#4,.T.);
#6=IFCFACE((#5));
#7=IFCCLOSEDSHELL((#6));
#8=IFCFACETEDBREP(#7);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	m, err := Extract(es, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Vertices) != 3 || len(m.Indices) != 3 {
		t.Fatalf("got %d vertices/%d indices", len(m.Vertices), len(m.Indices))
	}
}
func TestTriangulatedFaceSet(t *testing.T) {
	src := `#1=IFCCARTESIANPOINTLIST3D(((0.,0.,0.),(1.,0.,0.),(0.,1.,0.)));
#2=IFCTRIANGULATEDFACESET(#1,$,$,$,((1,2,3)));`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	m, err := Extract(es, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Vertices) != 3 || len(m.Indices) != 3 {
		t.Fatalf("got %d vertices/%d indices", len(m.Vertices), len(m.Indices))
	}
}
