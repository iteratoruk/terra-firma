# Design

The reasoning behind the decisions. `CLAUDE.md` states the rules; this explains why
they exist, so the rules can be applied with judgement rather than cargo-culted. If a
rule here and reality conflict, this document is wrong and should be updated — it is a
record of decisions, not scripture.

## The core inversion

The genre (Settlers, Anno, Civ, Factorio) makes **logistics** the drama and treats
resources as effectively infinite. This game keeps logistics as the *mechanism* but
makes **finitude** the *subject*. No prior Settlers-like makes resource exhaustion and
regeneration the central tragedy; they make transport the constraint and the pile
bottomless. Inverting that is the entire pitch and the thing to protect when trading
off everything else.

## Game vs. instrument

This is a **toy that is fun to operate**, not a pure model. "Fun to operate" is the
ruthless filter: a mechanic that doesn't survive contact with playability gets cut,
even if it's theoretically faithful. The political-economic ideas are what the game is
*about*; they are not a lecture and not a simulation-for-its-own-sake. Legibility and
playability win ties.

## The flow atom

The irreducible unit is a **flow**, not a stockpile. Resources are physical, located,
and must be moved by carrier agents who are themselves a constraint (the Settlers
donkey-bottleneck lineage). Transport cost is a *real subtracted quantity*, not a
hand-wave — if it isn't tangible, the simulation fails to describe physical limits,
which is the whole point. The value/influence layers in later versions ride on *who
controls the flows and chokepoints*, so flow has to be the substrate from tick one.
The smallest meaningful first build is **one log walking from forest to warehouse**,
deterministically, observable in the snapshot.

## The recurring move: properties and rules, never tables of outcomes

This is the spine of the whole design and it recurs in every subsystem below. Wherever
the genre would hand-author a list of outcomes (crop X grows in biome Y; ideology Z
gives bonus W), this game instead models a small set of **properties** and a **rule**
that computes the interaction, and lets the outcomes emerge. Crops are property-ranges
matched against tile properties; technologies are modifications to functions; labour is
a cost/demand vector; ideology is a policy over a space. The payoff is that content
cost stays sub-linear — you extend the game by adding properties, not by authoring
every pairwise interaction — and the game can produce configurations you never
explicitly designed. When tempted to write a lookup table of outcomes, stop: model the
properties and the rule instead.

A second, related caution learned the hard way in design: **classify by what the tick
has to do, not by what the player has to do.** Player-facing similarity and
engine-facing distinction frequently pull apart (e.g. a driven herd and a carried log
both look like "a thing a person moves," but the tick must run the herd's autonomous
dynamics and do nothing to the log). The snapshot exists precisely to let those two
views diverge: classify in the engine by tick-behaviour, let the snapshot present
player-facing similarities.

## The resource taxonomy: three modes of being

Everything in the economy is one of three kinds of thing, distinguished by **how
location attaches** — equivalently, by **what `Tick()` must do to it**. This taxonomy
is the backbone; stocks, flow, surplus, capital, the herd, and (later) money all hang
off it. NB this is a *conceptual* taxonomy to guide data shape — NOT a mandate for a
class hierarchy. In Go, prefer distinct types with transitions as plain functions; a
thing's "mode of being" is a fact about which type holds it, not a field on a shared
base. Do not build a UML tree and fall in love with it.

