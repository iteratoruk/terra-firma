package engine

// Carrier — an actor (DESIGN.md taxonomy): has autonomous tick-dynamics; once
// given a destination it moves itself each tick. Speed is NOT a stored field
// on the carrier; it's computed by the speed() rule from the carrier's type
// and the current tile's transit condition. Tech later extends the rule's
// domain, not a scalar (see CLAUDE.md "tech-modifiable rates").

// CarrierSpec is the construction-time description of one carrier. V1
// vocabulary: Type "porter". Destination is optional (nil = no destination).
type CarrierSpec struct {
	Type string
	Hex  Hex
}

// carrier is the engine's internal, mutable carrier.
type carrier struct {
	typ string
	hex Hex
}

// CarrierSnapshot is one carrier's observable state.
type CarrierSnapshot struct {
	Type string `json:"type"`
	Q    int    `json:"q"`
	R    int    `json:"r"`
}

// lessCarrier is a total order on carriers used to make snapshot iteration
// deterministic regardless of construction order.
func lessCarrier(a, b *carrier) bool {
	if a.hex.Q != b.hex.Q {
		return a.hex.Q < b.hex.Q
	}
	if a.hex.R != b.hex.R {
		return a.hex.R < b.hex.R
	}
	return a.typ < b.typ
}

func sortCarriers(cs []*carrier) {
	for i := 1; i < len(cs); i++ {
		for j := i; j > 0 && lessCarrier(cs[j], cs[j-1]); j-- {
			cs[j], cs[j-1] = cs[j-1], cs[j]
		}
	}
}
