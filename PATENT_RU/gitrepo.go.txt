// Package gitrepo clones and updates the tracked repository using the
// system git binary, so authentication relies on whatever credential
// helpers are already configured in the environment rather than requiring
// the orchestrator to handle tokens or SSH keys itself.
package gitrepo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// run executes git with args in dir and returns trimmed combined stdout.
// stderr is captured into the returned error for diagnostics.
func run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// isRepo reports whether dir is the root of a git working tree.
func isRepo(ctx context.Context, dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	_, err = run(ctx, dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// Clone clones repoURL at branch into destDir, which must not already
// exist.
func Clone(ctx context.Context, repoURL, branch, destDir string) error {
	if _, err := run(ctx, "", "clone", "--branch", branch, "--single-branch", repoURL, destDir); err != nil {
		return fmt.Errorf("clone %q: %w", repoURL, err)
	}
	return nil
}

// EnsureRepo makes sure destDir contains an up-to-date checkout of repoURL
// at branch: cloning it if absent, or fetching and fast-forwarding it if
// present. It reports whether new commits were pulled.
func EnsureRepo(ctx context.Context, repoURL, branch, destDir string) (updated bool, err error) {
	if !isRepo(ctx, destDir) {
		if err := Clone(ctx, repoURL, branch, destDir); err != nil {
			return false, err
		}
		return true, nil
	}
	return Update(ctx, destDir, branch)
}

// Update fetches branch from origin and fast-forwards destDir's working
// tree to match. It reports whether the local HEAD moved. Update refuses
// to proceed (returning an error) if the fast-forward would be unsafe,
// e.g. due to local commits diverging from origin.
func Update(ctx context.Context, destDir, branch string) (updated bool, err error) {
	before, err := run(ctx, destDir, "rev-parse", "HEAD")
	if err != nil {
		return false, fmt.Errorf("resolve current HEAD: %w", err)
	}

	if _, err := run(ctx, destDir, "fetch", "origin", branch); err != nil {
		return false, fmt.Errorf("fetch origin/%s: %w", branch, err)
	}

	if _, err := run(ctx, destDir, "checkout", branch); err != nil {
		return false, fmt.Errorf("checkout %s: %w", branch, err)
	}

	if _, err := run(ctx, destDir, "merge", "--ff-only", "origin/"+branch); err != nil {
		return false, fmt.Errorf("fast-forward to origin/%s: %w", branch, err)
	}

	after, err := run(ctx, destDir, "rev-parse", "HEAD")
	if err != nil {
		return false, fmt.Errorf("resolve updated HEAD: %w", err)
	}

	return before != after, nil
}
