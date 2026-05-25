package engine

// This file is the engine root. It deliberately knows nothing about rendering,
// pixels, or I/O. The outside world observes state ONLY through Snapshot() and
// changes it ONLY through Command — see CLAUDE.md invariants.

// Command is the only way an external actor changes the world. Applied at a tick
// boundary. V1 has no commands yet; the type exists so the seam is in the right
// place from the start.
type Command interface {
	apply(*World)
}

// TileSpec is the configuration for one tile's single resource at world setup.
// (V1: one resource per tile. The model will generalise to several stocks per
// tile later; the spec stays the construction-time description.)
//
// TransitCondition is one of the determinants of carrier speed (see carrier.go
// and the speed() rule). V1 vocabulary: "unimproved". Defaults to "unimproved"
// when unset so existing tile configurations keep their meaning; later infra
// (paved road, etc.) will require explicit setting.
type TileSpec struct {
	Hex              Hex
	Resource         string
	Value            int
	Capacity         int
	Regen            int
	Harvest          int
	TransitCondition string
}

// Config is the full deterministic description of a world's initial state.
// Same Config => same world, forever.
type Config struct {
	Seed        int64
	Tiles       []TileSpec
	Goods       []GoodSpec
	Carriers    []CarrierSpec
	Populations []PopulationSpec
}

// tile is the engine's internal, mutable tile. A tile is a location that bears a
// resource process; it is NOT an actor and has no autonomous mobility. Resources
// here are tile-bound (see DESIGN.md taxonomy).
type tile struct {
	hex      Hex
	resource string
	stock    *Stock
	transit  string
}

// World is the authoritative, mutable simulation state. It is a value the engine
// owns; nothing outside mutates it directly.
//
// consumed counts the goods that have left the world by being eaten (the only
// way a good leaves the world in V1 — moving and dropping are mode changes, not
// destructions). It generalises the #4 conservation invariant:
// free + held + consumed == initial, for any kind of good.
type World struct {
	tick        uint64
	rng         *RNG
	tiles       []*tile       // kept in canonical hex order for deterministic iteration
	goods       []*good       // kept in canonical order, same reason
	carriers    []*carrier    // kept in canonical order, same reason
	populations []*population // kept in canonical order, same reason
	consumed    int
}

// NewWorld builds a world from a Config. Tiles are sorted into canonical hex
// order so that every iteration over them is deterministic, satisfying the
// no-order-dependent-state invariant.
func NewWorld(cfg Config) *World {
	tiles := make([]*tile, 0, len(cfg.Tiles))
	for _, ts := range cfg.Tiles {
		s := NewStock(ts.Value, ts.Capacity)
		s.SetRegen(ts.Regen)
		s.SetHarvest(ts.Harvest)
		transit := ts.TransitCondition
		if transit == "" {
			transit = "unimproved"
		}
		tiles = append(tiles, &tile{hex: ts.Hex, resource: ts.Resource, stock: s, transit: transit})
	}
	// Deterministic order independent of how the caller listed the tiles.
	sortTiles(tiles)
	goods := make([]*good, 0, len(cfg.Goods))
	for _, gs := range cfg.Goods {
		goods = append(goods, &good{kind: gs.Kind, hex: gs.Hex})
	}
	sortGoods(goods)
	carriers := make([]*carrier, 0, len(cfg.Carriers))
	for _, cs := range cfg.Carriers {
		var dest *Hex
		if cs.Destination != nil {
			d := *cs.Destination // copy: we own a private value
			dest = &d
		}
		carriers = append(carriers, &carrier{typ: cs.Type, hex: cs.Hex, destination: dest})
	}
	sortCarriers(carriers)
	populations := make([]*population, 0, len(cfg.Populations))
	for _, ps := range cfg.Populations {
		populations = append(populations, &population{
			hex:             ps.Hex,
			reserve:         ps.Reserve,
			metabolism:      ps.Metabolism,
			starvationLimit: ps.StarvationLimit,
		})
	}
	sortPopulations(populations)
	return &World{
		tick:        0,
		rng:         NewRNG(cfg.Seed),
		tiles:       tiles,
		goods:       goods,
		carriers:    carriers,
		populations: populations,
	}
}

// Tick advances the world by one step. Order of operations is fixed and
// deterministic. V1:
//   - every tile's resource stock steps (regen minus harvest);
//   - every carrier with a destination advances toward it by speed() tiles;
//   - every population eats one co-located free edible good if any is available,
//     refilling its reserve by the good's calorie value (#6);
//   - every population's subsistence reserve falls by its metabolism, floored
//     at zero (the cliff: death yet to come in #7);
//   - inert goods (w.goods) are deliberately NOT stepped — inertness is what
//     makes them inert. They change only when labour acts on them (later slice).
func (w *World) Tick() {
	for _, t := range w.tiles {
		t.stock.Step()
	}
	for _, c := range w.carriers {
		w.stepCarrier(c)
	}
	for _, p := range w.populations {
		w.populationEat(p)
	}
	for _, p := range w.populations {
		p.reserve -= p.metabolism
		if p.reserve < 0 {
			p.reserve = 0
		}
	}
	w.updateStarvation()
	w.tick++
}

// updateStarvation runs the cliff at the end of each tick (after eat-then-
// metabolise): a population at zero reserve accumulates a streak; one above
// zero resets to zero (recovery is real); a streak that has passed the limit
// removes the population from the world. Death is an absence from the
// populations slice — no "dead" flag, no ghost.
func (w *World) updateStarvation() {
	alive := make([]*population, 0, len(w.populations))
	for _, p := range w.populations {
		if p.reserve == 0 {
			p.starvationTicks++
		} else {
			p.starvationTicks = 0
		}
		if p.starvationTicks > p.starvationLimit {
			continue
		}
		alive = append(alive, p)
	}
	w.populations = alive
}

