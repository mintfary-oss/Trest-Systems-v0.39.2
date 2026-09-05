package bim

import "testing"

func TestMeshElementManifest(t *testing.T) {
	m := MeshElementManifest{Version: "1", Ranges: []MeshElementRange{{ElementExternalID: "wall-1", IFCGlobalID: "g1", FirstIndex: 0, IndexCount: 6}, {ElementExternalID: "slab-1", FirstIndex: 6, IndexCount: 3}}}
	if err := m.Validate(9); err != nil {
		t.Fatal(err)
	}
	r, ok := m.FindIndex(7)
	if !ok || r.ElementExternalID != "slab-1" {
		t.Fatalf("unexpected range: %+v %v", r, ok)
	}
	if _, ok := m.FindIndex(9); ok {
		t.Fatal("out of range matched")
	}
}
