package termgui

import (
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/mods"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// LoadMenu implements a menu to select existing saves.
type LoadMenu struct {
	list termui.List // List of saves
}

// NewLoadMenu creates a new load menu for use.
func NewLoadMenu(s termui.TerminalDriver) *LoadMenu {
	var lm *LoadMenu
	var items []string
	var saveInfos []*game.SaveInfo
	for _, si := range game.Saves {
		items = append(items, si.Name)
		saveInfos = append(saveInfos, si)
	}
	lm = &LoadMenu{
		list: termui.List{
			Boxed: true,
			Title: "Load Save",
			Items: items,
			Selected: func(td termui.TerminalDriver, i int) error {
				si := saveInfos[i]
				if err := mods.LoadMods(si.Mods); err != nil {
					panic(err)
				}
				if err := game.LoadSave(si.ID); err != nil {
					panic(err)
				}
				m := game.NewCityMap()
				m.LoadCityPlan()
				termui.RunMode(s, NewGameMode(m, util.NewPoint(0*game.ChunkWidth+game.ChunkWidth/2, 0*game.ChunkHeight+game.ChunkHeight/2)))
				game.CloseSave()
				return termui.ErrorQuit
			},
		},
	}
	return lm
}

// HandleEvent implements the termui.Mode interface.
func (m *LoadMenu) HandleEvent(s termui.TerminalDriver, e any) error {
	switch ev := e.(type) {
	case *termui.EventKey:
		if ev.Key == '\033' {
			return termui.ErrorQuit
		}
	case *termui.EventQuit:
		return termui.ErrorQuit
	}
	return m.list.HandleInput(s, e)
}

// Draw implements the termui.Mode interface.
func (m *LoadMenu) Draw(s termui.TerminalDriver) {
	w, h := s.Size()
	m.list.Bounds = util.NewRectWH(w, h).CenterRect(22, len(m.list.Items)+2)
	m.list.Draw(s)
}
