package converter

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

func EncodeGLB(w interface{ Write([]byte) (int, error) }, m Mesh) error {
	var bin bytes.Buffer
	put := func(v uint32) { _ = binary.Write(&bin, binary.LittleEndian, v) }
	for _, v := range m.Positions {
		put(math.Float32bits(float32(v)))
	}
	posBytes := bin.Len()
	for _, v := range m.Indices {
		put(v)
	}
	buf := bin.Bytes()
	if rem := len(buf) % 4; rem != 0 {
		buf = append(buf, bytes.Repeat([]byte{0}, 4-rem)...)
	}
	j := map[string]interface{}{"asset": map[string]string{"version": "2.0", "generator": "Trest Systems BIM Converter"}, "scene": 0, "scenes": []interface{}{map[string]interface{}{"nodes": []int{0}}}, "nodes": []interface{}{map[string]interface{}{"mesh": 0}}, "meshes": []interface{}{map[string]interface{}{"primitives": []interface{}{map[string]interface{}{"attributes": map[string]int{"POSITION": 0}, "indices": 1, "mode": 4}}}}, "buffers": []interface{}{map[string]interface{}{"byteLength": len(buf)}}, "bufferViews": []interface{}{map[string]interface{}{"buffer": 0, "byteOffset": 0, "byteLength": posBytes, "target": 34962}, map[string]interface{}{"buffer": 0, "byteOffset": posBytes, "byteLength": len(buf) - posBytes, "target": 34963}}, "accessors": []interface{}{map[string]interface{}{"bufferView": 0, "componentType": 5126, "count": len(m.Positions) / 3, "type": "VEC3"}, map[string]interface{}{"bufferView": 1, "componentType": 5125, "count": len(m.Indices), "type": "SCALAR"}}}
	jb, _ := json.Marshal(j)
	if rem := len(jb) % 4; rem != 0 {
		jb = append(jb, bytes.Repeat([]byte{' '}, 4-rem)...)
	}
	length := 12 + 8 + len(jb) + 8 + len(buf)
	h := make([]byte, 12)
	copy(h, "glTF")
	binary.LittleEndian.PutUint32(h[4:], 2)
	binary.LittleEndian.PutUint32(h[8:], uint32(length))
	if _, e := w.Write(h); e != nil {
		return e
	}
	ch := make([]byte, 8)
	binary.LittleEndian.PutUint32(ch, uint32(len(jb)))
	binary.LittleEndian.PutUint32(ch[4:], 0x4E4F534A)
	if _, e := w.Write(ch); e != nil {
		return e
	}
	if _, e := w.Write(jb); e != nil {
		return e
	}
	binary.LittleEndian.PutUint32(ch, uint32(len(buf)))
	binary.LittleEndian.PutUint32(ch[4:], 0x004E4942)
	if _, e := w.Write(ch); e != nil {
		return e
	}
	_, e := w.Write(buf)
	return e
}
func ValidateGLBHeader(b []byte) error {
	if len(b) < 12 {
		return fmt.Errorf("GLB too short")
	}
	if string(b[:4]) != "glTF" || binary.LittleEndian.Uint32(b[4:8]) != 2 {
		return fmt.Errorf("invalid GLB header")
	}
	return nil
}

