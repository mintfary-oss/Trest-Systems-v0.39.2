package ifcgeometry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
type Mesh struct {
	Vertices       []Vec3   `json:"vertices"`
	Indices        []uint32 `json:"indices"`
	SourceEntityID int      `json:"source_entity_id"`
	SourceType     string   `json:"source_type"`
}

// Extract supports two portable IFC geometry forms without an external SDK:
// IfcFacetedBrep/IfcClosedShell/IfcFace/IfcPolyLoop and IfcTriangulatedFaceSet.
// It intentionally does not pretend to tessellate arbitrary swept/BRep geometry.
func Extract(entities []ifc.Entity, geometryEntityID int) (Mesh, error) {
	byID := map[int]ifc.Entity{}
	for _, e := range entities {
		byID[e.ID] = e
	}
	root, ok := byID[geometryEntityID]
	if !ok {
		return Mesh{}, fmt.Errorf("geometry entity #%d not found", geometryEntityID)
	}
	switch strings.ToUpper(root.Type) {
	case "IFCFACETEDBREP":
		return facetedBrep(byID, root)
	case "IFCTRIANGULATEDFACESET":
		return triangulatedFaceSet(byID, root)
	case "IFCPOLYGONALFACESET":
		return polygonalFaceSet(byID, root)
	case "IFCEXTRUDEDAREASOLID":
		return extrudedAreaSolid(byID, root)
	default:
		return Mesh{}, fmt.Errorf("unsupported IFC geometry type %s", root.Type)
	}
}

func facetedBrep(byID map[int]ifc.Entity, root ifc.Entity) (Mesh, error) {
	if len(root.Attributes) < 1 {
		return Mesh{}, fmt.Errorf("IfcFacetedBrep #%d has no shell", root.ID)
	}
	shellID, err := ref(root.Attributes[0])
	if err != nil {
		return Mesh{}, err
	}
	shell, ok := byID[shellID]
	if !ok || strings.ToUpper(shell.Type) != "IFCCLOSEDSHELL" {
		return Mesh{}, fmt.Errorf("shell #%d is not IfcClosedShell", shellID)
	}
	faceIDs := refs(strings.Join(shell.Attributes, ","))
	var out Mesh
	out.SourceEntityID = root.ID
	out.SourceType = root.Type
	for _, fid := range faceIDs {
		face, ok := byID[fid]
		if !ok || len(face.Attributes) < 1 {
			continue
		}
		bounds := refs(face.Attributes[0])
		if len(bounds) == 0 {
			continue
		}
		bound, ok := byID[bounds[0]]
		if !ok || len(bound.Attributes) < 1 {
			continue
		}
		loopID, err := ref(bound.Attributes[0])
		if err != nil {
			continue
		}
		loop, ok := byID[loopID]
		if !ok || len(loop.Attributes) < 1 {
			continue
		}
		pointIDs := refs(loop.Attributes[0])
		poly := make([]Vec3, 0, len(pointIDs))
		for _, pid := range pointIDs {
			p, ok := byID[pid]
			if !ok {
				continue
			}
			v, err := cartesianPoint(p)
			if err == nil {
				poly = append(poly, v)
			}
		}
		appendFan(&out, poly)
	}
	if len(out.Indices) == 0 {
		return Mesh{}, fmt.Errorf("IfcFacetedBrep #%d produced no triangles", root.ID)
	}
	return out, nil
}

func triangulatedFaceSet(byID map[int]ifc.Entity, root ifc.Entity) (Mesh, error) {
	if len(root.Attributes) < 2 {
		return Mesh{}, fmt.Errorf("IfcTriangulatedFaceSet #%d has insufficient attributes", root.ID)
	}
	coordID, err := ref(root.Attributes[0])
	if err != nil {
		return Mesh{}, err
	}
	coords, ok := byID[coordID]
	if !ok || !strings.EqualFold(coords.Type, "IFCCARTESIANPOINTLIST3D") {
		return Mesh{}, fmt.Errorf("coordinates #%d are not IfcCartesianPointList3D", coordID)
	}
	verts, err := pointList3D(coords)
	if err != nil {
		return Mesh{}, err
	}
	idxAttr := root.Attributes[len(root.Attributes)-1]
	tuples := nestedIntTuples(idxAttr)
	if len(tuples) == 0 {
		return Mesh{}, fmt.Errorf("IfcTriangulatedFaceSet #%d has no CoordIndex", root.ID)
	}
	var out Mesh
	out.SourceEntityID = root.ID
	out.SourceType = root.Type
	out.Vertices = verts
	for _, t := range tuples {
		if len(t) != 3 {
			continue
		}
		for _, v := range t {
			if v < 1 || v > len(verts) {
				return Mesh{}, fmt.Errorf("CoordIndex %d out of range", v)
			}
			out.Indices = append(out.Indices, uint32(v-1))
		}
	}
	if len(out.Indices) == 0 {
		return Mesh{}, fmt.Errorf("IfcTriangulatedFaceSet #%d produced no triangles", root.ID)
	}
	return out, nil
}

