package converter

import (
	"strings"
	"testing"
)

func TestParseDXF3DFace(t *testing.T) {
	s := `0\n3DFACE\n10\n0\n20\n0\n30\n0\n11\n1\n21\n0\n31\n0\n12\n0\n22\n1\n32\n0\n`
	m, e := ParseDXF(strings.NewReader(strings.ReplaceAll(s, "\\n", "\n")))
	if e != nil || len(m.Indices) != 3 {
		t.Fatalf("%v %+v", e, m)
	}
}
