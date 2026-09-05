package ifcsemantic

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
)

var refRE = regexp.MustCompile(`#([0-9]+)`)

type Graph struct {
	Out map[int][]int
	In  map[int][]int
}

// BuildGraph builds a lightweight IFC reference graph from raw STEP attributes.
// It deliberately keeps duplicate references out while preserving first-seen order.
func BuildGraph(entities []ifc.Entity) Graph {
	g := Graph{Out: map[int][]int{}, In: map[int][]int{}}
	for _, e := range entities {
		seen := map[int]bool{}
		for _, a := range e.Attributes {
			for _, m := range refRE.FindAllStringSubmatch(a, -1) {
				id, _ := strconv.Atoi(m[1])
				if id == e.ID || seen[id] {
					continue
				}
				seen[id] = true
				g.Out[e.ID] = append(g.Out[e.ID], id)
				g.In[id] = append(g.In[id], e.ID)
			}
		}
	}
	return g
}

func (g Graph) References(id int) []int   { return append([]int(nil), g.Out[id]...) }
func (g Graph) ReferencedBy(id int) []int { return append([]int(nil), g.In[id]...) }

func IsSpatialContainer(t string) bool {
	switch strings.ToUpper(t) {
	case "IFCPROJECT", "IFCSITE", "IFCBUILDING", "IFCBUILDINGSTOREY", "IFCSPACE":
		return true
	}
	return false
}