func cartesianPoint(e ifc.Entity) (Vec3, error) {
	if len(e.Attributes) < 1 {
		return Vec3{}, fmt.Errorf("point #%d has no coordinates", e.ID)
	}
	a := strings.Trim(e.Attributes[0], "() ")
	p := strings.Split(a, ",")
	if len(p) < 3 {
		return Vec3{}, fmt.Errorf("point #%d needs 3 coordinates", e.ID)
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(p[0]), 64)
	if err != nil {
		return Vec3{}, err
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(p[1]), 64)
	if err != nil {
		return Vec3{}, err
	}
	z, err := strconv.ParseFloat(strings.TrimSpace(p[2]), 64)
	if err != nil {
		return Vec3{}, err
	}
	return Vec3{x, y, z}, nil
}
func pointList3D(e ifc.Entity) ([]Vec3, error) {
	raw := strings.TrimSpace(e.Attributes[0])
	raw = strings.TrimPrefix(raw, "(")
	raw = strings.TrimSuffix(raw, ")")
	tuples := nestedFloatTuples(raw)
	out := make([]Vec3, 0, len(tuples))
	for _, t := range tuples {
		if len(t) < 3 {
			continue
		}
		out = append(out, Vec3{t[0], t[1], t[2]})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("point list #%d is empty", e.ID)
	}
	return out, nil
}
func appendFan(m *Mesh, p []Vec3) {
	if len(p) < 3 {
		return
	}
	base := uint32(len(m.Vertices))
	m.Vertices = append(m.Vertices, p...)
	for i := 1; i+1 < len(p); i++ {
		m.Indices = append(m.Indices, base, base+uint32(i), base+uint32(i+1))
	}
}
func ref(s string) (int, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	return strconv.Atoi(s)
}
func refs(s string) []int {
	var out []int
	for _, tok := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '(' || r == ')' || r == ' ' || r == '\n' || r == '\t' }) {
		if strings.HasPrefix(tok, "#") {
			if n, e := ref(tok); e == nil {
				out = append(out, n)
			}
		}
	}
	return out
}
func nestedIntTuples(s string) [][]int {
	re := regexp.MustCompile(`\(([-+0-9]+)\s*,\s*([-+0-9]+)\s*,\s*([-+0-9]+)\)`)
	ms := re.FindAllStringSubmatch(s, -1)
	out := make([][]int, 0, len(ms))
	for _, m := range ms {
		a, e1 := strconv.Atoi(m[1])
		b, e2 := strconv.Atoi(m[2])
		c, e3 := strconv.Atoi(m[3])
		if e1 == nil && e2 == nil && e3 == nil {
			out = append(out, []int{a, b, c})
		}
	}
	return out
}
func nestedFloatTuples(s string) [][]float64 {
	re := regexp.MustCompile(`\(([-+0-9.eE]+)\s*,\s*([-+0-9.eE]+)\s*,\s*([-+0-9.eE]+)\)`)
	ms := re.FindAllStringSubmatch(s, -1)
	out := make([][]float64, 0, len(ms))
	for _, m := range ms {
		a, e1 := strconv.ParseFloat(m[1], 64)
		b, e2 := strconv.ParseFloat(m[2], 64)
		c, e3 := strconv.ParseFloat(m[3], 64)
		if e1 == nil && e2 == nil && e3 == nil {
			out = append(out, []float64{a, b, c})
		}
	}
	return out
}

func polygonalFaceSet(byID map[int]ifc.Entity, root ifc.Entity) (Mesh, error) {
	if len(root.Attributes) < 3 {
		return Mesh{}, fmt.Errorf("IfcPolygonalFaceSet #%d has insufficient attributes", root.ID)
	}
	coordID, err := ref(root.Attributes[0])
	if err != nil {
		return Mesh{}, err
	}
	coords, ok := byID[coordID]
	if !ok || !strings.EqualFold(coords.Type, "IFCCARTESIANPOINTLIST3D") {
		return Mesh{}, fmt.Errorf("coordinates #%d are not IfcCartesianPointList3D", coordID)
	}
	verts, err := pointList3D(coords)
	if err != nil {
		return Mesh{}, err
	}
	faceIDs := refs(root.Attributes[2])
	if len(faceIDs) == 0 {
		return Mesh{}, fmt.Errorf("IfcPolygonalFaceSet #%d has no faces", root.ID)
	}
	var out Mesh
	out.Vertices = verts
	out.SourceEntityID = root.ID
	out.SourceType = root.Type
	for _, fid := range faceIDs {
		f, ok := byID[fid]
		if !ok || !strings.EqualFold(f.Type, "IFCINDEXEDPOLYGON") || len(f.Attributes) < 1 {
			continue
		}
		ids := plainInts(f.Attributes[0])
		if len(ids) < 3 {
			continue
		}
		poly := make([]Vec3, 0, len(ids))
		for _, n := range ids {
			if n < 1 || n > len(verts) {
				return Mesh{}, fmt.Errorf("polygon index %d out of range", n)
			}
			poly = append(poly, verts[n-1])
		}
		appendFanIndexed(&out, poly)
	}
	if len(out.Indices) == 0 {
		return Mesh{}, fmt.Errorf("IfcPolygonalFaceSet #%d produced no triangles", root.ID)
	}
	return out, nil
}

