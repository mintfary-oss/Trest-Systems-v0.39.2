package drawings

import (
	"fmt"
	"math"

	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/building"
	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/dxf"
	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/gost"
)

// SitePlanConfig controls site plan generation.
type SitePlanConfig struct {
	Scale int    // e.g. 500 for 1:500
	Title string // override title
}

// DrawSitePlan generates a site plan (СПОЗУ) per ГОСТ 21.204-93.
func DrawSitePlan(d *dxf.Document, b *building.Building, cfg SitePlanConfig) {
	scale := cfg.Scale
	if scale == 0 {
		scale = 500
	}

	ox := gost.ContentX0 + 50.0
	oy := gost.ContentY0 + 35.0

	plotW := px(b.Plot.WidthM, scale)
	plotH := px(b.Plot.DepthM, scale)

	// ── Plot boundary (граница ЗУ) ──
	d.Polyline("ГЕНПЛАН", dxf.LW050, true, dxf.Rect(ox, oy, ox+plotW, oy+plotH))

	// Plot area label
	d.Text("ТЕКСТ", "GOST", 2.8, ox+2.0, oy+plotH+4.0,
		fmt.Sprintf("Площадь земельного участка: %.4f га (%.1f м²)",
			b.Plot.AreaHa, b.Plot.AreaHa*10000))

	// ── Building footprint ──
	bldOffX := px(b.Plot.SetbackWM, scale)
	bldOffY := px(b.Plot.SetbackSM, scale)
	bldW := px(b.Axes.TotalWidth()+2*b.Dims.WallExtM, scale)
	bldH := px(b.Axes.TotalDepth()+2*b.Dims.WallExtM, scale)

	bldX0 := ox + bldOffX
	bldY0 := oy + bldOffY
	bldX1 := bldX0 + bldW
	bldY1 := bldY0 + bldH

	bldPts := dxf.Rect(bldX0, bldY0, bldX1, bldY1)
	d.Polyline("ГЕНПЛАН", dxf.LW070, true, bldPts)
	d.SolidHatch("ШТРИХОВКА", dxf.Color254, bldPts)

	// Building label
	bldCX := (bldX0 + bldX1) / 2.0
	bldCY := (bldY0 + bldY1) / 2.0
	d.Text("ТЕКСТ", "GOST", 3.0, bldCX-px(6.0, scale), bldCY+2.0, "Здание суда")
	d.Text("ТЕКСТ", "GOST", 2.5, bldCX-px(6.0, scale), bldCY-3.0, "(проектируемое)")

	// Footprint area
	d.Text("ТЕКСТ", "GOST", 2.3, bldCX-px(6.0, scale), bldCY-8.0,
		fmt.Sprintf("Пл. застройки: %.1f м²",
			(b.Axes.TotalWidth()+2*b.Dims.WallExtM)*(b.Axes.TotalDepth()+2*b.Dims.WallExtM)))

	// ── Driveway and entrance ──
	roadW := px(6.0, scale)
	entryX := bldX0 + bldW/2.0 - roadW/2.0

	// Main driveway (от ворот до входа)
	roadY0 := oy
	roadY1 := bldY0
	d.Polyline("ДОРОГИ", dxf.LW035, true,
		dxf.Rect(entryX, roadY0, entryX+roadW, roadY1))

	// Road hatch (лёгкая штриховка)
	d.Hatch("ШТРИХОВКА", "ANSI37", 0, 1.5, dxf.ColorGray,
		dxf.Rect(entryX, roadY0, entryX+roadW, roadY1))
	d.Text("ДОРОГИ", "GOST", 2.0, entryX+roadW+1.0, (roadY0+roadY1)/2.0, "Въезд/\nвыезд")

	// ── Parking area ──
	parkY0 := oy + px(3.0, scale)
	parkY1 := oy + px(11.0, scale)
	parkX0 := bldX0
	parkX1 := bldX0 + px(30.0, scale)
	d.Polyline("ДОРОГИ", dxf.LW035, true, dxf.Rect(parkX0, parkY0, parkX1, parkY1))

	// Parking spaces (разметка)
	spaceW := px(2.5, scale)
	spaceH := parkY1 - parkY0 - px(1.0, scale)
	nSpaces := int((parkX1 - parkX0) / spaceW)
	for i := 0; i < nSpaces; i++ {
		sx0 := parkX0 + float64(i)*spaceW
		sx1 := sx0 + spaceW
		d.Polyline("ДОРОГИ", dxf.LW013, true,
			dxf.Rect(sx0, parkY0+px(0.5, scale), sx1, parkY0+px(0.5, scale)+spaceH))
	}
	d.Text("ТЕКСТ", "GOST", 2.5,
		parkX0+2.0, (parkY0+parkY1)/2.0-1.5,
		fmt.Sprintf("Парковка (%d м/м)", nSpaces))

	// ── Pedestrian paths ──
	pathW := px(2.0, scale)
	// South entry path
	d.Polyline("ДОРОГИ", dxf.LW018, true,
		dxf.Rect(bldX0+bldW*0.35, oy+px(3.0, scale),
			bldX0+bldW*0.35+pathW, bldY0))
	d.Polyline("ДОРОГИ", dxf.LW018, true,
		dxf.Rect(bldX0+bldW*0.6, oy+px(3.0, scale),
			bldX0+bldW*0.6+pathW, bldY0))

	// Perimeter path
	perpW := px(1.5, scale)
	d.Polyline("ДОРОГИ", dxf.LW018, false, [][2]float64{
		{bldX0 - perpW, bldY0 - perpW},
		{bldX1 + perpW, bldY0 - perpW},
		{bldX1 + perpW, bldY1 + perpW},
		{bldX0 - perpW, bldY1 + perpW},
		{bldX0 - perpW, bldY0 - perpW},
	})

	// ── Trees and landscaping ──
	treeR := px(2.0, scale)
	treePositions := buildTreePositions(b, ox, oy, plotW, plotH, scale)
	for _, tp := range treePositions {
		drawTree(d, tp[0], tp[1], treeR)
	}

	// Green area labels
	d.Text("РАСТИТЕЛЬНОСТЬ", "GOST", 2.3,
		ox+2.0, bldY1+px(5.0, scale), "Газон с посевом трав")

	// ── Refuse area ──
	refX := ox + plotW - px(8.0, scale)
	refY := oy + plotH - px(8.0, scale)
	d.Polyline("ГЕНПЛАН", dxf.LW018, true,
		dxf.Rect(refX, refY, refX+px(4.0, scale), refY+px(4.0, scale)))
	d.Text("ТЕКСТ", "GOST", 2.0, refX, refY-3.5, "Площадка ТКО")

	// ── Gate / entrance to plot ──
	gateX := entryX
	gateW := roadW
	// Gate pillars
	pillarSz := px(0.4, scale)
	d.Polyline("ГЕНПЛАН", dxf.LW035, true,
		dxf.Rect(gateX-pillarSz, oy-pillarSz, gateX, oy))
	d.Polyline("ГЕНПЛАН", dxf.LW035, true,
		dxf.Rect(gateX+gateW, oy-pillarSz, gateX+gateW+pillarSz, oy))
	// Gate panels
	d.Line("ГЕНПЛАН", dxf.LW025, gateX, oy, gateX+gateW*0.45, oy-gateW*0.2)
	d.Line("ГЕНПЛАН", dxf.LW025, gateX+gateW, oy, gateX+gateW*0.55, oy-gateW*0.2)

	// ── Situation plan (small inset) ──
	drawSituationPlan(d, b, ox, oy, plotW, plotH)

	// ── North arrow ──
	drawNorthArrow(d, gost.ContentX1-30.0, gost.ContentY0+30.0)

	// ── Legend ──
	drawSiteLegend(d, gost.ContentX0+5.0, gost.ContentY0+5.0)

	// ── Dimensions ──
	DrawTotalDimH(d, ox, oy, ox+plotW, 14.0, scale)
	DrawTotalDimV(d, oy, ox, oy+plotH, 12.0, scale)

	// Building setback dimensions
	drawLinearDimV(d, ox+bldOffX, oy, ox+bldOffX, bldY0, ox+bldOffX-px(2.0, scale), scale)

	// ── Title ──
	title := cfg.Title
	if title == "" {
		title = "СХЕМА ПЛАНИРОВОЧНОЙ ОРГАНИЗАЦИИ ЗЕМЕЛЬНОГО УЧАСТКА"
	}
	gost.DrawSheetTitle(d, title, fmt.Sprintf("1:%d", scale))
	gost.DrawNote(d, "Координатная система местная. Абсолютные отметки приведены в Балтийской системе высот 1977 г.")
}

