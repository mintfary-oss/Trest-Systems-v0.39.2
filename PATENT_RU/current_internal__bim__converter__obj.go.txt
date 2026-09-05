package converter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type Vec3 struct{ X, Y, Z float64 }
type Mesh struct {
	Positions []float64 `json:"positions"`
	Indices   []uint32  `json:"indices"`
}
type glTF struct {
	Asset       asset        `json:"asset"`
	Scene       int          `json:"scene"`
	Scenes      []scene      `json:"scenes"`
	Nodes       []node       `json:"nodes"`
	Meshes      []mesh       `json:"meshes"`
	Buffers     []buffer     `json:"buffers"`
	BufferViews []bufferView `json:"bufferViews"`
	Accessors   []accessor   `json:"accessors"`
}
type asset struct {
	Version   string `json:"version"`
	Generator string `json:"generator"`
}
type scene struct {
	Nodes []int `json:"nodes"`
}
type node struct {
	Mesh int `json:"mesh"`
}
type mesh struct {
	Primitives []primitive `json:"primitives"`
}
type primitive struct {
	Attributes map[string]int `json:"attributes"`
	Indices    int            `json:"indices"`
	Mode       int            `json:"mode"`
}
type buffer struct {
	URI        string `json:"uri"`
	ByteLength int    `json:"byteLength"`
}
type bufferView struct {
	Buffer     int `json:"buffer"`
	ByteOffset int `json:"byteOffset"`
	ByteLength int `json:"byteLength"`
	Target     int `json:"target,omitempty"`
}
type accessor struct {
	BufferView    int       `json:"bufferView"`
	ComponentType int       `json:"componentType"`
	Count         int       `json:"count"`
	Type          string    `json:"type"`
	Min           []float64 `json:"min,omitempty"`
	Max           []float64 `json:"max,omitempty"`
}

func ParseOBJ(r io.Reader) (Mesh, error) {
	var verts []Vec3
	var idx []uint32
	s := bufio.NewScanner(r)
	line := 0
	for s.Scan() {
		line++
		f := strings.Fields(s.Text())
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "v":
			if len(f) < 4 {
				return Mesh{}, fmt.Errorf("line %d: invalid vertex", line)
			}
			x, e1 := strconv.ParseFloat(f[1], 64)
			y, e2 := strconv.ParseFloat(f[2], 64)
			z, e3 := strconv.ParseFloat(f[3], 64)
			if e1 != nil || e2 != nil || e3 != nil {
				return Mesh{}, fmt.Errorf("line %d: invalid vertex", line)
			}
			verts = append(verts, Vec3{x, y, z})
		case "f":
			if len(f) < 4 {
				return Mesh{}, fmt.Errorf("line %d: face needs 3 vertices", line)
			}
			face := make([]uint32, 0, len(f)-1)
			for _, tok := range f[1:] {
				p := strings.Split(tok, "/")
				n, e := strconv.Atoi(p[0])
				if e != nil || n == 0 {
					return Mesh{}, fmt.Errorf("line %d: invalid face index", line)
				}
				if n < 0 {
					n = len(verts) + n + 1
				}
				if n < 1 || n > len(verts) {
					return Mesh{}, fmt.Errorf("line %d: face index out of range", line)
				}
				face = append(face, uint32(n-1))
			}
			for i := 1; i+1 < len(face); i++ {
				idx = append(idx, face[0], face[i], face[i+1])
			}
		}
	}
	if err := s.Err(); err != nil {
		return Mesh{}, err
	}
	if len(verts) == 0 || len(idx) == 0 {
		return Mesh{}, fmt.Errorf("OBJ contains no renderable geometry")
	}
	pos := make([]float64, 0, len(verts)*3)
	min, max := []float64{verts[0].X, verts[0].Y, verts[0].Z}, []float64{verts[0].X, verts[0].Y, verts[0].Z}
	for _, v := range verts {
		pos = append(pos, v.X, v.Y, v.Z)
		if v.X < min[0] {
			min[0] = v.X
		}
		if v.Y < min[1] {
			min[1] = v.Y
		}
		if v.Z < min[2] {
			min[2] = v.Z
		}
		if v.X > max[0] {
			max[0] = v.X
		}
		if v.Y > max[1] {
			max[1] = v.Y
		}
		if v.Z > max[2] {
			max[2] = v.Z
		}
	}
	return Mesh{Positions: pos, Indices: idx}, nil
}

func EncodeGLTF(w io.Writer, m Mesh) error {
	if len(m.Positions)%3 != 0 {
		return fmt.Errorf("positions must be XYZ")
	}
	b := make([]byte, 0, len(m.Positions)*4+len(m.Indices)*4)
	put32 := func(v uint32) { b = append(b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24)) }
	for _, v := range m.Positions {
		f := float32(v)
		put32(math.Float32bits(f))
	}
	posBytes := len(b)
	for _, v := range m.Indices {
		put32(v)
	}
	ibOffset := posBytes
	payload := make([]byte, len(b))
	copy(payload, b)
	uri := "data:application/octet-stream;base64," + base64(payload)
	g := glTF{Asset: asset{"2.0", "Trest Systems BIM Converter"}, Scene: 0, Scenes: []scene{{Nodes: []int{0}}}, Nodes: []node{{Mesh: 0}}, Meshes: []mesh{{Primitives: []primitive{{Attributes: map[string]int{"POSITION": 0}, Indices: 1, Mode: 4}}}}, Buffers: []buffer{{URI: uri, ByteLength: len(payload)}}, BufferViews: []bufferView{{Buffer: 0, ByteOffset: 0, ByteLength: posBytes, Target: 34962}, {Buffer: 0, ByteOffset: ibOffset, ByteLength: len(payload) - ibOffset, Target: 34963}}, Accessors: []accessor{{BufferView: 0, ComponentType: 5126, Count: len(m.Positions) / 3, Type: "VEC3"}, {BufferView: 1, ComponentType: 5125, Count: len(m.Indices), Type: "SCALAR"}}}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(g)
}
func base64(b []byte) string {
	const t = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var o strings.Builder
	for i := 0; i < len(b); i += 3 {
		var n uint32
		rem := len(b) - i
		n = uint32(b[i]) << 16
		if rem > 1 {
			n |= uint32(b[i+1]) << 8
		}
		if rem > 2 {
			n |= uint32(b[i+2])
		}
		o.WriteByte(t[(n>>18)&63])
		o.WriteByte(t[(n>>12)&63])
		if rem > 1 {
			o.WriteByte(t[(n>>6)&63])
		} else {
			o.WriteByte('=')
		}
		if rem > 2 {
			o.WriteByte(t[n&63])
		} else {
			o.WriteByte('=')
		}
	}
	return o.String()
}
