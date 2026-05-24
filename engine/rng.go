package engine

import "math/rand/v2"

// RNG is the engine's only source of randomness. Everything stochastic in the
// world draws from an injected RNG seeded from Config.Seed, so that the same
// seed reproduces the same world exactly. Never use package-level rand or
// time-based seeds anywhere in the engine.
type RNG struct {
	src *rand.Rand
}

// NewRNG creates a deterministic RNG from a seed. Two RNGs with the same seed
// produce identical sequences.
func NewRNG(seed int64) *RNG {
	// PCG with a fixed second stream constant: seed fully determines output.
	return &RNG{src: rand.New(rand.NewPCG(uint64(seed), 0x9E3779B97F4A7C15))}
}

// IntN returns a deterministic pseudo-random int in [0, n).
func (r *RNG) IntN(n int) int { return r.src.IntN(n) }

// Float64 returns a deterministic pseudo-random float in [0.0, 1.0).
func (r *RNG) Float64() float64 { return r.src.Float64() }
