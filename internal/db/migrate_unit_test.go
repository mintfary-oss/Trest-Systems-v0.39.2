package db

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationFilesCanonicalNamesAndChecksums(t *testing.T) {
	d := t.TempDir()
	a := []byte("SELECT 1;\nSELECT 2;\n")
	if err := os.WriteFile(filepath.Join(d, "0002_b.sql"), a, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "0001_a.sql"), a, 0600); err != nil {
		t.Fatal(err)
	}
	files, err := loadMigrations(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Name != "0001_a.sql" || files[1].Name != "0002_b.sql" {
		t.Fatalf("%+v", files)
	}
	if files[0].Checksum != fmt.Sprintf("%x", sha256.Sum256(a)) {
		t.Fatal("checksum")
	}
	if files[0].Body != string(a) {
		t.Fatal("SQL bytes changed")
	}
}
func TestMigrationEmptyDirectoryFails(t *testing.T) {
	if _, err := loadMigrations(t.TempDir()); err == nil {
		t.Fatal("empty migration directory accepted")
	}
}
func TestMigrationUnsafeFilenameFails(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "x';drop.sql"), []byte("SELECT 1"), 0600)
	if _, err := loadMigrations(d); err == nil {
		t.Fatal("unsafe SQL filename accepted")
	}
}
func TestMigrationSymlinkFails(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(t.TempDir(), "target.sql")
	os.WriteFile(p, []byte("SELECT 1"), 0600)
	if err := os.Symlink(p, filepath.Join(d, "0001_link.sql")); err != nil {
		t.Skip(err)
	}
	if _, err := loadMigrations(d); err == nil {
		t.Fatal("symlink accepted")
	}
}
