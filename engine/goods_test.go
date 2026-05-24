package engine

import "testing"

// Inert goods are one of three modes of being (DESIGN.md taxonomy): a thing
// with a location but no autonomous tick-dynamics. These tests pin the issue
// #1 acceptance contract: a free log appears in the snapshot at its tile, and
// Tick() does nothing to it.

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
