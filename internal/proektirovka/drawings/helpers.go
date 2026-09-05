// Package drawings generates construction drawing content per Russian GOST standards.
//
// Standards: ГОСТ 21.501-2018, ГОСТ 21.201-2011, ГОСТ 2.303-68
package drawings

import (
	"fmt"
	"math"

	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/building"
	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/dxf"
)

// px converts meters to paper millimeters at a given scale.
// Example: px(6.0, 100) = 60.0 mm
func px(meters float64, scale int) float64 {
	return meters * 1000.0 / float64(scale)
}

// AxisConfig holds resolved axis positions for a drawing.
type AxisConfig struct {
	OX, OY float64 // origin of axis grid on paper (mm) — where axis 1/А intersects
	Scale  int
	XPos   []float64 // X positions on paper for each axis
	YPos   []float64 // Y positions on paper for each axis
	XLbls  []string
	YLbls  []string
	BW     float64 // building width on paper (mm)
	BH     float64 // building depth on paper (mm)
}

// NewAxisConfig builds paper coordinates for the axis grid.
func NewAxisConfig(b *building.Building, ox, oy float64, scale int) AxisConfig {
	xc := b.Axes.XCoords()
	yc := b.Axes.YCoords()

	xpos := make([]float64, len(xc))
	for i, v := range xc {
		xpos[i] = ox + px(v, scale)
	}
	ypos := make([]float64, len(yc))
	for i, v := range yc {
		ypos[i] = oy + px(v, scale)
	}

	return AxisConfig{
		OX:    ox,
		OY:    oy,
		Scale: scale,
		XPos:  xpos,
		YPos:  ypos,
		XLbls: b.Axes.XLabels,
		YLbls: b.Axes.YLabels,
		BW:    px(b.Axes.TotalWidth(), scale),
		BH:    px(b.Axes.TotalDepth(), scale),
	}
}

// DrawAxes draws the structural axis grid with bubble circles per ГОСТ 21.501-2018.
func DrawAxes(d *dxf.Document, ac AxisConfig) {
	ext := px(1.2, ac.Scale)  // extension beyond building outline
	bub := px(0.35, ac.Scale) // bubble radius (~350mm → 3.5mm at 1:100)
	txtH := math.Max(2.5, px(0.22, ac.Scale))

	// X-axes (vertical lines, numbered 1…N)
	for i, x := range ac.XPos {
		// Axis line
		d.Line("ОСИ", dxf.LW018, x, ac.OY-ext, x, ac.OY+ac.BH+ext)

		lbl := ""
		if i < len(ac.XLbls) {
			lbl = ac.XLbls[i]
		}

		// Bottom bubble
		cy := ac.OY - ext - bub*1.2
		d.Circle("ТЕКСТ_ОСЕЙ", dxf.LW018, x, cy, bub)
		d.Text("ТЕКСТ_ОСЕЙ", "GOST", txtH, x-txtH*0.4, cy-txtH*0.45, lbl)

		// Top bubble
		cy2 := ac.OY + ac.BH + ext + bub*1.2
		d.Circle("ТЕКСТ_ОСЕЙ", dxf.LW018, x, cy2, bub)
		d.Text("ТЕКСТ_ОСЕЙ", "GOST", txtH, x-txtH*0.4, cy2-txtH*0.45, lbl)
	}

	// Y-axes (horizontal lines, lettered А…N)
	for j, y := range ac.YPos {
		d.Line("ОСИ", dxf.LW018, ac.OX-ext, y, ac.OX+ac.BW+ext, y)

		lbl := ""
		if j < len(ac.YLbls) {
			lbl = ac.YLbls[j]
		}

		// Left bubble
		cx := ac.OX - ext - bub*1.2
		d.Circle("ТЕКСТ_ОСЕЙ", dxf.LW018, cx, y, bub)
		d.Text("ТЕКСТ_ОСЕЙ", "GOST", txtH, cx-txtH*0.4, y-txtH*0.45, lbl)

		// Right bubble
		cx2 := ac.OX + ac.BW + ext + bub*1.2
		d.Circle("ТЕКСТ_ОСЕЙ", dxf.LW018, cx2, y, bub)
		d.Text("ТЕКСТ_ОСЕЙ", "GOST", txtH, cx2-txtH*0.4, y-txtH*0.45, lbl)
	}
}

