package engine

// Stock is a quantity of a single resource that knows its own rates of change.
// It is deliberately NOT a bare integer: the regen/harvest/net rates are part of
// the object so the snapshot can report a trend ("falling 3/tick"), which the
// design treats as a legibility requirement, not an afterthought.
//
// A Stock lives somewhere — on a tile (tile-bound resource) or carried as an
// inert good — but the Stock type itself is agnostic about its bearer. Bearer
// semantics belong to whatever holds the Stock.
type Stock struct {
	value    int
	capacity int
	regen    int // units added per Step (natural regeneration)
	harvest  int // units removed per Step (extraction pressure)
}

// NewStock creates a stock with an initial value and a capacity ceiling.
func NewStock(value, capacity int) *Stock {
	if value > capacity {
		value = capacity
	}
	if value < 0 {
		value = 0
	}
	return &Stock{value: value, capacity: capacity}
}

func (s *Stock) Value() int    { return s.value }
func (s *Stock) Capacity() int { return s.capacity }
func (s *Stock) Regen() int    { return s.regen }
func (s *Stock) Harvest() int  { return s.harvest }

// Net is the change that will be applied on the next Step. Exposed so callers
// (and the snapshot) can read the trend without stepping.
func (s *Stock) Net() int { return s.regen - s.harvest }

// SetRegen / SetHarvest adjust the rates. These are how the world expresses
// "this tile regenerates at R" and "a harvester is drawing H from this stock".
// Rates persist until changed, so a stock keeps regenerating/ depleting tick
// over tick until something detaches the pressure.
func (s *Stock) SetRegen(r int)   { s.regen = clampNonNeg(r) }
func (s *Stock) SetHarvest(h int) { s.harvest = clampNonNeg(h) }

// Step applies one tick of net change, clamped to [0, capacity]. The floor at
// zero is load-bearing: a stock hitting zero is the local, immediate
// consequence the design depends on (exhausted soil, a felled-out forest).
func (s *Stock) Step() {
	s.value += s.Net()
	if s.value < 0 {
		s.value = 0
	}
	if s.value > s.capacity {
		s.value = s.capacity
	}
}

func clampNonNeg(x int) int {
	if x < 0 {
		return 0
	}
	return x
}
