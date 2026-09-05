package orders

import "testing"

func TestOrderTransitions(t *testing.T) {
	valid := [][2]string{{Draft, Published}, {Published, Matching}, {Matching, Contracted}, {Contracted, InProgress}, {InProgress, QualityCheck}, {QualityCheck, Completed}, {QualityCheck, InProgress}, {Draft, Cancelled}}
	for _, p := range valid {
		if !CanTransition(p[0], p[1]) {
			t.Fatalf("expected valid transition %s -> %s", p[0], p[1])
		}
	}
	invalid := [][2]string{{Completed, InProgress}, {Cancelled, Published}, {Draft, Completed}, {Published, Completed}}
	for _, p := range invalid {
		if CanTransition(p[0], p[1]) {
			t.Fatalf("expected invalid transition %s -> %s", p[0], p[1])
		}
	}
}
