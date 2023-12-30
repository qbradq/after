package mods

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/qbradq/after/internal/chunkgen"
	"github.com/qbradq/after/internal/game"
	"github.com/qbradq/after/internal/itemgen"
	"github.com/qbradq/after/internal/tilegen"
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
	tilegen.TileGens = map[string]tilegen.TileGen{}
	itemgen.ItemGens = map[string]itemgen.ItemGen{}
	chunkgen.ChunkGens = map[string]*chunkgen.ChunkGen{}
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
		mod := mods[id]
		if err := mod.loadItemGens(); err != nil {
			return err
		}
	}
	// Actors
	for _, id := range ids {
		mod := mods[id]
		if err := mod.loadActors(); err != nil {
			return err
		}
	}
	// Tiles
	for _, id := range ids {
		mod := mods[id]
		if err := mod.loadTiles(); err != nil {
			return err
		}
	}
	// TileGens
	for _, id := range ids {
		mod := mods[id]
		if err := mod.loadTileGens(); err != nil {
			return err
		}
	}
	// ChunkGens
	for _, id := range ids {
		mod := mods[id]
		if err := mod.loadChunkGens(); err != nil {
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
		var gens map[string]tilegen.TileGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			return err
		}
		for k, gen := range gens {
			if _, found := tilegen.TileGens[k]; found {
				return fmt.Errorf("duplicate tile generator %s", k)
			}
			tilegen.TileGens[k] = gen
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
		var gens []*chunkgen.ChunkGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			return err
		}
		for _, g := range gens {
			if len(g.ID) < 1 {
				return errors.New("chunk generator with no ID given")
			}
			if _, found := chunkgen.ChunkGens[g.ID]; found {
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
			chunkgen.ChunkGens[g.ID] = g
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
		var gens map[string]itemgen.ItemGen
		err = json.Unmarshal(d, &gens)
		if err != nil {
			panic(err)
		}
		for k, gen := range gens {
			if _, found := itemgen.ItemGens[k]; found {
				return fmt.Errorf("duplicate item generator definition %s", k)
			}
			itemgen.ItemGens[k] = gen
		}
	}
	return nil
}
