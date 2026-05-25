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

func TestPopulationDoesNotEatGoodOnDifferentTile(t *testing.T) {
	// Scenario 2 — the co-location predicate. Grain is at (1,0); pop is at
	// (0,0). After 1 tick: the grain is still there, the population has only
	// metabolised (3 - 1 = 2). A naive "eat any free grain" implementation
	// breaks here.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1},
		},
		Goods: []GoodSpec{
			{Kind: "grain", Hex: NewHex(1, 0)},
		},
	})

	w.Tick()

	snap := w.Snapshot()
	found := false
	for _, g := range snap.Goods {
		if g.Kind == "grain" && !g.Held && g.Q == 1 && g.R == 0 {
			found = true
		}
	}
	if !found {
		t.Errorf("grain at (1,0) should be untouched, got %+v", snap.Goods)
	}
	if got := snap.Populations[0].Reserve; got != 2 {
		t.Errorf("reserve should be just metabolised (3 - 1 = 2), got %d", got)
	}
}

func TestNonEdibleGoodsAreNotConsumed(t *testing.T) {
	// Scenario 5 — the edibility predicate. A log has calorie value 0 in V1;
	// the population cannot eat it. After 1 tick: the log is still free at
	// (0,0) and the population has only metabolised (3 - 1 = 2). Discriminator
	// against an "eat any free good at my hex" implementation that ignores
	// calorieValue.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	w.Tick()

	snap := w.Snapshot()
	if len(snap.Goods) != 1 {
		t.Fatalf("expected exactly 1 log still free at (0,0), got %d (%+v)", len(snap.Goods), snap.Goods)
	}
	g := snap.Goods[0]
	if g.Kind != "log" || g.Held || g.Q != 0 || g.R != 0 {
		t.Errorf("log should still be free at (0,0), got %+v", g)
	}
	if got := snap.Populations[0].Reserve; got != 2 {
		t.Errorf("reserve should be just metabolised (3 - 1 = 2), got %d", got)
	}
}

func TestPopulationEatsAtMostOneGoodPerTick(t *testing.T) {
	// Scenario 4 — the one-per-tick rule. Two free grains at the same hex; the
	// population takes one and leaves the other for next tick. A naive "eat all
	// edibles at my hex" implementation breaks here (reserve would be 12, and
	// no grain would remain). Validated by transient mutation: replacing the
	// `return` in populationEat with `continue` makes this test bite.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1},
		},
		Goods: []GoodSpec{
			{Kind: "grain", Hex: NewHex(0, 0)},
			{Kind: "grain", Hex: NewHex(0, 0)},
		},
	})

	w.Tick()

	snap := w.Snapshot()
	n := 0
	for _, g := range snap.Goods {
		if g.Kind == "grain" && !g.Held && g.Q == 0 && g.R == 0 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("exactly one free grain should remain at (0,0), got %d (%+v)", n, snap.Goods)
	}
	if got := snap.Populations[0].Reserve; got != 7 {
		t.Errorf("reserve after eating exactly one grain then metabolising: want 7, got %d", got)
	}
}

func TestPopulationDoesNotEatHeldGood(t *testing.T) {
	// Scenario 3 — the free-good predicate. A grain held by a carrier is not
	// "on the tray". After 1 tick the carrier is still holding it and the
	// population has only metabolised (3 - 1 = 2). A naive "eat anything at my
	// hex" implementation that ignores the holder breaks here.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0)},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1},
		},
		Goods: []GoodSpec{
			{Kind: "grain", Hex: NewHex(0, 0)},
		},
	})

	w.Apply(PickUp{Carrier: NewHex(0, 0), Good: NewHex(0, 0)})
	w.Tick()

	snap := w.Snapshot()
	if len(snap.Goods) != 1 {
		t.Fatalf("expected exactly 1 grain still in world (held), got %d (%+v)", len(snap.Goods), snap.Goods)
	}
	g := snap.Goods[0]
	if !g.Held {
		t.Errorf("grain should still be held by carrier, got %+v", g)
	}
	if got := snap.Populations[0].Reserve; got != 2 {
		t.Errorf("reserve should be just metabolised (3 - 1 = 2), got %d", got)
	}
}
