package engine

import "sort"

// Hex is an axial coordinate (Q, R). The third cube coordinate S is derived as
// -Q-R and never stored, which keeps the cube constraint Q+R+S=0 true by
// construction. All spatial reasoning in the engine uses these; offset
// coordinates are a renderer concern and must not appear in engine logic.
type Hex struct {
	Q int
	R int
}

// NewHex constructs an axial coordinate.
func NewHex(q, r int) Hex { return Hex{Q: q, R: r} }

// S returns the derived third cube coordinate.
func (h Hex) S() int { return -h.Q - h.R }

// hexDirections are the six unit steps in cube space. Uniform adjacency: every
// neighbour is exactly one of these vectors away, all at distance 1.
var hexDirections = [6]Hex{
	{Q: 1, R: 0}, {Q: 1, R: -1}, {Q: 0, R: -1},
	{Q: -1, R: 0}, {Q: -1, R: 1}, {Q: 0, R: 1},
}

// Add returns the vector sum of two hexes.
func (h Hex) Add(o Hex) Hex { return Hex{Q: h.Q + o.Q, R: h.R + o.R} }

// Neighbours returns the six adjacent hexes in a fixed, deterministic order.
func (h Hex) Neighbours() []Hex {
	out := make([]Hex, 6)
	for i, d := range hexDirections {
		out[i] = h.Add(d)
	}
	return out
}

// Distance is the hex (cube) distance: half the L1 norm of the cube-coordinate
// difference, which reduces to the standard single-formula hex distance.
func (h Hex) Distance(o Hex) int {
	dq := abs(h.Q - o.Q)
	dr := abs(h.R - o.R)
	ds := abs(h.S() - o.S())
	return (dq + dr + ds) / 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// LessHex is a total order on hexes (by Q then R) used to make any iteration
// over a set of hexes deterministic, satisfying the no-order-dependent-state
// invariant.
func LessHex(a, b Hex) bool {
	if a.Q != b.Q {
		return a.Q < b.Q
	}
	return a.R < b.R
}

// SortHexes sorts a slice into the canonical order in place.
func SortHexes(hs []Hex) {
	sort.Slice(hs, func(i, j int) bool { return LessHex(hs[i], hs[j]) })
}
