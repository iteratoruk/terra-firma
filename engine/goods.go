package engine

// Inert good — one of the three modes of being (DESIGN.md taxonomy). It has a
// location but no autonomous tick-dynamics: Tick() does nothing to it. Goods
// change only when labour acts on them (a later slice). A good's location is
// a property of the good itself, NOT of the tile — goods sit on the world
// alongside tiles, not inside them.

// GoodSpec is the construction-time description of one inert good.
type GoodSpec struct {
	Kind string
	Hex  Hex
}

// good is the engine's internal, mutable inert good. When holder is non-nil
// the good is "held" by that carrier and its location is *derived* from the
// carrier's position rather than read from hex. Holding does not promote a
// good to a new mode of being — it is still inert; only its locator changes.
type good struct {
	kind   string
	hex    Hex
	holder *carrier
}

// GoodSnapshot is one good's observable state. Held distinguishes a carried
// good from a free one even when both are at the same hex (the discriminator
// that makes the relation observable to a renderer).
type GoodSnapshot struct {
	Kind string `json:"kind"`
	Q    int    `json:"q"`
	R    int    `json:"r"`
	Held bool   `json:"held"`
}

// PickUp links a carrier to a co-located free good. Identifying both by hex
// keeps the command honest about intent ("this carrier picks up that good")
// without hiding the precondition that they must be on the same tile. If
// either is missing, or the good is already held, or they are on different
// tiles, the command is a no-op — the world is unchanged.
type PickUp struct {
	Carrier Hex
	Good    Hex
}

func (p PickUp) apply(w *World) {
	if p.Carrier != p.Good {
		return
	}
	var c *carrier
	for _, x := range w.carriers {
		if x.hex == p.Carrier {
			c = x
			break
		}
	}
	if c == nil {
		return
	}
	for _, g := range w.goods {
		if g.holder == nil && g.hex == p.Good {
			g.holder = c
			return
		}
	}
}

// Drop is the inverse of PickUp: it unlinks the good held by the carrier at
// the given hex and places it back on the world at the carrier's current
// position. V1 carriers hold at most one good, so the command identifies what
// to drop only by *who is dropping it*; a multi-item carrier will need a
// richer identifier. If no carrier is at the hex or it isn't holding anything,
// the command is a no-op.
type Drop struct {
	Carrier Hex
}

func (d Drop) apply(w *World) {
	var c *carrier
	for _, x := range w.carriers {
		if x.hex == d.Carrier {
			c = x
			break
		}
	}
	if c == nil {
		return
	}
	for _, g := range w.goods {
		if g.holder == c {
			g.hex = c.hex
			g.holder = nil
			return
		}
	}
}

// calorieValue is the rule that returns the calorie value of one good of the
// given kind. V1: "grain" is 5; every other kind is 0 (inert as far as eating
// is concerned). Future food variety extends the rule's domain — never a
// per-instance field on a good (see [[feedback_tech_modifiable_rates]] and the
// speed() precedent in carrier.go).
func calorieValue(kind string) int {
	if kind == "grain" {
		return 5
	}
	return 0
}

// lessGood is a total order on goods used to make snapshot iteration
// deterministic regardless of construction order.
func lessGood(a, b *good) bool {
	if a.hex.Q != b.hex.Q {
		return a.hex.Q < b.hex.Q
	}
	if a.hex.R != b.hex.R {
		return a.hex.R < b.hex.R
	}
	return a.kind < b.kind
}

func sortGoods(gs []*good) {
	for i := 1; i < len(gs); i++ {
		for j := i; j > 0 && lessGood(gs[j], gs[j-1]); j-- {
			gs[j], gs[j-1] = gs[j-1], gs[j]
		}
	}
}
