// Package mods provides facilities for loading game modifications. Game
// modifications define every tile, chunk, item and actor in the game.
package mods

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/qbradq/after/internal/citygen"
	"github.com/qbradq/after/internal/game"
)

func init() {
	// Load all mods
	files, err := os.ReadDir("mods")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		p := &Mod{
			Path: path.Join("mods", f.Name()),
		}
		d, err := os.ReadFile(path.Join(p.Path, "mod.json"))
		if err != nil {
			panic(err)
		}
		json.Unmarshal(d, p)
		if len(p.ID) < 1 {
			panic(fmt.Errorf("mod with no ID specified %s", p.Path))
		}
		if _, found := mods[p.ID]; found {
			panic(fmt.Errorf("duplicate mod ID %s from %s", p.ID, p.Path))
		}
		mods[p.ID] = p
	}
}

// Global map of mods
var mods = map[string]*Mod{}

// Mod represents a single content bundle for the After engine.
type Mod struct {
	Path        string // Base path to the mod
	ID          string // Unique ID
	Name        string // Display name
	Description string // Descriptive sentence
	MinIndex    int    // Minimum index value this mod uses
	MaxIndex    int    // Maximum index value this mod uses
}

// UnloadAllMods unloads all currently loaded mods.
func UnloadAllMods() {
	game.HelpPages = map[string]*game.HelpPage{}
	game.TileDefs = []*game.TileDef{}
	game.TileRefs = map[string]game.TileRef{}
	game.TileCrossRefs = []*game.TileDef{}
	game.TileCrossRefForRef = map[game.TileRef]game.TileCrossRef{}
	game.TileGens = map[string]game.TileGen{}
	game.ItemGens = map[string]game.ItemGen{}
	citygen.ChunkGenGroups = map[string]*citygen.ChunkGenGroup{}
	citygen.Scenarios = map[string]*citygen.Scenario{}
	game.ItemDefs = map[string]*game.Item{}
	game.ActorDefs = map[string]*game.Actor{}
}

// LoadMods loads all of the listed mods.
func LoadMods(ids []string) error {
	UnloadAllMods()
	// Help pages
	for _, id := range ids {
		mod, found := mods[id]
		if !found {
			return fmt.Errorf("mod %s not found", id)
		}
		if err := mod.loadHelp(); err != nil {
			return err
		}
	}
	// Items
	for _, id := range ids {
		if err := mods[id].loadItems(); err != nil {
			return err
		}
	}
	// ItemGens
	for _, id := range ids {
		if err := mods[id].loadItemGens(); err != nil {
			return err
		}
	}
	// Compile content statements
	for _, i := range game.ItemDefs {
		if err := i.CacheContentStatements(); err != nil {
			return err
		}
	}
	// Actors
	for _, id := range ids {
		if err := mods[id].loadActors(); err != nil {
			return err
		}
	}
	// Compile equipment statements
	for _, a := range game.ActorDefs {
		if err := a.CacheEquipmentStatements(); err != nil {
			return err
		}
	}
	// Tiles
	for _, id := range ids {
		if err := mods[id].loadTiles(); err != nil {
			return err
		}
	}
	// TileGens
	for _, id := range ids {
		if err := mods[id].loadTileGens(); err != nil {
			return err
		}
	}
	// ChunkGens
	for _, id := range ids {
		if err := mods[id].loadChunkGens(); err != nil {
			return err
		}
	}
	// Scenarios
	for _, id := range ids {
		if err := mods[id].loadScenarios(); err != nil {
			return err
		}
	}
	return nil
}

// loadHelp loads the mod's help pages.
func (m *Mod) loadHelp() error {
	doFile := func(hp, dp string, f fs.DirEntry) error {
		fp := path.Join(dp, f.Name())
		d, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		hp += "/" + f.Name()
		lines := strings.Split(string(d), "\n")
		if len(lines) < 1 {
			return fmt.Errorf("malformed help file %s", fp)
		}
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], "\n")
		}
		game.HelpPages[hp] = &game.HelpPage{
			Path:     hp,
			Title:    lines[0],
			Contents: lines[1:],
		}
		return nil
	}
	var doDir func(string, string) error
	doDir = func(hp, dp string) error {
		files, err := os.ReadDir(dp)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		for _, f := range files {
			if f.IsDir() {
				if err := doDir(hp+"/"+f.Name(), path.Join(dp, f.Name())); err != nil {
					return err
				}
			} else {
				if err := doFile(hp, dp, f); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return doDir(m.ID, path.Join(m.Path, "help"))
}

// loadTiles loads the mod's tile definitions.
func (m *Mod) loadTiles() error {
	files, err := os.ReadDir(path.Join(m.Path, "tiles"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "tiles", f.Name()))
		if err != nil {
			return err
		}
		var defs map[string]*game.TileDef
		err = json.Unmarshal(d, &defs)
		if err != nil {
			return err
		}
		for k, def := range defs {
			id := len(game.TileDefs)
			if _, found := game.TileRefs[k]; found {
				return fmt.Errorf("duplicate tile definition %s", k)
			}
			def.ID = k
			def.BackRef = game.TileRef(id)
			game.TileDefs = append(game.TileDefs, def)
			game.TileRefs[k] = game.TileRef(id)
		}
	}
	return nil
}

// loadTileGens loads the mod's tile generators.
func (m *Mod) loadTileGens() error {
	files, err := os.ReadDir(path.Join(m.Path, "tilegens"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "tilegens", f.Name()))
		if err != nil {
			return err
		}
		var gens map[string]game.TileGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			return err
		}
		for k, gen := range gens {
			if _, found := game.TileGens[k]; found {
				return fmt.Errorf("duplicate tile generator %s", k)
			}
			game.TileGens[k] = gen
		}
	}
	return nil
}

