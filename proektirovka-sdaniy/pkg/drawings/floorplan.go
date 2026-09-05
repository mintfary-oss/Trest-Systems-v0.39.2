package drawings

import (
	"fmt"

	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/building"
	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/dxf"
	"github.com/mintfary-oss/Proektirovka-sdaniy/pkg/gost"
)

// FloorPlanConfig controls how a floor plan is drawn.
type FloorPlanConfig struct {
	Scale       int    // e.g. 100 for 1:100
	FloorNumber int    // which floor to draw
	Title       string // override title (auto-generated if empty)
}

// DrawFloorPlan generates a complete floor plan drawing.
// The document must have its preamble already written (d.WritePreamble called).
func DrawFloorPlan(d *dxf.Document, b *building.Building, cfg FloorPlanConfig) {
	scale := cfg.Scale
	if scale == 0 {
		scale = 100
	}

	// ── Paper layout ──
	// Building origin on paper: leaves space for axes, labels, dimensions
	ox := gost.ContentX0 + 85.0
	oy := gost.ContentY0 + 60.0

	ac := NewAxisConfig(b, ox, oy, scale)
	ew := px(b.Dims.WallExtM, scale)
	iw := px(b.Dims.WallIntM, scale)
	pw := px(b.Dims.WallPrtM, scale)

	// ── Axes ──
	DrawAxes(d, ac)

	// ── Outer walls ──
	DrawOuterWalls(d, ac, b.Dims.WallExtM)

	// ── Get floor data ──
	floor := b.FloorByNumber(cfg.FloorNumber)

	// ── Internal walls from axis grid ──
	// Draw internal bearing walls at axis intersections based on building type
	xc := b.Axes.XCoords()
	yc := b.Axes.YCoords()

	// Transverse bearing walls (parallel to X-axis) — every other Y axis (except first and last)
	for j := 1; j < len(yc)-1; j++ {
		y := ac.OY + px(yc[j], scale)
		DrawWallH(d, "СТЕНЫ_НЕСУЩИЕ", ac.OX, y, ac.OX+ac.BW, iw)
	}

	// Longitudinal bearing walls (parallel to Y-axis) — middle axes
	midStart := 1
	midEnd := len(xc) - 2
	step := (len(xc) - 1) / 3
	if step < 1 {
		step = 1
	}
	for i := midStart; i <= midEnd; i += step {
		if i == 0 || i == len(xc)-1 {
			continue
		}
		x := ac.OX + px(xc[i], scale)
		DrawWallV(d, "СТЕНЫ_НЕСУЩИЕ", x, ac.OY, ac.OY+ac.BH, iw)
	}

	// ── Rooms from floor data ──
	if floor != nil {
		for _, room := range floor.Rooms {
			rx0 := ac.OX + px(room.X, scale) + ew
			ry0 := ac.OY + px(room.Y, scale) + ew
			rx1 := rx0 + px(room.Width, scale) - ew
			ry1 := ry0 + px(room.Height, scale) - ew

			switch room.Type {
			case building.RoomStaircase:
				DrawStaircase(d, rx0, ry0, rx1, ry1, 12)
			case building.RoomElevator:
				DrawElevator(d, rx0, ry0, rx1, ry1)
			case building.RoomToilet:
				drawToilet(d, rx0, ry0, rx1, ry1)
			default:
				// Draw room partition boundary if not aligned with bearing walls
				if room.Type != building.RoomCorridor {
					drawRoomPartitions(d, room, ac, scale, pw)
				}
			}

			// Room label
			cx := (rx0 + rx1) / 2.0
			cy := (ry0 + ry1) / 2.0
			DrawRoomLabel(d, cx, cy, room.ID, room.Name, room.Area())
		}
	} else {
		// No floor data: draw generic layout placeholder
		drawGenericFloorLayout(d, b, ac, scale, ew, iw, pw, cfg.FloorNumber)
	}

	// ── Windows ──
	drawWindowsOnAxes(d, b, ac, scale)

	// ── Section marks ──
	secX := ac.XPos[0] + px(3.0, scale)
	AddSectionMark(d, secX, ac.OY-ew-10.0, "1")
	AddSectionMark(d, secX, ac.OY+ac.BH+ew+10.0, "1")
	d.Line("РАЗМЕРЫ", dxf.LW013, secX, ac.OY-ew, secX, ac.OY+ac.BH+ew)

	// ── Chain dimensions ──
	// Horizontal chain (between X axes)
	DrawChainDimsH(d, ac.XPos, ac.OY, 22.0, scale)
	// Total horizontal
	DrawTotalDimH(d, ac.XPos[0], ac.OY, ac.XPos[len(ac.XPos)-1], 36.0, scale)

	// Vertical chain (between Y axes)
	DrawChainDimsV(d, ac.YPos, ac.OX, 14.0, scale)
	// Total vertical
	DrawTotalDimV(d, ac.YPos[0], ac.OX, ac.YPos[len(ac.YPos)-1], 28.0, scale)

	// ── Note ──
	gost.DrawNote(d, "Примечание: все размеры в мм, отметки в м. Не масштабировать.")

	// ── Title ──
	title := cfg.Title
	if title == "" {
		floor2 := b.FloorByNumber(cfg.FloorNumber)
		if floor2 != nil {
			title = fmt.Sprintf("ПЛАН %s  (Отм. %+.3f)", floor2.Name, floor2.Elevation)
		} else {
			title = fmt.Sprintf("ПЛАН %d-ГО ЭТАЖА", cfg.FloorNumber)
		}
	}
	gost.DrawSheetTitle(d, title, fmt.Sprintf("1:%d", scale))
}

