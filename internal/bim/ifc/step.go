package ifc

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Entity is a lossless-enough representation of one IFC STEP entity.
type Entity struct {
	ID         int
	Type       string
	Attributes []string
}

var entityRE = regexp.MustCompile(`^#([0-9]+)\s*=\s*([A-Za-z0-9_]+)\s*\((.*)\)\s*;\s*$`)

// ParseSTEP parses the physical IFC STEP entity records without depending on a third-party IFC SDK.
// It intentionally preserves attribute text; semantic decoding is layered on top.
func ParseSTEP(r io.Reader) ([]Entity, error) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 8*1024*1024)
	var out []Entity
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "HEADER") || strings.HasPrefix(line, "DATA") || line == "ENDSEC;" || line == "END-ISO-10303-21;" || line == "ISO-10303-21;" {
			continue
		}
		m := entityRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid IFC entity id: %w", err)
		}
		out = append(out, Entity{ID: id, Type: strings.ToUpper(m[2]), Attributes: splitAttributes(m[3])})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func splitAttributes(s string) []string {
	var out []string
	start, depth := 0, 0
	quoted := false
	for i, r := range s {
		switch r {
		case '\'':
			quoted = !quoted
		case '(':
			if !quoted {
				depth++
			}
		case ')':
			if !quoted && depth > 0 {
				depth--
			}
		case ',':
			if !quoted && depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if start < len(s) {
		out = append(out, strings.TrimSpace(s[start:]))
	}
	return out
}
