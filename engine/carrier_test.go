package engine

import "testing"

// Carriers are actors (DESIGN.md taxonomy): self-locating, with autonomous
// tick-dynamics (they move on their own each tick once given a destination).
// These tests pin the issue #2 acceptance contract.
//
// A keystone design point (see MEMORY: tech-modifiable rates are emergent):
// speed is NOT a stored field on the carrier — it is computed by a rule from
// the carrier's type and the tile's transit condition. V1 implements the
// single combination "porter × unimproved → 1 tile/tick"; future tech extends
// the rule's domain, not a number on the carrier.

func TestSnapshotExposesCarrierTypeAndTileTransitCondition(t *testing.T) {
	// Scenario 3: the snapshot exposes carrier type and tile transit condition.
	// Both are first-class stored properties with their V1 vocabulary
	// ("porter", "unimproved"). The snapshot is the only window in, so these
	// have to be visible there or they don't exist for any observer.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "forest", Capacity: 100, TransitCondition: "unimproved"},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0)},
		},
	})

	snap := w.Snapshot()

	if len(snap.Tiles) != 1 || snap.Tiles[0].TransitCondition != "unimproved" {
		t.Errorf("tile transit condition not exposed: %+v", snap.Tiles)
	}
	if len(snap.Carriers) != 1 || snap.Carriers[0].Type != "porter" {
		t.Errorf("carrier type not exposed: %+v", snap.Carriers)
	}
	c := snap.Carriers[0]
	if c.Q != 0 || c.R != 0 {
		t.Errorf("carrier position wrong: got (%d,%d), want (0,0)", c.Q, c.R)
	}
}

func TestPorterAdvancesOneTilePerTickOnUnimproved(t *testing.T) {
	// Scenario 1: the discriminating movement test. A porter on unimproved
	// ground advances 1 tile/tick toward its destination, and the destination
	// is cleared on arrival. This pins both the rule output (porter ×
	// unimproved → 1) and the tick-step mechanic (move toward destination,
	// stop on arrival).
	dest := NewHex(2, 0)
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(1, 0), Resource: "soil", Capacity: 10},
			{Hex: NewHex(2, 0), Resource: "soil", Capacity: 10},
		},
		Carriers: []CarrierSpec{
			{Type: "porter", Hex: NewHex(0, 0), Destination: &dest},
		},
	})

	w.Tick()
	if c := w.Snapshot().Carriers[0]; c.Q != 1 || c.R != 0 {
		t.Fatalf("after 1 tick: want porter at (1,0), got (%d,%d)", c.Q, c.R)
	}

	w.Tick()
	c := w.Snapshot().Carriers[0]
	if c.Q != 2 || c.R != 0 {
		t.Fatalf("after 2 ticks: want porter at (2,0), got (%d,%d)", c.Q, c.R)
	}
	if c.Destination != nil {
		t.Errorf("destination should be cleared on arrival, got %+v", c.Destination)
	}
}
