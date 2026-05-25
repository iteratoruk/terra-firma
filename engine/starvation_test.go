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

func TestPopulationAliveAtLimitDeadOneTickLater(t *testing.T) {
	// Scenario 2 — the cliff itself. With reserve 3, metabolism 1, limit 3:
	//   tick 3: reserve hits 0, streak → 1
	//   tick 4: reserve still 0,  streak → 2
	//   tick 5: reserve still 0,  streak → 3  (== limit, still alive)
	//   tick 6: streak would be 4 (> limit), population removed
	// The "alive at the limit" half forbids a one-tick guillotine (the streak
	// must be observable at its peak BEFORE death — issue note legibility);
	// the "gone after one more" half forbids a never-dies cliff.
	w := NewWorld(Config{
		Tiles: []TileSpec{
			{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
		},
		Populations: []PopulationSpec{
			{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1, StarvationLimit: 3},
		},
	})

	for i := 0; i < 5; i++ {
		w.Tick()
	}

	snap := w.Snapshot()
	if len(snap.Populations) != 1 {
		t.Fatalf("after 5 ticks: want exactly 1 population (at the limit, still alive), got %d", len(snap.Populations))
	}
	p := snap.Populations[0]
	if p.Q != 0 || p.R != 0 {
		t.Errorf("population should still be at (0,0), got (%d,%d)", p.Q, p.R)
	}
	if p.StarvationTicks != 3 {
		t.Errorf("starvation_ticks at the limit: want 3, got %d", p.StarvationTicks)
	}

	w.Tick()

	snap = w.Snapshot()
	if n := countPopulationsAt(snap, NewHex(0, 0)); n != 0 {
		t.Errorf("after 1 more tick: want 0 populations at (0,0) (removed by the cliff), got %d (%+v)", n, snap.Populations)
	}
}

func TestRecoveryResetsStarvationStreak(t *testing.T) {
	// Scenario 3 — the cliff is escapable from its edge. With reserve 3,
	// metabolism 1, limit 3, the streak is 1 after 3 ticks. A grain
	// (calorie 5) appearing at (0,0) is eaten on tick 4: +5 (reserve = 5)
	// then -1 metabolise (reserve = 4); reserve > 0 at end of tick, so the
	// streak resets to 0. Without the reset, the streak would persist at 1
	// and the player would be punished for a recovery that has already
	// happened — recovery is real, not a pause-button.
	//
	// The grain is added mid-run by reaching into the engine's slice from
	// this same-package test. There is no public good-spawn command in V1;
	// when one arrives the test can switch to Apply(...).
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

	mid := w.Snapshot().Populations[0]
	if mid.Reserve != 0 || mid.StarvationTicks != 1 {
		t.Fatalf("mid-run setup wrong: want reserve 0 + streak 1, got reserve %d streak %d", mid.Reserve, mid.StarvationTicks)
	}

	w.goods = append(w.goods, &good{kind: "grain", hex: NewHex(0, 0)})
	sortGoods(w.goods)

	w.Tick()

	snap := w.Snapshot()
	if len(snap.Populations) != 1 {
		t.Fatalf("recovery should keep the population alive, got %d", len(snap.Populations))
	}
	p := snap.Populations[0]
	if p.Reserve != 4 {
		t.Errorf("reserve after eat(+5) then metabolise(-1): want 4, got %d", p.Reserve)
	}
	if p.StarvationTicks != 0 {
		t.Errorf("streak after recovery: want 0, got %d", p.StarvationTicks)
	}
}

func TestEndToEndSameSeedDifferentOutcomes(t *testing.T) {
	// Scenario 4 — the first game-shaped test in the project. Two worlds with
	// identical seeds and population specs; the ONLY difference is the food
	// supply. The seeded run resolves to two different outcomes after the same
	// 20 ticks: the well-fed population survives, the unfed one is gone.
	//
	// With reserve 3 / metabolism 1 / limit 3:
	//   World B (no grain): dies at the end of tick 6 (streak 4 > 3).
	//   World A (5 grain @ (0,0)): eats +5, metabolises -1 each of ticks 1–5
	//     (reserve climbs to 23); ticks 6–20 just metabolise (15 × -1),
	//     leaving reserve = 8 at tick 20 — alive, reserve > 0.
	//
	// The seed is identical between the two worlds and is the same one used
	// by the empty world above for the determinism story. Death (or survival)
	// is a function of the configuration, not the seed; same seed AND same
	// food = same outcome forever (determinism invariant).
	const seed = 7
	specs := []PopulationSpec{
		{Hex: NewHex(0, 0), Reserve: 3, Metabolism: 1, StarvationLimit: 3},
	}
	tiles := []TileSpec{
		{Hex: NewHex(0, 0), Resource: "soil", Capacity: 10},
	}

	worldA := NewWorld(Config{
		Seed:        seed,
		Tiles:       tiles,
		Populations: specs,
		Goods: []GoodSpec{
			{Kind: "grain", Hex: NewHex(0, 0)},
			{Kind: "grain", Hex: NewHex(0, 0)},
			{Kind: "grain", Hex: NewHex(0, 0)},
			{Kind: "grain", Hex: NewHex(0, 0)},
			{Kind: "grain", Hex: NewHex(0, 0)},
		},
	})
	worldB := NewWorld(Config{
		Seed:        seed,
		Tiles:       tiles,
		Populations: specs,
	})

	for i := 0; i < 20; i++ {
		worldA.Tick()
		worldB.Tick()
	}

	snapA := worldA.Snapshot()
	if len(snapA.Populations) != 1 {
		t.Fatalf("world A after 20 ticks: want population still alive, got %d populations", len(snapA.Populations))
	}
	if snapA.Populations[0].Reserve <= 0 {
		t.Errorf("world A after 20 ticks: want reserve > 0 (well-fed survives), got %d", snapA.Populations[0].Reserve)
	}

	snapB := worldB.Snapshot()
	if len(snapB.Populations) != 0 {
		t.Errorf("world B after 20 ticks: want population dead (no food), got %+v", snapB.Populations)
	}
}

func countPopulationsAt(s Snapshot, h Hex) int {
	n := 0
	for _, p := range s.Populations {
		if p.Q == h.Q && p.R == h.R {
			n++
		}
	}
	return n
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
