package engine

import "testing"

// A Stock is a first-class object that knows its own deltas. CLAUDE.md: "Stocks
// know their own deltas ... not a bare integer differenced after the fact." These
// tests pin: net = regen - harvest, the floor at zero, the cap, and that the
// rates are observable for the legibility requirement.

func TestStockNetIsRegenMinusHarvest(t *testing.T) {
	s := NewStock(50, 100) // value 50, capacity 100
	s.SetRegen(5)
	s.SetHarvest(3)
	if got := s.Net(); got != 2 {
		t.Errorf("Net()=%d, want 2 (regen 5 - harvest 3)", got)
	}
}

func TestStockStepAppliesNet(t *testing.T) {
	s := NewStock(50, 100)
	s.SetRegen(5)
	s.SetHarvest(3)
	s.Step()
	if s.Value() != 52 {
		t.Errorf("after Step value=%d, want 52", s.Value())
	}
	// Rates persist across a step so the snapshot can report a trend; they are
	// only changed by explicit Set calls (e.g. a harvester attaching/detaching).
	if s.Regen() != 5 || s.Harvest() != 3 {
		t.Errorf("rates should persist across Step, got regen=%d harvest=%d", s.Regen(), s.Harvest())
	}
}

func TestStockCannotGoBelowZero(t *testing.T) {
	s := NewStock(2, 100)
	s.SetHarvest(5) // harvesting faster than it has
	s.Step()
	if s.Value() != 0 {
		t.Errorf("stock floored at 0, got %d", s.Value())
	}
	// Depletion to exactly zero is the local, immediate consequence the design
	// relies on (a farm whose soil hits zero dies). No negative stocks.
	s.Step()
	if s.Value() != 0 {
		t.Errorf("stock stays at 0, got %d", s.Value())
	}
}

func TestStockCannotExceedCapacity(t *testing.T) {
	s := NewStock(98, 100)
	s.SetRegen(5)
	s.Step()
	if s.Value() != 100 {
		t.Errorf("stock capped at capacity 100, got %d", s.Value())
	}
}

func TestStockDepletionIsDeterministic(t *testing.T) {
	// Two identically-configured stocks stepped the same number of times must
	// end identical — the determinism invariant at the smallest scale.
	mk := func() *Stock {
		s := NewStock(40, 100)
		s.SetRegen(2)
		s.SetHarvest(7)
		return s
	}
	a, b := mk(), mk()
	for i := 0; i < 20; i++ {
		a.Step()
		b.Step()
	}
	if a.Value() != b.Value() {
		t.Errorf("determinism broken: %d vs %d", a.Value(), b.Value())
	}
}
