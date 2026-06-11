package utils

import "github.com/JackQuincy/GoWorldsCollide/internal/random"

// ShuffleIf shuffles only the elements of values for which condition returns
// true, using the package-level project RNG.
func ShuffleIf[T any](values []T, condition func(T) bool) {
	ShuffleIfFrom(shuffleFunc(random.Shuffle[int]), values, condition)
}

// ShuffleIfFrom shuffles only the elements of values for which condition
// returns true, using r.
func ShuffleIfFrom[T any](r Shuffler, values []T, condition func(T) bool) {
	indices := make([]int, 0, len(values))
	elements := make([]T, 0, len(values))
	for i, value := range values {
		if condition(value) {
			indices = append(indices, i)
			elements = append(elements, value)
		}
	}

	if len(indices) == 0 {
		panic("utils: shuffle_if with no matching elements")
	}

	r.Shuffle(indices)
	for i, element := range elements {
		values[indices[i]] = element
	}
}

// Shuffler is the random behavior required by ShuffleIfFrom.
type Shuffler interface {
	Shuffle([]int)
}

type shuffleFunc func([]int)

func (f shuffleFunc) Shuffle(values []int) {
	f(values)
}