func extrudedAreaSolid(byID map[int]ifc.Entity, root ifc.Entity) (Mesh, error) {
	if len(root.Attributes) < 4 {
		return Mesh{}, fmt.Errorf("IfcExtrudedAreaSolid #%d has insufficient attributes", root.ID)
	}
	profileID, err := ref(root.Attributes[0])
	if err != nil {
		return Mesh{}, err
	}
	profile, ok := byID[profileID]
	if !ok {
		return Mesh{}, fmt.Errorf("profile #%d not found", profileID)
	}
	dirID, err := ref(root.Attributes[2])
	if err != nil {
		return Mesh{}, err
	}
	dir, ok := byID[dirID]
	if !ok {
		return Mesh{}, fmt.Errorf("direction #%d not found", dirID)
	}
	d, err := direction(dir)
	if err != nil {
		return Mesh{}, err
	}
	d, err = normalize(d)
	if err != nil {
		return Mesh{}, err
	}
	depth, err := parseFloat(root.Attributes[3])
	if err != nil || depth <= 0 {
		return Mesh{}, fmt.Errorf("invalid extrusion depth")
	}
	base, err := profilePolygon(byID, profile)
	if err != nil {
		return Mesh{}, err
	}
	// Apply profile Position when available; default is identity.
	if len(profile.Attributes) > 1 && strings.TrimSpace(profile.Attributes[1]) != "$" {
		pid, e := ref(profile.Attributes[1])
		if e != nil {
			return Mesh{}, e
		}
		tr, e := axisPlacement(byID, pid)
		if e != nil {
			return Mesh{}, e
		}
		for i := range base {
			base[i] = tr.Point(base[i])
		}
	}
	var out Mesh
	out.SourceEntityID = root.ID
	out.SourceType = root.Type
	n := len(base)
	if n < 3 {
		return Mesh{}, fmt.Errorf("profile #%d has fewer than 3 points", profileID)
	}
	out.Vertices = append(out.Vertices, base...)
	for _, p := range base {
		out.Vertices = append(out.Vertices, add(p, mul(d, depth)))
	}
	for i := 1; i+1 < n; i++ {
		out.Indices = append(out.Indices, 0, uint32(i+1), uint32(i))
		out.Indices = append(out.Indices, uint32(n), uint32(n+i), uint32(n+i+1))
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		out.Indices = append(out.Indices, uint32(i), uint32(j), uint32(n+j))
		out.Indices = append(out.Indices, uint32(i), uint32(n+j), uint32(n+i))
	}
	return out, nil
}

func profilePolygon(byID map[int]ifc.Entity, p ifc.Entity) ([]Vec3, error) {
	switch strings.ToUpper(p.Type) {
	case "IFCRECTANGLEPROFILEDEF":
		if len(p.Attributes) < 4 {
			return nil, fmt.Errorf("rectangle profile #%d incomplete", p.ID)
		}
		x, e := parseFloat(p.Attributes[2])
		if e != nil {
			return nil, e
		}
		y, e := parseFloat(p.Attributes[3])
		if e != nil {
			return nil, e
		}
		return []Vec3{{-x / 2, -y / 2, 0}, {x / 2, -y / 2, 0}, {x / 2, y / 2, 0}, {-x / 2, y / 2, 0}}, nil
	case "IFCARBITRARYCLOSEDPROFILEDEF":
		if len(p.Attributes) < 3 {
			return nil, fmt.Errorf("arbitrary profile #%d incomplete", p.ID)
		}
		cid, e := ref(p.Attributes[2])
		if e != nil {
			return nil, e
		}
		c, ok := byID[cid]
		if !ok {
			return nil, fmt.Errorf("profile curve #%d not found", cid)
		}
		if strings.EqualFold(c.Type, "IFCPOLYLINE") && len(c.Attributes) > 0 {
			ids := refs(c.Attributes[0])
			out := make([]Vec3, 0, len(ids))
			for _, id := range ids {
				v, ok := byID[id]
				if !ok {
					return nil, fmt.Errorf("point #%d not found", id)
				}
				q, e := cartesianPoint(v)
				if e != nil {
					return nil, e
				}
				out = append(out, q)
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("unsupported profile type %s", p.Type)
}

func appendFanIndexed(m *Mesh, p []Vec3) {
	if len(p) < 3 {
		return
	}
	base := uint32(len(m.Vertices))
	m.Vertices = append(m.Vertices, p...)
	for i := 1; i+1 < len(p); i++ {
		m.Indices = append(m.Indices, base, base+uint32(i), base+uint32(i+1))
	}
}

func plainInts(s string) []int {
	var out []int
	for _, tok := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '(' || r == ')' || r == ' ' || r == '\n' || r == '\t' }) {
		if n, err := strconv.Atoi(strings.TrimSpace(tok)); err == nil {
			out = append(out, n)
		}
	}
	return out
}
