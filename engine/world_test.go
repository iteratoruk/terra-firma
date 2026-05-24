package engine

import (
	"encoding/json"
	"testing"
)

// World is the engine's root. These tests pin the architectural invariants from
// CLAUDE.md: determinism under a seed, the snapshot as the only window in
// (immutable + serialisable), and commands as the only way in.

func TestSameSeedSameRNGSequence(t *testing.T) {
	a := NewRNG(42)
	b := NewRNG(42)
	for i := 0; i < 100; i++ {
		if a.IntN(1000) != b.IntN(1000) {
			t.Fatalf("same seed diverged at draw %d", i)
		}
	}
	c := NewRNG(43)
	diverged := false
	for i := 0; i < 100; i++ {
		if NewRNG(42).IntN(1000) != c.IntN(1000) {
			diverged = true
			break
		}
	}
	if !diverged {
		t.Error("different seeds produced identical sequence (suspicious)")
	}
}

func TestWorldTickIsDeterministic(t *testing.T) {
	// Same seed + same (empty) command stream => identical snapshots, tick for
	// tick. This is the highest-priority invariant.
	runA := runWorld(t, 7, 50)
	runB := runWorld(t, 7, 50)
	if runA != runB {
		t.Fatal("same seed produced different snapshots after 50 ticks")
	}
	// NOTE: we deliberately do NOT assert that different seeds diverge. The V1
	// tick is pure arithmetic (stock += regen-harvest) with no stochastic
	// element, so the seed currently has nothing to influence and identical
	// non-seed inputs SHOULD produce identical worlds regardless of seed. That
	// is correct, not a bug. Reinstate a divergence assertion only once the tick
	// actually consumes the RNG (e.g. weather, herd events). Until then the RNG
	// is tested directly in TestSameSeedSameRNGSequence.
}

func TestSnapshotRoundTrips(t *testing.T) {
	w := NewWorld(Config{Seed: 1, Tiles: demoTiles()})
	w.Tick()
	snap := w.Snapshot()

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("snapshot not serialisable: %v", err)
	}
	var back Snapshot
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("snapshot not deserialisable: %v", err)
	}
	again, _ := json.Marshal(back)
	if string(data) != string(again) {
		t.Error("snapshot did not survive a JSON round-trip unchanged")
	}
}

func TestSnapshotExposesStockAndRates(t *testing.T) {
	// Legibility requirement: the snapshot must carry the trend, not just the
	// level, so a renderer can show "falling 3/tick".
	w := NewWorld(Config{Seed: 1, Tiles: []TileSpec{
		{Hex: NewHex(0, 0), Resource: "forest", Value: 40, Capacity: 100, Regen: 1, Harvest: 4},
	}})
	snap := w.Snapshot()
	if len(snap.Tiles) != 1 {
		t.Fatalf("expected 1 tile in snapshot, got %d", len(snap.Tiles))
	}
	ts := snap.Tiles[0]
	if ts.Value != 40 || ts.Regen != 1 || ts.Harvest != 4 || ts.Net != -3 {
		t.Errorf("snapshot stock/rates wrong: %+v (want value40 regen1 harvest4 net-3)", ts)
	}
}

func TestTileDepletesToZeroOverTicks(t *testing.T) {
	// The Slice 1 headline behaviour: harvested past its regen, a tile's
	// resource drains and floors at zero — locally, immediately, visibly.
	w := NewWorld(Config{Seed: 1, Tiles: []TileSpec{
		{Hex: NewHex(0, 0), Resource: "forest", Value: 10, Capacity: 100, Regen: 1, Harvest: 4},
	}})
	for i := 0; i < 100; i++ {
		w.Tick()
	}
	if got := w.Snapshot().Tiles[0].Value; got != 0 {
		t.Errorf("over-harvested tile should be exhausted (0), got %d", got)
	}
}

// runWorld builds a deterministic demo world and returns its serialised snapshot
// after n ticks, as a string for easy comparison.
func runWorld(t *testing.T, seed int64, n int) string {
	t.Helper()
	w := NewWorld(Config{Seed: seed, Tiles: demoTiles()})
	for i := 0; i < n; i++ {
		w.Tick()
	}
	data, err := json.Marshal(w.Snapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(data)
}

func demoTiles() []TileSpec {
	return []TileSpec{
		{Hex: NewHex(0, 0), Resource: "forest", Value: 60, Capacity: 100, Regen: 2, Harvest: 5},
		{Hex: NewHex(1, 0), Resource: "soil", Value: 80, Capacity: 80, Regen: 1, Harvest: 0},
		{Hex: NewHex(0, 1), Resource: "forest", Value: 30, Capacity: 100, Regen: 2, Harvest: 0},
	}
}
