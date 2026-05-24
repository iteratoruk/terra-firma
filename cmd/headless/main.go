// Command headless runs the engine for a fixed number of ticks and prints each
// snapshot as a line of text. It is the first "front end": a SEPARATE package
// that imports the engine and observes it only through Snapshot(). No network
// hop — in-process, the snapshot contract IS the boundary. This is a debug
// harness, not the real renderer.
package main

import (
	"flag"
	"fmt"

	"github.com/terrafirma/terrafirma/engine"
)

func main() {
	seed := flag.Int64("seed", 1, "world seed (same seed => same world)")
	ticks := flag.Int("ticks", 20, "number of ticks to run")
	flag.Parse()

	w := engine.NewWorld(engine.Config{
		Seed: *seed,
		Tiles: []engine.TileSpec{
			// A small demo: two forests under harvest pressure and a recovering
			// soil tile. Watch the over-harvested forest drain to zero.
			{Hex: engine.NewHex(0, 0), Resource: "forest", Value: 60, Capacity: 100, Regen: 2, Harvest: 5},
			{Hex: engine.NewHex(1, 0), Resource: "forest", Value: 30, Capacity: 100, Regen: 2, Harvest: 6},
			{Hex: engine.NewHex(0, 1), Resource: "soil", Value: 20, Capacity: 80, Regen: 3, Harvest: 0},
		},
	})

	fmt.Printf("Terra Firma — seed %d, %d ticks\n", *seed, *ticks)
	printSnapshot(w.Snapshot())
	for i := 0; i < *ticks; i++ {
		w.Tick()
	}
	fmt.Println("...")
	printSnapshot(w.Snapshot())
}

func printSnapshot(s engine.Snapshot) {
	fmt.Printf("tick %d:\n", s.Tick)
	for _, t := range s.Tiles {
		trend := "steady"
		switch {
		case t.Net > 0:
			trend = fmt.Sprintf("recovering +%d/t", t.Net)
		case t.Net < 0:
			ttl := "—"
			if t.Net < 0 {
				ttl = fmt.Sprintf("empty in %d", ceilDiv(t.Value, -t.Net))
			}
			trend = fmt.Sprintf("falling %d/t (%s)", t.Net, ttl)
		}
		fmt.Printf("  (%2d,%2d) %-7s %3d/%-3d  %s\n",
			t.Q, t.R, t.Resource, t.Value, t.Capacity, trend)
	}
}

func ceilDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return (a + b - 1) / b
}
