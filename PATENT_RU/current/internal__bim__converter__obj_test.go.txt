package converter

import (
	"strings"
	"testing"
)

func TestOBJToGLTF(t *testing.T) {
	m, e := ParseOBJ(strings.NewReader("v 0 0 0\nv 1 0 0\nv 0 1 0\nf 1 2 3\n"))
	if e != nil {
		t.Fatal(e)
	}
	var b strings.Builder
	if e = EncodeGLTF(&b, m); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), `"version":"2.0"`) {
		t.Fatal("not glTF")
	}
}
