package ifcsemantic

import "strings"

// CoreClass identifies the IFC classes used by the construction hierarchy and common building elements.
type CoreClass string

const (
	ClassProject  CoreClass = "IfcProject"
	ClassSite     CoreClass = "IfcSite"
	ClassBuilding CoreClass = "IfcBuilding"
	ClassStorey   CoreClass = "IfcBuildingStorey"
	ClassSpace    CoreClass = "IfcSpace"
	ClassWall     CoreClass = "IfcWall"
	ClassSlab     CoreClass = "IfcSlab"
	ClassRoof     CoreClass = "IfcRoof"
	ClassDoor     CoreClass = "IfcDoor"
	ClassWindow   CoreClass = "IfcWindow"
	ClassColumn   CoreClass = "IfcColumn"
	ClassBeam     CoreClass = "IfcBeam"
)

var coreClasses = map[string]CoreClass{
	"IFCPROJECT": ClassProject, "IFCSITE": ClassSite, "IFCBUILDING": ClassBuilding,
	"IFCBUILDINGSTOREY": ClassStorey, "IFCSPACE": ClassSpace,
	"IFCWALL": ClassWall, "IFCWALLSTANDARDCASE": ClassWall,
	"IFCSLAB": ClassSlab, "IFCROOF": ClassRoof, "IFCDOOR": ClassDoor,
	"IFCWINDOW": ClassWindow, "IFCCOLUMN": ClassColumn, "IFCBEAM": ClassBeam,
}

func ClassOf(typeName string) (CoreClass, bool) {
	c, ok := coreClasses[strings.ToUpper(strings.TrimSpace(typeName))]
	return c, ok
}

// CoreElements returns only schema-independent core construction classes.
func (m Model) CoreElements() []Element {
	out := make([]Element, 0)
	for _, e := range m.Elements {
		if _, ok := ClassOf(e.Type); ok {
			out = append(out, e)
		}
	}
	return out
}
