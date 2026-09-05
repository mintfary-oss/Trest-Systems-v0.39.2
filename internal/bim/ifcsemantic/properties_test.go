package ifcsemantic

import (
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestDecodeProperties(t *testing.T) {
	src := `#10=IFCWALL('wall1',$,'Wall',$,$,$,$,$,$,.NOTDEFINED.);
#20=IFCPROPERTYSET('ps1',$,'Pset_WallCommon',$,(#21,#22));
#21=IFCPROPERTYSINGLEVALUE('FireRating',$,'2h', 'HOUR');
#22=IFCPROPERTYSINGLEVALUE('IsExternal',$,.T.,$);
#30=IFCRELDEFINESBYPROPERTIES('rel',$,(#10),#20);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	p := DecodeProperties(es)[10]
	if p["Pset_WallCommon"]["FireRating"].Value != "2h" {
		t.Fatalf("bad string property: %#v", p)
	}
	if p["Pset_WallCommon"]["IsExternal"].Value != true {
		t.Fatalf("bad bool property: %#v", p)
	}
}

func TestQuantityDecode(t *testing.T) {
	src := `#1=IFCWALL('g',$,'Wall',$,$,$,$,$,$,.NOTDEFINED.);
#2=IFCELEMENTQUANTITY('q',$,'BaseQuantities',$,$,(#3,#4));
#3=IFCQUANTITYLENGTH('Length',$,$,12.5,$);
#4=IFCQUANTITYAREA('Area',$,$,20.0,$);
#5=IFCRELDEFINESBYPROPERTIES('r',$,(#1),#2);`
	es, err := ifc.ParseSTEP(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	p := DecodeProperties(es)[1]
	if p["BaseQuantities"]["Length"].Value != 12.5 {
		t.Fatalf("bad length: %#v", p)
	}
	if p["BaseQuantities"]["Area"].Value != 20.0 {
		t.Fatalf("bad area: %#v", p)
	}
}