// DrawWallH draws a horizontal wall segment centered on y, from x0 to x1.
// layer: "СТЕНЫ_НЕСУЩИЕ" for load-bearing, "ПЕРЕГОРОДКИ" for partitions.
func DrawWallH(d *dxf.Document, layer string, x0, y, x1, thick float64) {
	half := thick / 2.0
	pts := dxf.Rect(x0, y-half, x1, y+half)
	d.Polyline(layer, dxf.LW070, true, pts)
	if layer != "ПЕРЕГОРОДКИ" {
		d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254, pts)
	}
}

// DrawWallV draws a vertical wall segment centered on x, from y0 to y1.
func DrawWallV(d *dxf.Document, layer string, x, y0, y1, thick float64) {
	half := thick / 2.0
	pts := dxf.Rect(x-half, y0, x+half, y1)
	d.Polyline(layer, dxf.LW070, true, pts)
	if layer != "ПЕРЕГОРОДКИ" {
		d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254, pts)
	}
}

// DrawOuterWalls draws the four outer walls of the building outline.
func DrawOuterWalls(d *dxf.Document, ac AxisConfig, wallExt float64) {
	ew := px(wallExt, ac.Scale)
	ox, oy := ac.OX, ac.OY
	bw, bh := ac.BW, ac.BH

	DrawWallH(d, "СТЕНЫ_НЕСУЩИЕ", ox-ew, oy, ox+bw+ew, ew)    // South (Axis А)
	DrawWallH(d, "СТЕНЫ_НЕСУЩИЕ", ox-ew, oy+bh, ox+bw+ew, ew) // North (Axis last)
	DrawWallV(d, "СТЕНЫ_НЕСУЩИЕ", ox, oy-ew, oy+bh+ew, ew)    // West (Axis 1)
	DrawWallV(d, "СТЕНЫ_НЕСУЩИЕ", ox+bw, oy-ew, oy+bh+ew, ew) // East (Axis last)
}

// DrawDoor draws a door symbol per ГОСТ 21.201-2011:
// a line (door panel) + arc (swing path).
// hingeX, hingeY — hinge point; width — door width (mm on paper);
// angleDeg — direction the panel lies when closed (degrees, 0=right);
// openAngle — swing angle (typically 90°).
func DrawDoor(d *dxf.Document, hingeX, hingeY, width, angleDeg, openAngle float64) {
	rad := angleDeg * math.Pi / 180.0
	// Closed panel endpoint
	ex := hingeX + width*math.Cos(rad)
	ey := hingeY + width*math.Sin(rad)
	// Panel line
	d.Line("ПРОЕМЫ", dxf.LW018, hingeX, hingeY, ex, ey)
	// Swing arc
	d.Arc("ПРОЕМЫ", dxf.LW013, hingeX, hingeY, width, angleDeg, angleDeg+openAngle)
}

// DrawWindowH draws a window symbol in a horizontal wall (3 parallel lines per ГОСТ 21.201-2011).
// x0, x1 — extent of window opening; y — wall centerline; wallThick — wall thickness (paper mm).
func DrawWindowH(d *dxf.Document, x0, y, x1, wallThick float64) {
	half := wallThick / 2.0
	d.Line("ПРОЕМЫ", dxf.LW025, x0, y+half, x1, y+half) // outer face
	d.Line("ПРОЕМЫ", dxf.LW018, x0, y, x1, y)           // glass line
	d.Line("ПРОЕМЫ", dxf.LW025, x0, y-half, x1, y-half) // inner face
}

// DrawWindowV draws a window symbol in a vertical wall.
func DrawWindowV(d *dxf.Document, x, y0, y1, wallThick float64) {
	half := wallThick / 2.0
	d.Line("ПРОЕМЫ", dxf.LW025, x+half, y0, x+half, y1)
	d.Line("ПРОЕМЫ", dxf.LW018, x, y0, x, y1)
	d.Line("ПРОЕМЫ", dxf.LW025, x-half, y0, x-half, y1)
}

