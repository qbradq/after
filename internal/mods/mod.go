// Package mods provides facilities for loading game modifications. Game
// modifications define every tile, chunk, item and actor in the game.
package mods

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

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
	game.TileDefs = []*game.TileDef{}
	game.TileRefs = map[string]game.TileRef{}
	game.TileCrossRefs = []*game.TileDef{}
	game.TileCrossRefForRef = map[game.TileRef]game.TileCrossRef{}
	citygen.TileGens = map[string]citygen.TileGen{}
	citygen.ItemGens = map[string]citygen.ItemGen{}
	citygen.ChunkGens = map[string]*citygen.ChunkGen{}
	citygen.Scenarios = map[string]*citygen.Scenario{}
	game.ItemDefs = map[string]*game.Item{}
	game.ActorDefs = map[string]*game.Actor{}
}

// LoadMods loads all of the listed mods.
func LoadMods(ids []string) error {
	UnloadAllMods()
	// Items
	for _, id := range ids {
		mod, found := mods[id]
		if !found {
			return fmt.Errorf("mod %s not found", id)
		}
		if err := mod.loadItems(); err != nil {
			return err
		}
	}
	// ItemGens
	for _, id := range ids {
		if err := mods[id].loadItemGens(); err != nil {
			return err
		}
	}
	// Actors
	for _, id := range ids {
		if err := mods[id].loadActors(); err != nil {
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
		var gens map[string]citygen.TileGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			return err
		}
		for k, gen := range gens {
			if _, found := citygen.TileGens[k]; found {
				return fmt.Errorf("duplicate tile generator %s", k)
			}
			citygen.TileGens[k] = gen
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
			if len(g.ID) < 1 {
				return errors.New("chunk generator with no ID given")
			}
			if _, found := citygen.ChunkGens[g.ID]; found {
				return fmt.Errorf("duplicate chunk generator %s", g.ID)
			}
			for iGenMap, genMap := range g.Maps {
				if len(genMap) != g.Height*game.ChunkHeight || len(genMap[0]) != g.Width*game.ChunkWidth {
					return fmt.Errorf("chunk generator map %s has the wrong dimensions", g.ID)
				}
				for iRow, row := range genMap {
					for iCol, r := range row {
						if _, found := g.Tiles[string(r)]; !found {
							return fmt.Errorf("chunk generator %s map %d:%dx%d references tile %s not in tiles list", g.ID, iGenMap, iCol, iRow, string(r))
						}
					}
				}
			}
			citygen.ChunkGens[g.ID] = g
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
		var gens map[string]citygen.ItemGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			panic(err)
		}
		for k, gen := range gens {
			if _, found := citygen.ItemGens[k]; found {
				return fmt.Errorf("duplicate item generator definition %s", k)
			}
			citygen.ItemGens[k] = gen
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
