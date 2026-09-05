package ifcgeometry

import (
	"fmt"
	"math"
)

type Ray struct{ Origin, Direction Vec3 }
type Hit struct {
	Triangle int     `json:"triangle"`
	Distance float64 `json:"distance"`
	Position Vec3    `json:"position"`
}

type aabb struct{ Min, Max Vec3 }
type bvhNode struct {
	Bounds      aabb
	Left, Right int
	Start, End  int
}
type BVH struct {
	Mesh      Mesh
	nodes     []bvhNode
	triangles []int
}

func BuildBVH(mesh Mesh) (*BVH, error) {
	if len(mesh.Indices)%3 != 0 {
		return nil, fmt.Errorf("indices length must be divisible by 3")
	}
	b := &BVH{Mesh: mesh, triangles: make([]int, len(mesh.Indices)/3)}
	for i := range b.triangles {
		b.triangles[i] = i
	}
	if len(b.triangles) > 0 {
		b.build(0, len(b.triangles))
	}
	return b, nil
}
func (b *BVH) build(start, end int) int {
	idx := len(b.nodes)
	b.nodes = append(b.nodes, bvhNode{Start: start, End: end, Left: -1, Right: -1})
	bounds := emptyAABB()
	cent := emptyAABB()
	for i := start; i < end; i++ {
		tb := b.triBounds(b.triangles[i])
		bounds = union(bounds, tb)
		cent = unionPoint(cent, center(tb))
	}
	b.nodes[idx].Bounds = bounds
	if end-start <= 4 {
		return idx
	}
	axis := largestAxis(cent)
	mid := (start + end) / 2
	sortTriangles(b, start, end, axis)
	b.nodes[idx].Left = b.build(start, mid)
	b.nodes[idx].Right = b.build(mid, end)
	b.nodes[idx].Start = 0
	b.nodes[idx].End = 0
	return idx
}
func (b *BVH) Pick(ray Ray) (Hit, bool) {
	d, err := normalize(ray.Direction)
	if err != nil {
		return Hit{}, false
	}
	ray.Direction = d
	best := Hit{Distance: math.Inf(1), Triangle: -1}
	found := false
	if len(b.nodes) == 0 {
		return Hit{}, false
	}
	stack := []int{0}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := b.nodes[n]
		if t, ok := intersectAABB(ray, node.Bounds, best.Distance); !ok || t > best.Distance {
			continue
		}
		if node.Left < 0 {
			for i := node.Start; i < node.End; i++ {
				tri := b.triangles[i]
				a, c, d := b.tri(tri)
				if t, p, ok := intersectTriangle(ray, [3]Vec3{a, c, d}); ok && t < best.Distance {
					best = Hit{Triangle: tri, Distance: t, Position: p}
					found = true
				}
			}
		} else {
			stack = append(stack, node.Left, node.Right)
		}
	}
	return best, found
}
func (b *BVH) tri(i int) (Vec3, Vec3, Vec3) {
	a := b.Mesh.Indices[i*3 : i*3+3]
	return b.Mesh.Vertices[a[0]], b.Mesh.Vertices[a[1]], b.Mesh.Vertices[a[2]]
}
func (b *BVH) triBounds(i int) aabb {
	a, c, d := b.tri(i)
	r := emptyAABB()
	r = unionPoint(r, a)
	r = unionPoint(r, c)
	r = unionPoint(r, d)
	return r
}
func emptyAABB() aabb {
	return aabb{Min: Vec3{math.Inf(1), math.Inf(1), math.Inf(1)}, Max: Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}}
}
func union(a, b aabb) aabb {
	return aabb{Min: Vec3{math.Min(a.Min.X, b.Min.X), math.Min(a.Min.Y, b.Min.Y), math.Min(a.Min.Z, b.Min.Z)}, Max: Vec3{math.Max(a.Max.X, b.Max.X), math.Max(a.Max.Y, b.Max.Y), math.Max(a.Max.Z, b.Max.Z)}}
}
func unionPoint(a aabb, p Vec3) aabb {
	return aabb{Min: Vec3{math.Min(a.Min.X, p.X), math.Min(a.Min.Y, p.Y), math.Min(a.Min.Z, p.Z)}, Max: Vec3{math.Max(a.Max.X, p.X), math.Max(a.Max.Y, p.Y), math.Max(a.Max.Z, p.Z)}}
}
func center(a aabb) Vec3 {
	return Vec3{(a.Min.X + a.Max.X) / 2, (a.Min.Y + a.Max.Y) / 2, (a.Min.Z + a.Max.Z) / 2}
}
func largestAxis(a aabb) int {
	dx, dy, dz := a.Max.X-a.Min.X, a.Max.Y-a.Min.Y, a.Max.Z-a.Min.Z
	if dy > dx && dy >= dz {
		return 1
	}
	if dz > dx && dz > dy {
		return 2
	}
	return 0
}
func sortTriangles(b *BVH, s, e, axis int) {
	for i := s + 1; i < e; i++ {
		v := b.triangles[i]
		j := i - 1
		cv := axisValue(center(b.triBounds(v)), axis)
		for j >= s && axisValue(center(b.triBounds(b.triangles[j])), axis) > cv {
			b.triangles[j+1] = b.triangles[j]
			j--
		}
		b.triangles[j+1] = v
	}
}
func axisValue(v Vec3, a int) float64 {
	if a == 1 {
		return v.Y
	}
	if a == 2 {
		return v.Z
	}
	return v.X
}
func intersectAABB(r Ray, b aabb, maxT float64) (float64, bool) {
	tmin := 0.0
	tmax := maxT
	for i := 0; i < 3; i++ {
		o, d, mn, mx := r.Origin.X, r.Direction.X, b.Min.X, b.Max.X
		if i == 1 {
			o, d, mn, mx = r.Origin.Y, r.Direction.Y, b.Min.Y, b.Max.Y
		}
		if i == 2 {
			o, d, mn, mx = r.Origin.Z, r.Direction.Z, b.Min.Z, b.Max.Z
		}
		if math.Abs(d) < 1e-15 {
			if o < mn || o > mx {
				return 0, false
			}
			continue
		}
		inv := 1 / d
		t1, t2 := (mn-o)*inv, (mx-o)*inv
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t1 > tmin {
			tmin = t1
		}
		if t2 < tmax {
			tmax = t2
		}
		if tmin > tmax {
			return 0, false
		}
	}
	return tmin, true
}
func intersectTriangle(r Ray, tri [3]Vec3) (float64, Vec3, bool) {
	e1 := sub(tri[1], tri[0])
	e2 := sub(tri[2], tri[0])
	h := cross(r.Direction, e2)
	a := dot(e1, h)
	if math.Abs(a) < 1e-12 {
		return 0, Vec3{}, false
	}
	f := 1 / a
	s := sub(r.Origin, tri[0])
	u := f * dot(s, h)
	if u < 0 || u > 1 {
		return 0, Vec3{}, false
	}
	q := cross(s, e1)
	v := f * dot(r.Direction, q)
	if v < 0 || u+v > 1 {
		return 0, Vec3{}, false
	}
	t := f * dot(e2, q)
	if t <= 1e-12 {
		return 0, Vec3{}, false
	}
	return t, add(r.Origin, mul(r.Direction, t)), true
}
func add(a, b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
