package diff

import "math"

type Result struct {
	AddedVertices    int     `json:"added_vertices"`
	RemovedVertices  int     `json:"removed_vertices"`
	MovedVertices    int     `json:"moved_vertices"`
	AddedTriangles   int     `json:"added_triangles"`
	RemovedTriangles int     `json:"removed_triangles"`
	MaxDisplacement  float64 `json:"max_displacement"`
}

func Compare(oldPos []float64, oldIdx []uint32, newPos []float64, newIdx []uint32, tolerance float64) Result {
	if tolerance < 0 {
		tolerance = 0
	}
	r := Result{AddedVertices: max0(len(newPos)/3 - len(oldPos)/3), RemovedVertices: max0(len(oldPos)/3 - len(newPos)/3), AddedTriangles: max0(len(newIdx)/3 - len(oldIdx)/3), RemovedTriangles: max0(len(oldIdx)/3 - len(newIdx)/3)}
	n := len(oldPos)
	if len(newPos) < n {
		n = len(newPos)
	}
	for i := 0; i+2 < n; i += 3 {
		dx := newPos[i] - oldPos[i]
		dy := newPos[i+1] - oldPos[i+1]
		dz := newPos[i+2] - oldPos[i+2]
		d := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if d > r.MaxDisplacement {
			r.MaxDisplacement = d
		}
		if d > tolerance {
			r.MovedVertices++
		}
	}
	return r
}
func max0(v int) int {
	if v > 0 {
		return v
	}
	return 0
}
