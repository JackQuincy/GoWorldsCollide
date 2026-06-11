package utils

import (
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/random"
)

func TestWeightedRandomIsDeterministic(t *testing.T) {
	first := random.New("weighted")
	second := random.New("weighted")
	weights := []int{25, 15, 3, 1, 0}

	for i := 0; i < 20; i++ {
		got := WeightedRandomFrom(first, weights)
		want := WeightedRandomFrom(second, weights)
		if got != want {
			t.Fatalf("draw %d differs: got %d, want %d", i, got, want)
		}
	}
}

func TestWeightedRandomSkipsZeroWeights(t *testing.T) {
	r := random.New("zero weights")
	weights := []float64{0, 0, 4, 0}

	for i := 0; i < 20; i++ {
		if got := WeightedRandomFrom(r, weights); got != 2 {
			t.Fatalf("got index %d, want 2", got)
		}
	}
}

func TestWeightedRandomRejectsInvalidWeights(t *testing.T) {
	tests := []struct {
		name    string
		weights []float64
	}{
		{name: "empty", weights: nil},
		{name: "all zero", weights: []float64{0, 0}},
		{name: "negative", weights: []float64{1, -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			WeightedRandomFrom(random.New("invalid"), tt.weights)
		})
	}
}
