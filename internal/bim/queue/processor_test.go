package queue

import "testing"

func TestProcessorExists(t *testing.T) {
	var p Processor
	if p.DB != nil {
		t.Fatal("unexpected db")
	}
}
