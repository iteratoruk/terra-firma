package engine

import "testing"

// Population is an actor in the DESIGN.md taxonomy: a self-locating thing with
// autonomous tick-dynamics. The dynamic introduced by issue #5 is biological
// rather than locomotive — the subsistence reserve depletes each tick under its
// own rule regardless of player input. These tests pin the issue #5 acceptance
// contract; refill (#6) and death (#7) are out of scope.

func TestPopulationReserveDecreasesByMetabolismEachTick(t *testing.T) {
	// Scenario 1: the discriminating test. The reserve must fall by metabolism
	// every tick, with no external trigger. If the tick fails to draw it down,
	// the whole "deadline, not a free ride" premise is decoration.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 10, Metabolism: 1},
		},
	})

	w.Tick()
	if got := w.Snapshot().Populations[0].Reserve; got != 9 {
		t.Fatalf("after 1 tick: want reserve 9, got %d", got)
	}

	for i := 0; i < 4; i++ {
		w.Tick()
	}
	if got := w.Snapshot().Populations[0].Reserve; got != 5 {
		t.Fatalf("after 5 ticks: want reserve 5, got %d", got)
	}
}

func TestPopulationReserveFloorsAtZero(t *testing.T) {
	// Scenario 2: the reserve must not go negative — death (#7) belongs to a
	// later slice, and "subsistence" stops meaning anything once the bar wraps
	// around. Run well past the point where an unfloored draw would go
	// negative. Validated by transient mutation (removing the floor guard
	// makes this test bite with a deeply negative reserve) — see MEMORY:
	// conservation test validation.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 10, Metabolism: 1},
		},
	})

	for i := 0; i < 100; i++ {
		w.Tick()
	}

	if got := w.Snapshot().Populations[0].Reserve; got != 0 {
		t.Errorf("after 100 ticks: want reserve 0 (floored), got %d", got)
	}
}

func TestHeavierMetabolismDepletesFaster(t *testing.T) {
	// Scenario 3: pins metabolism as a *per-population* rate, not a global
	// constant. Two populations on different tiles with the same starting
	// reserve diverge over 5 ticks because their metabolisms differ. If the
	// tick draws every population down by the same hard-coded amount, this
	// test bites.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 10, Metabolism: 1},
			{Hex: NewHex(1, 0), Reserve: 10, Metabolism: 2},
		},
	})

	for i := 0; i < 5; i++ {
		w.Tick()
	}

	got := populationsByHex(w.Snapshot())
	if got[NewHex(0, 0)].Reserve != 5 {
		t.Errorf("(0,0) metabolism 1 after 5 ticks: want reserve 5, got %d", got[NewHex(0, 0)].Reserve)
	}
	if got[NewHex(1, 0)].Reserve != 0 {
		t.Errorf("(1,0) metabolism 2 after 5 ticks: want reserve 0, got %d", got[NewHex(1, 0)].Reserve)
	}
}

func TestSnapshotExposesPopulationLocationReserveAndMetabolism(t *testing.T) {
	// Scenario 4: legibility requirement. The snapshot is the only window in,
	// so location, reserve, and metabolism must be visible there or they don't
	// exist for any observer. Exactly one population is exposed, at (0,0),
	// with the constructor's reserve and metabolism faithfully reported.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 10, Metabolism: 1},
		},
	})

	ps := w.Snapshot().Populations
	if len(ps) != 1 {
		t.Fatalf("want exactly 1 population in snapshot, got %d", len(ps))
	}
	p := ps[0]
	if p.Q != 0 || p.R != 0 {
		t.Errorf("position wrong: got (%d,%d), want (0,0)", p.Q, p.R)
	}
	if p.Reserve != 10 {
		t.Errorf("reserve wrong: got %d, want 10", p.Reserve)
	}
	if p.Metabolism != 1 {
		t.Errorf("metabolism wrong: got %d, want 1", p.Metabolism)
	}
}

func populationsByHex(s Snapshot) map[Hex]PopulationSnapshot {
	out := make(map[Hex]PopulationSnapshot, len(s.Populations))
	for _, p := range s.Populations {
		out[NewHex(p.Q, p.R)] = p
	}
	return out
}
