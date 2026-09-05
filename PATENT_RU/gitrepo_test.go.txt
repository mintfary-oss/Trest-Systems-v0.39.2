package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mustRun runs git in dir and fails the test on error.
func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// newRemote creates a local repository with an initial commit on branch
// and returns its filesystem path, usable as a clone/fetch source.
func newRemote(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()

	mustRun(t, dir, "init", "--initial-branch="+branch)
	mustRun(t, dir, "config", "user.email", "test@example.com")
	mustRun(t, dir, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	mustRun(t, dir, "add", "README.md")
	mustRun(t, dir, "commit", "-m", "initial")

	return dir
}

func TestEnsureRepoClonesWhenMissing(t *testing.T) {
	remote := newRemote(t, "main")
	dest := filepath.Join(t.TempDir(), "checkout")

	updated, err := EnsureRepo(context.Background(), remote, "main", dest)
	if err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}
	if !updated {
		t.Error("EnsureRepo: updated = false on first clone, want true")
	}
	if !isRepo(context.Background(), dest) {
		t.Error("EnsureRepo: dest is not a git repository after clone")
	}
}

func TestEnsureRepoUpdatesWhenPresent(t *testing.T) {
	remote := newRemote(t, "main")
	dest := filepath.Join(t.TempDir(), "checkout")

	if _, err := EnsureRepo(context.Background(), remote, "main", dest); err != nil {
		t.Fatalf("EnsureRepo (initial clone): %v", err)
	}

	// Add a new commit to the remote after the initial clone.
	if err := os.WriteFile(filepath.Join(remote, "README.md"), []byte("v2\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	mustRun(t, remote, "add", "README.md")
	mustRun(t, remote, "commit", "-m", "second")

	updated, err := EnsureRepo(context.Background(), remote, "main", dest)
	if err != nil {
		t.Fatalf("EnsureRepo (update): %v", err)
	}
	if !updated {
		t.Error("EnsureRepo: updated = false after new remote commit, want true")
	}

	got, err := os.ReadFile(filepath.Join(dest, "README.md"))
	if err != nil {
		t.Fatalf("read updated README: %v", err)
	}
	if string(got) != "v2\n" {
		t.Errorf("README.md = %q, want %q", got, "v2\n")
	}
}

func TestUpdateNoOpWhenAlreadyCurrent(t *testing.T) {
	remote := newRemote(t, "main")
	dest := filepath.Join(t.TempDir(), "checkout")

	if _, err := EnsureRepo(context.Background(), remote, "main", dest); err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}

	updated, err := Update(context.Background(), dest, "main")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated {
		t.Error("Update: updated = true with no new remote commits, want false")
	}
}
