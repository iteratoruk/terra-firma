package engine

import (
	"sort"
	"testing"
)

// Hex coordinates are axial (q, r) with implicit s = -q - r, i.e. cube coords
// constrained to q+r+s = 0. These tests pin the invariants CLAUDE.md requires:
// uniform six-neighbour adjacency and a single clean distance formula. Offset
// coordinates must never appear here — that is a display concern only.

func TestHexCubeConstraint(t *testing.T) {
	// s is derived, never stored; q+r+s must always be zero.
	cases := []Hex{
		NewHex(0, 0),
		NewHex(1, -1),
		NewHex(-2, 3),
		NewHex(5, 5),
	}
	for _, h := range cases {
		if got := h.Q + h.R + h.S(); got != 0 {
			t.Errorf("Hex%v violates cube constraint q+r+s=0: got %d", h, got)
		}
	}
}

func TestHexNeighboursAreSixAndUniform(t *testing.T) {
	h := NewHex(2, -1)
	ns := h.Neighbours()
	if len(ns) != 6 {
		t.Fatalf("expected exactly 6 neighbours, got %d", len(ns))
	}
	// Every neighbour must be at distance 1 — that is what "uniform adjacency"
	// means and it is the whole reason for choosing hexes.
	seen := map[Hex]bool{}
	for _, n := range ns {
		if d := h.Distance(n); d != 1 {
			t.Errorf("neighbour %v is at distance %d, want 1", n, d)
		}
		if seen[n] {
			t.Errorf("duplicate neighbour %v", n)
		}
		seen[n] = true
	}
}

func TestHexNeighboursAreReciprocal(t *testing.T) {
	// If b is a neighbour of a, then a must be a neighbour of b. Adjacency is
	// symmetric; a bug in the direction vectors would break this.
	a := NewHex(0, 0)
	for _, b := range a.Neighbours() {
		found := false
		for _, back := range b.Neighbours() {
			if back == a {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("adjacency not reciprocal: %v has neighbour %v but not vice versa", a, b)
		}
	}
}

func TestHexDistance(t *testing.T) {
	cases := []struct {
		a, b Hex
		want int
	}{
		{NewHex(0, 0), NewHex(0, 0), 0},
		{NewHex(0, 0), NewHex(1, 0), 1},
		{NewHex(0, 0), NewHex(-1, 0), 1},
		{NewHex(0, 0), NewHex(2, -1), 2},
		{NewHex(0, 0), NewHex(-3, 1), 3},
		{NewHex(1, -2), NewHex(-2, 2), 4}, // symmetric, non-origin
	}
	for _, c := range cases {
		if got := c.a.Distance(c.b); got != c.want {
			t.Errorf("Distance(%v,%v)=%d, want %d", c.a, c.b, got, c.want)
		}
		// Distance must be symmetric.
		if got := c.b.Distance(c.a); got != c.want {
			t.Errorf("Distance not symmetric: Distance(%v,%v)=%d, want %d", c.b, c.a, got, c.want)
		}
	}
}

// Hex must be usable as a deterministic map key with a stable sort order, because
// CLAUDE.md forbids order-dependent map iteration affecting state. We provide a
// canonical sort so any iteration over a set of hexes can be made deterministic.
func TestHexCanonicalSortIsDeterministic(t *testing.T) {
	hexes := []Hex{NewHex(1, 0), NewHex(0, 1), NewHex(-1, 0), NewHex(0, 0), NewHex(1, -1)}
	a := append([]Hex(nil), hexes...)
	b := append([]Hex(nil), hexes...)
	// Shuffle b into a different starting order.
	b[0], b[4] = b[4], b[0]

	SortHexes(a)
	SortHexes(b)

	if !sort.SliceIsSorted(a, func(i, j int) bool { return LessHex(a[i], a[j]) }) {
		t.Fatal("SortHexes did not produce sorted output")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("sort not deterministic across input orderings at %d: %v vs %v", i, a[i], b[i])
		}
	}
}
