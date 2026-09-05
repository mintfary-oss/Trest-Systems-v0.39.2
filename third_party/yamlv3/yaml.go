// Package yaml implements the YAML subset used by Trest Systems configuration
// files. It intentionally supports mappings, sequences, quoted/plain scalars,
// inline lists/maps, comments, and indentation-based nesting. The public API
// mirrors yaml.v3's Unmarshal entry point used by this project.
package yaml

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type line struct {
	indent int
	text   string
	number int
}

type parser struct {
	lines []line
	pos   int
}

// Unmarshal parses the supported YAML subset into out.
func Unmarshal(data []byte, out any) error {
	if out == nil {
		return fmt.Errorf("yaml: nil output")
	}
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("yaml: output must be a non-nil pointer")
	}
	lines, err := tokenize(string(data))
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}
	p := &parser{lines: lines}
	node, err := p.parseBlock(lines[0].indent)
	if err != nil {
		return err
	}
	if p.pos != len(p.lines) {
		return fmt.Errorf("yaml: unexpected content at line %d", p.lines[p.pos].number)
	}
	return assign(v.Elem(), node, "")
}

func tokenize(src string) ([]line, error) {
	src = strings.TrimPrefix(src, "\ufeff")
	raw := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	out := make([]line, 0, len(raw))
	for i, original := range raw {
		if strings.ContainsRune(original, '\t') {
			return nil, fmt.Errorf("yaml: tabs are not supported (line %d)", i+1)
		}
		trimmedComment := stripComment(original)
		if strings.TrimSpace(trimmedComment) == "" || strings.TrimSpace(trimmedComment) == "---" || strings.TrimSpace(trimmedComment) == "..." {
			continue
		}
		indent := 0
		for indent < len(trimmedComment) && trimmedComment[indent] == ' ' {
			indent++
		}
		out = append(out, line{indent: indent, text: strings.TrimSpace(trimmedComment), number: i + 1})
	}
	return out, nil
}

func stripComment(s string) string {
	var quote rune
	escaped := false
	depthSquare, depthCurly := 0, 0
	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case '[':
			depthSquare++
		case ']':
			if depthSquare > 0 {
				depthSquare--
			}
		case '{':
			depthCurly++
		case '}':
			if depthCurly > 0 {
				depthCurly--
			}
		case '#':
			if depthSquare == 0 && depthCurly == 0 && (i == 0 || unicode.IsSpace(rune(s[i-1]))) {
				return strings.TrimRight(s[:i], " \t")
			}
		}
	}
	return strings.TrimRight(s, " \t")
}

func (p *parser) parseBlock(indent int) (any, error) {
	if p.pos >= len(p.lines) {
		return nil, nil
	}
	if p.lines[p.pos].indent < indent {
		return nil, nil
	}
	if p.lines[p.pos].indent > indent {
		return nil, fmt.Errorf("yaml: unexpected indentation at line %d", p.lines[p.pos].number)
	}
	if isSequenceLine(p.lines[p.pos].text) {
		return p.parseSequence(indent)
	}
	return p.parseMapping(indent)
}

func isSequenceLine(text string) bool { return text == "-" || strings.HasPrefix(text, "- ") }

func (p *parser) parseMapping(indent int) (map[string]any, error) {
	result := make(map[string]any)
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		if ln.indent < indent {
			break
		}
		if ln.indent > indent {
			return nil, fmt.Errorf("yaml: unexpected indentation at line %d", ln.number)
		}
		if isSequenceLine(ln.text) {
			break
		}
		key, raw, ok := splitKeyValue(ln.text)
		if !ok || key == "" {
			return nil, fmt.Errorf("yaml: expected key: value at line %d", ln.number)
		}
		keyValue, err := parseKey(key)
		if err != nil {
			return nil, fmt.Errorf("yaml: line %d: %w", ln.number, err)
		}
		if _, exists := result[keyValue]; exists {
			return nil, fmt.Errorf("yaml: duplicate key %q at line %d", keyValue, ln.number)
		}
		p.pos++
		if strings.TrimSpace(raw) != "" {
			value, err := parseInline(raw)
			if err != nil {
				return nil, fmt.Errorf("yaml: line %d: %w", ln.number, err)
			}
			result[keyValue] = value
			continue
		}
		if p.pos >= len(p.lines) || p.lines[p.pos].indent <= indent {
			result[keyValue] = nil
			continue
		}
		value, err := p.parseBlock(p.lines[p.pos].indent)
		if err != nil {
			return nil, err
		}
		result[keyValue] = value
	}
	return result, nil
}

