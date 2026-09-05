// Package gost implements GOST R 21.101-2020 drawing frame and title block.
// Стандарт: ГОСТ Р 21.101-2020 «Система проектной документации для строительства»
package gost

import (
	"fmt"

	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/dxf"
)

// Sheet dimensions per ГОСТ Р 21.101-2020 (mm, A1 landscape)
const (
	SheetW = 841.0 // A1 width
	SheetH = 594.0 // A1 height

	// Frame margins (ГОСТ Р 21.101-2020, п. 5.2)
	MarginLeft   = 25.0 // поле подшивки 20 + рамка 5
	MarginRight  = 5.0
	MarginTop    = 5.0
	MarginBottom = 5.0

	// Frame corners
	FrameX0 = MarginLeft
	FrameY0 = MarginBottom
	FrameX1 = SheetW - MarginRight
	FrameY1 = SheetH - MarginTop

	// Title block (основная надпись форма 3, прил. Ж ГОСТ Р 21.101-2020)
	StampW  = 185.0
	StampH  = 55.0
	StampX0 = FrameX1 - StampW
	StampY0 = FrameY0
	StampX1 = FrameX1
	StampY1 = FrameY0 + StampH

	// Content area (рабочее поле — над штампом, внутри рамки)
	ContentX0 = FrameX0
	ContentY0 = StampY1
	ContentX1 = FrameX1
	ContentY1 = FrameY1
	ContentW  = ContentX1 - ContentX0 // 811 mm
	ContentH  = ContentY1 - ContentY0 // 534 mm
)

// TitleBlockData holds data for filling the title block (основная надпись).
type TitleBlockData struct {
	Org      string // Наименование организации
	Project  string // Наименование проекта/объекта строительства
	Address  string // Адрес объекта
	Drawing  string // Наименование чертежа
	Stage    string // Стадия (П / РД)
	Mark     string // Марка комплекта (АР, КР, ...)
	Sheet    string // Номер листа
	Sheets   string // Всего листов
	Scale    string // Масштаб
	Year     string // Год
	Designer string // Разработал
	Checker  string // Проверил
	NormCtrl string // Нормоконтроль
	ChiefEng string // Главный инженер
	Approver string // Утвердил
	DocCode  string // Шифр документа
}

// DrawFrame draws the A1 sheet border and inner frame.
func DrawFrame(d *dxf.Document) {
	// Outer sheet border (thin line 0.18mm)
	d.Polyline("РАМКА", dxf.LW018, true, dxf.Rect(0, 0, SheetW, SheetH))

	// Inner frame (main line 0.7mm)
	d.Polyline("РАМКА", dxf.LW070, true, dxf.Rect(FrameX0, FrameY0, FrameX1, FrameY1))
}