// DrawStaircase draws a staircase symbol per ГОСТ 21.201-2011:
// steps as parallel lines + direction arrow.
func DrawStaircase(d *dxf.Document, x0, y0, x1, y1 float64, nSteps int) {
	w := x1 - x0
	h := y1 - y0
	stepH := h / float64(nSteps)

	// Outline
	d.Polyline("ЛЕСТНИЦЫ", dxf.LW025, true, dxf.Rect(x0, y0, x1, y1))

	// Step lines (horizontal)
	for i := 1; i < nSteps; i++ {
		sy := y0 + float64(i)*stepH
		d.Line("ЛЕСТНИЦЫ", dxf.LW013, x0, sy, x1, sy)
	}

	// Direction arrow (up)
	mx := (x0 + x1) / 2.0
	arrowLen := h * 0.8
	ay0 := y0 + h*0.1
	ay1 := ay0 + arrowLen
	d.Line("ЛЕСТНИЦЫ", dxf.LW025, mx, ay0, mx, ay1)
	as := math.Min(w, h) * 0.08
	d.Line("ЛЕСТНИЦЫ", dxf.LW025, mx, ay1, mx-as, ay1-as*1.5)
	d.Line("ЛЕСТНИЦЫ", dxf.LW025, mx, ay1, mx+as, ay1-as*1.5)

	_ = w // suppress unused warning if not used elsewhere
}

// DrawElevator draws an elevator shaft symbol (rectangle with diagonal cross).
func DrawElevator(d *dxf.Document, x0, y0, x1, y1 float64) {
	d.Polyline("ЛЕСТНИЦЫ", dxf.LW025, true, dxf.Rect(x0, y0, x1, y1))
	d.Line("ЛЕСТНИЦЫ", dxf.LW013, x0, y0, x1, y1)
	d.Line("ЛЕСТНИЦЫ", dxf.LW013, x1, y0, x0, y1)
}

// DrawRoomLabel writes room number and area at the room center.
func DrawRoomLabel(d *dxf.Document, cx, cy float64, number, name string, area float64) {
	txtH := 2.5
	// Room number
	d.Text("ТЕКСТ", "GOST", txtH, cx-float64(len(number))*txtH*0.35, cy+txtH*0.5, number)
	// Room name (smaller)
	if name != "" {
		d.Text("ТЕКСТ", "GOST", txtH*0.75, cx-float64(len([]rune(name)))*txtH*0.28, cy-txtH*0.8, name)
	}
	// Area
	if area > 0 {
		areaStr := fmt.Sprintf("%.1f м²", area)
		d.Text("ТЕКСТ", "GOST", txtH*0.7, cx-float64(len([]rune(areaStr)))*txtH*0.27, cy-txtH*2.0, areaStr)
	}
}

// DrawChainDimsH draws a row of horizontal chain dimensions below a baseline y.
// points: X coordinates on paper (mm) to dimension between.
// offsetY: distance below y for the dimension line (positive = further below).
// scale: drawing scale for label calculation (real mm = paper mm * scale).
func DrawChainDimsH(d *dxf.Document, points []float64, baseY, offsetY float64, scale int) {
	if len(points) < 2 {
		return
	}
	dimY := baseY - offsetY

	for i := 0; i < len(points)-1; i++ {
		x1 := points[i]
		x2 := points[i+1]
		drawLinearDimH(d, x1, baseY, x2, baseY, dimY, scale)
	}
}

// DrawTotalDimH draws the overall horizontal dimension.
func DrawTotalDimH(d *dxf.Document, x0, baseY, x1, offsetY float64, scale int) {
	drawLinearDimH(d, x0, baseY, x1, baseY, baseY-offsetY, scale)
}

// DrawChainDimsV draws a row of vertical chain dimensions to the left of a baseline x.
func DrawChainDimsV(d *dxf.Document, points []float64, baseX, offsetX float64, scale int) {
	if len(points) < 2 {
		return
	}
	dimX := baseX - offsetX

	for i := 0; i < len(points)-1; i++ {
		y1 := points[i]
		y2 := points[i+1]
		drawLinearDimV(d, baseX, y1, baseX, y2, dimX, scale)
	}
}

// DrawTotalDimV draws the overall vertical dimension.
func DrawTotalDimV(d *dxf.Document, y0, baseX, y1, offsetX float64, scale int) {
	drawLinearDimV(d, baseX, y0, baseX, y1, baseX-offsetX, scale)
}

