package ifcsemantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

// PropertyValue is a decoded IFC property/quantity value while retaining its raw STEP token.
type PropertyValue struct {
	Value any    `json:"value"`
	Unit  string `json:"unit,omitempty"`
	Raw   string `json:"raw,omitempty"`
}

// PropertySets contains PropertySet and ElementQuantity data keyed by name.
type PropertySets map[string]map[string]PropertyValue

// DecodeProperties extracts common IfcPropertySet and IfcElementQuantity data and attaches
// them to related objects through IfcRelDefinesByProperties. Unknown schema constructs are preserved as Raw.
func DecodeProperties(entities []ifc.Entity) map[int]PropertySets {
	byID := make(map[int]ifc.Entity, len(entities))
	for _, e := range entities {
		byID[e.ID] = e
	}
	defs := make(map[int]PropertySets)
	for _, e := range entities {
		t := strings.ToUpper(e.Type)
		if t != "IFCPROPERTYSET" && t != "IFCELEMENTQUANTITY" {
			continue
		}
		if len(e.Attributes) < 2 {
			continue
		}
		nameIdx := 2
		if len(e.Attributes) <= nameIdx {
			continue
		}
		name := decodeString(e.Attributes[nameIdx])
		if name == "" {
			name = t
		}
		ps := PropertySets{name: {}}
		refs := referencedIDs(e.Attributes)
		for _, rid := range refs {
			p, ok := byID[rid]
			if !ok {
				continue
			}
			pt := strings.ToUpper(p.Type)
			if t == "IFCPROPERTYSET" && pt == "IFCPROPERTYSINGLEVALUE" {
				addSingle(ps[name], p)
			}
			if t == "IFCELEMENTQUANTITY" && strings.HasPrefix(pt, "IFCQUANTITY") {
				addQuantity(ps[name], p)
			}
		}
		defs[e.ID] = ps
	}
	out := make(map[int]PropertySets)
	for _, r := range entities {
		if strings.ToUpper(r.Type) != "IFCRELDEFINESBYPROPERTIES" {
			continue
		}
		refs := referencedIDs(r.Attributes)
		if len(refs) < 2 {
			continue
		}
		propDef := refs[len(refs)-1]
		ps, ok := defs[propDef]
		if !ok {
			continue
		}
		for _, obj := range refs[:len(refs)-1] {
			if _, exists := out[obj]; !exists {
				out[obj] = PropertySets{}
			}
			mergePropertySets(out[obj], ps)
		}
	}
	return out
}

func addSingle(dst map[string]PropertyValue, e ifc.Entity) {
	if len(e.Attributes) < 3 {
		return
	}
	name := decodeString(e.Attributes[0])
	if name == "" {
		return
	}
	v := decodeTypedValue(e.Attributes[2])
	if len(e.Attributes) > 3 && e.Attributes[3] != "$" {
		v.Unit = decodeString(e.Attributes[3])
	}
	dst[name] = v
}

func addQuantity(dst map[string]PropertyValue, e ifc.Entity) {
	if len(e.Attributes) < 3 {
		return
	}
	name := decodeString(e.Attributes[0])
	if name == "" {
		return
	}
	// Common IfcQuantity* entities place the measured value at attribute index 3.
	idx := 3
	if len(e.Attributes) <= idx {
		return
	}
	v := decodeTypedValue(e.Attributes[idx])
	if len(e.Attributes) > 2 && e.Attributes[2] != "$" {
		v.Unit = decodeString(e.Attributes[2])
	}
	dst[name] = v
}

func mergePropertySets(dst, src PropertySets) {
	for set, vals := range src {
		if _, ok := dst[set]; !ok {
			dst[set] = map[string]PropertyValue{}
		}
		for k, v := range vals {
			dst[set][k] = v
		}
	}
}

func decodeTypedValue(raw string) PropertyValue {
	raw = strings.TrimSpace(raw)
	v := PropertyValue{Raw: raw}
	if raw == "$" || raw == "*" {
		return v
	}
	if strings.HasPrefix(raw, "IFC") && strings.Contains(raw, "(") && strings.HasSuffix(raw, ")") {
		i := strings.IndexByte(raw, '(')
		typ := strings.ToUpper(raw[:i])
		inner := raw[i+1 : len(raw)-1]
		v.Value = decodeScalar(inner)
		v.Unit = typ
		return v
	}
	v.Value = decodeScalar(raw)
	return v
}

func decodeScalar(raw string) any {
	raw = strings.TrimSpace(raw)
	if raw == "$" || raw == "*" {
		return nil
	}
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") {
		return decodeString(raw)
	}
	upper := strings.ToUpper(raw)
	if upper == ".T." {
		return true
	}
	if upper == ".F." {
		return false
	}
	if strings.HasPrefix(raw, ".") && strings.HasSuffix(raw, ".") {
		return strings.Trim(raw, ".")
	}
	if i, err := strconv.Atoi(raw); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f
	}
	return raw
}

func PropertyJSON(entities []ifc.Entity, id int) map[string]any {
	p := DecodeProperties(entities)[id]
	out := map[string]any{}
	for set, vals := range p {
		inner := map[string]any{}
		for name, value := range vals {
			inner[name] = value
		}
		out[set] = inner
	}
	return out
}

func ValidatePropertySetName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("property set name is empty")
	}
	return nil
}