- **Actors** — self-locating things with **autonomous tick-dynamics**. They have state
  that evolves under their own rules every tick regardless of player input: humans
  (metabolise, age, can starve), livestock/herds (graze the tile they stand on, breed,
  starve, can die to weather/predation). Location is a property *of the thing* and the
  thing's state changes on its own. This is the expensive category. Population-held
  stocks (the nomad's herd travels with the population, not the tile) live here.

- **Tile-bound processes** — no independent location; a tile owns them and they draw on
  that tile's properties to do something over time: crops, structures, improvements. A
  field *is at* its tile, cannot be picked up, interacts with the tile's conditions.
  `Tick()` runs the tile's process. Location is a property *of the tile*.

- **Inert goods** — have a location but **no autonomous dynamics**: left alone for 50
  ticks they do nothing. Moved only by an actor expending labour. Logs, harvested
  grain, seed-in-its-bag. `Tick()` does nothing to them; they change only when the
  labour system acts on them. Location is a property *of the thing* but the thing
  cannot change it.

**The criterion is autonomous tick-dynamics, not player-relocation-cost.** "Must be
driven by labour to be relocated on command" is a *property some actors have*, not a
reason to reclassify them as goods. A herd is an actor (it grazes and breeds on its
own — that autonomy IS the pastoral economy) that is *also* labour-coupled for
commanded movement. Filing it as an inert good would force the tick loop to special-
case its grazing/breeding back in, i.e. an "inert" good that isn't — the exact
impedance mismatch to avoid. Bonus: actor-plus-labour-coupling gives herding vs.
droving for free (a grazing herd needs little labour and gains condition; a driven herd
needs heavy labour and loses condition).

### Lifecycle hinges = transitions between modes

Every interesting economic event is a thing changing which mode it's in. Most are
physical transformations; one is a pure allocation decision:

- **Seed → planted crop**: inert good → tile-bound process. (Act: planting.)
- **Crop → harvest**: tile-bound process → inert good. (Act: harvesting.)
- **Livestock → meat**: actor → inert good. (Act: slaughter; *irreversible* — forecloses
  the breeding the actor offered. The eat-it-or-breed-it tension as a one-way hinge.)
- **Grain → seed**: inert good → inert good — *no mode change, no physical change*, only
  a change of intended project (eat vs. plant). The purest expression of the surplus
  question: the drama is entirely in the labelling of intent. The model must express
  both transformation hinges AND pure-allocation hinges.

## Finitude: two kinds of limit, deliberately separated

1. **Stock depletion with a regeneration rule** (V1, the spine). Every resource is a
   stock per tile with a regen function and a harvest function. Harvest > regen →
   collapse. Cheap, legible, and the heart of the game. A *technology* is modelled as
   a modification to these functions — e.g. crop rotation changes the soil-fertility
   regen function so fallow tiles recover faster. This is a clean, honest model of
   what a technology actually *is*: a change to the rules governing a process.

2. **Emergent spatial cascades** (V1.5+, deferred). "Fell trees here → flooding/erosion
   there" is a cross-tile hydrological effect, not a per-tile number. It is
   thematically perfect (it's half of why several Tainter collapses happened) but has
   a hard **legibility problem**: the consequence is displaced in space and time, so
   the player may not connect cause to effect, and it reads as arbitrary. That breaks
   the tangible-limits goal. Do not build it until the feedback/legibility problem is
   solved. V1 consequences are **local and immediate**: this tile's fertility is gone,
   so this farm dies, now.

The fail state — **village death** — is in from day one. Without it, depletion has no
teeth and the physical-limits thesis is decorative.

## Biome = difficulty

Starting conditions are abstract and dateless (no fixed "6000 BC"; the era is
notional flavour, not simulated content). You start with *some* resources and the
biome constrains which directions are viable. This means **difficulty and biome are
the same lever**: a hard start is a thin biome with slow regen; an easy start is a
rich one. Strategic identity (forestry-and-trade vs. grain-surplus) emerges from the
stocks and regen rates of the tiles a player settles, not from scripted rules. Free
strategic depth out of the resource maths.

## Needs: not one mechanic

"Need" is several distinct mechanics and must not collapse into one bar, or the game
stops being about *place*:

- Needs met by **consuming a stock** (food/calories always; water when scarce).
- Needs met by **being in a state** (sheltered, warm, safe) — conferred by a persisting
  structure, not a stock you deplete.
- Needs met by **proximity to a feature** (adjacent to river → water is free).

Where you settle determines which needs are free and which are expensive. This is why
placement is the heart of the early game, and why hex adjacency earns its rent ("within
shelter range?", "adjacent to water?" are uniform-neighbour queries).

### Build order for the economic layers

Each layer is inert until the one below works; each is where one theme first becomes
expressible. Do not build a higher layer before the lower is solid.

0. **Subsistence stock + deadline.** A person draws down an internal reserve each tick;
   eating refills it; at zero, escalating consequences then death. One abstract calorie
   unit. No food types, no preferences. This is the cliff — the whole game for a while.
1. **Food as substance with properties.** Distinct foodstuffs with calorie density and —
   add this *before* preference — **perishability**. Perishability is a political
   property masquerading as a physical one: it makes storage a tech worth researching,
   makes surplus *hard to hold* (capital has a decay rate), and is the first wedge
   between physical stock and value (what lasts is worth more than what rots,
   independent of calories).
2. **Capital/process resources.** Seeds, cuttings, livestock — not consumed (or not
   only; livestock is both). Capital + suitable tile + time → flow of consumable.
   "Suitable" via property-match (tile temp/moisture/fertility vs. crop tolerated
   ranges + yield curve), NOT a crop×biome table. Gives "grows badly" (survives,
   low-yield) for free, which makes placement a real decision.
3. **Needs that aren't stocks.** Shelter-as-state, water-as-hybrid (free adjacent to
   water, carried stock otherwise). Makes placement central.
4. **Preference and desirability.** ONLY now. Once subsistence is reliable and a surplus
   of calories exists, *which* calories — variety has value, some foods desirable beyond
   physical properties. Birth of use-value-beyond-subsistence; first place the social
   layer peels from the physical. Meaningless before 0–3 (preference without subsistence
   is decoration on a corpse).

## Labour, and the two ledgers

Labour-as-energy is a physical resource: a depleting reserve, replenished by sustenance,
drawn down faster by heavier work. It lives in the **material ledger** beside calories.
It is also *finite per tick* (a person does one thing per tick), which is what makes the
economy real: competing uses for the same hands, opportunity cost, the basis on which
surplus extraction will later mean something. Labour enters around layer 2.

But "military work needs social capital, dirty jobs demand acknowledgement, spectacle
satisfies desire without material recompense" is NOT "labour is a resource." It is the
claim that **work has a social cost not reducible to its caloric cost, paid in a
different currency.** The correct abstraction is **two ledgers with different
conservation laws** — and the WRONG abstraction (reject it) is a second "social energy"
stock that works like calories. That wrong move treats a social relation as a physical
substance and makes the social layer merely calories-with-a-different-colour, which
makes the spectacle endgame *inexpressible*.

- **Material ledger**: physical, conserved, legible. Calories, goods, labour-as-energy.
  Can't get out more than went in.
- **Social ledger**: non-conserved, relational. NOT a stock a person "has" but a
  *standing held in relation to others/the settlement*. Recognition, status, legitimacy.
  Defining property: **it need not balance against the material ledger.** You can pay in
  status instead of calories; extract material surplus and return recognition; or — the
  spectacle move — satisfy a social want with a representation costing almost nothing
  material.

Every kind of work has a **cost vector** and a **demand vector**, each with a material
and a social component, and these need not match:

- Ordinary labour: material cost (energy), material demand (sustenance), social ~0.
- Dirty work: material energy cost *plus* a social cost to the worker (lowers standing)
  unless offset by issued recognition. Unmet → refusal/unrest. Legibly accounted.
- Military labour: material sustenance (more of it) *plus* a social prerequisite — can
  only be performed by those holding sufficient standing, or status must be conferred.
  So military power is gated behind the social ledger: a society that can *feed* soldiers
  but cannot *legitimate* them cannot field them. (This is why the education trap below
  works.)

**Spectacle (late tech) = a technology that issues social-ledger credit at low material
cost.** Desire is a social-ledger demand. Normally satisfied by recognition backed by
real material contribution; spectacle satisfies the same demand with representation of
minimal material backing. Its hollowness is *emergent*, not authored: the material
ledger grinds on unsatisfied while the social one reports contentment, and that gap is
visible in the dashboard. The game says it without saying it.

**Type-level commitment now, mechanism deferred:** model work as a task with a
cost/demand vector over both ledgers, with the social components simply **zero** in the
early game. The social columns exist in the type and go unused. Do not build a task model
that hard-codes social credit as a fixed task property — that's the version that can
never grow a politics. Social credit issued by a task is *computed by the society's
recognition-policy applied to the task* (see ideology), not a property of the task.

**LEGIBILITY CONSTRAINT (raised stakes):** social causation is even harder to make
visible than hydrology. **No social-ledger effect ships without its legibility story** —
the dashboard must be able to show "this stratum's standing fell because this policy
stopped recognising this task," or the politics becomes mystique rather than mechanic.
Same rule as environmental cascades, higher stakes.

## Ideology = a recognition-policy over a continuous space

An ideology is **not a set of bonuses** (that's the Civ-civics mush — a label on a
+10%). An ideology is a **policy over the social ledger: a function deciding who gets
recognised for what.** The material ledger runs identically under every ideology
(calories are calories, a sword costs the same iron). What changes is the *allocation of
social-ledger credit*: which tasks/strata are honoured. Communism routes recognition to
the dirty/productive; a war-state to the martial; a religious-pacifist order to the
devotional and lets the martial starve socially.

The killer consequence, which is *emergent not authored*: **the pacifist can't field
soldiers** — not because a rule forbids war, but because military labour's social
prerequisite is never met, so no one will be a warrior, so the role goes unfilled. The
society has, through its values, rendered itself unable to produce people willing to die
for it. That is *enacting* pacifism, not *representing* it (a disabled attack button).

Three disciplines to keep it correct, not mushy:

- **NOT a named menu.** "Communist/fascist/pacifist" as selectable options is the mushy
  version and beneath what this builds. Model the *space* (recognition-policy as a
  distribution over which tasks/strata accrue standing); named ideologies are merely
  legible *regions* of it (like named regions of a colour wheel), surfaced for UI only.
  Players drift into configurations, build hybrids, or find degenerate ones never named.
  Politics emerges from *where in the space* a society sits.
- **Stickiness / held claims.** Ideology must cost, or it's just a menu you re-pick for
  local optimum. Recognition once issued is a *held claim*: strata you've elevated
  become constituencies that resist redistribution of esteem. Withdrawing recognition is
  the politically expensive, destabilising act. Same Tainter-flavoured rigidity as
  hegemony: the structure that solved one era's problem defeats you in the next. The
  war-state cannot easily become the trading republic because the warriors won't have it.
- **Ethics as tragic structure, not a morality meter.** The mechanics are value-neutral;
  they just propagate consequences. Ethics emerges from living inside your configuration
  *next to others*. The pacifist society may be internally stable, materially content,
  genuinely a better place to be a little human — and get annihilated by the neighbour
  who immiserated his own population to feed an honoured war-machine. The game doesn't
  editorialise; it runs both ledgers forward and lets you watch the better society lose,
  or survive only by compromising what made it better.

## Surplus, NOT points/money

Reject a single fungible "points" currency generated by surplus and spendable across
education/military/influence/tech. It is a **converter between ledgers**, and the
converter dissolves the incommensurability that is the entire point — it smuggles the
named-menu back in through the treasury and makes fascist and commune just different
shopping baskets bought with the same money. Embedding money as a first-class starting
concept is precisely how the genre writes "capitalism is the natural state" into the
base layer.

- **Surplus is real, not a score.** It is the material ledger's un-consumed remainder —
  grain not eaten, labour not spent on subsistence. Physical, perishable, located. You
  *direct* it; the cost is always **opportunity cost in the material ledger**, never an
  abstract spend.
- **Education** = people *not* producing because they teach/learn instead. Cost = foregone
  present labour + consumed surplus. Effect: lowers labour-cost of processes and the cost
  of advancing tech (you still choose which tech paths). It is a *bet* that future
  productivity outruns present cost — that bet *is* the game, and a points abstraction
  would hide it.
- **THE EDUCATION TRAP (keystone).** Education buys *material* capability only. Fielding
  an army is gated behind the *social* ledger. So pouring surplus into education in
  pursuit of soldiers gets you the equipment and not the will: well-fed, well-tooled,
  educated people who won't fight, versus a poorer/dumber neighbour whose social ledger
  honours warriors. Investing in the wrong ledger for your goal is the thesis in
  miniature, and it burns the player legibly (schoolhouses full, barracks empty).
- **Influence** (the multiplayer ally-attractor) is NOT bought either — it is an
  *emergent reading of your configuration* (material surplus + an attractive
  recognition-policy + low spectacle-gap project influence). You *are* it or you aren't.
- **Money** is a *late, emergent technology*, not a starting balance. It is what happens
  when perishable/located/incommensurable surpluses need exchange and someone invents a
  token that *stands in* for value — itself already a kind of spectacle (a representation
  mediating real production, letting value detach from physical backing). Its arrival
  switches on the value/price layer, enables speculation, and lets surplus finally escape
  perishability and locality to become hoardable *capital*. A momentous tech-tree event;
  make the player invent it and make inventing it change the game's physics.

**General rule:** every time the design is offered a *fungible abstract quantity that
converts between domains* (social-energy stock, ideology-menu, points/money), refuse the
converter; model the real incommensurable thing and let conversion be impossible or a
momentous invented event. The incommensurability is not a problem to smooth over — it IS
the game. (UI scalars reading distinct gauges: fine and necessary. A single scalar
*spendable across domains*: the trap.)

## Modes of production: one physics, multiple strategies

The genre writes agrarian sedentism into the base layer not through economics but
through **spatiality**: the tile you improve is the tile you stand on, value accrues to
fixed locations you then defend. That installs the agrarian state before any farming
rule is written.

But the per-tile depletion model already *contains* nomadism — no capital-mobility
restructuring needed (an earlier over-reach, withdrawn). Same physics, opposite strategy:

- **Sedentary**: cannot leave, so must keep tiles above their depletion threshold
  indefinitely — fights depletion *in place* (fallow, rotation, managed forestry).
- **Nomadic**: can leave, so draws a tile down and *moves on*, returning only once it has
  naturally regenerated — *outruns* depletion *across space*. Without natural
  regeneration the nomad could never return; the regen function is what *makes* nomadism
  work, not what fights it.

Finitude survives the generalisation and gets stronger: it's not a fact about farming
but about *any* mode of drawing life from finite land. The overgrazing nomad dies as the
no-fallow farmer dies.

**The biome selects the mode**, emergently, via arithmetic not authored rules. Nomadism
trades intensity for extent: more land per person (relying on natural, not managed,
regeneration), so it needs abundant land and reliable regeneration (steppe, savanna) and
is *strictly dominated* by sedentism in a rich tight valley. Extent-vs-intensity also
gives "less contention over land" → nomadism suits the margins, and the historical
friction between sown and steppe falls out the moment two modes share a map.

Military consequence falls out of the two-ledger model with no new machinery: the
pastoralist's mode *is* mobile and martial in one — producer and warrior are the same
honoured competence, so no separate warrior stratum to feed and legitimate. The
sedentary state must *extract* a surplus to feed a *separate* warrior class and then
*socially legitimate* it — two expensive ledger operations the nomad doesn't incur. The
game would *discover* the Mongol advantage, not script it.

**What survives as cheap type-level discipline (NOT the withdrawn restructuring):**

- A herd is an **actor with population-held stock** (travels with the population, grazes/
  breeds/starves) — a modest addition (one more kind of stock-bearer), not a new
  ontology.
- **"Settlement" and "defend your location" are NOT primitives.** Build the root noun as
  *a population organised around a mode of drawing subsistence from the land*; a
  settlement is the *sited* sibling, a band the *mobile* sibling — siblings, not
  parent-and-special-case. Defence is a cost that *sited* capital incurs *because it's
  immobile*, never a universal. In V1 everyone is sedentary and this is invisible, but
  keep the verbs general ("this population draws from these tiles" — true whether it
  stays or moves), so the nomadic strategy is never *framed out* even though the physics
  always permitted it. Cheap now; the retrofit is expensive.
- Honest risk: the nomadic mode's *legibility* may not be worth its build cost for V2 and
  may slide to V3 or "someday." Fine. What's not fine is letting the agrarian base layer
  harden so as to foreclose it.

## Hegemony as complexity load (V2+)

The strong idea, and the solution to the genre's runaway-leader / snowball problem.
Dominance accumulates influence across economic/political/cultural/military registers,
but is wired to Tainter: hegemony is a **complexity load** that costs more to maintain
as it grows, with *declining marginal returns* on each increment. The runaway leader
is therefore the player carrying the heaviest overhead. The empire falls not because
someone out-produces it but because maintenance outruns extraction. This is
self-balancing and supplies late-game pace: spend down your position or be crushed by
its weight. Note: switches on only in the agrarian-and-up regime — foragers have no
surplus, no price layer, no hegemony — which is another reason V1's playable economy
starts at settlement, not at the romantic Ice-Age prologue.

## Conflict: Settlers, not Anno (V2+)

No click-to-attack, no god-view, no power fantasy. You arrange conditions and watch
what your people can do; ultimate powerlessness on the battlefield is the point. Also
far less code than active RTS combat (no contested pathfinding, target selection,
feel-tuning) and a better fit for an agent model — closer to a physics readout than a
control system.

## The value/price layer (later)

Goods have a physical-conservation layer (logs, grain — conserved, legible, Settlers-
style) AND a value/price layer that is emergent, social, non-conserved, and can come
apart from the physical layer. Surplus extraction, speculation, the use-value /
exchange-value gap become *visible game objects*. This maps onto a classification
ontology already worked out elsewhere: physical stock, value, and price are distinct
schemes with separate governance, and their value-equivalence is architecturally
irrelevant. Not V1 — but the V1 data model (stocks as first-class objects) should not
foreclose adding a value layer on top later.

## Architecture: the snapshot as boundary

Engine is a deterministic, seeded, headless Go library. State advances via `Tick()`.
The outside sees state only via an immutable serialisable `Snapshot()`; it acts only
via `Apply(Command)`. The renderer lives on the far side of that contract and the
engine never knows it exists.

Why a boundary but not (yet) a process/network boundary: the engine and renderer have
genuinely different shapes (deterministic/testable vs. soft/taste-driven), so the
separation enforces discipline and keeps rendering concerns from leaking into the
tick. But a long-running-server + client + wire-protocol setup for a single-player
local toy is a distributed-systems problem volunteered for no benefit. So:

- Start with a **clean Go package + narrow API**, called **in-process**. The snapshot
  contract *is* the boundary. Full speed, trivial debugging.
- The **snapshot pattern** is the real load-bearing idea, more than the process split.
  If `Snapshot()` is immutable + serialisable and the renderer can only see snapshots
  and only act via commands, the contract is identical whether the renderer is
  in-process, across a pipe, or in a browser over a websocket. The "where does the
  renderer live" question can be deferred indefinitely because the contract doesn't
  change.
- Serialisable snapshots also hand TDD a free superpower: **golden-file testing of
  whole runs** (see CLAUDE.md). A test is just a renderer that asserts on snapshots
  instead of drawing them.
- Push the real process/network boundary out until a concrete need pulls it in
  (browser client to show people; a JVM/LibGDX renderer; running the sim elsewhere).
  "It feels more properly architected" is not such a need.

There is a real correspondence worth noting once and not labouring: the snapshot is
the engine's *spectacle* — the representation the observer consumes, separated from the
production that generates it, with action back on production only through a constrained
command channel. Architecture and theme specified by the same interface.

## Why Go, why tick-based, why not actors

A single-player economic sim's "agents" are not autonomous processes; they are data
iterated over on a clock. The honest model is a **synchronous tick**: a global clock,
each tick advances the world deterministically, agents read last tick's state and
write this tick's. Determinism is a gift to your future self — reproducible bugs,
trivial save/load, tractable tests. Go fits this exactly; its concurrency is a
*performance tool* (parallelise agent updates within a tick, join, commit), not part
of the model's semantics.

An actor model (Scala + Pekko) was considered and rejected. Actors give async,
nondeterministic message ordering — poison for a game you want to test, balance, save,
and reason about. Actors earn their keep when concurrency is irreducibly part of the
domain (distributed systems); here it is not. There is also a thematic irony: actors
would model each little person as a genuinely autonomous locus of agency, which is
exactly *wrong* for a game whose point is that the people are caught in a system they
don't control. The god-clock is the mechanically honest representation of the thesis.

## Hex coordinates

Hex grid because flow-between-adjacent and spread-across-neighbours are the substrate,
and hex gives **uniform, unambiguous adjacency** (exactly six edge-sharing neighbours,
all equidistant) — no diagonal-distance lie, no edge-or-corner special-casing in every
spatial rule. Internally **axial/cube** coordinates (a hex grid as a diagonal slice of
a 3D cube grid where coordinates sum to zero): neighbours become six vector additions,
distance a single formula. Naive offset-array storage is a trap — it turns
"give me the neighbours" into odd/even-row branching and off-by-one bugs. Offset stays
strictly a display concern. Cube also makes the eventual hydrology ("flow to the lowest
of six uniform neighbours") merely hard rather than miserable. Reference: Red Blob
Games hex grid guide.
