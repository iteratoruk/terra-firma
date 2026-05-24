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

// good is the engine's internal, mutable inert good.
type good struct {
	kind string
	hex  Hex
}

// GoodSnapshot is one good's observable state.
type GoodSnapshot struct {
	Kind string `json:"kind"`
	Q    int    `json:"q"`
	R    int    `json:"r"`
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
