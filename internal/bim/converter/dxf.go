package converter

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseDXF reads a conservative ASCII DXF subset (POINT and 3DFACE) into a triangle mesh.
func ParseDXF(r io.Reader) (Mesh, error) {
	s := bufio.NewScanner(r)
	var codes []string
	for s.Scan() {
		codes = append(codes, strings.TrimSpace(s.Text()))
	}
	if err := s.Err(); err != nil {
		return Mesh{}, err
	}
	var p []float64
	var idx []uint32
	for i := 0; i+1 < len(codes); i += 2 {
		code, val := codes[i], codes[i+1]
		if code != "0" {
			continue
		}
		typ := strings.ToUpper(val)
		if typ == "3DFACE" {
			var xyz [12]float64
			found := 0
			for j := i + 2; j+1 < len(codes) && codes[j] != "0"; j += 2 {
				c := codes[j]
				v := codes[j+1]
				n, e := strconv.ParseFloat(v, 64)
				if e != nil {
					continue
				}
				switch c {
				case "10":
					xyz[0] = n
					found |= 1
				case "20":
					xyz[1] = n
					found |= 1
				case "30":
					xyz[2] = n
				case "11":
					xyz[3] = n
					found |= 2
				case "21":
					xyz[4] = n
					found |= 2
				case "31":
					xyz[5] = n
				case "12":
					xyz[6] = n
					found |= 4
				case "22":
					xyz[7] = n
					found |= 4
				case "32":
					xyz[8] = n
				case "13":
					xyz[9] = n
					found |= 8
				case "23":
					xyz[10] = n
				case "33":
					xyz[11] = n
				}
			}
			if found&7 == 7 {
				base := uint32(len(p) / 3)
				p = append(p, xyz[:]...)
				idx = append(idx, base, base+1, base+2)
				if found&8 != 0 {
					p = append(p, xyz[9], xyz[10], xyz[11])
					idx = append(idx, base, base+2, base+3)
				}
			}
		}
	}
	if len(idx) == 0 {
		return Mesh{}, fmt.Errorf("DXF contains no supported 3DFACE geometry")
	}
	return Mesh{Positions: p, Indices: idx}, nil
}
