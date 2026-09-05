package ifc

import (
	"strings"
	"testing"
)

func TestParseSTEP(t *testing.T) {
	input := `ISO-10303-21;
DATA;
#1=IFCPROJECT('abc',$,'Project',('desc,with comma'),$,$,$,$,$);
#2=IFCWALLSTANDARDCASE('wall',#1,'Wall',(#3,#4),$,$);
ENDSEC;
END-ISO-10303-21;`
	got, err := ParseSTEP(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != 1 || got[0].Type != "IFCPROJECT" {
		t.Fatalf("unexpected entities: %#v", got)
	}
	if len(got[0].Attributes) != 9 {
		t.Fatalf("attrs=%d want 9: %#v", len(got[0].Attributes), got[0].Attributes)
	}
	if got[0].Attributes[3] != "('desc,with comma')" {
		t.Fatalf("nested attribute split: %#v", got[0].Attributes)
	}
}
