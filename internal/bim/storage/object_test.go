package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalStore(t *testing.T) {
	s := LocalStore{Root: t.TempDir()}
	o, err := s.Put(context.Background(), "models/a.ifc", strings.NewReader("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if o.Size != 3 || o.SHA256 == "" {
		t.Fatal(o)
	}
	r, err := s.Get(context.Background(), o.Key)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "abc" {
		t.Fatal(string(b))
	}
}
