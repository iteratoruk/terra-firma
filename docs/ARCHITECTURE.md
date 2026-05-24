# Architecture

The structural map of the code, for navigation and safe modification. This is
the *what and how* of the implementation. For the *why* of the game's mechanics
(finitude, ledgers, ideology), see DESIGN.md. For settled decisions with rejected
alternatives, see docs/adr/. For build/test commands, see the Makefile.

## System context

Terra Firma is a single-player, offline, deterministic economic simulation. There
is no server, no network, no database, no external dependency beyond the Go
standard library. "Users" of the engine are: the headless harness (today), and a
graphical renderer (later). Both are *observers* that consume snapshots; neither
reaches into engine state.

```
        commands                         snapshots
  observer ───────────▶  ┌───────────┐  ───────────▶  observer
  (renderer/harness)     │  engine   │                (renderer/harness)
                         │  (World)  │
                         └───────────┘
```

The whole architecture is one boundary: **the engine is observed only through an
immutable serialisable Snapshot and mutated only through Commands applied at tick
boundaries.** Everything else follows from holding that line.

## Packages (container level)

| Path            | Responsibility | May import |
| --------------- | -------------- | ---------- |
| `engine/`       | The authoritative simulation. Deterministic, headless, pure of I/O and rendering. The only package that mutates world state. | stdlib only |
| `cmd/headless/` | A debug front-end: runs the engine and prints snapshots as text. The first "observer". | `engine`, stdlib |

Dependency rule (enforced by convention + review, see CLAUDE.md): **`engine`
imports nothing of the project; observers import `engine`. Never the reverse.**
The engine does not know it is being watched. If `engine` ever needs to import a
`cmd/` or rendering package, something has gone wrong.

## Inside `engine/` (component level)

Flat package, organised by file. Flatness is deliberate (Go idiom + agentize C1.1
favours accessible monoliths over clever hierarchies). Split into sub-packages
only when a real compile-time boundary demands it, and then by *domain concept*,
not by *layer*.

| File          | Component | Notes |
| ------------- | --------- | ----- |
| `world.go`    | `World`, `Config`, `TileSpec`, `Snapshot`, `TileSnapshot`, `Command` | The root. Owns all state; defines the boundary types. |
| `hex.go`      | `Hex` and hex math | Axial/cube coordinates. Neighbours, distance, canonical sort. The spatial substrate. |
| `stock.go`    | `Stock` | A resource quantity that tracks its own regen/harvest/net deltas. |
| `rng.go`      | `RNG` | The engine's single source of seeded randomness. |
| `*_test.go`   | tests | Co-located with the code they test (Go convention). |
| `testdata/`   | golden files | Serialised snapshots for golden-file run tests. Ignored by the go tool. |

### The resource taxonomy in code

DESIGN.md defines three modes of being (actors / tile-bound processes / inert
goods), distinguished by *what the tick does to them*. This is a conceptual guide
to data shape, NOT a class hierarchy. Current mapping:

- **Tile-bound process** — a resource `Stock` owned by a `tile` (in `world.go`).
  `Tick()` steps it against its regen/harvest. (Implemented.)
- **Inert good** — a quantity moved only by labour. (Not yet — arrives in Slice 2
  as the carried log.)
- **Actor** — self-locating thing with autonomous tick-dynamics (humans,
  livestock). (Not yet — arrives Slice 2/3.)

When actors and goods arrive, expect transitions modelled as plain functions
(`Plant`, `Harvest`, `Slaughter`), not polymorphic methods on a shared base.

## Critical flow: one tick

The single most important flow to understand before modifying the engine.

1. Caller invokes `World.Tick()`.
2. The world iterates its tiles **in canonical hex order** (sorted at
   construction; never map-iteration order — that would break determinism).
3. Each tile's `Stock.Step()` applies `net = regen - harvest`, clamped to
   `[0, capacity]`. A stock hitting `0` is the local, immediate exhaustion the
   design depends on.
4. `world.tick` increments.

No wall-clock time is read. No unseeded randomness is drawn. Given the same
`Config` and the same sequence of `Tick()`/`Apply()` calls, the world is
bit-identical every run. This is the highest-priority invariant.

## Critical flow: observation

1. Caller invokes `World.Snapshot()`.
2. The world builds a fresh `Snapshot` value, copying out tile state **including
   precomputed `Net`** so observers can show a trend without reaching back in.
3. Tiles appear in canonical hex order, so the serialised form is stable — which
   is what makes golden-file testing possible.

The `Snapshot` shares no mutable state with the `World`; an observer cannot mutate
the world by holding a snapshot.

## Determinism: how it is protected

- Single seeded `RNG` (`rng.go`), injected from `Config.Seed`. No package-level
  `rand`, no time-based seeds.
- Tiles iterated in canonical hex order, never map order.
- All stock arithmetic is integer and order-independent within a tick.
- Verified by `TestWorldTickIsDeterministic` and run under `-race` in CI.
- (Today the tick consumes no randomness, so the seed is plumbed but inert. When
  stochastic behaviour arrives, the same-seed/same-world property must still hold;
  add a different-seed/different-world assertion at that point.)

## Testing strategy

- **Unit tests**, co-located, table-driven where natural. Current coverage ~92%.
- **Property tests** for invariants (e.g. net = regen − harvest; stock floored at
  zero; conservation across a transport step — the last arrives with Slice 2).
- **Golden-file run tests** (Slice 2 onward): seed a world, run N ticks, serialise
  the snapshot, compare to a committed `testdata/*.golden.json`. A changed golden
  file is a reviewable record that behaviour changed. Regenerate deliberately via
  `make golden`, never blindly.

## What deliberately does not exist yet

Tracked so their absence is understood as a choice, not an oversight (see ROADMAP):
no AI opponents, no multiplayer, no real-world map sources, no renderer beyond the
text harness, no network/process boundary, no social ledger, no money, no emergent
environmental cascades. The architecture is shaped so none of these requires
tearing up the boundary when it arrives.