// drawGenericFloorLayout draws a generic floor layout when no room data is available.
// This creates a realistic-looking plan based on the building dimensions and type.
func drawGenericFloorLayout(
	d *dxf.Document,
	b *building.Building,
	ac AxisConfig,
	scale int,
	ew, iw, pw float64,
	floorNum int,
) {
	xc := b.Axes.XCoords()
	yc := b.Axes.YCoords()

	if len(xc) < 2 || len(yc) < 2 {
		return
	}

	// Partition walls creating rooms
	// Create logical zones based on floor number
	nX := len(xc) - 1 // number of X spans
	nY := len(yc) - 1 // number of Y spans

	// Vertical partitions at intermediate X positions (every other span)
	for i := 1; i < nX; i += 2 {
		x := ac.OX + px(xc[i], scale)
		if i%3 == 0 {
			continue // leave open for corridors
		}
		// Partial partition (south half)
		midY := ac.OY + ac.BH/2.0
		d.Line("ПЕРЕГОРОДКИ", dxf.LW025, x, ac.OY+ew, x, midY)
	}

	// Horizontal partition creating corridor
	if nY >= 3 {
		corrY := ac.OY + px(yc[nY/2], scale)
		DrawWallH(d, "ПЕРЕГОРОДКИ", ac.OX+ew, corrY, ac.OX+ac.BW-ew, pw)
	}

	// Staircases (at 1/4 and 3/4 of X width, north section)
	if nX >= 4 && nY >= 2 {
		// Staircase 1
		lkIdx := nX / 4
		lk1x0 := ac.OX + px(xc[lkIdx], scale) + iw/2
		lk1x1 := ac.OX + px(xc[lkIdx+1], scale) - iw/2
		lk1y0 := ac.OY + px(yc[nY/2], scale) + iw/2
		lk1y1 := ac.OY + ac.BH - ew
		DrawStaircase(d, lk1x0, lk1y0, lk1x1, lk1y1, 12)

		// Staircase 2
		lkIdx2 := 3 * nX / 4
		lk2x0 := ac.OX + px(xc[lkIdx2], scale) + iw/2
		lk2x1 := ac.OX + px(xc[lkIdx2+1], scale) - iw/2
		DrawStaircase(d, lk2x0, lk1y0, lk2x1, lk1y1, 12)

		// Elevator (between staircases)
		elIdx := nX / 2
		elx0 := ac.OX + px(xc[elIdx], scale) + iw/2
		elx1 := elx0 + px(3.0, scale)
		DrawElevator(d, elx0, lk1y0, elx1, lk1y0+px(4.0, scale))
	}

	// Room labels for generic layout
	roomLabels := genericRoomLabels(floorNum, nX, nY)
	for _, rl := range roomLabels {
		cx := ac.OX + px(xc[0], scale) + px(rl.xSpan*b.Axes.TotalWidth()/float64(nX), scale)
		cy := ac.OY + px(yc[0], scale) + px(rl.ySpan*b.Axes.TotalDepth()/float64(nY), scale)
		DrawRoomLabel(d, cx, cy, rl.id, rl.name, rl.area)
	}
}

type roomLabelData struct {
	id, name string
	area     float64
	xSpan    float64 // relative X position (0..1)
	ySpan    float64 // relative Y position (0..1)
}

