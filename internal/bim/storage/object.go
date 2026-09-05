package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Object struct {
	Key    string `json:"key"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type Store interface {
	Put(context.Context, string, io.Reader) (Object, error)
	Get(context.Context, string) (io.ReadCloser, error)
}

type LocalStore struct{ Root string }

func (s LocalStore) Put(ctx context.Context, key string, r io.Reader) (Object, error) {
	select {
	case <-ctx.Done():
		return Object{}, ctx.Err()
	default:
	}
	if key == "" {
		return Object{}, fmt.Errorf("key is empty")
	}
	path := filepath.Join(s.Root, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Object{}, err
	}
	f, err := os.Create(path)
	if err != nil {
		return Object{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), r)
	if err != nil {
		return Object{}, err
	}
	return Object{Key: key, Size: n, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}
func (s LocalStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return os.Open(filepath.Join(s.Root, filepath.Clean("/"+key)))
}
