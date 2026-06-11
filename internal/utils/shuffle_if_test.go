package utils

import (
	"reflect"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/random"
)

func TestShuffleIfOnlyMovesMatchingElements(t *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6}
	original := append([]int(nil), values...)
	r := random.New("shuffle-if")

	ShuffleIfFrom(shuffleFunc(func(indices []int) {
		random.ShuffleFrom(r, indices)
	}), values, func(value int) bool {
		return value%2 == 0
	})

	for i, value := range values {
		if original[i]%2 != 0 && value != original[i] {
			t.Fatalf("non-matching value at index %d changed from %d to %d", i, original[i], value)
		}
	}

	gotEven := matching(values, func(value int) bool { return value%2 == 0 })
	wantEven := []int{2, 4, 6}
	if !sameElements(gotEven, wantEven) {
		t.Fatalf("matching elements changed: got %v, want elements %v", gotEven, wantEven)
	}
}

func TestShuffleIfIsDeterministic(t *testing.T) {
	first := []int{0, 1, 2, 3, 4, 5, 6, 7}
	second := append([]int(nil), first...)
	firstRNG := random.New("shuffle-if deterministic")
	secondRNG := random.New("shuffle-if deterministic")

	ShuffleIfFrom(shuffleFunc(func(indices []int) {
		random.ShuffleFrom(firstRNG, indices)
	}), first, func(value int) bool { return value%2 == 0 })
	ShuffleIfFrom(shuffleFunc(func(indices []int) {
		random.ShuffleFrom(secondRNG, indices)
	}), second, func(value int) bool { return value%2 == 0 })

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different results: %v != %v", first, second)
	}
}

func TestShuffleIfRejectsNoMatches(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	values := []int{1, 3, 5}
	r := random.New("no matches")
	ShuffleIfFrom(shuffleFunc(func(indices []int) {
		random.ShuffleFrom(r, indices)
	}), values, func(value int) bool { return value%2 == 0 })
}

func matching[T any](values []T, condition func(T) bool) []T {
	result := make([]T, 0, len(values))
	for _, value := range values {
		if condition(value) {
			result = append(result, value)
		}
	}
	return result
}

func sameElements(values, expected []int) bool {
	counts := make(map[int]int, len(expected))
	for _, value := range expected {
		counts[value]++
	}
	for _, value := range values {
		counts[value]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return len(values) == len(expected)
}
