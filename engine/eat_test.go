package engine

import "testing"

// Eating — the first lifecycle hinge where a good leaves the world by being
// *used up* rather than just moved. These tests pin the issue #6 acceptance
// contract.
//
// Eating is autonomous, not commanded (DESIGN.md: a population's mode of being
// is metabolism). The player decides where food *is*; the population eats what
// is available, on its own.
//
// Tick order: stocks → carriers → populations eat → populations metabolise.
// The scenarios' arithmetic depends on eat-before-metabolise.

func TestPopulationEatsColocatedFreeGrainAndReserveRefills(t *testing.T) {
	// Scenario 1 — the discriminating test. Pop reserve 3, metabolism 1; a free
	// grain (calorie 5) at the same hex. After 1 tick: eat (+5)=8, then
	// metabolise (-1)=7. The grain has been destroyed (consumed, not moved).
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1},
		},
		Goods: []GoodSpec{
			{Kind: "grain", Hex: NewHex(0, 0)},
		},
	})

	w.Tick()

	snap := w.Snapshot()
	for _, g := range snap.Goods {
		if g.Kind == "grain" && !g.Held && g.Q == 0 && g.R == 0 {
			t.Errorf("free grain should have been consumed at (0,0), still present: %+v", g)
		}
	}
	if got := snap.Populations[0].Reserve; got != 7 {
		t.Errorf("reserve after eat(+5) then metabolise(-1): want 7, got %d", got)
	}
}
