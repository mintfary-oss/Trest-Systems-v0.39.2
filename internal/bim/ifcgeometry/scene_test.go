package ifcgeometry

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestBuildSceneGeometryStableGlobalIDRanges(t *testing.T) {
	src := `#1=IFCCARTESIANPOINTLIST3D(((0.,0.,0.),(1.,0.,0.),(0.,1.,0.)));
#2=IFCTRIANGULATEDFACESET(#1,$,$,$,((1,2,3)));
#3=IFCSHAPEREPRESENTATION($,'Body','Tessellation',(#2));
#4=IFCPRODUCTDEFINITIONSHAPE($,$,(#3));
#5=IFCCARTESIANPOINT((10.,0.,0.));
#6=IFCAXIS2PLACEMENT3D(#5,$,$);
#7=IFCLOCALPLACEMENT(#6,$);
#20=IFCWALL('gid-B',$,'B',$,#7,#4,$);
#30=IFCWALL('gid-A',$,'A',$,#7,#4,$);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	s, err := BuildSceneGeometry(es, []int{30, 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Ranges) != 2 || s.Ranges[0].GlobalID != "gid-B" || s.Ranges[1].GlobalID != "gid-A" {
		t.Fatalf("unexpected ranges %+v", s.Ranges)
	}
	if s.Ranges[0].TriangleStart != 0 || s.Ranges[1].TriangleStart != 1 {
		t.Fatalf("unexpected starts %+v", s.Ranges)
	}
}
