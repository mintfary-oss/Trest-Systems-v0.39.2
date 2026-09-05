package config

import (
	"os"
	"path/filepath"
)

// Paths contains the local trestctl state directories.
type Paths struct {
	Root        string
	Reports     string
	Diagnostics string
}

// DefaultPaths returns a deterministic per-user state directory. Root can be
// overridden with TREST_STATE_DIR for system installations.
func DefaultPaths() Paths {
	root := os.Getenv("TREST_STATE_DIR")
	if root == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			root = filepath.Join(home, ".trest")
		} else {
			root = ".trest"
		}
	}
	return Paths{
		Root:        root,
		Reports:     filepath.Join(root, "reports"),
		Diagnostics: filepath.Join(root, "diagnostics"),
	}
}

// EnsureDirs creates all state directories with owner-only permissions where
// the platform supports them.
func EnsureDirs(paths Paths) error {
	for _, dir := range []string{paths.Root, paths.Reports, paths.Diagnostics} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}
