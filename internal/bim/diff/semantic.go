package diff

import "sort"

type SemanticElement struct {
	GlobalID    string
	ElementType string
	Name        string
	Properties  map[string]any
}
type SemanticChange struct {
	GlobalID string           `json:"global_id"`
	Change   string           `json:"change"`
	Before   *SemanticElement `json:"before,omitempty"`
	After    *SemanticElement `json:"after,omitempty"`
}
type SemanticResult struct {
	Added   []SemanticChange `json:"added"`
	Removed []SemanticChange `json:"removed"`
	Changed []SemanticChange `json:"changed"`
}

// CompareSemantic compares IFC elements by stable GlobalId instead of vertex position.
func CompareSemantic(oldElements, newElements []SemanticElement) SemanticResult {
	om, nm := map[string]SemanticElement{}, map[string]SemanticElement{}
	for _, e := range oldElements {
		if e.GlobalID != "" {
			om[e.GlobalID] = e
		}
	}
	for _, e := range newElements {
		if e.GlobalID != "" {
			nm[e.GlobalID] = e
		}
	}
	r := SemanticResult{}
	for id, e := range nm {
		if o, ok := om[id]; !ok {
			x := e
			r.Added = append(r.Added, SemanticChange{GlobalID: id, Change: "added", After: &x})
		} else if o.ElementType != e.ElementType || o.Name != e.Name || !sameProperties(o.Properties, e.Properties) {
			a, b := o, e
			r.Changed = append(r.Changed, SemanticChange{GlobalID: id, Change: "changed", Before: &a, After: &b})
		}
	}
	for id, e := range om {
		if _, ok := nm[id]; !ok {
			x := e
			r.Removed = append(r.Removed, SemanticChange{GlobalID: id, Change: "removed", Before: &x})
		}
	}
	sort.Slice(r.Added, func(i, j int) bool { return r.Added[i].GlobalID < r.Added[j].GlobalID })
	sort.Slice(r.Removed, func(i, j int) bool { return r.Removed[i].GlobalID < r.Removed[j].GlobalID })
	sort.Slice(r.Changed, func(i, j int) bool { return r.Changed[i].GlobalID < r.Changed[j].GlobalID })
	return r
}
func sameProperties(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
