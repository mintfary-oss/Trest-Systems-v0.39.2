package state

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmptyState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	st, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if st.LastCompletedStep != "" {
		t.Errorf("LastCompletedStep = %q, want empty", st.LastCompletedStep)
	}
	if len(st.Steps) != 0 {
		t.Errorf("Steps = %v, want empty", st.Steps)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	st := New()
	st.MarkSuccess("clone")
	st.MarkFailure("build", errors.New("boom"))

	if err := st.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.LastCompletedStep != "clone" {
		t.Errorf("LastCompletedStep = %q, want %q", loaded.LastCompletedStep, "clone")
	}
	if !loaded.IsCompleted("clone") {
		t.Error("IsCompleted(clone) = false, want true")
	}
	if loaded.IsCompleted("build") {
		t.Error("IsCompleted(build) = true, want false")
	}
	if got := loaded.Steps["build"].Error; got != "boom" {
		t.Errorf("Steps[build].Error = %q, want %q", got, "boom")
	}
}

func TestMarkFailureDoesNotAdvanceLastCompletedStep(t *testing.T) {
	st := New()
	st.MarkSuccess("clone")
	st.MarkFailure("build", errors.New("boom"))

	if st.LastCompletedStep != "clone" {
		t.Errorf("LastCompletedStep = %q, want %q", st.LastCompletedStep, "clone")
	}
}