func DecodeGLB(data []byte) (Mesh, error) {
	if err := ValidateGLBHeader(data); err != nil {
		return Mesh{}, err
	}
	decl := int(binary.LittleEndian.Uint32(data[8:12]))
	if decl > len(data) || decl < 12 {
		return Mesh{}, fmt.Errorf("invalid GLB length")
	}
	off := 12
	var doc map[string]any
	var bin []byte
	for off+8 <= decl {
		n := int(binary.LittleEndian.Uint32(data[off : off+4]))
		typ := binary.LittleEndian.Uint32(data[off+4 : off+8])
		off += 8
		if n < 0 || off+n > decl {
			return Mesh{}, fmt.Errorf("invalid GLB chunk")
		}
		chunk := data[off : off+n]
		off += n
		switch typ {
		case 0x4E4F534A:
			if err := json.Unmarshal(bytes.TrimSpace(chunk), &doc); err != nil {
				return Mesh{}, fmt.Errorf("invalid GLB JSON: %w", err)
			}
		case 0x004E4942:
			bin = append([]byte(nil), chunk...)
		}
	}
	if doc == nil || len(bin) == 0 {
		return Mesh{}, fmt.Errorf("GLB missing JSON or BIN chunk")
	}
	meshes, _ := doc["meshes"].([]any)
	if len(meshes) == 0 {
		return Mesh{}, fmt.Errorf("mesh not found")
	}
	m0, _ := meshes[0].(map[string]any)
	ps, _ := m0["primitives"].([]any)
	if len(ps) == 0 {
		return Mesh{}, fmt.Errorf("primitive not found")
	}
	p, _ := ps[0].(map[string]any)
	attrs, _ := p["attributes"].(map[string]any)
	pi, ok := attrs["POSITION"].(float64)
	if !ok {
		return Mesh{}, fmt.Errorf("POSITION accessor missing")
	}
	ii, ok := p["indices"].(float64)
	if !ok {
		return Mesh{}, fmt.Errorf("indices accessor missing")
	}
	accessors, _ := doc["accessors"].([]any)
	views, _ := doc["bufferViews"].([]any)
	read := func(ai int, index bool) ([]float64, []uint32, error) {
		if ai < 0 || ai >= len(accessors) {
			return nil, nil, fmt.Errorf("accessor out of range")
		}
		a, _ := accessors[ai].(map[string]any)
		vi, ok := a["bufferView"].(float64)
		if !ok {
			return nil, nil, fmt.Errorf("bufferView missing")
		}
		if int(vi) >= len(views) {
			return nil, nil, fmt.Errorf("bufferView out of range")
		}
		v, _ := views[int(vi)].(map[string]any)
		bo := 0
		if x, ok := v["byteOffset"].(float64); ok {
			bo = int(x)
		}
		ao := 0
		if x, ok := a["byteOffset"].(float64); ok {
			ao = int(x)
		}
		start := bo + ao
		count := int(a["count"].(float64))
		ct := int(a["componentType"].(float64))
		typ, _ := a["type"].(string)
		if index {
			out := make([]uint32, count)
			for i := 0; i < count; i++ {
				q := start
				switch ct {
				case 5121:
					q = start + i
					if q+1 > len(bin) {
						return nil, nil, fmt.Errorf("index data truncated")
					}
					out[i] = uint32(bin[q])
				case 5123:
					q = start + i*2
					if q+2 > len(bin) {
						return nil, nil, fmt.Errorf("index data truncated")
					}
					out[i] = uint32(binary.LittleEndian.Uint16(bin[q : q+2]))
				case 5125:
					q = start + i*4
					if q+4 > len(bin) {
						return nil, nil, fmt.Errorf("index data truncated")
					}
					out[i] = binary.LittleEndian.Uint32(bin[q : q+4])
				default:
					return nil, nil, fmt.Errorf("unsupported index componentType %d", ct)
				}
			}
			return nil, out, nil
		}
		if ct != 5126 || typ != "VEC3" {
			return nil, nil, fmt.Errorf("unsupported POSITION accessor")
		}
		out := make([]float64, count*3)
		for i := range out {
			q := start + i*4
			if q+4 > len(bin) {
				return nil, nil, fmt.Errorf("position data truncated")
			}
			out[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bin[q : q+4])))
		}
		return out, nil, nil
	}
	pos, _, err := read(int(pi), false)
	if err != nil {
		return Mesh{}, err
	}
	_, idx, err := read(int(ii), true)
	if err != nil {
		return Mesh{}, err
	}
	return Mesh{Positions: pos, Indices: idx}, nil
}