// loadChunkGens loads the mod's chunk generators.
func (m *Mod) loadChunkGens() error {
	// Load chunk generators
	files, err := os.ReadDir(path.Join(m.Path, "chunks"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "chunks", f.Name()))
		if err != nil {
			return err
		}
		var gens []*citygen.ChunkGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			return err
		}
		for _, g := range gens {
			if len(g.Group) < 1 {
				return errors.New("chunk generator with no group given")
			}
			if len(g.Variant) < 1 {
				return errors.New("chunk generator with no variant given")
			}
			if len(g.Map) != g.Height*game.ChunkHeight || len(g.Map[0]) != g.Width*game.ChunkWidth {
				return fmt.Errorf("chunk generator group %s variant %s has the wrong dimensions", g.Group, g.Variant)
			}
			for iRow, row := range g.Map {
				for iCol, r := range row {
					if _, found := g.Tiles[string(r)]; !found {
						return fmt.Errorf("chunk generator %s at %dx%d references tile %s not in tiles list", g.Group, iCol, iRow, string(r))
					}
				}
			}
			if group, found := citygen.ChunkGenGroups[g.Group]; found {
				group.Add(g)
			} else {
				group = citygen.NewChunkGenGroup(g.Group)
				citygen.ChunkGenGroups[g.Group] = group
				group.Add(g)
			}
		}
	}
	return nil
}

// loadActors loads the mod's actor definitions.
func (m *Mod) loadActors() error {
	files, err := os.ReadDir(path.Join(m.Path, "actors"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "actors", f.Name()))
		if err != nil {
			return err
		}
		var actors map[string]*game.Actor
		err = json.Unmarshal(d, &actors)
		if err != nil {
			return err
		}
		for k, a := range actors {
			if _, found := game.ActorDefs[k]; found {
				return fmt.Errorf("duplicate actor definition %s", k)
			}
			a.TemplateID = k
			game.ActorDefs[k] = a
		}
	}
	return nil
}

// loadItems loads the mod's item definitions.
func (m *Mod) loadItems() error {
	files, err := os.ReadDir(path.Join(m.Path, "items"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "items", f.Name()))
		if err != nil {
			return err
		}
		var items map[string]*game.Item
		err = json.Unmarshal(d, &items)
		if err != nil {
			panic(err)
		}
		for k, i := range items {
			if _, found := game.ItemDefs[k]; found {
				return fmt.Errorf("duplicate item definition %s", k)
			}
			i.TemplateID = k
			game.ItemDefs[k] = i
		}
	}
	return nil
}

// loadItemGens loads the mod's item generators.
func (m *Mod) loadItemGens() error {
	files, err := os.ReadDir(path.Join(m.Path, "itemgens"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "itemgens", f.Name()))
		if err != nil {
			return err
		}
		var gens map[string]game.ItemGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			panic(err)
		}
		for k, gen := range gens {
			if _, found := game.ItemGens[k]; found {
				return fmt.Errorf("duplicate item generator definition %s", k)
			}
			game.ItemGens[k] = gen
		}
	}
	return nil
}

// loadScenarios loads the mod's scenario definitions.
func (m *Mod) loadScenarios() error {
	files, err := os.ReadDir(path.Join(m.Path, "scenarios"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	for _, f := range files {
		d, err := os.ReadFile(path.Join(m.Path, "scenarios", f.Name()))
		if err != nil {
			return err
		}
		var defs map[string]*citygen.Scenario
		err = json.Unmarshal(d, &defs)
		if err != nil {
			return err
		}
		for k, def := range defs {
			if _, found := citygen.Scenarios[k]; found {
				return fmt.Errorf("duplicate scenario definition %s", k)
			}
			citygen.Scenarios[k] = def
		}
	}
	return nil
}
