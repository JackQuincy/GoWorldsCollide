package memory

import (
	"errors"
	"reflect"
	"testing"
)

func TestBlockNormalizesEndpoints(t *testing.T) {
	block := NewBlock(20, 10)
	if block.Start != 10 || block.End != 20 || block.Size() != 11 {
		t.Fatalf("block = %+v, size %d", block, block.Size())
	}
}

func TestHeapAllocateUsesBestFit(t *testing.T) {
	var heap Heap
	heap.Free(100, 199)
	heap.Free(300, 349)
	heap.Free(500, 574)

	start, err := heap.Allocate(40)
	if err != nil {
		t.Fatal(err)
	}
	if start != 300 {
		t.Fatalf("start = %d, want 300", start)
	}
	if heap.Available() != 185 {
		t.Fatalf("available = %d, want 185", heap.Available())
	}
}

func TestHeapAllocateUsesBlockOrderForEqualFits(t *testing.T) {
	var heap Heap
	heap.Free(100, 109)
	heap.Free(200, 209)

	start, err := heap.Allocate(5)
	if err != nil {
		t.Fatal(err)
	}
	if start != 200 {
		t.Fatalf("start = %d, want 200", start)
	}
}

func TestHeapAllocateRemovesExactBlock(t *testing.T) {
	var heap Heap
	heap.Free(100, 109)

	start, err := heap.Allocate(10)
	if err != nil {
		t.Fatal(err)
	}
	if start != 100 || heap.Available() != 0 || len(heap.Blocks()) != 0 {
		t.Fatalf("start %d, available %d, blocks %v", start, heap.Available(), heap.Blocks())
	}
}

func TestHeapAllocateReportsOutOfMemory(t *testing.T) {
	var heap Heap
	heap.Free(0, 4)

	_, err := heap.Allocate(6)
	if !errors.Is(err, ErrOutOfMemory) {
		t.Fatalf("error = %v, want ErrOutOfMemory", err)
	}
}

func TestHeapFreeMergesAdjacentAndOverlappingBlocks(t *testing.T) {
	var heap Heap
	heap.Free(100, 109)
	heap.Free(120, 129)
	heap.Free(108, 121)

	want := []Block{{Start: 100, End: 129}}
	if got := heap.Blocks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("blocks = %v, want %v", got, want)
	}
	if heap.Available() != 30 {
		t.Fatalf("available = %d, want 30", heap.Available())
	}
}

func TestHeapFreeInsideExistingBlockDoesNothing(t *testing.T) {
	var heap Heap
	heap.Free(100, 199)
	heap.Free(120, 130)

	want := []Block{{Start: 100, End: 199}}
	if got := heap.Blocks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("blocks = %v, want %v", got, want)
	}
}

func TestHeapReserveSplitsBlock(t *testing.T) {
	var heap Heap
	heap.Free(100, 199)
	heap.Reserve(125, 174)

	want := []Block{
		{Start: 100, End: 124},
		{Start: 175, End: 199},
	}
	if got := heap.Blocks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("blocks = %v, want %v", got, want)
	}
	if heap.Available() != 50 {
		t.Fatalf("available = %d, want 50", heap.Available())
	}
}

func TestHeapReserveTrimsAndRemovesBlocks(t *testing.T) {
	var heap Heap
	heap.Free(100, 109)
	heap.Free(200, 219)
	heap.Free(300, 309)
	heap.Reserve(105, 304)

	want := []Block{
		{Start: 305, End: 309},
		{Start: 100, End: 104},
	}
	if got := heap.Blocks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("blocks = %v, want %v", got, want)
	}
	if heap.Available() != 10 {
		t.Fatalf("available = %d, want 10", heap.Available())
	}
}
