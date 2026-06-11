package memory

import (
	"errors"
	"testing"
)

func TestDefaultHeapsMatchPythonFreeMap(t *testing.T) {
	heaps := NewDefaultHeaps()

	if got, want := heaps.BlockCount(), 119; got != want {
		t.Fatalf("block count = %d, want %d", got, want)
	}
	if got, want := heaps.Available(), 1113686; got != want {
		t.Fatalf("available = %d, want %d", got, want)
	}
}

func TestDefaultHeapsIncludeExpandedBanks(t *testing.T) {
	heaps := NewDefaultHeaps()

	for bank := 0x30; bank < 0x40; bank++ {
		heap, err := heaps.Heap(bank)
		if err != nil {
			t.Fatal(err)
		}
		blocks := heap.Blocks()
		if len(blocks) != 1 {
			t.Fatalf("bank %#x has %d blocks, want 1", bank, len(blocks))
		}
		want := Block{
			Start: bank * BankSize,
			End:   bank*BankSize + BankSize - 1,
		}
		if blocks[0] != want {
			t.Fatalf("bank %#x block = %+v, want %+v", bank, blocks[0], want)
		}
	}
}

func TestHeapsAllocateAndReserveWithinBank(t *testing.T) {
	heaps := &Heaps{}
	if err := heaps.Free(0x012000, 0x0120ff); err != nil {
		t.Fatal(err)
	}
	if err := heaps.Reserve(0x012040, 0x01207f); err != nil {
		t.Fatal(err)
	}

	start, err := heaps.Allocate(0x01, 0x40)
	if err != nil {
		t.Fatal(err)
	}
	if start != 0x012000 {
		t.Fatalf("start = %#x, want 0x012000", start)
	}
}

func TestHeapsRejectCrossBankRanges(t *testing.T) {
	heaps := &Heaps{}
	if err := heaps.Free(0x00fff0, 0x010010); err == nil {
		t.Fatal("expected cross-bank error")
	}
}

func TestHeapsRejectInvalidBank(t *testing.T) {
	heaps := &Heaps{}
	if _, err := heaps.Heap(BankCount); err == nil {
		t.Fatal("expected invalid bank error")
	}
}

func TestDefaultHeapAllocationFailure(t *testing.T) {
	heaps := NewDefaultHeaps()
	_, err := heaps.Allocate(0x00, BankSize)
	if !errors.Is(err, ErrOutOfMemory) {
		t.Fatalf("error = %v, want ErrOutOfMemory", err)
	}
}
