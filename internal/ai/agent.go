package ai

import "strings"

func ValidateInstructions(v string) bool { return strings.TrimSpace(v) != "" }
func MustApprove(policy AgentPolicy, tool string) error {
	allowed, approval := policy.AllowsTool(tool)
	if !allowed {
		return ErrForbiddenTool
	}
	if approval {
		return ErrApprovalRequired
	}
	return nil
}
