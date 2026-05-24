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
type TileSpec struct {
	Hex      Hex
	Resource string
	Value    int
	Capacity int
	Regen    int
	Harvest  int
}

// Config is the full deterministic description of a world's initial state.
// Same Config => same world, forever.
type Config struct {
	Seed  int64
	Tiles []TileSpec
	Goods []GoodSpec
}

// tile is the engine's internal, mutable tile. A tile is a location that bears a
// resource process; it is NOT an actor and has no autonomous mobility. Resources
// here are tile-bound (see DESIGN.md taxonomy).
type tile struct {
	hex      Hex
	resource string
	stock    *Stock
}

// World is the authoritative, mutable simulation state. It is a value the engine
// owns; nothing outside mutates it directly.
type World struct {
	tick  uint64
	rng   *RNG
	tiles []*tile // kept in canonical hex order for deterministic iteration
	goods []*good // kept in canonical order, same reason
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
		tiles = append(tiles, &tile{hex: ts.Hex, resource: ts.Resource, stock: s})
	}
	// Deterministic order independent of how the caller listed the tiles.
	sortTiles(tiles)
	goods := make([]*good, 0, len(cfg.Goods))
	for _, gs := range cfg.Goods {
		goods = append(goods, &good{kind: gs.Kind, hex: gs.Hex})
	}
	sortGoods(goods)
	return &World{
		tick:  0,
		rng:   NewRNG(cfg.Seed),
		tiles: tiles,
		goods: goods,
	}
}

// Tick advances the world by one step. Order of operations is fixed and
// deterministic. V1: every tile's resource stock steps (regen minus harvest).
func (w *World) Tick() {
	for _, t := range w.tiles {
		t.stock.Step()
	}
	w.tick++
}

// Apply submits a command at the current tick boundary.
func (w *World) Apply(c Command) { c.apply(w) }

// --- Snapshot: the only window in ---

// Snapshot is an immutable, serialisable view of the world for observers
// (renderers, tests, dashboards). It carries levels AND rates, because trend is
// a legibility requirement. It shares no mutable state with the World.
type Snapshot struct {
	Tick  uint64         `json:"tick"`
	Tiles []TileSnapshot `json:"tiles"`
	Goods []GoodSnapshot `json:"goods"`
}

// TileSnapshot is one tile's observable state. Net is included precomputed so a
// renderer never has to reach back into the engine to show a trend.
type TileSnapshot struct {
	Q        int    `json:"q"`
	R        int    `json:"r"`
	Resource string `json:"resource"`
	Value    int    `json:"value"`
	Capacity int    `json:"capacity"`
	Regen    int    `json:"regen"`
	Harvest  int    `json:"harvest"`
	Net      int    `json:"net"`
}

// Snapshot produces the immutable view. Tiles come out in canonical hex order so
// the serialised form is stable (essential for golden-file tests).
func (w *World) Snapshot() Snapshot {
	out := Snapshot{
		Tick:  w.tick,
		Tiles: make([]TileSnapshot, 0, len(w.tiles)),
		Goods: make([]GoodSnapshot, 0, len(w.goods)),
	}
	for _, t := range w.tiles {
		out.Tiles = append(out.Tiles, TileSnapshot{
			Q:        t.hex.Q,
			R:        t.hex.R,
			Resource: t.resource,
			Value:    t.stock.Value(),
			Capacity: t.stock.Capacity(),
			Regen:    t.stock.Regen(),
			Harvest:  t.stock.Harvest(),
			Net:      t.stock.Net(),
		})
	}
	for _, g := range w.goods {
		out.Goods = append(out.Goods, GoodSnapshot{
			Kind: g.kind,
			Q:    g.hex.Q,
			R:    g.hex.R,
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