func genericRoomLabels(floorNum, nX, nY int) []roomLabelData {
	switch floorNum {
	case 0: // basement
		return []roomLabelData{
			{"Б-1", "Эл.щитовая", 36, 0.1, 0.3},
			{"Б-2", "Тепловой узел", 54, 0.35, 0.3},
			{"Б-3", "Насосная", 18, 0.55, 0.3},
			{"Б-4", "Вент. камера", 72, 0.75, 0.3},
			{"Б-5", "Хранилище", 108, 0.2, 0.75},
			{"Б-6", "Технич. помещ.", 36, 0.65, 0.75},
		}
	case 1: // ground floor
		return []roomLabelData{
			{"101", "Вестибюль", 72, 0.12, 0.25},
			{"102", "Пост охраны", 18, 0.25, 0.25},
			{"103", "Зал ожидания", 54, 0.45, 0.25},
			{"104", "Зал с/з №1", 108, 0.12, 0.65},
			{"105", "Зал с/з №2", 108, 0.50, 0.65},
			{"106", "Конвойный блок", 54, 0.82, 0.5},
			{"107", "Канцелярия", 36, 0.72, 0.25},
		}
	case 6: // top technical floor
		return []roomLabelData{
			{"601", "Архив раб. докум.", 72, 0.12, 0.25},
			{"602", "Серверная", 36, 0.4, 0.25},
			{"603", "Архив суд. дел", 72, 0.65, 0.25},
			{"604", "Вент. камера", 72, 0.15, 0.72},
			{"605", "Зал совещаний", 54, 0.55, 0.72},
			{"606", "АТС/диспетч.", 36, 0.78, 0.72},
		}
	default: // typical floor 2-5
		return []roomLabelData{
			{"201", "Зал с/з №1", 108, 0.12, 0.25},
			{"202", "Зал с/з №2", 108, 0.55, 0.25},
			{"203", "Каб. судьи", 30, 0.12, 0.72},
			{"204", "Каб. судьи", 30, 0.28, 0.72},
			{"205", "Совещ. комн.", 20, 0.42, 0.72},
			{"206", "Каб. судьи", 30, 0.55, 0.72},
			{"207", "Каб. судьи", 30, 0.70, 0.72},
			{"208", "Аппарат суда", 72, 0.82, 0.5},
		}
	}
}

func drawRoomPartitions(d *dxf.Document, room building.Room, ac AxisConfig, scale int, pw float64) {
	rx0 := ac.OX + px(room.X, scale)
	ry0 := ac.OY + px(room.Y, scale)
	rx1 := rx0 + px(room.Width, scale)
	ry1 := ry0 + px(room.Height, scale)

	pts := dxf.Rect(rx0, ry0, rx1, ry1)
	d.Polyline("ПЕРЕГОРОДКИ", dxf.LW025, true, pts)
}

func drawToilet(d *dxf.Document, x0, y0, x1, y1 float64) {
	d.Polyline("САНУЗЛЫ", dxf.LW025, true, dxf.Rect(x0, y0, x1, y1))
	// Simplified toilet bowl symbol
	w := x1 - x0
	h := y1 - y0
	// Toilet bowl (ellipse approximated with circle)
	d.Circle("САНУЗЛЫ", dxf.LW013, x0+w*0.3, y0+h*0.3, w*0.18)
	// Sink rectangle
	sinkW := w * 0.35
	sinkH := h * 0.25
	sinkX := x0 + w*0.5
	sinkY := y0 + h*0.6
	d.Polyline("САНУЗЛЫ", dxf.LW013, true, dxf.Rect(sinkX, sinkY, sinkX+sinkW, sinkY+sinkH))
}

func drawWindowsOnAxes(d *dxf.Document, b *building.Building, ac AxisConfig, scale int) {
	ew := px(b.Dims.WallExtM, scale)
	winW := px(1.5, scale) // typical window width 1500mm

	xc := b.Axes.XCoords()
	yc := b.Axes.YCoords()

	// Windows on south and north walls (at midpoints of X spans)
	for i := 0; i < len(xc)-1; i++ {
		midX := (xc[i] + xc[i+1]) / 2.0
		x := ac.OX + px(midX, scale)
		// Skip entrance area (middle spans of south wall)
		mid := len(xc) / 2
		if i == mid-1 || i == mid {
			continue
		}
		DrawWindowH(d, x-winW/2, ac.OY, x+winW/2, ew)       // south
		DrawWindowH(d, x-winW/2, ac.OY+ac.BH, x+winW/2, ew) // north
	}

	// Windows on west wall (at midpoints of Y spans, skip stairwell spans)
	winW2 := px(1.2, scale)
	for j := 0; j < len(yc)-1; j++ {
		midY := (yc[j] + yc[j+1]) / 2.0
		y := ac.OY + px(midY, scale)
		DrawWindowV(d, ac.OX, y-winW2/2, y+winW2/2, ew)       // west
		DrawWindowV(d, ac.OX+ac.BW, y-winW2/2, y+winW2/2, ew) // east
	}
}
