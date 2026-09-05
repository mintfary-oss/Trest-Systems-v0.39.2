package converter

import (
	"bytes"
	"fmt"
)

func ExampleDecodeGLB() {
	var b bytes.Buffer
	_ = EncodeGLB(&b, Mesh{Positions: []float64{0, 0, 0, 1, 0, 0, 0, 1, 0}, Indices: []uint32{0, 1, 2}})
	m, err := DecodeGLB(b.Bytes())
	if err != nil {
		panic(err)
	}
	fmt.Println(len(m.Positions), len(m.Indices))
	// Output: 9 3
}