// DrawTitleBlock draws the complete title block (основная надпись форма 3).
//
// Layout (185×55 mm, bottom-right of frame):
//
//	┌──────────────────────────────────────────────────────────────────┐ y=55
//	│  Наименование организации         │  Шифр документа             │
//	├──────────────────────────────────────────────────────────────────┤ y=49
//	│  Наименование объекта строительства                              │
//	├──────────────────────────────────────────────────────────────────┤ y=43
//	│  Наименование чертежа                                            │
//	├────────────┬──────────┬────────┬───────────────────────────────┤ y=35
//	│  Масштаб   │  Стадия  │  Марка │  Лист  │  Листов             │
//	├─────┬──────┬────┬─────┼────────────────────────────────────────┤ y=27
//	│Разр.│  ФИО │Пдп.│Дата │                                        │
//	├─────┼──────┼────┼─────┤              (правый блок)             │
//	│Пров.│  ФИО │Пдп.│Дата │                                        │
//	├─────┼──────┼────┼─────┤                                        │
//	│Н.к. │  ФИО │Пдп.│Дата │                                        │
//	├─────┼──────┼────┼─────┤                                        │
//	│Г.и. │  ФИО │Пдп.│Дата │                                        │
//	├─────┼──────┼────┼─────┤                                        │
//	│Утв. │  ФИО │Пдп.│Дата │                                        │
//	└─────┴──────┴────┴─────┴────────────────────────────────────────┘ y=0
func DrawTitleBlock(d *dxf.Document, tb TitleBlockData) {
	x0, y0 := StampX0, StampY0
	x1, y1 := StampX1, StampY1

	hl := func(y, xa, xb float64, lw int) {
		d.Line("ШТАМП", lw, xa, y0+y, xb, y0+y)
	}
	vl := func(x, ya, yb float64, lw int) {
		d.Line("ШТАМП", lw, x0+x, y0+ya, x0+x, y0+yb)
	}
	txt := func(s string, x, y, h float64) {
		if s == "" {
			return
		}
		d.Text("ШТАМП", "GOST", h, x0+x, y0+y, s)
	}

	// ── Outer border (thick) ──
	d.Polyline("ШТАМП", dxf.LW070, true, dxf.Rect(x0, y0, x1, y1))

	// ── Row heights from bottom (total = 55mm) ──
	// 5 staff rows × 5.5mm + attrs 8mm + drawing name 8mm + object 7mm + org 6.5mm = 55mm
	const (
		rowStaff = 5.5 // personnel row height
		rowAttr  = 8.0 // attributes (scale, stage, mark, sheet)
		rowDraw  = 8.0 // drawing name
		rowObj   = 7.0 // object name
		rowOrg   = 6.5 // org name
	)
	staffTop := rowStaff * 5.0 // top of last staff row = 27.5mm
	attrTop := staffTop + rowAttr
	drawTop := attrTop + rowDraw
	objTop := drawTop + rowObj
	// orgTop  := objTop + rowOrg  // = 55mm = StampH

	// Horizontal lines
	lw := dxf.LW018
	for i := 1; i <= 4; i++ {
		hl(float64(i)*rowStaff, 0, StampW, lw)
	}
	hl(staffTop, 0, StampW, lw)
	hl(attrTop, 0, StampW, lw)
	hl(drawTop, 0, StampW, lw)
	hl(objTop, 0, StampW, lw)

	// ── Staff columns: Role(20) | Name(45) | Sign(15) | Date(15) | Right(90) ──
	const (
		cRole = 20.0
		cName = 65.0 // role + name = 65
		cSign = 80.0 // + sign = 80
		cDate = 95.0 // + date = 95
	)
	for _, cx := range []float64{cRole, cName, cSign, cDate} {
		vl(cx, 0, staffTop, lw)
	}

	// ── Attribute row columns: Scale(25) | Stage(20) | Mark(20) | Sheet(20) | Sheets ──
	const (
		aScale  = 25.0
		aStage  = 45.0
		aMark   = 65.0
		aSheet  = 85.0
		aSheets = 105.0
	)
	for _, ax := range []float64{aScale, aStage, aMark, aSheet, aSheets} {
		vl(ax, staffTop, attrTop, lw)
	}

	// ── TEXT CONTENT ──

	// Column headers (tiny labels at top of each column, row 5)
	hdrH := 1.8
	hdrY := staffTop - rowStaff*0.3
	for _, col := range [][2]interface{}{
		{1.0, "Должность"},
		{cRole + 1.0, "Фамилия"},
		{cName + 1.0, "Подпись"},
		{cSign + 1.0, "Дата"},
	} {
		txt(col[1].(string), col[0].(float64), hdrY, hdrH)
	}

	// Staff rows
	staff := []struct{ role, name string }{
		{"Разработал", tb.Designer},
		{"Проверил", tb.Checker},
		{"Н.контроль", tb.NormCtrl},
		{"Гл.инженер", tb.ChiefEng},
		{"Утвердил", tb.Approver},
	}
	for i, s := range staff {
		ry := float64(i)*rowStaff + rowStaff*0.3
		txt(s.role, 1.0, ry, 2.3)
		txt(s.name, cRole+1.0, ry, 2.3)
	}

	// Attribute row
	attrs := []struct {
		label, val string
		x          float64
	}{
		{"Масштаб", tb.Scale, 0},
		{"Стадия", tb.Stage, aScale},
		{"Марка", tb.Mark, aStage},
		{"Лист", tb.Sheet, aMark},
		{"Листов", tb.Sheets, aSheet},
	}
	for _, a := range attrs {
		txt(a.label, a.x+1.0, staffTop+rowAttr*0.6, 1.8)
		txt(a.val, a.x+2.0, staffTop+rowAttr*0.1, 3.2)
	}

	// Drawing name (bold, large)
	txt(tb.Drawing, 2.0, attrTop+rowDraw*0.3, 3.5)

	// Object name
	obj := tb.Address
	if len(obj) > 90 {
		obj = obj[:90] + "..."
	}
	txt(obj, 2.0, drawTop+rowObj*0.25, 2.3)

	// Project name (2nd line in object row)
	proj := tb.Project
	if len(proj) > 70 {
		proj = proj[:70] + "..."
	}
	txt(proj, 2.0, drawTop+rowObj*0.65, 2.5)

	// Org name (top row)
	txt(tb.Org, 2.0, objTop+rowOrg*0.3, 2.8)

	// Document code (top-right)
	code := tb.DocCode
	if code == "" {
		code = fmt.Sprintf("68-%s-%s-%s", tb.Year, tb.Mark, tb.Sheet)
	}
	txt(code, 2.0, objTop+rowOrg*0.65, 2.0)

	// Right-side identification block
	// (правый блок: инвентарные номера — упрощённо)
	txt("Инв. № подл.", cDate+2.0, staffTop-rowStaff*1.5, 2.0)
	txt("Взам. инв. №", cDate+2.0, staffTop-rowStaff*3.0, 2.0)
	txt(tb.Year+" г.", cDate+2.0, staffTop-rowStaff*4.0, 2.3)
}

// DrawSheetTitle draws the drawing title above the content (per ГОСТ 21.501-2018 п.4.10).
// The title is underlined.
func DrawSheetTitle(d *dxf.Document, text, scaleStr string) {
	x := ContentX0 + 5.0
	y := ContentY1 - 15.0
	d.Text("ТЕКСТ", "GOST", 5.0, x, y, text)
	// Underline
	tw := float64(len([]rune(text))) * 5.0 * 0.65
	d.Line("ТЕКСТ", dxf.LW025, x, y-0.5, x+tw, y-0.5)
	// Scale label
	if scaleStr != "" {
		d.Text("ТЕКСТ", "GOST", 3.0, x, y-8.0,
			fmt.Sprintf("Масштаб %s", scaleStr))
	}
}

// DrawNote draws a note text at the bottom-left of the content area.
func DrawNote(d *dxf.Document, note string) {
	d.Text("ТЕКСТ", "GOST", 2.0, ContentX0+5.0, ContentY0+2.0, note)
}
