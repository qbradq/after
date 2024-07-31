package game

import (
	"io"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/qbradq/after/lib/termui"
	"github.com/qbradq/after/lib/util"
)

const (
	ChunkWidth  int = 16 // Width of a chunk in tiles
	ChunkHeight int = 16 // Height of a chunk in tiles
)

// ChunkGen is the interface all chunk generators must implement for the game
// library to interact with them.
type ChunkGen interface {
	// Generate handles all initial procedural generation for the chunk
	Generate(*Chunk, *CityMap)
	// AssignStaticInfo inserts all of the non-procedurally generated bits into
	// the chunk, such as name and map rune.
	AssignStaticInfo(*Chunk)
	// GetGroup returns the group ID of the generator.
	GetGroup() string
	// GetVariant returns the variant name.
	GetVariant() string
}

// ChunkFlag encodes various bit flags of a chunk.
type ChunkFlags uint8

const (
	ChunkFlagsNone     ChunkFlags = 0b00000000 // No flags enabled
	ChunkFlagsOccupied ChunkFlags = 0b00000001 // Chunk was explicitly placed during the city gen process
)

// Chunk represents the smallest unit of city planning and contains the tiles,
// items and actors within its bounds.
type Chunk struct {
	// Values persisted at the CityMap level
	Generator      ChunkGen    // The chunk generator responsible for procedural generation
	ChunkGenOffset util.Point  // Offset from the top-left corner of the chunk generator
	Facing         util.Facing // Facing of the chunk during generation
	Flags          ChunkFlags  // Flags
	// Persistent values
	Tiles    []*TileDef    // Tile matrix
	Items    []*Item       // All items within the chunk
	Actors   []*Actor      // All actors within the chunk
	Vehicles []*Vehicle    // All vehicles who's Northwest corner are in this chunk
	HasSeen  bitmap.Bitmap // Bitmap of all spaces that have been previously viewed by the player
	// Reconstituted values
	Position          util.Point   // Position of the chunk on the city map in chunks
	Ref               uint32       // Reference index for the chunk
	Bounds            util.Rect    // Bounds of the chunk
	Name              string       // Descriptive name of the chunk
	MinimapRune       string       // Rune to display on the minimap
	MinimapForeground termui.Color // Foreground color of the rune on the minimap
	MinimapBackground termui.Color // Background color of the rune on the minimap
	// Working values
	Loaded       time.Time     // Time this chunk was loaded, the zero value means it is not in memory
	BlocksWalk   bitmap.Bitmap // Bitmap of all spaces that are blocked for walking
	BlocksVis    bitmap.Bitmap // Bitmap of all spaces that are blocked for visibility
	BlocksClimb  bitmap.Bitmap // Bitmap of all spaces that can be climbed
	bitmapsDirty bool          // If true the BlocksWalk and BlocksVis bitmaps need to be rebuilt before use

}

// NewChunk allocates and returns a new Chunk struct. Note that this struct does
// *not* have the Generator field set yet and all of the tile pointers are nil.
// See Load().
func NewChunk(x, y int, r uint32) *Chunk {
	c := &Chunk{
		Position:          util.NewPoint(x, y),
		Ref:               r,
		Bounds:            util.NewRectXYWH(x*ChunkWidth, y*ChunkHeight, ChunkWidth, ChunkHeight),
		Name:              "an error",
		MinimapRune:       "!",
		MinimapForeground: termui.ColorWhite,
		MinimapBackground: termui.ColorRed,
	}
	return c
}

func (c *Chunk) relOfs(p util.Point) uint32 {
	return uint32((p.Y-c.Bounds.TL.Y)*ChunkWidth + (p.X - c.Bounds.TL.X))
}

// Write writes the chunk to w.
func (c *Chunk) Write(w io.Writer) {
	util.PutUint32(w, 0)        // Version
	for _, t := range c.Tiles { // Tile map
		util.PutUint16(w, uint16(getTileCrossRef(t.BackRef)))
	}
	util.PutUint16(w, uint16(len(c.Items))) // Number of items
	for _, i := range c.Items {             // Items
		i.Write(w)
	}
	util.PutUint16(w, uint16(len(c.Actors))) // Number of actors
	for _, a := range c.Actors {             // Actors
		a.Write(w)
	}
	util.PutUint16(w, uint16(len(c.Vehicles))) // Number of vehicles
	for _, v := range c.Vehicles {             // Vehicles
		v.Write(w)
	}
	c.HasSeen.WriteTo(w) // Remembered bitmap
}