func (p *parser) parseSequence(indent int) ([]any, error) {
	result := make([]any, 0)
	for p.pos < len(p.lines) {
		ln := p.lines[p.pos]
		if ln.indent < indent {
			break
		}
		if ln.indent != indent || !isSequenceLine(ln.text) {
			break
		}
		rest := strings.TrimSpace(strings.TrimPrefix(ln.text, "-"))
		p.pos++
		if rest == "" {
			if p.pos >= len(p.lines) || p.lines[p.pos].indent <= indent {
				result = append(result, nil)
				continue
			}
			value, err := p.parseBlock(p.lines[p.pos].indent)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
			continue
		}

		if key, raw, ok := splitKeyValue(rest); ok && !strings.HasPrefix(rest, "{") {
			item := make(map[string]any)
			parsedKey, err := parseKey(key)
			if err != nil {
				return nil, fmt.Errorf("yaml: line %d: %w", ln.number, err)
			}
			if strings.TrimSpace(raw) == "" {
				if p.pos < len(p.lines) && p.lines[p.pos].indent > indent {
					value, err := p.parseBlock(p.lines[p.pos].indent)
					if err != nil {
						return nil, err
					}
					item[parsedKey] = value
				} else {
					item[parsedKey] = nil
				}
			} else {
				value, err := parseInline(raw)
				if err != nil {
					return nil, fmt.Errorf("yaml: line %d: %w", ln.number, err)
				}
				item[parsedKey] = value
			}

			// A sequence mapping item commonly continues with sibling keys
			// indented beneath the dash (for example, name/type/path).
			if p.pos < len(p.lines) && p.lines[p.pos].indent > indent && !isSequenceLine(p.lines[p.pos].text) {
				continuationIndent := p.lines[p.pos].indent
				continuation, err := p.parseMapping(continuationIndent)
				if err != nil {
					return nil, err
				}
				for k, v := range continuation {
					if _, exists := item[k]; exists {
						return nil, fmt.Errorf("yaml: duplicate key %q in sequence item", k)
					}
					item[k] = v
				}
			}
			result = append(result, item)
			continue
		}

		value, err := parseInline(rest)
		if err != nil {
			return nil, fmt.Errorf("yaml: line %d: %w", ln.number, err)
		}
		result = append(result, value)
	}
	return result, nil
}

func splitKeyValue(s string) (string, string, bool) {
	var quote rune
	escaped := false
	square, curly := 0, 0
	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case '[':
			square++
		case ']':
			square--
		case '{':
			curly++
		case '}':
			curly--
		case ':':
			if square == 0 && curly == 0 {
				return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), true
			}
		}
	}
	return "", "", false
}