func drawTree(d *dxf.Document, cx, cy, r float64) {
	d.Circle("РАСТИТЕЛЬНОСТЬ", dxf.LW013, cx, cy, r)
	// Cross inside (per ГОСТ 21.204-93)
	d.Line("РАСТИТЕЛЬНОСТЬ", dxf.LW013, cx-r*0.6, cy, cx+r*0.6, cy)
	d.Line("РАСТИТЕЛЬНОСТЬ", dxf.LW013, cx, cy-r*0.6, cx, cy+r*0.6)
}

func buildTreePositions(b *building.Building, ox, oy, plotW, plotH float64, scale int) [][2]float64 {
	var pts [][2]float64
	spacing := px(5.0, scale)
	margin := px(2.5, scale)

	// West side trees
	for y := oy + margin; y < oy+plotH-margin; y += spacing {
		pts = append(pts, [2]float64{ox + margin, y})
	}
	// East side trees
	for y := oy + margin; y < oy+plotH-margin; y += spacing {
		pts = append(pts, [2]float64{ox + plotW - margin, y})
	}
	// North side trees
	for x := ox + margin; x < ox+plotW-margin; x += spacing {
		pts = append(pts, [2]float64{x, oy + plotH - margin})
	}
	return pts
}

func drawNorthArrow(d *dxf.Document, x, y float64) {
	shaftH := 12.0
	headH := 4.0
	headW := 3.0

	// Shaft
	d.Line("ГЕНПЛАН", dxf.LW025, x, y-shaftH/2, x, y+shaftH/2)

	// Arrowhead (filled triangle)
	tipY := y + shaftH/2
	pts := [][2]float64{
		{x, tipY + headH},
		{x - headW, tipY},
		{x + headW, tipY},
	}
	d.Polyline("ГЕНПЛАН", dxf.LW025, true, pts)
	d.SolidHatch("ГЕНПЛАН", dxf.ColorBlack, pts)

	// "С" label
	d.Text("ТЕКСТ", "GOST", 5.0, x-3.0, tipY+headH+2.0, "С")

	// Circle around arrow
	d.Circle("ГЕНПЛАН", dxf.LW018, x, y, shaftH/2+headH*0.5)
}