// Unload frees chunk-level persistent memory
func (c *Chunk) Unload() {
	c.Tiles = nil
	c.Items = nil
	c.Actors = nil
	c.Vehicles = nil
	c.HasSeen = nil
	c.Loaded = time.Time{}
}

// Read allocates memory and reads the chunk from r.
func (c *Chunk) Read(r io.Reader) {
	c.Tiles = make([]*TileDef, ChunkWidth*ChunkHeight)
	_ = util.GetUint32(r)    // Version
	for i := range c.Tiles { // Tile map
		x := TileCrossRef(util.GetUint16(r))
		c.Tiles[i] = TileCrossRefs[x]
	}
	n := int(util.GetUint16(r)) // Number of items
	for i := 0; i < n; i++ {    // Items
		c.Items = append(c.Items, NewItemFromReader(r))
	}
	n = int(util.GetUint16(r)) // Number of actors
	for i := 0; i < n; i++ {   // Actors
		c.Actors = append(c.Actors, NewActorFromReader(r))
	}
	n = int(util.GetUint16(r)) // Number of vehicles
	for i := 0; i < n; i++ {   // Vehicles
		c.Vehicles = append(c.Vehicles, NewVehicleFromReader(r))
	}
	c.HasSeen.ReadFrom(r) // Remembered bitmap
}

// RebuildBitmaps must be called after chunk load or generation in order to
// rebuild the walk and vis bitmap caches. The HasSeen bitmap is persistent.
func (c *Chunk) RebuildBitmaps() {
	// Clear bitmaps
	c.BlocksVis.Clear()
	c.BlocksWalk.Clear()
	c.BlocksClimb.Clear()
	// Consider tiles
	for i, t := range c.Tiles {
		if t.BlocksVis {
			c.BlocksVis.Set(uint32(i))
		}
		if t.BlocksWalk {
			c.BlocksWalk.Set(uint32(i))
			if !t.Climbable {
				c.BlocksClimb.Set(uint32(i))
			}
		}
	}
	// Consider items
	for _, i := range c.Items {
		if i.BlocksVis {
			c.BlocksVis.Set(c.relOfs(i.Position))
		}
		if i.BlocksWalk {
			c.BlocksWalk.Set(c.relOfs(i.Position))
			if !i.Climbable {
				c.BlocksClimb.Set(c.relOfs(i.Position))
			}
		}
	}
	// Consider actors
	for _, a := range c.Actors {
		c.BlocksWalk.Set(c.relOfs(a.Position))
		c.BlocksClimb.Set(c.relOfs(a.Position))
	}
	c.bitmapsDirty = false
}

// PlaceItemRelative adds the item to the chunk and adjusts the
// position from chunk-relative to absolute.
func (c *Chunk) PlaceItemRelative(i *Item) {
	i.Position.X += c.Bounds.TL.X
	i.Position.Y += c.Bounds.TL.Y
	c.PlaceItem(i, true)
}

// CanStack returns true if and item can be stacked on the given point.
func (c *Chunk) CanStack(p util.Point) bool {
	if !c.Bounds.Contains(p) {
		return false
	}
	t := c.Tiles[c.relOfs(p)]
	if t.BlocksStack {
		return false
	}
	for _, i := range c.Items {
		if i.Position != p {
			continue
		}
		if i.BlocksStack {
			return false
		}
	}
	return true
}

// PlaceItem places the item within the chunk. This is a no-op if the item's
// current position lies outside the chunk.
func (c *Chunk) PlaceItem(i *Item, force bool) bool {
	for _, o := range c.Items {
		if o == i {
			return false
		}
	}
	if !force {
		if !c.Bounds.Contains(i.Position) {
			return false
		}
		if !c.CanStack(i.Position) {
			return false
		}
	}
	if i.BlocksVis {
		c.BlocksVis.Set(c.relOfs(i.Position))
	}
	if i.BlocksWalk {
		c.BlocksWalk.Set(c.relOfs(i.Position))
	}
	if !i.Climbable {
		ref := c.relOfs(i.Position)
		c.BlocksClimb.Set(ref)
	}
	c.Items = append(c.Items, i)
	return true
}

