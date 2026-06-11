package random

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	rand "math/rand/v2"
)

var defaultRNG *RNG

// RNG wraps the project's pseudorandom generator.
//
// It mirrors WorldsCollide's rng.py API from the worlds-divided branch:
// seed strings are hashed with SHA-256, the first 16 bytes are interpreted as a
// big-endian integer, and the resulting value initializes a PCG-backed
// generator.
type RNG struct {
	r *rand.Rand
}

// Seed initializes the package-level RNG from seedText.
func Seed(seedText string) {
	defaultRNG = New(seedText)
}

// New returns a deterministic RNG initialized from seedText.
func New(seedText string) *RNG {
	seed1, seed2 := seedsFromString(seedText)
	return NewPCG(seed1, seed2)
}

// NewPCG returns a deterministic RNG initialized with two 64-bit PCG seeds.
func NewPCG(seed1, seed2 uint64) *RNG {
	return &RNG{r: rand.New(rand.NewPCG(seed1, seed2))}
}

// Choice returns a random element from seq.
func Choice[T any](seq []T) T {
	return ChoiceFrom(mustDefault(), seq)
}

// Sample returns k unique random elements from population.
func Sample[T any](population []T, k int) []T {
	return SampleFrom(mustDefault(), population, k)
}

// Shuffle pseudo-randomizes lst in place.
func Shuffle[T any](lst []T) {
	ShuffleFrom(mustDefault(), lst)
}

// Randint returns a random integer in the inclusive range [a, b].
func Randint(a, b int) int {
	return mustDefault().Randint(a, b)
}

// Randrange returns a random integer in [0, stop).
func Randrange(stop int) int {
	return mustDefault().Randrange(stop)
}

// RandrangeBetween returns a random integer in [start, stop).
func RandrangeBetween(start, stop int) int {
	return mustDefault().RandrangeBetween(start, stop)
}

// Random returns a random float64 in [0.0, 1.0).
func Random() float64 {
	return mustDefault().Random()
}

// Triangular returns a sample from the triangular distribution.
func Triangular(low, high, mode float64) float64 {
	return mustDefault().Triangular(low, high, mode)
}

// Gauss returns a sample from the normal distribution.
func Gauss(mu, sigma float64) float64 {
	return mustDefault().Gauss(mu, sigma)
}

// Uint64 returns a pseudo-random uint64.
func (r *RNG) Uint64() uint64 {
	return r.r.Uint64()
}

// ChoiceFrom returns a random element from seq using r.
func ChoiceFrom[T any](r *RNG, seq []T) T {
	if len(seq) == 0 {
		panic("random: choice from empty sequence")
	}
	return seq[r.r.IntN(len(seq))]
}

// SampleFrom returns k unique random elements from population using r.
func SampleFrom[T any](r *RNG, population []T, k int) []T {
	if k < 0 || k > len(population) {
		panic("random: sample larger than population or negative")
	}

	indices := r.r.Perm(len(population))[:k]
	result := make([]T, k)
	for i, index := range indices {
		result[i] = population[index]
	}
	return result
}

// ShuffleFrom pseudo-randomizes lst in place using r.
func ShuffleFrom[T any](r *RNG, lst []T) {
	r.r.Shuffle(len(lst), func(i, j int) {
		lst[i], lst[j] = lst[j], lst[i]
	})
}

// Randint returns a random integer in the inclusive range [a, b].
func (r *RNG) Randint(a, b int) int {
	if b < a {
		panic("random: empty range for randint")
	}
	return r.r.IntN(b-a+1) + a
}

// Randrange returns a random integer in [0, stop).
func (r *RNG) Randrange(stop int) int {
	return r.r.IntN(stop)
}

// RandrangeBetween returns a random integer in [start, stop).
func (r *RNG) RandrangeBetween(start, stop int) int {
	if stop <= start {
		panic("random: empty range for randrange")
	}
	return r.r.IntN(stop-start) + start
}

// Random returns a random float64 in [0.0, 1.0).
func (r *RNG) Random() float64 {
	return r.r.Float64()
}

// Triangular returns a sample from the triangular distribution.
func (r *RNG) Triangular(low, high, mode float64) float64 {
	if high < low {
		panic("random: triangular high must be >= low")
	}
	if mode < low || mode > high {
		panic("random: triangular mode outside range")
	}
	if high == low {
		return low
	}

	u := r.r.Float64()
	c := (mode - low) / (high - low)
	if u <= c {
		return low + math.Sqrt(u*(high-low)*(mode-low))
	}
	return high - math.Sqrt((1-u)*(high-low)*(high-mode))
}

// Gauss returns a sample from the normal distribution.
func (r *RNG) Gauss(mu, sigma float64) float64 {
	return mu + sigma*r.r.NormFloat64()
}

func seedsFromString(seedText string) (uint64, uint64) {
	sum := sha256.Sum256([]byte(seedText))
	return binary.BigEndian.Uint64(sum[:8]), binary.BigEndian.Uint64(sum[8:16])
}

func mustDefault() *RNG {
	if defaultRNG == nil {
		panic("random: Seed must be called before using package-level RNG")
	}
	return defaultRNG
}
