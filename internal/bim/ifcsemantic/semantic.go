package ifcsemantic

import (
	"fmt"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

type Element struct {
	ID             int
	Type           string
	GlobalID       string
	Name           string
	Description    string
	ObjectType     string
	PredefinedType string
	RawAttributes  []string
}

type Model struct {
	Elements   []Element
	ByID       map[int]Element
	ByGlobalID map[string]Element
}

func Build(entities []ifc.Entity) Model {
	m := Model{ByID: map[int]Element{}, ByGlobalID: map[string]Element{}}
	for _, e := range entities {
		if !strings.HasPrefix(strings.ToUpper(e.Type), "IFC") {
			continue
		}
		el := Element{ID: e.ID, Type: e.Type, RawAttributes: append([]string(nil), e.Attributes...)}
		if len(e.Attributes) > 0 {
			el.GlobalID = decodeString(e.Attributes[0])
		}
		if len(e.Attributes) > 2 {
			el.Name = decodeString(e.Attributes[2])
		}
		if len(e.Attributes) > 3 {
			el.Description = decodeString(e.Attributes[3])
		}
		if len(e.Attributes) > 4 {
			el.ObjectType = decodeString(e.Attributes[4])
		}
		if len(e.Attributes) > 8 {
			el.PredefinedType = decodeEnum(e.Attributes[len(e.Attributes)-1])
		}
		m.Elements = append(m.Elements, el)
		m.ByID[el.ID] = el
		if el.GlobalID != "" && el.GlobalID != "$" {
			m.ByGlobalID[el.GlobalID] = el
		}
	}
	return m
}

func (m Model) Get(id int) (Element, bool)            { e, ok := m.ByID[id]; return e, ok }
func (m Model) GetGlobalID(id string) (Element, bool) { e, ok := m.ByGlobalID[id]; return e, ok }
func decodeString(v string) string {
	v = strings.TrimSpace(v)
	if v == "$" || v == "*" {
		return ""
	}
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return strings.ReplaceAll(v[1:len(v)-1], "''", "'")
	}
	return v
}
func decodeEnum(v string) string { v = strings.TrimSpace(v); return strings.Trim(v, ".") }
func RequireGlobalID(e Element) error {
	if e.GlobalID == "" {
		return fmt.Errorf("IFC entity #%d has no GlobalId", e.ID)
	}
	return nil
}
