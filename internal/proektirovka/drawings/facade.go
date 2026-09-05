package drawings

import (
	"fmt"
	"math"

	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/building"
	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/dxf"
	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/gost"
)

// FacadeConfig controls how a facade drawing is generated.
type FacadeConfig struct {
	Scale int    // e.g. 100 for 1:100
	Side  string // "south" | "north" | "west" | "east"
	Title string // override title
}

// DrawFacade generates a complete building facade drawing.
// ГОСТ 21.501-2018: фасады, отметки уровней, оси.
func DrawFacade(d *dxf.Document, b *building.Building, cfg FacadeConfig) {
	scale := cfg.Scale
	if scale == 0 {
		scale = 100
	}

	// ── Determine facade orientation ──
	isEW := cfg.Side == "west" || cfg.Side == "east" // east-west facade (shows Y-depth)
	var facadeLen float64                            // horizontal extent (m)
	var axisCoords []float64
	var axisLabels []string

	if isEW {
		facadeLen = b.Axes.TotalDepth()
		axisCoords = b.Axes.YCoords()
		axisLabels = b.Axes.YLabels
	} else {
		facadeLen = b.Axes.TotalWidth()
		axisCoords = b.Axes.XCoords()
		axisLabels = b.Axes.XLabels
	}

	// ── Paper layout ──
	ox := gost.ContentX0 + 90.0 // building facade left edge
	oy := gost.ContentY0 + 50.0 // ground level (±0.000)

	flMM := px(facadeLen, scale) // facade length on paper
	ew := px(b.Dims.WallExtM, scale)
	fh := px(b.Dims.FloorHM, scale)
	bsmH := px(b.Dims.BasementHM, scale)
	totH := float64(b.NFloors) * fh

	// ── Ground line (тонкая линия уровня земли) ──
	groundExt := px(3.0, scale)
	d.Line("СТЕНЫ_НЕСУЩИЕ", dxf.LW070, ox-groundExt, oy, ox+flMM+groundExt, oy)

	// Ground hatch (грунт)
	soilH := px(1.0, scale)
	soilPts := dxf.Rect(ox-groundExt, oy-soilH, ox+flMM+groundExt, oy)
	d.Hatch("ШТРИХОВКА", "ANSI37", 45, 0.5, dxf.Color254, soilPts)

	// ── Building outline (контур здания) ──
	bldPts := dxf.Rect(ox-ew, oy, ox+flMM+ew, oy+totH)
	d.Polyline("СТЕНЫ_НЕСУЩИЕ", dxf.LW070, true, bldPts)

	// ── Floor lines (линии перекрытий, тонкие) ──
	for fl := 1; fl <= b.NFloors; fl++ {
		yfl := oy + float64(fl)*fh
		d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW025, ox-ew, yfl, ox+flMM+ew, yfl)
	}

	// ── Windows on facade ──
	winW := px(1.5, scale) // window width (mm)
	winH := px(1.8, scale) // window height (mm)
	sill := px(0.9, scale) // sill height
	xgap := px(0.5, scale) // gap from axis to window edge

	// Calculate window X positions (at mid-span)
	winXPositions := make([]float64, 0, len(axisCoords)-1)
	for i := 0; i < len(axisCoords)-1; i++ {
		mid := (axisCoords[i] + axisCoords[i+1]) / 2.0
		winXPositions = append(winXPositions, ox+px(mid, scale))
	}

	for fl := 0; fl < b.NFloors; fl++ {
		ySill := oy + float64(fl)*fh + sill
		for wi, wx := range winXPositions {
			// Skip middle windows on ground floor for main entrance
			if fl == 0 && isMainEntrance(wi, len(winXPositions)) {
				continue
			}
			// Window rectangle
			wx0 := wx - winW/2
			wx1 := wx + winW/2
			d.Polyline("ПРОЕМЫ", dxf.LW035, true,
				dxf.Rect(wx0, ySill, wx1, ySill+winH))
			// Horizontal transom
			d.Line("ПРОЕМЫ", dxf.LW013,
				wx0, ySill+winH*0.55, wx1, ySill+winH*0.55)
			// Vertical mullion
			d.Line("ПРОЕМЫ", dxf.LW013,
				(wx0+wx1)/2, ySill, (wx0+wx1)/2, ySill+winH)
		}
		_ = xgap
	}

	// ── Main entrance (1st floor, south or west facade) ──
	if !isEW || cfg.Side == "west" {
		midX := ox + flMM/2.0
		entW := px(4.5, scale) // entrance width
		entH := px(3.0, scale)
		ex0 := midX - entW/2
		ex1 := midX + entW/2
		// Entrance portal
		d.Polyline("ПРОЕМЫ", dxf.LW050, true,
			dxf.Rect(ex0, oy, ex1, oy+entH))
		// Door divisions (double door)
		d.Line("ПРОЕМЫ", dxf.LW025, midX, oy, midX, oy+entH)
		// Canopy (козырёк)
		canopyY := oy + entH
		d.Polyline("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW035, true,
			dxf.Rect(ex0-px(1.0, scale), canopyY,
				ex1+px(1.0, scale), canopyY+px(0.3, scale)))
		// Steps
		for si := 0; si < 3; si++ {
			stepOff := px(float64(si)*0.3, scale)
			stepY := oy - stepOff - px(0.15, scale)
			d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW018,
				ex0-stepOff, stepY, ex1+stepOff, stepY)
		}
	}

	// ── Parapet ──
	parH := px(0.7, scale)
	d.Polyline("СТЕНЫ_НЕСУЩИЕ", dxf.LW035, true,
		dxf.Rect(ox-ew, oy+totH, ox+flMM+ew, oy+totH+parH))

	// ── Axis lines on facade ──
	axBubR := px(0.35, scale)
	axTxtH := math.Max(2.5, px(0.22, scale))
	for i, axX := range axisCoords {
		x := ox + px(axX, scale)
		// Short axis tick at bottom
		d.Line("ОСИ", dxf.LW018, x, oy-px(0.8, scale), x, oy)
		// Bubble below
		cy := oy - px(0.8, scale) - axBubR*1.3
		d.Circle("ТЕКСТ_ОСЕЙ", dxf.LW018, x, cy, axBubR)
		lbl := ""
		if i < len(axisLabels) {
			lbl = axisLabels[i]
		}
		d.Text("ТЕКСТ_ОСЕЙ", "GOST", axTxtH,
			x-axTxtH*0.4, cy-axTxtH*0.45, lbl)
	}

	// ── Horizontal dimensions ──
	DrawChainDimsH(d, axisXPositions(ox, axisCoords, scale),
		oy, 16.0, scale)
	DrawTotalDimH(d,
		ox+px(axisCoords[0], scale), oy,
		ox+px(axisCoords[len(axisCoords)-1], scale),
		28.0, scale)

	// ── Elevation marks (отметки уровней) ──
	markX := ox + flMM + ew + px(3.0, scale)
	elevations := buildElevations(b)
	for _, el := range elevations {
		ya := oy + px(el.valueM, scale)
		AddElevationMark(d, markX, ya, el.valueM)
		if el.label != "" {
			d.Text("ТЕКСТ", "GOST", 2.3,
				markX+px(1.2, scale), ya+0.5, el.label)
		}
		// Horizontal reference line
		d.Line("ОТМЕТКИ", dxf.LW013, ox-ew, ya, markX-px(0.5, scale), ya)
	}

	// Basement mark
	if b.Basement {
		ya := oy - bsmH
		AddElevationMark(d, markX, ya, -b.Dims.BasementHM)
		d.Line("ОТМЕТКИ", dxf.LW013, ox-ew, ya, markX-px(0.5, scale), ya)
	}

	// ── Floor height dimension (left side) ──
	dimLX := ox - ew - px(4.0, scale)
	for fl := 0; fl < b.NFloors; fl++ {
		y1 := oy + float64(fl)*fh
		y2 := oy + float64(fl+1)*fh
		drawLinearDimV(d, dimLX+px(2.0, scale), y1, dimLX+px(2.0, scale), y2,
			dimLX, scale)
	}

	// ── Title ──
	title := cfg.Title
	if title == "" {
		switch cfg.Side {
		case "south":
			last := len(axisLabels) - 1
			if last >= 0 {
				title = fmt.Sprintf("ФАСАД ГЛАВНЫЙ  Оси %s–%s",
					axisLabels[0], axisLabels[last])
			} else {
				title = "ФАСАД ГЛАВНЫЙ"
			}
		case "north":
			title = "ФАСАД ДВОРОВЫЙ"
		case "west":
			title = "ФАСАД ЛЕВЫЙ ТОРЦЕВОЙ"
		case "east":
			title = "ФАСАД ПРАВЫЙ ТОРЦЕВОЙ"
		default:
			title = "ФАСАД"
		}
	}
	gost.DrawSheetTitle(d, title, fmt.Sprintf("1:%d", scale))
	gost.DrawNote(d, "Примечание: отметки в метрах. Размеры в мм.")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type elevMark struct {
	valueM float64
	label  string
}

func buildElevations(b *building.Building) []elevMark {
	marks := make([]elevMark, 0, b.NFloors+2)
	marks = append(marks, elevMark{0.0, "±0,000  Уровень чистого пола 1 эт."})
	for fl := 1; fl <= b.NFloors; fl++ {
		elev := float64(fl) * b.Dims.FloorHM
		if fl < b.NFloors {
			marks = append(marks, elevMark{elev, fmt.Sprintf("+%.3f", elev)})
		}
	}
	// Roof / parapet
	roofElev := float64(b.NFloors) * b.Dims.FloorHM
	marks = append(marks, elevMark{roofElev, fmt.Sprintf("+%.3f  Верх перекрытия", roofElev)})
	marks = append(marks, elevMark{roofElev + 0.7, fmt.Sprintf("+%.3f  Верх парапета", roofElev+0.7)})
	return marks
}

func axisXPositions(ox float64, axisCoords []float64, scale int) []float64 {
	out := make([]float64, len(axisCoords))
	for i, v := range axisCoords {
		out[i] = ox + px(v, scale)
	}
	return out
}

func isMainEntrance(idx, total int) bool {
	mid := total / 2
	return idx == mid-1 || idx == mid
}
