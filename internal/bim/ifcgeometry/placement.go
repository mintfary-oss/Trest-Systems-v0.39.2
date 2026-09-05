package ifcgeometry

import (
	"fmt"
	"math"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

// Mat4 is a row-major 4x4 affine transform. Points are column vectors.
type Mat4 [4][4]float64

func Identity() Mat4 {
	return Mat4{{1, 0, 0, 0}, {0, 1, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}}
}

func (m Mat4) Mul(n Mat4) Mat4 {
	var r Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return r
}

func (m Mat4) Point(v Vec3) Vec3 {
	return Vec3{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3],
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3],
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3],
	}
}

type Transform struct {
	Matrix Mat4 `json:"matrix"`
	Depth  int  `json:"depth"`
}

// ResolvePlacement follows IfcLocalPlacement -> PlacementRelTo recursively.
func ResolvePlacement(entities []ifc.Entity, placementID int) (Transform, error) {
	byID := indexEntities(entities)
	return resolvePlacement(byID, placementID, map[int]bool{})
}

func resolvePlacement(byID map[int]ifc.Entity, id int, visiting map[int]bool) (Transform, error) {
	if id == 0 {
		return Transform{Matrix: Identity()}, nil
	}
	if visiting[id] {
		return Transform{}, fmt.Errorf("placement cycle at #%d", id)
	}
	e, ok := byID[id]
	if !ok {
		return Transform{}, fmt.Errorf("placement #%d not found", id)
	}
	if !strings.EqualFold(e.Type, "IFCLOCALPLACEMENT") || len(e.Attributes) < 1 {
		return Transform{}, fmt.Errorf("entity #%d is not IfcLocalPlacement", id)
	}
	visiting[id] = true
	defer delete(visiting, id)
	rp, err := ref(e.Attributes[0])
	if err != nil {
		return Transform{}, fmt.Errorf("placement #%d relative: %w", id, err)
	}
	local, err := axisPlacement(byID, rp)
	if err != nil {
		return Transform{}, err
	}
	parent := Identity()
	depth := 1
	if len(e.Attributes) > 1 && strings.TrimSpace(e.Attributes[1]) != "$" {
		pid, err := ref(e.Attributes[1])
		if err != nil {
			return Transform{}, err
		}
		p, err := resolvePlacement(byID, pid, visiting)
		if err != nil {
			return Transform{}, err
		}
		parent = p.Matrix
		depth += p.Depth
	}
	return Transform{Matrix: parent.Mul(local), Depth: depth}, nil
}

func axisPlacement(byID map[int]ifc.Entity, id int) (Mat4, error) {
	e, ok := byID[id]
	if !ok {
		return Mat4{}, fmt.Errorf("axis placement #%d not found", id)
	}
	if !strings.EqualFold(e.Type, "IFCAXIS2PLACEMENT3D") || len(e.Attributes) < 1 {
		return Mat4{}, fmt.Errorf("entity #%d is not IfcAxis2Placement3D", id)
	}
	locID, err := ref(e.Attributes[0])
	if err != nil {
		return Mat4{}, err
	}
	loc, ok := byID[locID]
	if !ok {
		return Mat4{}, fmt.Errorf("location #%d not found", locID)
	}
	t, err := cartesianPoint(loc)
	if err != nil {
		return Mat4{}, err
	}
	z := Vec3{0, 0, 1}
	x := Vec3{1, 0, 0}
	if len(e.Attributes) > 1 && strings.TrimSpace(e.Attributes[1]) != "$" {
		id, er := ref(e.Attributes[1])
		if er != nil {
			return Mat4{}, er
		}
		z, err = direction(byID[id])
		if err != nil {
			return Mat4{}, err
		}
	}
	if len(e.Attributes) > 2 && strings.TrimSpace(e.Attributes[2]) != "$" {
		id, er := ref(e.Attributes[2])
		if er != nil {
			return Mat4{}, er
		}
		x, err = direction(byID[id])
		if err != nil {
			return Mat4{}, err
		}
	}
	z, err = normalize(z)
	if err != nil {
		return Mat4{}, err
	}
	// Make X orthogonal to Z, then derive right-handed Y.
	x = sub(x, mul(z, dot(x, z)))
	x, err = normalize(x)
	if err != nil {
		return Mat4{}, fmt.Errorf("invalid axis placement #%d: %w", id, err)
	}
	y, err := normalize(cross(z, x))
	if err != nil {
		return Mat4{}, err
	}
	x = cross(y, z)
	return Mat4{{x.X, y.X, z.X, t.X}, {x.Y, y.Y, z.Y, t.Y}, {x.Z, y.Z, z.Z, t.Z}, {0, 0, 0, 1}}, nil
}

func direction(e ifc.Entity) (Vec3, error) {
	if !strings.EqualFold(e.Type, "IFCDIRECTION") || len(e.Attributes) < 1 {
		return Vec3{}, fmt.Errorf("entity #%d is not IfcDirection", e.ID)
	}
	a := strings.Trim(e.Attributes[0], "() ")
	p := strings.Split(a, ",")
	if len(p) < 3 {
		return Vec3{}, fmt.Errorf("direction #%d needs 3 ratios", e.ID)
	}
	v := Vec3{}
	var err error
	if v.X, err = parseFloat(p[0]); err != nil {
		return Vec3{}, err
	}
	if v.Y, err = parseFloat(p[1]); err != nil {
		return Vec3{}, err
	}
	if v.Z, err = parseFloat(p[2]); err != nil {
		return Vec3{}, err
	}
	return v, nil
}
func parseFloat(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscan(strings.TrimSpace(s), &v)
	return v, err
}
func dot(a, b Vec3) float64      { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func sub(a, b Vec3) Vec3         { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func mul(a Vec3, s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func cross(a, b Vec3) Vec3       { return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X} }
func normalize(v Vec3) (Vec3, error) {
	n := math.Sqrt(dot(v, v))
	if n < 1e-15 {
		return Vec3{}, fmt.Errorf("zero direction vector")
	}
	return mul(v, 1/n), nil
}

func indexEntities(entities []ifc.Entity) map[int]ifc.Entity {
	m := make(map[int]ifc.Entity, len(entities))
	for _, e := range entities {
		m[e.ID] = e
	}
	return m
}
