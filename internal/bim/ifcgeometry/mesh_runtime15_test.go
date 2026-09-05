package ifcgeometry

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestPolygonalFaceSet(t *testing.T) {
	src := `#1=IFCCARTESIANPOINTLIST3D(((0.,0.,0.),(1.,0.,0.),(1.,1.,0.),(0.,1.,0.)));
#2=IFCINDEXEDPOLYGON((1,2,3,4));
#3=IFCPOLYGONALFACESET(#1,$,(#2),$);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	m, err := Extract(es, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Vertices) != 8 || len(m.Indices) != 6 {
		t.Fatalf("got %d/%d", len(m.Vertices), len(m.Indices))
	}
}

func TestExtrudedRectangle(t *testing.T) {
	src := `#1=IFCRECTANGLEPROFILEDEF(.AREA.,$,.2,.4);
#2=IFCDIRECTION((0.,0.,1.));
#3=IFCEXTRUDEDAREASOLID(#1,$,#2,3.);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	m, err := Extract(es, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Vertices) != 8 || len(m.Indices) != 36 {
		t.Fatalf("got %d/%d", len(m.Vertices), len(m.Indices))
	}
	if m.Vertices[4].Z != 3 {
		t.Fatalf("unexpected extrusion: %+v", m.Vertices[4])
	}
}
