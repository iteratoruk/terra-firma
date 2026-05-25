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

	// As of issue #4, w.rng is consumed by the randomised command generator in
	// goods_test.go — so the world's evolution genuinely depends on the seed
	// for the first time. The deferred divergence assertion (previously vacuous
	// because no part of world evolution touched the RNG) is now meaningful:
	// same seed under the same randomised harness must produce the same world,
	// AND different seeds must produce different worlds.
	randA := runWorldUnderRandomCommands(t, 7, 50)
	randB := runWorldUnderRandomCommands(t, 7, 50)
	if randA != randB {
		t.Fatal("same seed under randomised commands produced different snapshots")
	}
	randC := runWorldUnderRandomCommands(t, 8, 50)
	if randA == randC {
		t.Fatal("different seeds produced identical worlds under randomised commands — w.rng isn't driving evolution")
	}
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

// runWorldUnderRandomCommands drives a small world by the same w.rng-backed
// command generator used by the conservation property test (see
// applyRandomCommand in goods_test.go). The serialised snapshot after n steps
// is the comparable artefact: same seed → same string; different seed →
// different string (the divergence assertion in TestWorldTickIsDeterministic
// depends on the latter).
func runWorldUnderRandomCommands(t *testing.T, seed int64, n int) string {
	t.Helper()
	w := buildConservationWorld(seed, 3, 2)
	for i := 0; i < n; i++ {
		applyRandomCommand(w)
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
