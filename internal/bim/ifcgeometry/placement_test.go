package ifcgeometry

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"math"
	"strings"
	"testing"
)

func TestResolvePlacementChain(t *testing.T) {
	src := `#1=IFCCARTESIANPOINT((10.,20.,30.));
#2=IFCDIRECTION((0.,0.,1.));
#3=IFCDIRECTION((1.,0.,0.));
#4=IFCAXIS2PLACEMENT3D(#1,#2,#3);
#5=IFCLOCALPLACEMENT(#4,$);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ResolvePlacement(es, 5)
	if err != nil {
		t.Fatal(err)
	}
	p := tr.Matrix.Point(Vec3{1, 2, 3})
	if math.Abs(p.X-11) > 1e-9 || math.Abs(p.Y-22) > 1e-9 || math.Abs(p.Z-33) > 1e-9 {
		t.Fatalf("unexpected point: %+v", p)
	}
}

func TestNestedPlacement(t *testing.T) {
	src := `#1=IFCCARTESIANPOINT((1.,0.,0.));
#2=IFCDIRECTION((0.,0.,1.));
#3=IFCDIRECTION((1.,0.,0.));
#4=IFCAXIS2PLACEMENT3D(#1,#2,#3);
#5=IFCLOCALPLACEMENT(#4,$);
#6=IFCCARTESIANPOINT((0.,2.,0.));
#7=IFCAXIS2PLACEMENT3D(#6,#2,#3);
#8=IFCLOCALPLACEMENT(#7,#5);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ResolvePlacement(es, 8)
	if err != nil {
		t.Fatal(err)
	}
	p := tr.Matrix.Point(Vec3{0, 0, 0})
	if math.Abs(p.X-1) > 1e-9 || math.Abs(p.Y-2) > 1e-9 {
		t.Fatalf("unexpected point %+v", p)
	}
}
