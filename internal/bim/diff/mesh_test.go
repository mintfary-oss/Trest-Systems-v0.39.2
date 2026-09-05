package diff

import "testing"

func TestCompare(t *testing.T) {
	r := Compare([]float64{0, 0, 0, 1, 0, 0}, []uint32{0, 1, 1}, []float64{0, 0, 0, 2, 0, 0, 3, 0, 0}, []uint32{0, 1, 2}, .5)
	if r.AddedVertices != 1 || r.MovedVertices != 1 || r.MaxDisplacement != 1 {
		t.Fatalf("%+v", r)
	}
}
