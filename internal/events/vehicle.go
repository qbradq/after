package events

import (
	"strings"

	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/lib/util"
)

func init() {
	rve("OpenVehicleDoor", openVehicleDoor)
	rve("CloseVehicleDoor", closeVehicleDoor)
}

func openVehicleDoor(v *game.Vehicle, l *game.VehicleLocation, i *game.Item, pp util.Point, src *game.Actor, m *game.CityMap) error {
	ff := util.FloodFill{
		Matches: func(p util.Point) bool {
			l := v.GetLocationAbsolute(p)
			if l == nil {
				return false
			}
			for _, p := range l.Parts {
				if p.TemplateID == i.TemplateID {
					return true
				}
			}
			return false
		},
		Set: func(p util.Point) {
			l := v.GetLocationAbsolute(p)
			if l == nil {
				return
			}
			for _, p := range l.Parts {
				if p.TemplateID == i.TemplateID {
					l.Remove(p)
					np := game.NewItem("Open"+p.TemplateID, m.Now, false)
					l.Add(np)
					m.FlagBitmapsForVehicle(v, v.Bounds)
				}
			}
		},
	}
	ff.Execute(pp)
	return nil
}

func closeVehicleDoor(v *game.Vehicle, l *game.VehicleLocation, i *game.Item, pp util.Point, src *game.Actor, m *game.CityMap) error {
	ff := util.FloodFill{
		Matches: func(p util.Point) bool {
			l := v.GetLocationAbsolute(p)
			if l == nil {
				return false
			}
			for _, p := range l.Parts {
				if p.TemplateID == i.TemplateID {
					return true
				}
			}
			return false
		},
		Set: func(p util.Point) {
			l := v.GetLocationAbsolute(p)
			if l == nil {
				return
			}
			for _, p := range l.Parts {
				if p.TemplateID == i.TemplateID {
					l.Remove(p)
					s, _ := strings.CutPrefix(i.TemplateID, "Open")
					np := game.NewItem(s, m.Now, false)
					np.Position = p.Position
					l.Add(np)
					m.FlagBitmapsForVehicle(v, v.Bounds)
				}
			}
		},
	}
	ff.Execute(pp)
	return nil
}
