package engine

import (
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