// populationEat finds an edible, free, co-located good and consumes it: the
// good is removed from the world and the population's reserve grows by the
// good's calorie value. At most one per tick per population; saturation is a
// follow-up (see issue #6 notes).
func (w *World) populationEat(p *population) {
	for i, g := range w.goods {
		if g.holder != nil {
			continue
		}
		if g.hex != p.hex {
			continue
		}
		cv := calorieValue(g.kind)
		if cv <= 0 {
			continue
		}
		p.reserve += cv
		w.goods = append(w.goods[:i], w.goods[i+1:]...)
		w.consumed++
		return
	}
}

// stepCarrier advances one carrier by speed(carrier, currentTile) hexes toward
// its destination, clearing the destination on arrival. Each step picks the
// first neighbour in canonical direction order that reduces hex distance to
// the destination — deterministic, and ambiguity-free for the V1 collinear
// paths the carrier story uses.
func (w *World) stepCarrier(c *carrier) {
	if c.destination == nil {
		return
	}
	t := w.tileAt(c.hex)
	if t == nil {
		return // off-map; the carrier story keeps carriers on tiled hexes
	}
	for i, s := 0, speed(c, t); i < s; i++ {
		if c.hex == *c.destination {
			break
		}
		currDist := c.hex.Distance(*c.destination)
		for _, d := range hexDirections {
			next := c.hex.Add(d)
			if next.Distance(*c.destination) < currDist {
				c.hex = next
				break
			}
		}
	}
	if c.hex == *c.destination {
		c.destination = nil
	}
}

// tileAt returns the tile at h, or nil if no tile exists there. Linear scan is
// fine at V1 sizes; canonical-sort lets us upgrade to a binary search later if
// it ever shows up in a profile.
func (w *World) tileAt(h Hex) *tile {
	for _, t := range w.tiles {
		if t.hex == h {
			return t
		}
	}
	return nil
}

// Apply submits a command at the current tick boundary.
func (w *World) Apply(c Command) { c.apply(w) }

// --- Snapshot: the only window in ---

// Snapshot is an immutable, serialisable view of the world for observers
// (renderers, tests, dashboards). It carries levels AND rates, because trend is
// a legibility requirement. It shares no mutable state with the World.
type Snapshot struct {
	Tick        uint64               `json:"tick"`
	Tiles       []TileSnapshot       `json:"tiles"`
	Goods       []GoodSnapshot       `json:"goods"`
	Carriers    []CarrierSnapshot    `json:"carriers"`
	Populations []PopulationSnapshot `json:"populations"`
}

// TileSnapshot is one tile's observable state. Net is included precomputed so a
// renderer never has to reach back into the engine to show a trend.
// TransitCondition is exposed so an observer can explain a carrier's speed
// without reaching back into the engine ("slow because the road is mud").
type TileSnapshot struct {
	Q                int    `json:"q"`
	R                int    `json:"r"`
	Resource         string `json:"resource"`
	TransitCondition string `json:"transit_condition"`
	Value            int    `json:"value"`
	Capacity         int    `json:"capacity"`
	Regen            int    `json:"regen"`
	Harvest          int    `json:"harvest"`
	Net              int    `json:"net"`
}

// Snapshot produces the immutable view. Tiles come out in canonical hex order so
// the serialised form is stable (essential for golden-file tests).
func (w *World) Snapshot() Snapshot {
	out := Snapshot{
		Tick:        w.tick,
		Tiles:       make([]TileSnapshot, 0, len(w.tiles)),
		Goods:       make([]GoodSnapshot, 0, len(w.goods)),
		Carriers:    make([]CarrierSnapshot, 0, len(w.carriers)),
		Populations: make([]PopulationSnapshot, 0, len(w.populations)),
	}
	for _, t := range w.tiles {
		out.Tiles = append(out.Tiles, TileSnapshot{
			Q:                t.hex.Q,
			R:                t.hex.R,
			Resource:         t.resource,
			TransitCondition: t.transit,
			Value:            t.stock.Value(),
			Capacity:         t.stock.Capacity(),
			Regen:            t.stock.Regen(),
			Harvest:          t.stock.Harvest(),
			Net:              t.stock.Net(),
		})
	}
	for _, g := range w.goods {
		h := g.hex
		held := false
		if g.holder != nil {
			h = g.holder.hex
			held = true
		}
		out.Goods = append(out.Goods, GoodSnapshot{
			Kind: g.kind,
			Q:    h.Q,
			R:    h.R,
			Held: held,
		})
	}
	for _, c := range w.carriers {
		var dest *Hex
		if c.destination != nil {
			d := *c.destination // copy: snapshot shares no mutable state with the world
			dest = &d
		}
		out.Carriers = append(out.Carriers, CarrierSnapshot{
			Type:        c.typ,
			Q:           c.hex.Q,
			R:           c.hex.R,
			Destination: dest,
		})
	}
	for _, p := range w.populations {
		out.Populations = append(out.Populations, PopulationSnapshot{
			Q:               p.hex.Q,
			R:               p.hex.R,
			Reserve:         p.reserve,
			Metabolism:      p.metabolism,
			StarvationTicks: p.starvationTicks,
			StarvationLimit: p.starvationLimit,
		})
	}
	return out
}

func sortTiles(ts []*tile) {
	// Insertion by canonical hex order; small N, stable, deterministic.
	for i := 1; i < len(ts); i++ {
		for j := i; j > 0 && LessHex(ts[j].hex, ts[j-1].hex); j-- {
			ts[j], ts[j-1] = ts[j-1], ts[j]
		}
	}
}
