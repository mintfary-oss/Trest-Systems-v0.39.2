package ai

import "errors"

var (
	ErrForbiddenTool    = errors.New("agent tool is not permitted")
	ErrApprovalRequired = errors.New("human approval required")
)

type ToolPolicy struct {
	Name             string `json:"name"`
	RequiresApproval bool   `json:"requires_approval"`
}
type AgentPolicy struct {
	Tools       []ToolPolicy   `json:"tools"`
	Permissions []string       `json:"permissions"`
	Memory      map[string]any `json:"memory"`
	Sandbox     map[string]any `json:"sandbox"`
}

func (p AgentPolicy) AllowsTool(name string) (bool, bool) {
	for _, t := range p.Tools {
		if t.Name == name {
			return true, t.RequiresApproval
		}
	}
	return false, false
}
