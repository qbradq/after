package game

import (
	"github.com/qbradq/after/lib/util"
)

const (
	CityMapWidth  int = 660 // Width of a city map in chunks
	CityMapHeight int = 660 // Height of a city map in chunks
)

// CityMap represents the entire world of the game in terms of which chunks go
// where.
type CityMap struct {
	Bounds     util.Rect // Bounds of the city map in chunks
	TileBounds util.Rect // Bounds of the city map in tiles
	Chunks     []*Chunk  // The chunks of the map
}

// NewCityMap allocates and returns a new CityMap structure.
func NewCityMap() *CityMap {
	m := &CityMap{
		Bounds:     util.NewRectWH(CityMapWidth, CityMapHeight),
		TileBounds: util.NewRectWH(CityMapWidth*ChunkWidth, CityMapHeight*ChunkHeight),
		Chunks:     make([]*Chunk, CityMapWidth*CityMapHeight),
	}
	for i := range m.Chunks {
		m.Chunks[i] = NewChunk()
	}
	return m
}

// GetChunkFromMapPoint returns the chunk definition at the given map location
// or nil if out of bounds. Note the point is in chunks not tiles.
func (m *CityMap) GetChunkFromMapPoint(p util.Point) *Chunk {
	if !m.Bounds.Contains(p) {
		return nil
	}
	return m.Chunks[p.Y*CityMapWidth+p.X]
}

// GetChunk returns the correct chunk for the given absolute tile point or nil
// if the point is out of bounds.
func (m *CityMap) GetChunk(p util.Point) *Chunk {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	return m.Chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
}

// GetTile returns the tile at the given absolute tile point or nil if the point
// is out of bounds.
func (m *CityMap) GetTile(p util.Point) *TileDef {
	if !m.TileBounds.Contains(p) {
		return nil
	}
	c := m.Chunks[(p.Y/ChunkHeight)*CityMapWidth+(p.X/ChunkWidth)]
	t := c.Tiles[(p.Y%ChunkHeight)*ChunkWidth+(p.X%ChunkWidth)]
	return t
}

// Load ensures that all chunks in the area given in absolute tile coordinates
// have been generated and are loaded into memory.
func (m *CityMap) Load(area util.Rect) {
	r := util.Rect{
		TL: util.Point{
			X: area.TL.X / ChunkWidth,
			Y: area.TL.Y / ChunkHeight,
		},
		BR: util.Point{
			X: area.BR.X / ChunkWidth,
			Y: area.BR.Y / ChunkHeight,
		},
	}
	if r.TL.X < 0 {
		r.TL.X = 0
	}
	if r.TL.Y < 0 {
		r.TL.Y = 0
	}
	if r.BR.X >= CityMapWidth {
		r.BR.X = CityMapWidth - 1
	}
	if r.BR.Y >= CityMapHeight {
		r.BR.Y = CityMapHeight - 1
	}
	var p util.Point
	for p.Y = r.TL.Y; p.Y <= r.BR.Y; p.Y++ {
		for p.X = r.TL.X; p.X <= r.BR.X; p.X++ {
			m.Chunks[p.Y*CityMapWidth+p.X].Load()
		}
	}
}
