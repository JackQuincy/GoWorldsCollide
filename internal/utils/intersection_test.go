package utils

import (
	"reflect"
	"testing"
)

func TestIntersectionPreservesLeftOrderAndDuplicates(t *testing.T) {
	got := Intersection(
		[]int{4, 2, 4, 1, 3, 2},
		[]int{2, 4},
	)
	want := []int{4, 2, 4, 2}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestIntersectionReturnsEmptySlice(t *testing.T) {
	got := Intersection([]string{"a", "b"}, []string{"c"})
	if got == nil {
		t.Fatal("got nil, want an empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty slice", got)
	}
}
