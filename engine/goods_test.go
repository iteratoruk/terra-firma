package engine

import (
	"fmt"
	"reflect"
	"testing"
)

// Inert goods are one of three modes of being (DESIGN.md taxonomy): a thing
// with a location but no autonomous tick-dynamics. These tests pin the issue
// #1 acceptance contract: a free log appears in the snapshot at its tile, and
// Tick() does nothing to it.

func TestInertGoodsDoNotChangeOnTheirOwnTick(t *testing.T) {
	// Inertness is the defining property of this mode of being. After N ticks
	// with no commands, the goods collection must be identical: same count,
	// same kinds, same positions, nothing new elsewhere.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "forest", Capacity: 100},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	before := w.Snapshot().Goods
	for i := 0; i < 10; i++ {
		w.Tick()
	}
	after := w.Snapshot().Goods

	if !reflect.DeepEqual(before, after) {
		t.Fatalf("goods changed across 10 idle ticks:\n before=%+v\n after =%+v", before, after)
	}
	if len(after) != 1 || after[0].Kind != "log" || after[0].Q != 0 || after[0].R != 0 {
		t.Fatalf("expected exactly 1 log at (0,0), got %+v", after)
	}
}

func TestCarrierPicksUpLogAtItsCurrentTile(t *testing.T) {
	// Scenario 1: a carrier commanded to pick up a co-located log holds it
	// after the next tick, and no free log remains at the tile. This forces
	// the holder relationship into existence and surfaces it on the snapshot.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0)},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	w.Apply(PickUp{Carrier: NewHex(0, 0), Good: NewHex(0, 0)})
	w.Tick()

	snap := w.Snapshot()
	if len(snap.Goods) != 1 {
		t.Fatalf("expected exactly 1 log in snapshot, got %d (%+v)", len(snap.Goods), snap.Goods)
	}
	g := snap.Goods[0]
	if !g.Held {
		t.Errorf("log should be held after pickup, got Held=false (%+v)", g)
	}
	if g.Q != 0 || g.R != 0 {
		t.Errorf("held log should be at (0,0), got (%d,%d)", g.Q, g.R)
	}
	for _, gs := range snap.Goods {
		if !gs.Held && gs.Q == 0 && gs.R == 0 {
			t.Errorf("no free log should remain at (0,0), found %+v", gs)
		}
	}
}

func TestHeldGoodFollowsCarrierAndCoLocatedFreeGoodStaysPut(t *testing.T) {
	// Scenario 3, with the discriminator from the #3 hand-off notes baked in:
	// two logs at (0,0), one becomes held, one stays free. After the carrier
	// moves to (1,0), the held log is at (1,0) (its location *derived* from
	// the carrier) while the free log remains at (0,0). This is the test that
	// rules out a "naive" implementation where Held is a label but position
	// is still read from the good — it forces the snapshot to derive.
	dest := NewHex(1, 0)
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0), Destination: &dest},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	w.Apply(PickUp{Carrier: NewHex(0, 0), Good: NewHex(0, 0)})
	w.Tick()

	snap := w.Snapshot()
	if c := snap.Carriers[0]; c.Q != 1 || c.R != 0 {
		t.Fatalf("carrier should be at (1,0), got (%d,%d)", c.Q, c.R)
	}
	if len(snap.Goods) != 2 {
		t.Fatalf("expected 2 logs in snapshot, got %d (%+v)", len(snap.Goods), snap.Goods)
	}
	var held, free *GoodSnapshot
	for i := range snap.Goods {
		g := &snap.Goods[i]
		if g.Held {
			held = g
		} else {
			free = g
		}
	}
	if held == nil || free == nil {
		t.Fatalf("expected one held and one free log, got %+v", snap.Goods)
	}
	if held.Q != 1 || held.R != 0 {
		t.Errorf("held log should be at carrier's (1,0), got (%d,%d)", held.Q, held.R)
	}
	if free.Q != 0 || free.R != 0 {
		t.Errorf("free log should still be at (0,0), got (%d,%d)", free.Q, free.R)
	}
}

