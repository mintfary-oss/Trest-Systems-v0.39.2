// Package building defines the data model for a building.
// Buildings are described in YAML configuration files.
package building

// RoomType classifies a room's function.
type RoomType string

const (
	RoomCourtroom    RoomType = "courtroom"    // зал судебных заседаний
	RoomJudge        RoomType = "judge_office" // кабинет судьи
	RoomOffice       RoomType = "office"       // кабинет/рабочее помещение
	RoomWaiting      RoomType = "waiting"      // зал ожидания
	RoomLobby        RoomType = "lobby"        // вестибюль
	RoomSecurity     RoomType = "security"     // пост охраны
	RoomStaircase    RoomType = "staircase"    // лестничная клетка
	RoomElevator     RoomType = "elevator"     // лифтовая шахта
	RoomToilet       RoomType = "toilet"       // санузел
	RoomTechnical    RoomType = "technical"    // техническое помещение
	RoomCorridor     RoomType = "corridor"     // коридор
	RoomEscort       RoomType = "escort"       // конвойный блок
	RoomHolding      RoomType = "holding"      // камера подсудимых
	RoomArchive      RoomType = "archive"      // архив
	RoomConference   RoomType = "conference"   // зал совещаний
	RoomServer       RoomType = "server"       // серверная
	RoomMechanical   RoomType = "mechanical"   // вентиляционная/механическая
	RoomDeliberation RoomType = "deliberation" // совещательная комната
)

// Room describes a single room or functional zone on a floor plan.
type Room struct {
	ID     string   `yaml:"id"`
	Name   string   `yaml:"name"`
	Type   RoomType `yaml:"type"`
	X      float64  `yaml:"x"`       // origin X from axis 1 (m)
	Y      float64  `yaml:"y"`       // origin Y from axis А (m)
	Width  float64  `yaml:"width"`   // width along X (m)
	Height float64  `yaml:"height"`  // depth along Y (m)
	AreaM2 float64  `yaml:"area_m2"` // area (m²), auto-calculated if 0
}

// Area returns the room area in m².
func (r Room) Area() float64 {
	if r.AreaM2 > 0 {
		return r.AreaM2
	}
	return r.Width * r.Height
}

// Floor describes one floor of the building.
type Floor struct {
	Number    int     `yaml:"number"`    // floor number (0=basement, 1=ground, ...)
	Elevation float64 `yaml:"elevation"` // floor elevation (m)
	Name      string  `yaml:"name"`      // e.g. "1-й этаж", "Типовой этаж"
	Rooms     []Room  `yaml:"rooms"`
}

// Wall describes a wall segment (internal).
type Wall struct {
	X0, Y0    float64 `yaml:"x0"`
	X1, Y1    float64 `yaml:"x1"`
	Thickness float64 `yaml:"thickness"`
	Load      bool    `yaml:"load"` // load-bearing?
}

// AxisGrid defines the structural axis grid.
type AxisGrid struct {
	XSpans  []float64 `yaml:"x_spans"`  // spans between X axes (m)
	YSpans  []float64 `yaml:"y_spans"`  // spans between Y axes (m)
	XLabels []string  `yaml:"x_labels"` // e.g. ["1","2","3",...]
	YLabels []string  `yaml:"y_labels"` // e.g. ["А","Б","В",...]
}

// XCoords returns cumulative X coordinates of axes.
func (g AxisGrid) XCoords() []float64 {
	coords := make([]float64, len(g.XSpans)+1)
	coords[0] = 0
	for i, s := range g.XSpans {
		coords[i+1] = coords[i] + s
	}
	return coords
}

// YCoords returns cumulative Y coordinates of axes.
func (g AxisGrid) YCoords() []float64 {
	coords := make([]float64, len(g.YSpans)+1)
	coords[0] = 0
	for i, s := range g.YSpans {
		coords[i+1] = coords[i] + s
	}
	return coords
}

// TotalWidth returns total building width along X (m).
func (g AxisGrid) TotalWidth() float64 {
	total := 0.0
	for _, s := range g.XSpans {
		total += s
	}
	return total
}

// TotalDepth returns total building depth along Y (m).
func (g AxisGrid) TotalDepth() float64 {
	total := 0.0
	for _, s := range g.YSpans {
		total += s
	}
	return total
}

// Organization describes the design organization.
type Organization struct {
	Name     string `yaml:"name"`
	Designer string `yaml:"designer"`
	Checker  string `yaml:"checker"`
	NormCtrl string `yaml:"norm_ctrl"`
	ChiefEng string `yaml:"chief_eng"`
	Approver string `yaml:"approver"`
}

// Client describes the client/owner.
type Client struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

// Metadata holds project-level metadata.
type Metadata struct {
	ObjectName string `yaml:"object_name"` // полное наименование объекта
	Address    string `yaml:"address"`     // адрес объекта
	Stage      string `yaml:"stage"`       // стадия (П, РД)
	Year       string `yaml:"year"`
	DocPrefix  string `yaml:"doc_prefix"` // шифр (напр. "68-2026")
}

// Dimensions holds the key building dimensions.
type Dimensions struct {
	WallExtM   float64 `yaml:"wall_ext_m"`   // наружная стена (м)
	WallIntM   float64 `yaml:"wall_int_m"`   // несущая внутренняя стена (м)
	WallPrtM   float64 `yaml:"wall_prt_m"`   // перегородка (м)
	FloorHM    float64 `yaml:"floor_h_m"`    // высота этажа (м)
	BasementHM float64 `yaml:"basement_h_m"` // высота подвала (м)
	SlabHM     float64 `yaml:"slab_h_m"`     // толщина перекрытия (м)
	FndDepthM  float64 `yaml:"fnd_depth_m"`  // глубина фундамента (м)
}

// Plot describes the land plot.
type Plot struct {
	AreaHa    float64 `yaml:"area_ha"`
	WidthM    float64 `yaml:"width_m"`
	DepthM    float64 `yaml:"depth_m"`
	SetbackNM float64 `yaml:"setback_n_m"` // отступ с севера
	SetbackSM float64 `yaml:"setback_s_m"` // отступ с юга
	SetbackEM float64 `yaml:"setback_e_m"` // отступ с востока
	SetbackWM float64 `yaml:"setback_w_m"` // отступ с запада
}

// Building is the top-level configuration for a building project.
type Building struct {
	Meta     Metadata     `yaml:"meta"`
	Org      Organization `yaml:"organization"`
	Client   Client       `yaml:"client"`
	Axes     AxisGrid     `yaml:"axes"`
	Dims     Dimensions   `yaml:"dimensions"`
	Plot     Plot         `yaml:"plot"`
	NFloors  int          `yaml:"n_floors"` // надземных этажей
	Basement bool         `yaml:"basement"` // есть ли подвал
	Floors   []Floor      `yaml:"floors"`
	Walls    []Wall       `yaml:"walls,omitempty"` // дополнительные стены
}

// FloorByNumber returns the floor with the given number, or nil.
func (b *Building) FloorByNumber(n int) *Floor {
	for i := range b.Floors {
		if b.Floors[i].Number == n {
			return &b.Floors[i]
		}
	}
	return nil
}

// TypicalFloor returns a representative typical floor (2nd floor if present).
func (b *Building) TypicalFloor() *Floor {
	f := b.FloorByNumber(2)
	if f == nil {
		f = b.FloorByNumber(1)
	}
	return f
}
