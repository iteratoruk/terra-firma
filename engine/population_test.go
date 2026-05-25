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