func drawSituationPlan(d *dxf.Document, b *building.Building, ox, oy, plotW, plotH float64) {
	// Small inset situation plan in the lower-left of content area
	sx := gost.ContentX0 + 5.0
	sy := gost.ContentY0 + 55.0
	sw := 55.0
	sh := 45.0

	// Border
	d.Polyline("ГЕНПЛАН", dxf.LW018, true, dxf.Rect(sx, sy, sx+sw, sy+sh))
	d.Text("ТЕКСТ", "GOST", 2.5, sx+2.0, sy+sh-6.0, "СИТУАЦИОННЫЙ ПЛАН")
	d.Text("ТЕКСТ", "GOST", 2.0, sx+2.0, sy+sh-10.5, "М 1:5000")

	// Schematic streets
	streetW := 4.0
	// Horizontal street (ул. Радищева)
	d.Polyline("ДОРОГИ", dxf.LW018, false, [][2]float64{
		{sx, sy + sh*0.3},
		{sx + sw, sy + sh*0.3},
	})
	d.Polyline("ДОРОГИ", dxf.LW018, false, [][2]float64{
		{sx, sy + sh*0.3 + streetW},
		{sx + sw, sy + sh*0.3 + streetW},
	})
	d.Text("ТЕКСТ", "GOST", 1.8, sx+2.0, sy+sh*0.3-3.5, "ул. Радищева")

	// Site block
	bsx := sx + sw*0.35
	bsy := sy + sh*0.3 + streetW + 3.0
	bsw := sw * 0.3
	bsh := sh * 0.35
	d.Polyline("ГЕНПЛАН", dxf.LW025, true, dxf.Rect(bsx, bsy, bsx+bsw, bsy+bsh))
	d.SolidHatch("ШТРИХОВКА", dxf.Color254, dxf.Rect(bsx, bsy, bsx+bsw, bsy+bsh))
	d.Text("ТЕКСТ", "GOST", 1.8, bsx+1.0, bsy+bsh/2.0-1.0, "Участок")

	// Legend entry
	d.Polyline("ГЕНПЛАН", dxf.LW025, true, dxf.Rect(sx+2.0, sy+4.0, sx+7.0, sy+8.0))
	d.SolidHatch("ШТРИХОВКА", dxf.Color254, dxf.Rect(sx+2.0, sy+4.0, sx+7.0, sy+8.0))
	d.Text("ТЕКСТ", "GOST", 2.0, sx+9.0, sy+5.0, "— Проектируемое здание")
}

func drawSiteLegend(d *dxf.Document, x, y float64) {
	items := []struct{ sym, desc string }{
		{"──────", "Ограждение территории"},
		{"- - - -", "Граница земельного участка"},
		{"●──●", "Дорожки, тротуары (асфальт)"},
		{"○+○", "Дерево лиственное"},
		{"□□□", "Парковочные места"},
	}
	d.Text("ТЕКСТ", "GOST", 2.8, x, y+float64(len(items))*5.5+4.0, "УСЛОВНЫЕ ОБОЗНАЧЕНИЯ:")
	for i, item := range items {
		iy := y + float64(len(items)-i-1)*5.0
		d.Text("ТЕКСТ", "GOST", 2.0, x, iy+1.5, item.sym)
		d.Text("ТЕКСТ", "GOST", 2.0, x+22.0, iy+1.5, item.desc)
	}

	_ = math.Pi // keep import
}
