package random

import (
	"reflect"
	"testing"
)

func TestNewReplaysSequence(t *testing.T) {
	first := sample(New("abc -flag"))
	second := sample(New("abc -flag"))

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different sequences: %v != %v", first, second)
	}
}

func TestDifferentSeedsDiverge(t *testing.T) {
	first := sample(New("seed-a"))
	second := sample(New("seed-b"))

	if reflect.DeepEqual(first, second) {
		t.Fatalf("different seeds produced the same sequence: %v", first)
	}
}

func TestSeedInitializesPackageRNG(t *testing.T) {
	Seed("abc -flag")
	first := []int{
		Randint(1, 10),
		Randrange(10),
		RandrangeBetween(3, 9),
	}

	Seed("abc -flag")
	second := []int{
		Randint(1, 10),
		Randrange(10),
		RandrangeBetween(3, 9),
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same package seed produced different sequences: %v != %v", first, second)
	}
}

func TestChoiceSampleAndShuffleAreDeterministic(t *testing.T) {
	firstRNG := New("abc -flag")
	secondRNG := New("abc -flag")
	first := []int{0, 1, 2, 3, 4, 5, 6, 7}
	second := []int{0, 1, 2, 3, 4, 5, 6, 7}

	firstChoice := ChoiceFrom(firstRNG, first)
	firstSample := SampleFrom(firstRNG, first, 3)
	ShuffleFrom(firstRNG, first)

	secondChoice := ChoiceFrom(secondRNG, second)
	secondSample := SampleFrom(secondRNG, second, 3)
	ShuffleFrom(secondRNG, second)

	if firstChoice != secondChoice {
		t.Fatalf("same seed produced different choices: %v != %v", firstChoice, secondChoice)
	}
	if !reflect.DeepEqual(firstSample, secondSample) {
		t.Fatalf("same seed produced different samples: %v != %v", firstSample, secondSample)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different shuffles: %v != %v", first, second)
	}
}

func TestSeedsFromStringMatchesPythonWrapperHashing(t *testing.T) {
	seed1, seed2 := seedsFromString("abc -flag")

	if seed1 != 0x24524a7cec3849f6 || seed2 != 0xed657fb46d6dc53d {
		t.Fatalf("unexpected seeds: %#x %#x", seed1, seed2)
	}
}

func sample(r *RNG) []uint64 {
	return []uint64{
		r.Uint64(),
		uint64(r.Randint(1, 1000)),
		uint64(r.Randrange(1000)),
		uint64(r.RandrangeBetween(100, 200)),
	}
}
