package engine

// Population — an actor (DESIGN.md taxonomy): a self-locating thing with
// autonomous tick-dynamics. The dynamic here is biological: the subsistence
// reserve depletes by metabolism each tick under its own rule, regardless of
// player input. This slice (#5) introduces the cliff; refill (#6) and death
// (#7) build on top of it.
//
// Naming discipline (CLAUDE.md "modes of production"): the type is Population,
// NOT Settlement. A population *is at a tile*; it does not own territory. The
// nomadic/sedentary distinction must remain expressible.
//
// Metabolism is a stored baseline (V1: one rate per individual). The
// labour-as-energy mechanic from DESIGN.md (heavier work draws faster) will
// later extend this rule's domain; here it is baseline draw-down with no work.

// PopulationSpec is the construction-time description of one population.
// StarvationLimit is the number of consecutive zero-reserve ticks the
// population can survive; once the streak passes the limit, death (#7). Per-
// population in V1; a future "resilience" tech would replace the stored value
// with a rule (same shape as the speed rule — see [[feedback_tech_modifiable_rates]]).
type PopulationSpec struct {
	Hex             Hex
	Reserve         int
	Metabolism      int
	StarvationLimit int
}

// population is the engine's internal, mutable population.
// starvationTicks counts consecutive ticks ending at zero reserve; it resets
// to zero the moment the reserve climbs back above zero (recovery is real).
type population struct {
	hex             Hex
	reserve         int
	metabolism      int
	starvationLimit int
	starvationTicks int
}

// PopulationSnapshot is one population's observable state. StarvationTicks is
// exposed BEFORE death (issue #7 legibility requirement — death is not allowed
// to surprise). A dead population is an absence from this slice, never a flag.
type PopulationSnapshot struct {
	Q               int `json:"q"`
	R               int `json:"r"`
	Reserve         int `json:"reserve"`
	Metabolism      int `json:"metabolism"`
	StarvationTicks int `json:"starvation_ticks"`
	StarvationLimit int `json:"starvation_limit"`
}

// lessPopulation is a total order on populations used to make snapshot
// iteration deterministic regardless of construction order.
func lessPopulation(a, b *population) bool {
	if a.hex.Q != b.hex.Q {
		return a.hex.Q < b.hex.Q
	}
	if a.hex.R != b.hex.R {
		return a.hex.R < b.hex.R
	}
	if a.metabolism != b.metabolism {
		return a.metabolism < b.metabolism
	}
	return a.reserve < b.reserve
}

func sortPopulations(ps []*population) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0 && lessPopulation(ps[j], ps[j-1]); j-- {
			ps[j], ps[j-1] = ps[j-1], ps[j]
		}
	}
}