func parseKey(s string) (string, error) {
	v, err := parseInline(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	switch x := v.(type) {
	case string:
		return x, nil
	default:
		return fmt.Sprint(x), nil
	}
}

func parseInline(raw string) (any, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	if strings.HasPrefix(s, "[") {
		if !strings.HasSuffix(s, "]") {
			return nil, fmt.Errorf("unterminated inline sequence")
		}
		body := strings.TrimSpace(s[1 : len(s)-1])
		if body == "" {
			return []any{}, nil
		}
		parts, err := splitTopLevel(body, ',')
		if err != nil {
			return nil, err
		}
		out := make([]any, 0, len(parts))
		for _, part := range parts {
			value, err := parseInline(part)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		return out, nil
	}
	if strings.HasPrefix(s, "{") {
		if !strings.HasSuffix(s, "}") {
			return nil, fmt.Errorf("unterminated inline mapping")
		}
		body := strings.TrimSpace(s[1 : len(s)-1])
		out := make(map[string]any)
		if body == "" {
			return out, nil
		}
		parts, err := splitTopLevel(body, ',')
		if err != nil {
			return nil, err
		}
		for _, part := range parts {
			key, valueRaw, ok := splitKeyValue(part)
			if !ok {
				return nil, fmt.Errorf("invalid inline mapping item %q", part)
			}
			parsedKey, err := parseKey(key)
			if err != nil {
				return nil, err
			}
			value, err := parseInline(valueRaw)
			if err != nil {
				return nil, err
			}
			out[parsedKey] = value
		}
		return out, nil
	}
	if strings.HasPrefix(s, "\"") {
		if !strings.HasSuffix(s, "\"") || len(s) < 2 {
			return nil, fmt.Errorf("unterminated double-quoted string")
		}
		v, err := strconv.Unquote(s)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	if strings.HasPrefix(s, "'") {
		if !strings.HasSuffix(s, "'") || len(s) < 2 {
			return nil, fmt.Errorf("unterminated single-quoted string")
		}
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), nil
	}
	switch strings.ToLower(s) {
	case "true", "yes", "on":
		return true, nil
	case "false", "no", "off":
		return false, nil
	case "null", "~":
		return nil, nil
	}
	if n, err := strconv.ParseInt(s, 0, 64); err == nil {
		return n, nil
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n, nil
	}
	return s, nil
}

func splitTopLevel(s string, separator rune) ([]string, error) {
	var quote rune
	escaped := false
	square, curly := 0, 0
	start := 0
	parts := make([]string, 0)
	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case '[':
			square++
		case ']':
			square--
		case '{':
			curly++
		case '}':
			curly--
		default:
			if r == separator && square == 0 && curly == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
		if square < 0 || curly < 0 {
			return nil, fmt.Errorf("unbalanced inline collection")
		}
	}
	if quote != 0 || square != 0 || curly != 0 {
		return nil, fmt.Errorf("unterminated inline collection")
	}
	parts = append(parts, strings.TrimSpace(s[start:]))
	return parts, nil
}

func assign(dst reflect.Value, src any, path string) error {
	if !dst.CanSet() {
		return fmt.Errorf("yaml: cannot set %s", path)
	}
	if dst.Kind() == reflect.Pointer {
		if src == nil {
			dst.SetZero()
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assign(dst.Elem(), src, path)
	}
	if src == nil {
		dst.SetZero()
		return nil
	}

	if dst.CanAddr() && dst.Addr().CanInterface() {
		if unmarshaler, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText([]byte(fmt.Sprint(src)))
		}
	}

	switch dst.Kind() {
	case reflect.Interface:
		dst.Set(reflect.ValueOf(src))
		return nil
	case reflect.Struct:
		m, ok := src.(map[string]any)
		if !ok {
			return typeError(path, "mapping", src)
		}
		fieldMap := structFields(dst.Type())
		for key, value := range m {
			index, found := fieldMap[key]
			if !found {
				// Match case-insensitively as a final convenience.
				for candidate, idx := range fieldMap {
					if strings.EqualFold(candidate, key) {
						index, found = idx, true
						break
					}
				}
			}
			if !found {
				continue // yaml.v3 ignores unknown fields by default.
			}
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			if err := assign(dst.Field(index), value, childPath); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		m, ok := src.(map[string]any)
		if !ok {
			return typeError(path, "mapping", src)
		}
		if dst.IsNil() {
			dst.Set(reflect.MakeMapWithSize(dst.Type(), len(m)))
		}
		for key, value := range m {
			k := reflect.New(dst.Type().Key()).Elem()
			if err := assign(k, key, path+".<key>"); err != nil {
				return err
			}
			v := reflect.New(dst.Type().Elem()).Elem()
			if err := assign(v, value, path+"."+key); err != nil {
				return err
			}
			dst.SetMapIndex(k, v)
		}
		return nil
	case reflect.Slice:
		items, ok := src.([]any)
		if !ok {
			return typeError(path, "sequence", src)
		}
		out := reflect.MakeSlice(dst.Type(), len(items), len(items))
		for i, item := range items {
			if err := assign(out.Index(i), item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		dst.Set(out)
		return nil
	case reflect.Array:
		items, ok := src.([]any)
		if !ok {
			return typeError(path, "sequence", src)
		}
		if len(items) != dst.Len() {
			return fmt.Errorf("yaml: %s: array length %d, got %d", path, dst.Len(), len(items))
		}
		for i, item := range items {
			if err := assign(dst.Index(i), item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		dst.SetString(fmt.Sprint(src))
		return nil
	case reflect.Bool:
		v, ok := src.(bool)
		if !ok {
			parsed, err := strconv.ParseBool(fmt.Sprint(src))
			if err != nil {
				return typeError(path, "boolean", src)
			}
			v = parsed
		}
		dst.SetBool(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		switch v := src.(type) {
		case int64:
			n = v
		case float64:
			n = int64(v)
		default:
			parsed, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
			if err != nil {
				return typeError(path, "integer", src)
			}
			n = parsed
		}
		if dst.OverflowInt(n) {
			return fmt.Errorf("yaml: %s: integer overflow", path)
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		var n uint64
		switch v := src.(type) {
		case int64:
			if v < 0 {
				return typeError(path, "unsigned integer", src)
			}
			n = uint64(v)
		case float64:
			if v < 0 {
				return typeError(path, "unsigned integer", src)
			}
			n = uint64(v)
		default:
			parsed, err := strconv.ParseUint(fmt.Sprint(v), 10, 64)
			if err != nil {
				return typeError(path, "unsigned integer", src)
			}
			n = parsed
		}
		if dst.OverflowUint(n) {
			return fmt.Errorf("yaml: %s: integer overflow", path)
		}
		dst.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		var n float64
		switch v := src.(type) {
		case float64:
			n = v
		case int64:
			n = float64(v)
		default:
			parsed, err := strconv.ParseFloat(fmt.Sprint(v), 64)
			if err != nil {
				return typeError(path, "number", src)
			}
			n = parsed
		}
		if dst.OverflowFloat(n) {
			return fmt.Errorf("yaml: %s: floating-point overflow", path)
		}
		dst.SetFloat(n)
		return nil
	default:
		return fmt.Errorf("yaml: unsupported destination type %s at %s", dst.Type(), path)
	}
}

func structFields(t reflect.Type) map[string]int {
	out := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		tag := field.Tag.Get("yaml")
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = toSnake(field.Name)
		}
		out[name] = i
		out[field.Name] = i
	}
	return out
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func typeError(path, expected string, got any) error {
	if path == "" {
		path = "<root>"
	}
	return fmt.Errorf("yaml: %s: expected %s, got %T", path, expected, got)
}
