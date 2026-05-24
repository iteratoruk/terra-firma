package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Golden-file run: seed a world, run N ticks, serialise the snapshot, compare
// to a committed expected output. A failure means behaviour changed; if that
// was intended, regenerate the golden via `make golden` and review the diff.

func TestGoldenDemoWorldSeed7_20Ticks(t *testing.T) {
	w := NewWorld(Config{Seed: 7, Tiles: demoTiles()})
	for i := 0; i < 20; i++ {
		w.Tick()
	}

	got, err := json.MarshalIndent(w.Snapshot(), "", "  ")
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	got = append(got, '\n')

	path := filepath.Join("testdata", "demo_world_seed7_20ticks.golden.json")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run `make golden` to create): %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("snapshot diverged from %s. If the change is intended, regenerate with `make golden` and review the diff.", path)
	}
}
