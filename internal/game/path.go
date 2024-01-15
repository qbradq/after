package game

import (
	"container/heap"
	"math"

	"github.com/qbradq/after/lib/util"
	"golang.org/x/tools/container/intsets"
)

const (
	MaxPath      int     = 64                                // Maximum number of steps along a path, we keep this short to short-circuit broad explorations
	MaxPathNodes int     = (MaxPath*2 + 1) * (MaxPath*2 + 1) // Maximum number of path nodes that could possibly be in memory at one time
	pnHScale     float64 = 1 + (1 / float64(MaxPath))        // Tie breaker scale amount as described in the "breaking ties" section of the Heuristics page of the article
)

var (
	npOpen       pathNodeQueue    = make(pathNodeQueue, 0, MaxPathNodes) // pathNodeQueue cache for NewPath's open queue
	npPool       []*pathNode      = make([]*pathNode, MaxPathNodes)      // Pool of pre-allocated path nodes
	npReturnPool []*pathNode      = make([]*pathNode, 0, MaxPathNodes)   // Pool of pre-allocated path nodes that need to be returned to npPool
	npRetBuf     []util.Direction = make([]util.Direction, 0, MaxPath)   // Buffer for the reverse-order return path
)

// Point offsets for neighbor checks
var aMapPointOfs = []struct {
	P util.Point
	D util.Direction
}{
	{P: util.Point{X: 0, Y: -1}, D: util.DirectionNorth},
	{P: util.Point{X: 1, Y: -1}, D: util.DirectionNorthEast},
	{P: util.Point{X: 1, Y: 0}, D: util.DirectionEast},
	{P: util.Point{X: 1, Y: 1}, D: util.DirectionSouthEast},
	{P: util.Point{X: 0, Y: 1}, D: util.DirectionSouth},
	{P: util.Point{X: -1, Y: 1}, D: util.DirectionSouthWest},
	{P: util.Point{X: -1, Y: 0}, D: util.DirectionWest},
	{P: util.Point{X: -1, Y: -1}, D: util.DirectionNorthWest},
}

func init() {
	for i := range npPool {
		npPool[i] = &pathNode{}
	}
}

// pathNode represents a node along the search path.
type pathNode struct {
	p *pathNode      // Pointer to the previous pathNode, or nil if root
	g float64        // Number of steps from the start notated as g(n) in the article
	h float64        // Heuristic distance to end notated as h(n) in the article
	f float64        // Cached sum of g and h as described in the article
	l util.Point     // Location of the path node
	d util.Direction // Direction of travel from parent node to this node
	i int            // Priority queue index
	s int            // Cached hash value of l
}

// pathNodeQueue implements a priority queue over pathNode elements.
type pathNodeQueue []*pathNode

func (q pathNodeQueue) Len() int { return len(q) }

func (q pathNodeQueue) Less(i, j int) bool {
	return q[i].f < q[j].f
}

func (q pathNodeQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].i = i
	q[j].i = j
}

func (q *pathNodeQueue) Push(x any) {
	n := len(*q)
	item := x.(*pathNode)
	item.i = n
	*q = append(*q, item)
}

func (q *pathNodeQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.i = -1
	*q = old[0 : n-1]
	return item
}

// Path represents a shortest path between two points on the map. The A/A*
type Path []util.Direction

// NewPath calculates a shortest path from p1 to p2 considering the current
// walk-ability of each location with regard to m. The path is appended to ret.
// The A/A* path-finding algorithm used here is presented in the following
// article: https://theory.stanford.edu/~amitp/GameProgramming/AStarComparison.html
func NewPath(p1, p2 util.Point, m *CityMap, ret *Path) {
	// Returns a pathNode pointer from the pool that is pre-calculated for the
	// given point on m.
	fn := func(p util.Point, parent *pathNode, d util.Direction) *pathNode {
		n := npPool[len(npPool)-1]
		npPool = npPool[:len(npPool)-1]
		npReturnPool = append(npReturnPool, n)
		n.p = parent
		n.l = p
		n.d = d
		n.g = 0
		if parent != nil {
			n.g = parent.g + 1
		}
		dx := math.Abs(float64(p.X - p2.X))
		dy := math.Abs(float64(p.Y - p2.Y))
		n.h = math.Max(dx, dy) * math.Min(dx, dy) * pnHScale
		n.f = n.g + n.h
		n.s = (p.X & 0x0000FFFF) | ((p.Y & 0x0000FFFF) << 16)
		return n
	}
	// Setup
	var n *pathNode
	var open, closed intsets.Sparse
	npOpen = npOpen[:0]
	npReturnPool = npReturnPool[:0]
	root := fn(p1, nil, util.DirectionInvalid)
	heap.Push(&npOpen, root)
	open.Insert(root.s)
	// Process nodes starting with the root
	for {
		// Out of nodes to consider and we never reached the goal so there was
		// no path found, so return without appending anything to ret
		if len(npOpen) == 0 {
			n = nil
			break
		}
		// Goal reached, exit loop and begin result portion
		n = heap.Pop(&npOpen).(*pathNode)
		if n.l == p2 {
			break
		}
		closed.Insert(n.s)
		// Next step would exceed maximum path length, do not consider it
		if int(n.g) >= MaxPath {
			continue
		}
		// Process all neighbors
		for _, ofs := range aMapPointOfs {
			// Skip tiles that are not loaded
			p := n.l.Add(ofs.P)
			if !m.loadBounds.Contains(p) {
				continue
			}
			// Skip previously-considered or to-be-considered tiles
			s := (p.X & 0x0000FFFF) | ((p.Y & 0x0000FFFF) << 16)
			if open.Has(s) || closed.Has(s) {
				continue
			}
			// Do not consider non-walkable tiles
			c := m.GetChunk(p)
			if c.BlocksWalk.Contains(c.relOfs(p)) {
				continue
			}
			// Construct neighbor node and insert into open set and queue
			nn := fn(p, n, ofs.D)
			open.Insert(nn.s)
			heap.Push(&npOpen, nn)
		}
	}
	// Construct result
	if n != nil {
		npRetBuf = npRetBuf[:0]
		for {
			// At root node, no more steps to take
			if n.d == util.DirectionInvalid {
				break
			}
			// Append the step and go to the parent
			npRetBuf = append(npRetBuf, n.d)
			n = n.p
		}
		// Append the path to ret in the correct order
		for i := len(npRetBuf) - 1; i >= 0; i-- {
			*ret = append(*ret, npRetBuf[i])
		}
	}
	// Return nodes to the pool
	npPool = append(npPool, npReturnPool...)
}