func TestPickUpIsRejectedWhenCarrierAndGoodAreOnDifferentTiles(t *testing.T) {
	// Scenario 2: pickup is a labour-mediated act — it can only happen where
	// the labour is. A command issued for a carrier and good on different
	// hexes must NOT link them; the good stays free at its tile. Validated by
	// transient mutation (remove the same-hex check, confirm the test bites).
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0)},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(1, 0)},
		},
	})

	w.Apply(PickUp{Carrier: NewHex(0, 0), Good: NewHex(1, 0)})
	w.Tick()

	snap := w.Snapshot()
	if len(snap.Goods) != 1 {
		t.Fatalf("expected 1 log in snapshot, got %d", len(snap.Goods))
	}
	g := snap.Goods[0]
	if g.Held {
		t.Errorf("log should NOT be held: carrier and log were on different tiles, got %+v", g)
	}
	if g.Q != 1 || g.R != 0 {
		t.Errorf("free log should still be at (1,0), got (%d,%d)", g.Q, g.R)
	}
}

func TestPickUpMoveDropPreservesLogCount(t *testing.T) {
	// Issue #4 Scenario 1: a pickup-move-drop preserves the total log count.
	// At every observed tick (and around each command boundary) free logs +
	// held logs == 1. The discriminating assertion that *forces* Drop's body
	// is the post-drop check that the log is free at the carrier's drop hex —
	// a stub Drop satisfies the count but not the held->free transition.
	dest := NewHex(3, 0)
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(2, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(3, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0), Destination: &dest},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	assertLogTotalIs(t, w.Snapshot(), 1, "initial")

	w.Apply(PickUp{Carrier: NewHex(0, 0), Good: NewHex(0, 0)})
	assertLogTotalIs(t, w.Snapshot(), 1, "after pickup")

	for i := 0; i < 3; i++ {
		w.Tick()
		assertLogTotalIs(t, w.Snapshot(), 1, fmt.Sprintf("after tick %d", i+1))
	}

	w.Apply(Drop{Carrier: NewHex(3, 0)})
	assertLogTotalIs(t, w.Snapshot(), 1, "after drop")

	snap := w.Snapshot()
	g := snap.Goods[0]
	if g.Held {
		t.Errorf("log should be free after drop, got Held=true (%+v)", g)
	}
	if g.Q != 3 || g.R != 0 {
		t.Errorf("dropped log should be at carrier's hex (3,0), got (%d,%d)", g.Q, g.R)
	}
}

// assertLogTotalIs counts free + held logs in the snapshot and fails if the
// sum differs from want. The conservation law (issue #4) cares about the sum,
// not the breakdown, so the assertion ignores Held.
func assertLogTotalIs(t *testing.T, snap Snapshot, want int, when string) {
	t.Helper()
	n := 0
	for _, g := range snap.Goods {
		if g.Kind == "log" {
			n++
		}
	}
	if n != want {
		t.Errorf("%s: free+held logs = %d, want %d (goods=%+v)", when, n, want, snap.Goods)
	}
}

func TestConservationHoldsAcrossArbitraryCommands(t *testing.T) {
	// Issue #4 Scenario Outline: free + held logs == K at every tick under any
	// sequence of pickup and drop commands, with no production or consumption.
	// (Per #4 discussion: "move" here is the carrier's autonomous tick-step
	// under a construction-time destination, not a command — so the generator
	// emits only PickUp and Drop.)
	//
	// The generator pulls from w.rng so failures are reproducible from the
	// seed alone — and so the world's evolution now genuinely depends on the
	// seed, which the divergence assertion in world_test.go relies on.
	//
	// Validated by transient mutation per MEMORY: in turn, neuter Drop's
	// holder-clear, then PickUp's holder-set-only (re-add the good as a free
	// duplicate), then confirm each makes the property bite with a count-off
	// message naming free+held.
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
			w := buildConservationWorld(17, tc.K, tc.C)
			assertLogTotalIs(t, w.Snapshot(), tc.K, "initial")
			for step := 0; step < 200; step++ {
				applyRandomCommand(w)
				assertLogTotalIs(t, w.Snapshot(), tc.K, fmt.Sprintf("step %d after command", step))
				w.Tick()
				assertLogTotalIs(t, w.Snapshot(), tc.K, fmt.Sprintf("step %d after tick", step))
				if t.Failed() {
					return // stop on first violation; the count alone is enough to diagnose
				}
			}
		})
	}
}

