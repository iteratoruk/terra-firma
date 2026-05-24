# Terra Firma

*(working title)*

An economic simulation game about **physical limits**. You lead a small group that
settles a place, draws resources from it, and either learns to live within what the
land can regenerate — or exhausts it and has to move on.

It takes the **Settlers** lineage as its starting point — resources are physical,
located, and must be *carried*, so logistics is the mechanism of play — but inverts
its subject. Where Settlers and its descendants treat resources as effectively
infinite and make transport the only real constraint, here **finitude is the point**.
Forests deplete. Soil loses fertility. A settlement that consumes faster than the
land regenerates dies, and you migrate. Logistics is *how* you play; living within
limits is *what the game is about*.

## What makes it not-just-Settlers

Three commitments distinguish it from the genre it borrows from:

- **Finitude as subject.** Every resource is a stock with a regeneration rule. Harvest
  faster than it recovers and it collapses — locally, visibly, consequentially. Crop
  rotation, fallow, and managed forestry are not flavour; they are the difference
  between a settlement that lasts and one that doesn't.

- **Hegemony as a cost, not a prize.** (Later versions.) Becoming dominant —
  economically, politically, culturally, militarily — accumulates *influence*, but
  also accumulates a maintenance burden with declining marginal returns. The runaway
  leader carries the heaviest overhead. Empires fall because the cost of holding the
  system outruns what the system can extract. The pacing comes from having to spend
  down your position or be slowly crushed by its weight.

- **Powerlessness on the battlefield.** Conflict follows Settlers, not Anno: no
  click-to-attack, no general's god-view. You arrange the conditions and watch what
  your people can do. The absence of the power fantasy is deliberate.

The intellectual furniture — value as a social relation distinct from physical stock,
complexity as a cost that compounds, the representation you consume being separated
from the production that generates it — is load-bearing in the design but stays out
of the player's way. The game should be legible and fun to operate first. The ideas
are what it's *about*, not what it lectures you with.

## Architecture in one breath

A deterministic, seeded simulation **engine** (Go, headless) advances the world a
**tick** at a time. The outside world observes it only through an immutable,
serialisable **snapshot**, and acts on it only through **commands**. The renderer is a
separate concern on the far side of that boundary — the engine never knows it is being
watched. The first "renderer" is a debug dashboard of coloured hexes; cute little
people come much later and slot into the same snapshot contract.

See `docs/DESIGN.md` for the decisions behind all of this and `docs/ROADMAP.md` for
the release sequence.

## Status

Pre-alpha. Building the V1 spine: a hex world of depleting/regenerating resource
stocks, a flow atom that carries goods at real cost, and village-death as a live fail
state. No opponents yet, no real-world maps yet — both are deliberately deferred.

## Running it

_(Nothing to run yet. This section gets filled in once the engine has a tick and the
debug dashboard exists.)_