// RemoveItem removes the item from the chunk. This is a no-op if the item's
// current position lies outside the chunk.
func (c *Chunk) RemoveItem(i *Item) bool {
	if !c.Bounds.Contains(i.Position) {
		return false
	}
	idx := -1
	for n, item := range c.Items {
		if item == i {
			idx = n
			break
		}
	}
	if idx < 0 {
		return false
	}
	// Remove from slice while maintaining order
	copy(c.Items[idx:], c.Items[idx+1:])
	c.Items[len(c.Items)-1] = nil
	c.Items = c.Items[:len(c.Items)-1]
	// Bitmap updates
	if i.BlocksVis || i.BlocksWalk || !i.Climbable {
		c.bitmapsDirty = true
	}
	return true
}

// PlaceVehicle adds the vehicle to this chunk's list of managed vehicles.
func (c *Chunk) PlaceVehicle(v *Vehicle) bool {
	if !c.Bounds.Contains(v.Bounds.TL) {
		return false
	}
	for _, o := range c.Vehicles {
		if v == o {
			return false
		}
	}
	c.Vehicles = append(c.Vehicles, v)
	return true
}

// RemoveVehicle removes the vehicle from this chunk's list of managed vehicles.
func (c *Chunk) RemoveVehicle(v *Vehicle) bool {
	if !c.Bounds.Contains(v.Bounds.TL) {
		return false
	}
	idx := -1
	for n, o := range c.Vehicles {
		if o == v {
			idx = n
			break
		}
	}
	if idx < 0 {
		return false
	}
	// Remove from slice while maintaining order
	copy(c.Vehicles[idx:], c.Vehicles[idx+1:])
	c.Vehicles[len(c.Vehicles)-1] = nil
	c.Vehicles = c.Vehicles[:len(c.Vehicles)-1]
	return true
}

// PlaceActorRelative adds the actor to the chunk and adjusts the position from
// chunk-relative to absolute. This function always allows standing on climbable
// locations.
func (c *Chunk) PlaceActorRelative(a *Actor) {
	a.Position.X += c.Bounds.TL.X
	a.Position.Y += c.Bounds.TL.Y
	c.PlaceActor(a, true)
}

// PlaceActor places the actor within the chunk. This is a no-op if the
// current position lies outside the chunk. If the climbing parameter is true
// the locations that allow climbing are also considered. The first return value
// is true if the actor can step on the location. The second return value is
// true only if the first return value is false and the location allowed
// climbing.
func (c *Chunk) PlaceActor(a *Actor, climbing bool) (bool, bool) {
	cw, cc := c.CanStep(a, a.Position)
	if !cw {
		if cc && climbing {
			c.Actors = append(c.Actors, a)
			c.BlocksWalk.Set(c.relOfs(a.Position))
			return false, true
		}
		return false, false
	}
	c.Actors = append(c.Actors, a)
	c.BlocksWalk.Set(c.relOfs(a.Position))
	c.BlocksClimb.Set(c.relOfs(a.Position))
	return true, false
}

// RemoveActor removes the Actor from the chunk. This is a no-op if the
// current position lies outside the chunk.
func (c *Chunk) RemoveActor(a *Actor) {
	if !c.Bounds.Contains(a.Position) {
		return
	}
	idx := -1
	for n, actor := range c.Actors {
		if actor == a {
			idx = n
			break
		}
	}
	if idx < 0 {
		return
	}
	// Remove from slice, order is not important
	c.Actors[idx] = c.Actors[len(c.Actors)-1]
	c.Actors[len(c.Actors)-1] = nil
	c.Actors = c.Actors[:len(c.Actors)-1]
	c.bitmapsDirty = true
}

// CanStep returns true if the location is valid for an actor to step. The
// second return value is true only if the first return value is false and the
// location allows climbing.
func (c *Chunk) CanStep(a *Actor, p util.Point) (bool, bool) {
	if !c.Bounds.Contains(p) {
		return false, false
	}
	if c.bitmapsDirty {
		c.RebuildBitmaps()
	}
	if c.BlocksWalk.Contains(c.relOfs(p)) {
		return false, !c.BlocksClimb.Contains(c.relOfs(p))
	}
	for _, a := range c.Actors {
		if a.Position == p {
			return false, false
		}
	}
	return true, false
}
