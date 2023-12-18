package mods

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/qbradq/after/internal/chunkgen"
	"github.com/qbradq/after/internal/game"
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

// LoadModByID loads the named mod.
func LoadModByID(id string) error {
	mod, found := mods[id]
	if !found {
		return fmt.Errorf("mod %s not found", id)
	}
	return mod.Load()
}

// Load loads the mod from disk and inserts all of the mod's data structures
// into the appropriate registries.
func (m *Mod) Load() error {
	// Load tiles
	files, err := os.ReadDir(path.Join(m.Path, "tiles"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		for _, f := range files {
			d, err := os.ReadFile(path.Join(m.Path, "tiles", f.Name()))
			if err != nil {
				return err
			}
			var defs map[string]*game.TileDef
			err = json.Unmarshal(d, &defs)
			if err != nil {
				panic(err)
			}
			for k, def := range defs {
				id := len(game.TileDefs)
				if _, found := game.TileRefs[k]; found {
					panic(fmt.Errorf("duplicate tile definition %s", k))
				}
				def.BackRef = game.TileRef(id)
				game.TileDefs = append(game.TileDefs, def)
				game.TileRefs[k] = game.TileRef(id)
			}
		}
	}
	// Load tile generators
	files, err = os.ReadDir(path.Join(m.Path, "tilegens"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		for _, f := range files {
			d, err := os.ReadFile(path.Join(m.Path, "tilegens", f.Name()))
			if err != nil {
				return err
			}
			var gens map[string]*tilegen.TileGen
			err = json.Unmarshal(d, &gens)
			if err != nil {
				panic(err)
			}
			for k, gen := range gens {
				if _, found := tilegen.TileGens[k]; found {
					panic(fmt.Errorf("duplicate tile generator %s", k))
				}
				tilegen.TileGens[k] = gen
			}
		}
	}
	// Load chunk generators
	files, err = os.ReadDir(path.Join(m.Path, "chunks"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		for _, f := range files {
			d, err := os.ReadFile(path.Join(m.Path, "chunks", f.Name()))
			if err != nil {
				return err
			}
			var gens []*chunkgen.ChunkGen
			err = json.Unmarshal(d, &gens)
			if err != nil {
				panic(err)
			}
			for _, g := range gens {
				if len(g.ID) < 1 {
					panic("chunk generator with no ID given")
				}
				if _, found := chunkgen.ChunkGens[g.ID]; found {
					panic(fmt.Errorf("duplicate chunk generator %s", g.ID))
				}
				if len(g.Map) != g.Height*game.ChunkHeight || len(g.Map[0]) != g.Width*game.ChunkWidth {
					panic(fmt.Errorf("chunk generator map %s has the wrong dimensions", g.ID))
				}
				for _, row := range g.Map {
					for _, r := range row {
						if _, found := g.Tiles[string(r)]; !found {
							panic(fmt.Errorf("chunk generator map %s references tile %s not in tiles list", g.ID, string(r)))
						}
					}
				}
				for _, t := range g.Tiles {
					if _, found := tilegen.TileGens[t]; !found {
						if _, found := game.TileRefs[t]; !found {
							panic(fmt.Errorf("chunk generator %s tiles list references non-existent tile generator or tile %s", g.ID, t))
						}
					}
				}
				chunkgen.ChunkGens[g.ID] = g
			}
		}
	}
	return nil
}
