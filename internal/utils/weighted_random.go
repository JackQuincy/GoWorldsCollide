package utils

import "github.com/JackQuincy/GoWorldsCollide/internal/random"

type Weight interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// WeightedRandom returns the index of a randomly chosen weight using the
// package-level project RNG.
func WeightedRandom[T Weight](weights []T) int {
	return WeightedRandomFrom(randomDefault{}, weights)
}

// WeightedRandomFrom returns the index of a randomly chosen weight using r.
func WeightedRandomFrom[T Weight](r RandomFloat64, weights []T) int {
	if len(weights) == 0 {
		panic("utils: weighted random from empty weights")
	}

	var total float64
	for _, weight := range weights {
		w := float64(weight)
		if w < 0 {
			panic("utils: weighted random with negative weight")
		}
		total += w
	}
	if total <= 0 {
		panic("utils: weighted random with no positive weights")
	}

	rnd := r.Random() * total
	for i, weight := range weights {
		rnd -= float64(weight)
		if rnd < 0 {
			return i
		}
	}

	return len(weights) - 1
}

// RandomFloat64 is the random behavior required by WeightedRandomFrom.
type RandomFloat64 interface {
	Random() float64
}

type randomDefault struct{}

func (randomDefault) Random() float64 {
	return random.Random()
}
