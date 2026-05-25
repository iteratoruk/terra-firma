package engine

import (
	"encoding/json"
	"testing"
)

// Starvation — issue #7: the cliff has teeth. A population whose reserve sits
// at zero accumulates a streak; when it passes the limit, the population dies
// (is removed from the snapshot). Recovery is real: a refill resets the streak.
//
// The streak is exposed in the snapshot BEFORE death (legibility), and the
// dead population is a snapshot ABSENCE, not a flag (a dead-but-present ghost
// invites "can I revive them?" — see issue notes).
//
// Tick order: stocks → carriers → populations eat → populations metabolise →
// starvation streak update → death removal. The scenarios' arithmetic depends
// on streak-update happening AFTER metabolise.

func TestStarvationStreakBeginsWhenReserveHitsZero(t *testing.T) {
	// Scenario 1 — the discriminating test for the streak counter. With reserve
	// 3 and metabolism 1, the reserve hits zero at the end of tick 3. The
	// streak rule (if reserve==0 at end of tick, ticks+=1) must then bump
	// starvation_ticks to 1. A streak that bumps a tick early or late breaks
	// the rest of the cliff arithmetic.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1, StarvationLimit: 3},
		},
	})

	for i := 0; i < 3; i++ {
		w.Tick()
	}

	snap := w.Snapshot()
	if len(snap.Populations) != 1 {
		t.Fatalf("expected population still alive at end of tick 3, got %d", len(snap.Populations))
	}
	p := snap.Populations[0]
	if p.Reserve != 0 {
		t.Errorf("reserve at end of tick 3: want 0, got %d", p.Reserve)
	}
	if p.StarvationTicks != 1 {
		t.Errorf("starvation_ticks at end of tick 3: want 1, got %d", p.StarvationTicks)
	}
}

func TestSnapshotExposesStarvationFieldsAndRoundTrips(t *testing.T) {
	// Legibility + snapshot-as-only-window requirement: the streak and the
	// limit are visible to observers, and the JSON form round-trips so the
	// renderer/dashboard contract holds.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1, StarvationLimit: 3},
		},
	})

	snap := w.Snapshot()
	if got := snap.Populations[0].StarvationTicks; got != 0 {
		t.Errorf("initial starvation_ticks: want 0, got %d", got)
	}
	if got := snap.Populations[0].StarvationLimit; got != 3 {
		t.Errorf("starvation_limit must round-trip the spec: want 3, got %d", got)
	}

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
