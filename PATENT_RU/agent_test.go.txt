package ai

import "testing"

func TestToolPolicy(t *testing.T) {
	p := AgentPolicy{Tools: []ToolPolicy{{Name: "estimate.read", RequiresApproval: false}, {Name: "order.approve", RequiresApproval: true}}}
	if err := MustApprove(p, "estimate.read"); err != nil {
		t.Fatal(err)
	}
	if err := MustApprove(p, "order.approve"); err != ErrApprovalRequired {
		t.Fatalf("got %v", err)
	}
	if err := MustApprove(p, "shell"); err != ErrForbiddenTool {
		t.Fatalf("got %v", err)
	}
}
