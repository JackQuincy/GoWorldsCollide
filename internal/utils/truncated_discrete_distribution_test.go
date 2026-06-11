package utils

import "testing"

func TestTruncatedDiscreteDistributionRetriesOutsideBounds(t *testing.T) {
	minimum := 2
	maximum := 4
	r := &gaussianSequence{values: []float64{1.4, 4.6, 3.2}}

	got := TruncatedDiscreteDistributionFrom(r, 0, 1, &minimum, &maximum)
	if got != 3 {
		t.Fatalf("got %d, want 3", got)
	}
	if r.calls != 3 {
		t.Fatalf("got %d calls, want 3", r.calls)
	}
}

func TestTruncatedDiscreteDistributionUsesPythonRounding(t *testing.T) {
	tests := []struct {
		value float64
		want  int
	}{
		{value: 0.5, want: 0},
		{value: 1.5, want: 2},
		{value: 2.5, want: 2},
		{value: -1.5, want: -2},
	}

	for _, tt := range tests {
		r := &gaussianSequence{values: []float64{tt.value}}
		if got := TruncatedDiscreteDistributionFrom(r, 0, 1, nil, nil); got != tt.want {
			t.Fatalf("round(%v) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestTruncatedDiscreteDistributionRejectsInvertedBounds(t *testing.T) {
	minimum := 5
	maximum := 4

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	TruncatedDiscreteDistributionFrom(
		&gaussianSequence{values: []float64{5}},
		0,
		1,
		&minimum,
		&maximum,
	)
}

type gaussianSequence struct {
	values []float64
	calls  int
}

func (r *gaussianSequence) Gauss(_, _ float64) float64 {
	value := r.values[r.calls]
	r.calls++
	return value
}
