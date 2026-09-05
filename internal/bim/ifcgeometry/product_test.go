package ifcgeometry

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"math"
	"strings"
	"testing"
)

func TestExtractProductGeometryAppliesPlacement(t *testing.T) {
	src := `#1=IFCCARTESIANPOINTLIST3D(((0.,0.,0.),(1.,0.,0.),(0.,1.,0.)));
#2=IFCTRIANGULATEDFACESET(#1,$,$,$,((1,2,3)));
#3=IFCSHAPEREPRESENTATION($,'Body','Tessellation',(#2));
#4=IFCPRODUCTDEFINITIONSHAPE($,$,(#3));
#5=IFCCARTESIANPOINT((10.,20.,30.));
#6=IFCDIRECTION((0.,0.,1.));
#7=IFCDIRECTION((1.,0.,0.));
#8=IFCAXIS2PLACEMENT3D(#5,#6,#7);
#9=IFCLOCALPLACEMENT(#8,$);
#10=IFCWALL('g',$,'Wall',$,#9,#4,$);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	g, err := ExtractProductGeometry(es, 10)
	if err != nil {
		t.Fatal(err)
	}
	if g.GeometryEntityID != 2 || g.PlacementEntityID != 9 {
		t.Fatalf("unexpected instance %+v", g)
	}
	p := g.Mesh.Vertices[0]
	if math.Abs(p.X-10) > 1e-9 || math.Abs(p.Y-20) > 1e-9 || math.Abs(p.Z-30) > 1e-9 {
		t.Fatalf("unexpected world point %+v", p)
	}
}
