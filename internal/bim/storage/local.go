package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manifest struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type Local struct{ Root string }

func NewLocal(root string) Local { return Local{Root: root} }
func (s Local) Put(key string, r io.Reader) (Manifest, error) {
	if s.Root == "" {
		return Manifest{}, fmt.Errorf("storage root is empty")
	}
	p := filepath.Join(s.Root, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return Manifest{}, err
	}
	f, err := os.Create(p)
	if err != nil {
		return Manifest{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), r)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{Path: key, Size: n, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}
func (s Local) Get(key string) (io.ReadCloser, error) {
	p := filepath.Join(s.Root, filepath.Clean("/"+key))
	return os.Open(p)
}
