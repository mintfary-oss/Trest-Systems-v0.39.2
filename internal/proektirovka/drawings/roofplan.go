package drawings

import (
	"fmt"

	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/building"
	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/dxf"
	"github.com/mintfary-oss/trest-sistems/internal/proektirovka/gost"
)

// RoofPlanConfig controls roof plan generation.
type RoofPlanConfig struct {
	Scale int
	Title string
}

// DrawRoofPlan generates the roof plan (план кровли) per ГОСТ 21.501-2018.
func DrawRoofPlan(d *dxf.Document, b *building.Building, cfg RoofPlanConfig) {
	scale := cfg.Scale
	if scale == 0 {
		scale = 200
	}

	ox := gost.ContentX0 + 70.0
	oy := gost.ContentY0 + 45.0

	ac := NewAxisConfig(b, ox, oy, scale)
	ew := px(b.Dims.WallExtM, scale)
	bw := ac.BW
	bh := ac.BH

	// ── Simplified axes (lines only, no bubbles for 1:200) ──
	xc := b.Axes.XCoords()
	yc := b.Axes.YCoords()
	for i, v := range xc {
		x := ox + px(v, scale)
		d.Line("ОСИ", dxf.LW013, x, oy-ew-px(1.0, scale), x, oy+bh+ew+px(1.0, scale))
		if i < len(b.Axes.XLabels) {
			d.Text("ТЕКСТ_ОСЕЙ", "GOST", 2.5, x-1.5, oy-ew-px(1.5, scale)-4.5, b.Axes.XLabels[i])
		}
	}
	for j, v := range yc {
		y := oy + px(v, scale)
		d.Line("ОСИ", dxf.LW013, ox-ew-px(1.0, scale), y, ox+bw+ew+px(1.0, scale), y)
		if j < len(b.Axes.YLabels) {
			d.Text("ТЕКСТ_ОСЕЙ", "GOST", 2.5, ox-ew-px(1.5, scale)-5.0, y-1.2, b.Axes.YLabels[j])
		}
	}

	// ── Parapet outer boundary ──
	d.Polyline("СТЕНЫ_НЕСУЩИЕ", dxf.LW070, true,
		dxf.Rect(ox-ew, oy-ew, ox+bw+ew, oy+bh+ew))

	// Parapet inner edge
	d.Polyline("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW035, true,
		dxf.Rect(ox, oy, ox+bw, oy+bh))

	// Parapet hatch (wall cross-section)
	parPts := dxf.Rect(ox-ew, oy-ew, ox+bw+ew, oy+bh+ew)
	d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254, parPts)

	// ── Roof drainage: slope lines and arrows ──
	// Flat roof with 1.5% slope to drain funnels
	// Divide roof into 4 drainage zones
	midX := ox + bw/2.0
	midY := oy + bh/2.0
	slopeAngle := 45.0

	// Ridge lines (конек / водораздел)
	d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW025, midX, oy, midX, oy+bh)
	d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW025, ox, midY, ox+bw, midY)

	// Slope arrows
	type arrow struct{ x0, y0, x1, y1 float64 }
	arrows := []arrow{
		{midX*0.5 + ox*0.5, midY*0.5 + oy*0.5, ox + px(3.0, scale), oy + px(3.0, scale)},
		{midX*0.5 + ox*0.5 + bw*0.5, midY*0.5 + oy*0.5, ox + bw - px(3.0, scale), oy + px(3.0, scale)},
		{midX*0.5 + ox*0.5, midY*0.5 + oy*0.5 + bh*0.5, ox + px(3.0, scale), oy + bh - px(3.0, scale)},
		{midX*0.5 + ox*0.5 + bw*0.5, midY*0.5 + oy*0.5 + bh*0.5, ox + bw - px(3.0, scale), oy + bh - px(3.0, scale)},
	}
	for _, ar := range arrows {
		d.Line("РАЗМЕРЫ", dxf.LW018, ar.x0, ar.y0, ar.x1, ar.y1)
		// Arrowhead
		dx := ar.x1 - ar.x0
		dy := ar.y1 - ar.y0
		ln := sqrt2(dx*dx + dy*dy)
		if ln > 0 {
			ux, uy := dx/ln, dy/ln
			perpX, perpY := -uy, ux
			as := 1.5
			d.Line("РАЗМЕРЫ", dxf.LW018, ar.x1, ar.y1,
				ar.x1-ux*as*2-perpX*as, ar.y1-uy*as*2-perpY*as)
			d.Line("РАЗМЕРЫ", dxf.LW018, ar.x1, ar.y1,
				ar.x1-ux*as*2+perpX*as, ar.y1-uy*as*2+perpY*as)
		}
		d.Text("РАЗМЕРЫ", "GOST", 2.5,
			(ar.x0+ar.x1)/2.0+1.5, (ar.y0+ar.y1)/2.0+1.0, "i=1,5%")
	}

	_ = slopeAngle

	// ── Roof drain funnels (водосточные воронки) ──
	funnelR := px(0.5, scale)
	funnelPositions := [][2]float64{
		{ox + bw*0.15, oy + bh*0.15},
		{ox + bw*0.85, oy + bh*0.15},
		{ox + bw*0.15, oy + bh*0.85},
		{ox + bw*0.85, oy + bh*0.85},
		{ox + bw*0.5, oy + bh*0.15},
		{ox + bw*0.5, oy + bh*0.85},
	}
	for _, fp := range funnelPositions {
		// Funnel circle with cross (per ГОСТ 21.204-93)
		d.Circle("САНУЗЛЫ", dxf.LW035, fp[0], fp[1], funnelR)
		d.Circle("САНУЗЛЫ", dxf.LW018, fp[0], fp[1], funnelR*0.4)
		d.Line("САНУЗЛЫ", dxf.LW025, fp[0]-funnelR*1.4, fp[1], fp[0]+funnelR*1.4, fp[1])
		d.Line("САНУЗЛЫ", dxf.LW025, fp[0], fp[1]-funnelR*1.4, fp[0], fp[1]+funnelR*1.4)
	}

	// ── Ventilation shafts ──
	vshW := px(1.2, scale)
	vshH := px(1.5, scale)
	vshPositions := [][2]float64{
		{ox + bw*0.2, oy + bh*0.5},
		{ox + bw*0.4, oy + bh*0.5},
		{ox + bw*0.6, oy + bh*0.5},
		{ox + bw*0.8, oy + bh*0.5},
	}
	for _, vp := range vshPositions {
		vPts := dxf.Rect(vp[0]-vshW/2, vp[1]-vshH/2, vp[0]+vshW/2, vp[1]+vshH/2)
		d.Polyline("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW035, true, vPts)
		d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254, vPts)
		d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW013,
			vp[0]-vshW/2, vp[1]-vshH/2, vp[0]+vshW/2, vp[1]+vshH/2)
	}

	// ── Elevator machine room ──
	emrW := px(3.0, scale)
	emrH := px(3.5, scale)
	emrX := ox + bw*0.5 - emrW/2
	emrY := oy + bh*0.4
	emrPts := dxf.Rect(emrX, emrY, emrX+emrW, emrY+emrH)
	d.Polyline("СТЕНЫ_НЕСУЩИЕ", dxf.LW070, true, emrPts)
	d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254, emrPts)
	d.Text("ТЕКСТ", "GOST", 2.0, emrX, emrY+emrH+1.5, "МО лифта")

	// ── Roof slope annotation ──
	d.Text("ТЕКСТ", "GOST", 2.8, ox+px(2.0, scale), oy+bh/2.0-3.0,
		"Кровля плоская эксплуатируемая")
	d.Text("ТЕКСТ", "GOST", 2.3, ox+px(2.0, scale), oy+bh/2.0-8.0,
		"Покрытие — ТПО-мембрана Logicroof V-RP 1,5мм")
	d.Text("ТЕКСТ", "GOST", 2.3, ox+px(2.0, scale), oy+bh/2.0-13.0,
		"Утеплитель — PIR плиты 200мм")

	// ── Legend ──
	lgX := gost.ContentX0 + 5.0
	lgY := gost.ContentY0 + 5.0
	d.Text("ТЕКСТ", "GOST", 2.8, lgX, lgY+30.0, "УСЛОВНЫЕ ОБОЗНАЧЕНИЯ:")
	d.Circle("САНУЗЛЫ", dxf.LW025, lgX+3.5, lgY+23.0, 3.5)
	d.Text("ТЕКСТ", "GOST", 2.3, lgX+10.0, lgY+22.0, "— Водосточная воронка")
	d.Polyline("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW025, true,
		dxf.Rect(lgX+1.0, lgY+12.0, lgX+8.0, lgY+16.5))
	d.Hatch("ШТРИХОВКА", "ANSI31", 45, 0.3, dxf.Color254,
		dxf.Rect(lgX+1.0, lgY+12.0, lgX+8.0, lgY+16.5))
	d.Text("ТЕКСТ", "GOST", 2.3, lgX+10.0, lgY+13.0, "— Вентиляционная шахта")
	d.Line("СТЕНЫ_НЕНЕСУЩИЕ", dxf.LW025, lgX+1.0, lgY+6.0, lgX+9.0, lgY+6.0)
	d.Text("ТЕКСТ", "GOST", 2.3, lgX+10.0, lgY+5.0, "— Водораздел")

	// ── Dimensions ──
	DrawTotalDimH(d, ox, oy, ox+bw, 14.0, scale)
	DrawTotalDimV(d, oy, ox, oy+bh, 12.0, scale)

	// ── Title ──
	title := cfg.Title
	if title == "" {
		roofElev := float64(b.NFloors) * b.Dims.FloorHM
		title = fmt.Sprintf("ПЛАН КРОВЛИ  Отм. +%.3f", roofElev)
	}
	gost.DrawSheetTitle(d, title, fmt.Sprintf("1:%d", scale))
	gost.DrawNote(d, "Примечание: размеры в мм, отметки в м. Уклон кровли i=1,5%.")
}

func sqrt2(v float64) float64 {
	if v < 0 {
		return 0
	}
	x := v
	for i := 0; i < 20; i++ {
		x = (x + v/x) / 2
	}
	return x
}
