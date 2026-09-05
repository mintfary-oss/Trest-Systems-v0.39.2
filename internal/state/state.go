// Package state persists the orchestrator's progress to state.json so a
// second invocation can resume after a non-fatal error instead of
// repeating already-completed steps.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StepResult records the outcome of a single orchestration step.
type StepResult struct {
	// Completed indicates the step finished without error.
	Completed bool `json:"completed"`
	// Error holds the last error message for this step, if any.
	Error string `json:"error,omitempty"`
	// UpdatedAt is when this step's result was last written.
	UpdatedAt time.Time `json:"updated_at"`
}

// State is the persisted orchestrator progress.
type State struct {
	// LastCompletedStep is the name of the most recently completed step,
	// used to decide where a resumed run should continue from.
	LastCompletedStep string `json:"last_completed_step"`
	// Steps maps step name to its most recent result.
	Steps map[string]StepResult `json:"steps"`
}

// New returns an empty State ready for use.
func New() *State {
	return &State{Steps: make(map[string]StepResult)}
}

// Load reads state from path. A missing file is not an error; it returns a
// fresh, empty State so first runs work without pre-creating the file.
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("read state %q: %w", path, err)
	}

	st := New()
	if err := json.Unmarshal(data, st); err != nil {
		return nil, fmt.Errorf("parse state %q: %w", path, err)
	}
	if st.Steps == nil {
		st.Steps = make(map[string]StepResult)
	}
	return st, nil
}

// Save writes state to path as indented JSON, replacing any prior content.
func (s *State) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state %q: %w", path, err)
	}
	return nil
}

// MarkSuccess records step as completed and updates LastCompletedStep.
func (s *State) MarkSuccess(step string) {
	s.Steps[step] = StepResult{Completed: true, UpdatedAt: time.Now().UTC()}
	s.LastCompletedStep = step
}

// MarkFailure records step as failed with err, without advancing
// LastCompletedStep.
func (s *State) MarkFailure(step string, err error) {
	s.Steps[step] = StepResult{Completed: false, Error: err.Error(), UpdatedAt: time.Now().UTC()}
}

// IsCompleted reports whether step previously completed successfully.
func (s *State) IsCompleted(step string) bool {
	return s.Steps[step].Completed
}
