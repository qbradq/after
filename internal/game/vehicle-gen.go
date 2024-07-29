package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// VehicleGen encapsulates all of the parts and top-level functionality to
// generate a vehicle.
type VehicleGen struct {
	Group    string            // Group that this vehicle generator belongs to
	Variant  string            // Generator variant name
	Name     string            // Name of the generated vehicle
	Width    int               // Width of the layout in parts
	Height   int               // Height of the layout in parts
	Map      []string          // The generator map
	Legend   map[string]string // Legend translating layer characters to parts
	genCache []ItemStatement   // Cache of parsed generator statements
}

// VehicleGenGroup represents a group of vehicle generators.
type VehicleGenGroup struct {
	ID          string                 // ID of the group.
	Variants    map[string]*VehicleGen // Map of vehicle gens by variant name.
	VariantList []*VehicleGen          // List of vehicle gens
}

// NewVehicleGenGroup creates a new VehicleGenGroup ready for use.
func NewVehicleGenGroup(id string) *VehicleGenGroup {
	return &VehicleGenGroup{
		ID:       id,
		Variants: map[string]*VehicleGen{},
	}
}

// Add adds a variant to the group.
func (g *VehicleGenGroup) Add(v *VehicleGen) error {
	if _, duplicate := g.Variants[v.Variant]; duplicate {
		return fmt.Errorf("duplicate variant %s in vehicle gen group %s", v.Variant, g.ID)
	}
	g.Variants[v.Variant] = v
	g.VariantList = append(g.VariantList, v)
	return nil
}

// Get returns a pointer to one of the variant vehicle generators at random.
func (g *VehicleGenGroup) Get() *VehicleGen {
	if len(g.VariantList) < 1 {
		return nil
	}
	return g.VariantList[util.Random(0, len(g.VariantList))]
}

// VehicleGenGroups is the map of all vehicle generator groups.
var VehicleGenGroups = map[string]*VehicleGenGroup{}

// CacheGens creates a cache of the part generators for this variant.
func (g *VehicleGen) CacheGens() error {
	// Validate input
	if g.Height != len(g.Map) {
		return fmt.Errorf("bad height in vehicle generator %s:%s", g.Group, g.Variant)
	}
	for iLine, line := range g.Map {
		if g.Width != len(line) {
			return fmt.Errorf("bad width in vehicle generator %s:%s, line %d", g.Group, g.Variant, iLine)
		}
	}
	// Create cache
	g.genCache = make([]ItemStatement, g.Width*g.Height)
	for iy, line := range g.Map {
		for ix, r := range line {
			k := string(r)
			t, found := g.Legend[k]
			if !found {
				return fmt.Errorf("bad legend reference in vehicle generator %s:%s at %dx%d", g.Group, g.Variant, ix, iy)
			}
			s := ItemStatement{}
			if err := json.Unmarshal([]byte("\""+t+"\""), &s); err != nil {
				return err
			}
			g.genCache[iy*g.Width+ix] = append(g.genCache[iy*g.Width+ix], s...)
		}
	}
	return nil
}

// Generate returns a new, procedurally generated vehicle.
func (g *VehicleGen) Generate(now time.Time) *Vehicle {
	// Basic generation
	ret := newVehicle(util.NewPoint(g.Width, g.Height))
	ret.Name = g.Name
	// Parts generation
	var p util.Point
	for p.Y = 0; p.Y < g.Height; p.Y++ {
		for p.X = 0; p.X < g.Width; p.X++ {
			s := g.genCache[p.Y*g.Width+p.X]
			for _, i := range s.Evaluate(now) {
				if !ret.Attach(i, p) {
					Log.Log(termui.ColorRed, "Failed to attach part in vehicle generator")
					return nil
				}
			}
		}
	}
	return ret
}
