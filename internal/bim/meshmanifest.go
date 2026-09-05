package bim

import (
	"encoding/json"
	"fmt"
	"sort"
)

// MeshElementRange binds a contiguous triangle-index range to a stable BIM element.
type MeshElementRange struct {
	ElementExternalID string `json:"element_external_id"`
	IFCGlobalID       string `json:"ifc_global_id,omitempty"`
	FirstIndex        int    `json:"first_index"`
	IndexCount        int    `json:"index_count"`
}

type MeshElementManifest struct {
	Version string             `json:"version"`
	Ranges  []MeshElementRange `json:"ranges"`
}

func (m MeshElementManifest) Validate(indexCount int) error {
	if m.Version == "" {
		return fmt.Errorf("manifest version is empty")
	}
	prev := 0
	for _, r := range m.Ranges {
		if r.ElementExternalID == "" {
			return fmt.Errorf("element_external_id is empty")
		}
		if r.FirstIndex < 0 || r.IndexCount <= 0 || r.FirstIndex+r.IndexCount > indexCount {
			return fmt.Errorf("invalid index range for %s", r.ElementExternalID)
		}
		if r.FirstIndex < prev {
			return fmt.Errorf("ranges are not sorted")
		}
		if r.IndexCount%3 != 0 {
			return fmt.Errorf("range for %s is not triangle-aligned", r.ElementExternalID)
		}
		prev = r.FirstIndex + r.IndexCount
	}
	return nil
}

func (m MeshElementManifest) FindIndex(index int) (MeshElementRange, bool) {
	for _, r := range m.Ranges {
		if index >= r.FirstIndex && index < r.FirstIndex+r.IndexCount {
			return r, true
		}
	}
	return MeshElementRange{}, false
}

func (m *MeshElementManifest) Sort() {
	sort.Slice(m.Ranges, func(i, j int) bool { return m.Ranges[i].FirstIndex < m.Ranges[j].FirstIndex })
}
func (m MeshElementManifest) Marshal() ([]byte, error) { return json.Marshal(m) }
func ParseMeshElementManifest(b []byte) (MeshElementManifest, error) {
	var m MeshElementManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	return m, nil
}
