# Roadmap

**Terra Firma** (working name).

Thin vertical slices. Each slice should leave the project green, runnable, and
slightly more of a game than before. Build top to bottom; resist reaching ahead.
Checkboxes are the source of truth for "what's done" across sessions.

Vocabulary (see DESIGN.md "resource taxonomy"): **actors** = self-locating things with
autonomous tick-dynamics (humans, livestock); **tile-bound processes** = things a tile
owns and runs (crops, structures); **inert goods** = things that do nothing until
labour moves them (logs, grain, seed). Slices below use these terms deliberately.

## The release arc

- **V1 — Sandbox.** Single player, no opponents. The full physical-limits spine: hex
  world, depleting/regenerating stocks, the flow atom, technology-as-function-change,
  village death. This is where we prove the economy is interesting *on its own merits*,
  before any rivalry can paper over a hollow core.
- **V2 — Competition.** One or more AI opponents; the hegemony-as-complexity-load
  mechanic; Settlers-style indirect conflict. Only start once V1 is genuinely fun solo.
- **V3 — Real-world maps.** A `MapSource` implementation backed by open DEM/OSM data
  (NOT Google Maps — see note below), draping biomes over real geography. Procedural
  generation remains the default; real geography is *just another `MapSource`*.

> **Note on V3 / map sources.** Build to a clean `MapSource` interface from V1 (it
> yields a grid of tiles with elevation + biome) and back it with a procedural
> generator. Real-world geography is a later *implementation* of that interface. Use
> open data (SRTM / Copernicus DEM for elevation, OpenStreetMap for land use, Natural
> Earth for coastlines), not Google Maps — Google's licensing does not permit ingesting
> tiles/terrain as redistributable game assets, and it's the wrong data anyway. This is
> a self-contained data-engineering project that must not be allowed to eat the engine
> work early.

---

## V1 slices

### Slice 0 — Skeleton
- [ ] Go module initialised; `engine` package; CI runs `go test ./...` + `gofmt` check.
- [ ] Seeded RNG injected into the engine; a trivial determinism test (same seed →
      same RNG sequence) passing.
- [ ] `World` value with a no-op `Tick()` and an empty `Snapshot()`; round-trip
      serialise/deserialise of the snapshot tested.

### Slice 1 — One tile, one stock, depletion + regen
- [ ] Hex tile addressed in axial/cube coords; neighbour + distance functions, tested.
- [ ] A resource `Stock` as a first-class object tracking value + regen/harvest/net
      deltas. Regen and harvest functions. Property test: net = regen − harvest.
- [ ] A single tile with one stock that regenerates each tick; depletes when harvested
      past regen; cannot go below zero. Golden-file test over N ticks.
- [ ] Snapshot exposes stock value *and* its rates (legibility requirement).

### Slice 2 — The flow atom (one inert good walks)
- [ ] Multiple tiles; a path between two of them over hex edges.
- [ ] An **actor** (carrier) that moves one **inert good** (a log) from a source tile to
      a store over ticks, at a real transport cost (a subtracted quantity). The good is
      inert: it does nothing on its own; the actor's labour relocates it.
- [ ] Property test: physical (material-ledger) stock conserved across a transport step
      unless a production/consumption event accounts for the delta.
- [ ] Snapshot exposes the actor's position so a renderer could draw it *moving*.

### Slice 3 — Tile-bound process + consumption + population death
- [ ] One **tile-bound process** (raw → product) running at a tile against that tile's
      properties, consuming input stock, producing output stock. (First lifecycle hinge
      in code: tile-bound process yields an inert good.)
- [ ] A **population** (one or more human actors) with subsistence consumption that draws
      stores down each tick. Use general verbs: a population *draws from tiles* — do NOT
      hard-code "settlement owns territory". (Sedentary case only in V1, but framed so
      the mobile case isn't foreclosed. See CLAUDE.md "Don't" / DESIGN.md modes.)
- [ ] **Population death:** a population that cannot maintain subsistence throughput
      dies. The fail state is live and tested. This is the cliff; it is in from day one.

### Slice 4 — The debug dashboard (first "renderer")
- [ ] A tiny local web page (served by a *separate* main, not the engine package)
      polling `Snapshot()` and drawing coloured hexes on a canvas: stock levels as
      colour, falling trends visible, the carrier moving.
- [ ] This is the most useful tool in the project — it lets us *see* the economy
      misbehave while it's still squares. Not the real renderer.

### Slice 5 — Technology as function-change
- [ ] A minimal tech: e.g. crop rotation / fallow that modifies the soil-fertility
      regen function so managed tiles recover and unmanaged ones collapse.
- [ ] Research as a command applied at a tick boundary; effect visible in the snapshot
      rates.

### Slice 6 — Biome / difficulty + `MapSource`
- [ ] `MapSource` interface yielding tiles with elevation + biome.
- [ ] A procedural generator behind it. Biome sets per-tile stock levels and regen
      rates → difficulty and biome are the same lever.
- [ ] Starting-location choice as a command; strategic identity emerges from the
      stocks of chosen tiles.

When all of V1 is checked and the sandbox is genuinely fun to operate solo, stop, play
it, and only then open V2.

---

## Parking lot (explicitly NOT now)

Captured so they're not lost and not built prematurely:

- Emergent spatial cascades: deforestation → erosion/flooding (needs the legibility
  problem solved first; needs the hydrology layer cube-coords are setting up).
- The value/price layer distinct from physical stock (surplus, speculation as game
  objects).
- Hegemony / complexity-load late game.
- Real-world `MapSource` from open DEM/OSM data.
- A real process/network boundary for the renderer (and any JVM/browser renderer).