// buildConservationWorld lays out a small hex region with K logs and C
// carriers in deterministic positions, each carrier with a construction-time
// destination so the random-command harness has something to transport.
// Layout is independent of the seed; randomness enters only via the
// w.rng-driven command generator, so divergence across seeds is purely a
// function of the commands chosen.
func buildConservationWorld(seed int64, K, C int) *World {
	tiles := []TileSpec{}
	for q := -3; q <= 3; q++ {
		for r := -3; r <= 3; r++ {
			tiles = append(tiles, TileSpec{Hex: NewHex(q, r), Resource: "soil", Capacity: 10})
		}
	}
	goods := make([]GoodSpec, K)
	for i := 0; i < K; i++ {
		goods[i] = GoodSpec{Kind: "log", Hex: NewHex(i-K/2, 0)}
	}
	carriers := make([]CarrierSpec, C)
	for i := 0; i < C; i++ {
		dest := NewHex(i-C/2, 1) // away from log row so carriers walk through pickup hexes
		carriers[i] = CarrierSpec{
			Type:        "porter",
			Hex:         NewHex(i-C/2, -1),
			Destination: &dest,
		}
	}
	return NewWorld(Config{Seed: seed, Tiles: tiles, Goods: goods, Carriers: carriers})
}

// applyRandomCommand draws from w.rng to issue one PickUp or Drop. Reads
// state via Snapshot() (the boundary applies to test code too) but consults
// w.rng directly because in-package tests are allowed to and the harness has
// to use the world's seeded RNG — see the divergence assertion in
// world_test.go that depends on it.
func applyRandomCommand(w *World) {
	snap := w.Snapshot()
	if len(snap.Carriers) == 0 {
		return
	}
	c := snap.Carriers[w.rng.IntN(len(snap.Carriers))]
	cHex := NewHex(c.Q, c.R)
	switch w.rng.IntN(3) {
	case 0:
		// PickUp at the carrier's own hex — succeeds iff a free good is
		// co-located. The bias toward same-hex attempts is what produces a
		// meaningful number of *successful* pickups in a 5×5 area.
		w.Apply(PickUp{Carrier: cHex, Good: cHex})
	case 1:
		// PickUp targeting a random known good's hex — usually fizzles
		// (carrier and good on different tiles) but exercises the rejection
		// path under the property.
		if len(snap.Goods) > 0 {
			g := snap.Goods[w.rng.IntN(len(snap.Goods))]
			w.Apply(PickUp{Carrier: cHex, Good: NewHex(g.Q, g.R)})
		}
	case 2:
		w.Apply(Drop{Carrier: cHex})
	}
}

func TestSnapshotListsFreeLogAtItsTile(t *testing.T) {
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "forest", Capacity: 100},
			{Hex: NewHex(1, 0), Resource: "forest", Capacity: 100},
		},
		Goods: []GoodSpec{
			{Kind: "log", Hex: NewHex(0, 0)},
		},
	})

	snap := w.Snapshot()

	logs := 0
	for _, g := range snap.Goods {
		if g.Kind == "log" {
			logs++
		}
	}
	if logs != 1 {
		t.Fatalf("expected exactly 1 log in snapshot, got %d (goods=%+v)", logs, snap.Goods)
	}
	if snap.Goods[0].Q != 0 || snap.Goods[0].R != 0 {
		t.Errorf("log not at (0,0): got (%d,%d)", snap.Goods[0].Q, snap.Goods[0].R)
	}
}