func drawLinearDimH(d *dxf.Document, x1, y1, x2, y2, dimY float64, scale int) {
	lbl := formatDimLabel(math.Abs(x2-x1), scale)
	arrow := 2.5
	lw := dxf.LW013

	// Dimension line
	d.Line("РАЗМЕРЫ", lw, x1, dimY, x2, dimY)
	// Extension lines
	d.Line("РАЗМЕРЫ", lw, x1, y1, x1, dimY)
	d.Line("РАЗМЕРЫ", lw, x2, y2, x2, dimY)
	// Tick marks (засечки) at 45°
	d.Line("РАЗМЕРЫ", lw, x1+arrow*0.6, dimY+arrow*0.6, x1-arrow*0.6, dimY-arrow*0.6)
	d.Line("РАЗМЕРЫ", lw, x2+arrow*0.6, dimY+arrow*0.6, x2-arrow*0.6, dimY-arrow*0.6)
	// Label above dimension line
	xm := (x1 + x2) / 2.0
	lblW := float64(len(lbl)) * 2.5 * 0.65
	d.Text("РАЗМЕРЫ", "GOST", 2.5, xm-lblW/2.0, dimY+1.2, lbl)
}

func drawLinearDimV(d *dxf.Document, x1, y1, x2, y2, dimX float64, scale int) {
	lbl := formatDimLabel(math.Abs(y2-y1), scale)
	arrow := 2.5
	lw := dxf.LW013

	d.Line("РАЗМЕРЫ", lw, dimX, y1, dimX, y2)
	d.Line("РАЗМЕРЫ", lw, x1, y1, dimX, y1)
	d.Line("РАЗМЕРЫ", lw, x2, y2, dimX, y2)
	d.Line("РАЗМЕРЫ", lw, dimX+arrow*0.6, y1+arrow*0.6, dimX-arrow*0.6, y1-arrow*0.6)
	d.Line("РАЗМЕРЫ", lw, dimX+arrow*0.6, y2+arrow*0.6, dimX-arrow*0.6, y2-arrow*0.6)

	ym := (y1 + y2) / 2.0
	d.TextRotated("РАЗМЕРЫ", "GOST", 2.5, dimX-4.5, ym-float64(len(lbl))*2.5*0.3, 90, lbl)
}

// formatDimLabel converts paper mm distance to real-world label string.
// Real dimension in mm = paper mm * scale. Output is always in mm (no units).
func formatDimLabel(paperMM float64, scale int) string {
	realMM := paperMM * float64(scale)
	// Round to nearest mm
	rounded := math.Round(realMM)
	return fmt.Sprintf("%d", int(rounded))
}

// AddElevationMark draws a level mark symbol (ГОСТ 21.501-2018).
// valueM — elevation in meters (e.g. +3.600 or -3.300).
func AddElevationMark(d *dxf.Document, x, y, valueM float64) {
	h := 4.0
	sign := "+"
	if valueM < 0 {
		sign = ""
	} else if valueM == 0 {
		sign = "±"
	}
	label := fmt.Sprintf("%s%.3f", sign, valueM)

	// Triangle pointing up
	pts := [][2]float64{
		{x, y},
		{x - h*0.5, y - h*0.9},
		{x + h*0.5, y - h*0.9},
	}
	d.Polyline("ОТМЕТКИ", dxf.LW018, true, pts)

	// Horizontal leader line
	d.Line("ОТМЕТКИ", dxf.LW013, x-h*1.5, y, x+h*4.0, y)

	// Label
	d.Text("ОТМЕТКИ", "GOST", 2.8, x+h*0.3, y+0.5, label)
}

// AddSectionMark draws a section cut mark (ГОСТ 21.501-2018).
func AddSectionMark(d *dxf.Document, x, y float64, label string) {
	r := 4.0
	d.Circle("РАЗМЕРЫ", dxf.LW025, x, y, r)
	d.Text("РАЗМЕРЫ", "GOST", 2.5, x-r*0.6, y-1.2, label+"-"+label)
	d.Line("РАЗМЕРЫ", dxf.LW025, x, y-r, x, y-r*2.5)
	// Arrow tip
	d.Line("РАЗМЕРЫ", dxf.LW025, x, y-r*2.5, x-1.5, y-r*2.0)
	d.Line("РАЗМЕРЫ", dxf.LW025, x, y-r*2.5, x+1.5, y-r*2.0)
}
