package engine

import (
	"fmt"
	"testing"
)

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

func TestGrainConservationUnderArbitraryCommandsAndEating(t *testing.T) {
	// Scenario 6 — conservation generalises. Grain that *leaves* the world by
	// being eaten is accounted for in w.consumed, so the invariant becomes
	// free + held + consumed == K at every tick under any sequence of
	// pickup/drop commands plus autonomous eating.
	//
	// Validated by transient mutation per [[feedback_conservation_test_validation]]:
	//   - drop the consumed++ from populationEat → consumed stays 0 while
	//     grain disappears; assertion bites with free+held+0 < K.
	//   - drop the slice-remove from populationEat → grain is eaten AND
	//     remains; assertion bites with free+held+consumed > K.
	cases := []struct {
		name string
		K, C int
	}{
		{"K=1_C=1", 1, 1},
		{"K=3_C=1", 3, 1},
		{"K=5_C=3", 5, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := buildEatingConservationWorld(23, tc.K, tc.C)
			assertGrainConserved(t, w, tc.K, "initial")
			for step := 0; step < 200; step++ {
				applyRandomCommand(w)
				assertGrainConserved(t, w, tc.K, fmt.Sprintf("step %d after command", step))
				w.Tick()
				assertGrainConserved(t, w, tc.K, fmt.Sprintf("step %d after tick", step))
				if t.Failed() {
					return
				}
			}
		})
	}
}

// buildEatingConservationWorld is the eating analogue of
// buildConservationWorld: grain (edible) instead of logs, plus a population at
// the centre that will actually eat anything dropped on its hex. Carriers each
// get a construction-time destination so the random-command harness can move
// grain around the same way it moves logs in the #4 conservation test.
// Metabolism is 0 so the test isolates the eating-conservation question from
// the reserve-arithmetic one.
func buildEatingConservationWorld(seed int64, K, C int) *World {
	tiles := []TileSpec{}
	for q := -3; q <= 3; q++ {
		for r := -3; r <= 3; r++ {
			tiles = append(tiles, TileSpec{Hex: NewHex(q, r), Resource: "soil", Capacity: 10})
		}
	}
	goods := make([]GoodSpec, K)
	for i := 0; i < K; i++ {
		goods[i] = GoodSpec{Kind: "grain", Hex: NewHex(i-K/2, 0)}
	}
	carriers := make([]CarrierSpec, C)
	for i := 0; i < C; i++ {
		dest := NewHex(i-C/2, 1)
		carriers[i] = CarrierSpec{
			Type:        "porter",
			Hex:         NewHex(i-C/2, -1),
			Destination: &dest,
		}
	}
	pops := []PopulationSpec{
		// StarvationLimit high enough to outlast the 200-step random harness:
		// reserve starts at 0 with metabolism 0, so the streak ticks up every
		// step it isn't fed. This test pins eating-conservation, not death.
		{Hex: NewHex(0, 0), Reserve: 0, Metabolism: 0, StarvationLimit: 1000},
	}
	return NewWorld(Config{Seed: seed, Tiles: tiles, Goods: goods, Carriers: carriers, Populations: pops})
}

// assertGrainConserved checks free + held + consumed == want for grain.
func assertGrainConserved(t *testing.T, w *World, want int, when string) {
	t.Helper()
	free, held := 0, 0
	for _, g := range w.Snapshot().Goods {
		if g.Kind != "grain" {
			continue
		}
		if g.Held {
			held++
		} else {
			free++
		}
	}
	total := free + held + w.consumed
	if total != want {
		t.Errorf("%s: free(%d)+held(%d)+consumed(%d)=%d, want %d", when, free, held, w.consumed, total, want)
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
