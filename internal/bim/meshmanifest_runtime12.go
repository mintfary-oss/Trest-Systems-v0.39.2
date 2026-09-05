package bim

import (
	"fmt"
	"sort"
)

// ElementMeshBatch describes one semantic element's contribution to a triangle mesh.
type ElementMeshBatch struct {
	ExternalID  string
	IFCGlobalID string
	IndexCount  int
}

// BuildMeshElementManifest assigns deterministic contiguous index ranges to element batches.
// It is intentionally geometry-agnostic: converters provide the triangle counts.
func BuildMeshElementManifest(version string, batches []ElementMeshBatch) (MeshElementManifest, error) {
	if version == "" {
		return MeshElementManifest{}, fmt.Errorf("version is empty")
	}
	cp := append([]ElementMeshBatch(nil), batches...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ExternalID < cp[j].ExternalID })
	m := MeshElementManifest{Version: version, Ranges: make([]MeshElementRange, 0, len(cp))}
	cursor := 0
	for _, b := range cp {
		if b.ExternalID == "" {
			return MeshElementManifest{}, fmt.Errorf("external id is empty")
		}
		if b.IndexCount <= 0 || b.IndexCount%3 != 0 {
			return MeshElementManifest{}, fmt.Errorf("invalid index count for %s", b.ExternalID)
		}
		m.Ranges = append(m.Ranges, MeshElementRange{ElementExternalID: b.ExternalID, IFCGlobalID: b.IFCGlobalID, FirstIndex: cursor, IndexCount: b.IndexCount})
		cursor += b.IndexCount
	}
	if err := m.Validate(cursor); err != nil {
		return MeshElementManifest{}, err
	}
	return m, nil
}
