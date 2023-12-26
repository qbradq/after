package game

import (
	"bytes"
	"fmt"

	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

// TileDefs is the global TileRef-to-*TileDef reference.
var TileDefs []*TileDef

// TileRefs is the global string-to-TileRef reference.
var TileRefs = map[string]TileRef{}

// TileRef represents a three foot by three foot area of the world and is a
// reference into the global TileDefs slice.
type TileRef uint16

// tileCrossRef is the value we write to the save database for persisting tile
// selections. See tileCrossRefs.
type tileCrossRef uint16

// tileCrossRefs indexes all valid tileCrossRef values in the save database to
// the tile defs.
var tileCrossRefs []*TileDef

// tileRefMap is a map of tileCrossRef values to tile IDs.
var tileRefMap = map[tileCrossRef]string{}

// tileCrossRefForRefs is a map of tileCrossRef associated TileRefs.
var tileCrossRefForRef = map[TileRef]tileCrossRef{}

// crossReferencesDirty is true when there have been additions made to the tile
// cross references since the last call to SaveTileRefs().
var crossReferencesDirty bool

// getTileCrossRef returns the tileCrossRef for the given TileRef. If this
// TileRef has never been cross-referenced before it will be added.
func getTileCrossRef(r TileRef) tileCrossRef {
	x, found := tileCrossRefForRef[r]
	if !found {
		t := TileDefs[r]
		x = tileCrossRef(len(tileCrossRefs))
		tileCrossRefs = append(tileCrossRefs, t)
		tileRefMap[x] = t.ID
		tileCrossRefForRef[r] = x
		crossReferencesDirty = true
	}
	return x
}

// SaveTileRefs saves tileRefMap.
func SaveTileRefs() {
	// Write out the map
	w := bytes.NewBuffer(nil)
	util.PutUint32(w, 0) // Version
	util.PutUint16(w, uint16(len(tileRefMap)))
	for k, v := range tileRefMap {
		util.PutUint16(w, uint16(k))
		util.PutString(w, v)
	}
	SaveValue("TileRefs", w.Bytes())
	// Flag as no longer dirty
	crossReferencesDirty = false
}

// LoadTileRefs loads tileRefMap and rebuilds tileCrossRefs.
func LoadTileRefs() {
	tileRefMap = make(map[tileCrossRef]string)
	// Read from database
	r := LoadValue("TileRefs")
	if r == nil {
		// TileRefs have not yet been written - probably a new save
		return
	}
	_ = util.GetUint32(r) // Version
	n := int(util.GetUint16(r))
	for i := 0; i < n; i++ {
		tileRefMap[tileCrossRef(util.GetUint16(r))] = util.GetString(r)
	}
	// Rebuild the cross references
	tileCrossRefs = make([]*TileDef, n)
	tileCrossRefForRef = map[TileRef]tileCrossRef{}
	for k, v := range tileRefMap {
		r, found := TileRefs[v]
		if !found {
			panic(fmt.Errorf("tile cross-reference referenced non-loaded tile %s", v))
		}
		t := TileDefs[r]
		tileCrossRefs[k] = t
		tileCrossRefForRef[r] = k
	}
}

// TileDef represents all of the data associated with a single tile.
type TileDef struct {
	BackRef TileRef      // The TileRef that indexes this TileDef within TileDefs, used to accelerate saving
	ID      string       // The unique ID of the tile
	Name    string       // Descriptive name of the tile
	Rune    string       // Map display rune
	Fg      termui.Color // Foreground display color
	Bg      termui.Color // Background display color
}
