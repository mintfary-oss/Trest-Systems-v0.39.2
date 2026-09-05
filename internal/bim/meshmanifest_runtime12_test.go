package bim

import "testing"

func TestBuildMeshElementManifest(t *testing.T) {
	m, err := BuildMeshElementManifest("1", []ElementMeshBatch{{ExternalID: "b", IFCGlobalID: "G2", IndexCount: 6}, {ExternalID: "a", IFCGlobalID: "G1", IndexCount: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Ranges) != 2 || m.Ranges[0].ElementExternalID != "a" || m.Ranges[1].FirstIndex != 3 {
		t.Fatalf("unexpected manifest: %+v", m)
	}
	if r, ok := m.FindIndex(4); !ok || r.IFCGlobalID != "G2" {
		t.Fatalf("pick mismatch: %+v %v", r, ok)
	}
}
