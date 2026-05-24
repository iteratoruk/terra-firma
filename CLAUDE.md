# CLAUDE.md

Operating rules for this project. Terse on purpose.

- *Why the game works as it does* (mechanics): @docs/DESIGN.md
- *How the code is structured* (navigation): @docs/ARCHITECTURE.md
- *Settled decisions with rejected alternatives*: @docs/adr/README.md (the index;
  read the relevant ADR before reversing anything)
- *What to build next*: @docs/ROADMAP.md
- *How to build/test/run/verify*: the Makefile (`make help`). This is the feedback
  loop — do not infer commands, look them up. **`make check` is the pre-commit
  gate; run it before every commit.** If it doesn't pass, the commit isn't ready.

Do not restate the contents of those files here.

## What this is

**Terra Firma** (working name). A deterministic, headless economic simulation engine in
Go. A game about physical resource limits. Rendering is a separate concern that does not
exist yet.

## Non-negotiable invariants

These are the rules that, if broken, break the project. Hold them in every session.

- **Determinism.** Same seed + same commands → identical world, tick for tick. No
  wall-clock time, no unsorted-map iteration that affects state, no non-seeded
  randomness anywhere in the engine. All randomness goes through the injected seeded
  RNG. If two runs with the same seed diverge, that is a bug of the highest priority.
- **The snapshot is the only window in.** Anything outside the engine observes state
  *only* through `Snapshot()`, which returns an immutable, serialisable value. Never
  expose mutable internal state. Never let a caller reach past the snapshot.
- **Commands are the only way in.** External actors change the world *only* by
  submitting a `Command` applied at a tick boundary. No setters, no direct mutation
  from outside the engine package.
- **The engine never imports a rendering, UI, or I/O-presentation concern.** No
  drawing, no pixels, no colours, no framerate, no HTTP handlers inside the engine
  package. The engine does not know it is being watched. If you are tempted to add a
  field "for the renderer," it belongs in the snapshot, not the engine state.
- **Stocks know their own deltas.** A resource stock is a first-class object that
  tracks its own rate of change (regen, harvest, net), not a bare integer differenced
  after the fact. The snapshot exposes stocks *and rates*. Legibility of trend is a
  design requirement, not an afterthought.
- **Coordinates are axial/cube internally.** Hex grid addressed in axial (cube)
  coordinates. Offset coordinates are a display-time concern only and must never leak
  into engine logic. Neighbour/distance/flow operations use cube math. Reference:
  Red Blob Games hex guide.
- **Classify by what the tick does, not by what the player does.** The resource
  taxonomy (actors / tile-bound processes / inert goods) is defined by *autonomous
  tick-dynamics*, not by player-relocation-cost. Actors (humans, livestock) have state
  that evolves on their own each tick; inert goods (logs, grain, seed) do nothing until
  labour acts on them; tile-bound things (crops, structures) run against their tile.
  Livestock are actors even though driving them costs labour — that's a property, not a
  reclassification. See DESIGN.md "resource taxonomy".
- **No fungible quantity that converts between domains.** There is no universal "points"
  or "money" the player spends across education/military/influence/tech. The material
  and social ledgers have different conservation laws and must NOT be made commensurable
  by a converter — the incommensurability is the game. Surplus is the material ledger's
  real un-consumed remainder (physical, perishable, located), directed by opportunity
  cost, never an abstract spend. Money, if it ever exists, is a late *emergent
  technology*, not a starting balance. UI scalars reading distinct gauges are fine; a
  single scalar spendable across domains is the trap. See DESIGN.md "surplus, not points".

## How we work

- **TDD.** Write the failing test first. The engine is pure and deterministic, so it
  is highly testable — there is no excuse not to. Prefer table-driven tests.
- **Golden-file tests for runs.** Seed a world, run N ticks, serialise the snapshot,
  commit it as the expected output. Behaviour changes show up as reviewable diffs.
  When a change *should* alter the golden file, regenerate it as a deliberate,
  reviewed step — never blindly.
- **Property-based tests for invariants.** e.g. physical stock is conserved across a
  transport step unless a production/consumption event accounts for the delta.
- Keep functions pure where possible; isolate the small mutable core.
- Small commits, each green. Conventional-commit style messages.

## Go conventions

- Standard `gofmt`/`goimports`; idiomatic Go, not Java-in-Go.
- No global mutable state. The world is a value the engine owns; dependencies (RNG,
  config) are injected.
- Errors are values; wrap with context. Panics only for genuinely-impossible states.
- Map iteration order is nondeterministic in Go — never iterate a map where order
  affects engine state. Use sorted keys or ordered slices for anything stateful.

## Don't

- Don't add the AI opponent, multiplayer, or real-world map sources yet. They are
  V2/V3. See the roadmap. Building them now is scope creep.
- Don't build emergent spatial cascades (deforestation → flooding) into V1. Local,
  immediate consequences only until the legibility problem is solved. See design doc.
- Don't ship any social-ledger effect without its legibility story. When the social
  ledger is eventually built, every effect must be explainable in the dashboard ("this
  stratum's standing fell because this policy stopped recognising this task"), or it
  becomes mystique not mechanic. Same rule as cascades, higher stakes. (Not a V1
  concern — the social columns of task vectors are zero in V1 — but don't build a task
  model that hard-codes social credit as a fixed task property.)
- Don't make "settlement" or "defend your location" engine primitives. The root concept
  is *a population organised around a mode of drawing subsistence from tiles* — a
  settlement is the sited case, a band the mobile case. Keep the verbs general ("this
  population draws from these tiles") so the nomadic strategy is never framed out, even
  though V1 builds only the sedentary case. See DESIGN.md "modes of production".
- Don't add a real network/process boundary for the renderer yet. In-process, the
  snapshot contract *is* the boundary. Push the real boundary out until a concrete
  need pulls it in.
- Don't write essays in comments. The design rationale lives in docs/, not inline.
