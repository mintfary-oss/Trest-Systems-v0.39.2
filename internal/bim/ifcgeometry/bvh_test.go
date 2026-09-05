package ifcgeometry

import "testing"

func TestBVHPicksNearestTriangle(t *testing.T) {
	m := Mesh{Vertices: []Vec3{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 0, 2}, {1, 0, 2}, {0, 1, 2}}, Indices: []uint32{0, 1, 2, 3, 4, 5}}
	b, err := BuildBVH(m)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := b.Pick(Ray{Origin: Vec3{.2, .2, 5}, Direction: Vec3{0, 0, -1}})
	if !ok {
		t.Fatal("no hit")
	}
	if h.Triangle != 1 || h.Distance != 3 {
		t.Fatalf("unexpected hit %+v", h)
	}
}
