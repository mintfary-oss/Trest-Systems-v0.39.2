package storage

import (
	"os"
	"strings"
	"testing"
)

func TestLocalPutGet(t *testing.T) {
	d := t.TempDir()
	s := NewLocal(d)
	m, err := s.Put("models/a.ifc", strings.NewReader("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if m.Size != 3 || m.SHA256 == "" {
		t.Fatalf("bad manifest: %+v", m)
	}
	r, err := s.Get("models/a.ifc")
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	if _, err = os.Stat(d + "/models/a.ifc"); err != nil {
		t.Fatal(err)
	}
}
