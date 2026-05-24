# ADR-0001: Go, tick-based simulation, not an actor model

Status: Accepted

## Context

The engine simulates many "agents" (people, later livestock) that have state and
behaviour. Two implementation paradigms were on the table:

1. A **synchronous tick-based** model: a global clock; each `Tick()` advances the
   whole world deterministically; agents are data iterated over on that clock.
2. An **actor model** (e.g. Scala + Pekko): each agent is an autonomous process
   exchanging asynchronous messages.

The actor model is superficially attractive because "lots of autonomous agents
sending messages" sounds like the domain.

## Decision

Use **Go, with a synchronous, deterministic, tick-based core.** Concurrency (Go
goroutines) is permitted only as a *performance optimisation within a tick*
(parallelise agent updates, join, commit) — never as part of the model's
semantics. There is one seeded RNG; the same seed and the same command stream
produce a bit-identical world.

## Alternatives rejected

**Actor model (Scala/Pekko).** Rejected because actors give asynchronous,
non-deterministic message ordering, which is poison for a game we want to test,
balance, save, and reason about. Reproducible bugs, trivial save/load, and
golden-file testing all depend on determinism that actors actively work against.
Actors earn their keep when concurrency is irreducibly part of the domain
(distributed systems); a single-player sim's agents are not that — they are data
on a clock.

There is also a thematic reason, which matters for this project specifically: the
actor model would represent each little person as a genuinely autonomous locus of
agency. That is exactly *wrong* for a game whose subject is that the people are
caught in a system they do not control. The god-clock is the mechanically honest
representation of the thesis.

## Consequences

- Determinism becomes the highest-priority invariant (see CLAUDE.md, ARCHITECTURE.md).
  No wall-clock reads, no unseeded randomness, no map-iteration-order affecting state.
- Save/load is trivial: persist the `Config` + seed + command log, replay.
- Testing is unusually strong: golden-file run tests become possible because runs
  are reproducible and snapshots are serialisable.
- Performance scaling, if ever needed, is intra-tick parallelism, not actors.
- An agent or contributor tempted to "modernise" toward actors/async should read
  this first: the non-determinism it introduces would break the test strategy,
  save system, and balance reasoning the whole project rests on.

## Notes

Migrated from DESIGN.md ("Why Go, why tick-based, why not actors") as the first
worked example of the ADR pattern. DESIGN.md retains the mechanics-level framing;
this ADR is the decision-of-record.
