package memory

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestManagerReserveAndAllocate(t *testing.T) {
	manager := NewManager(New(nil))
	if err := manager.Free(0x012000, 0x0120ff); err != nil {
		t.Fatal(err)
	}

	reserved, err := manager.Reserve(0x012040, 0x01207f, "reserved")
	if err != nil {
		t.Fatal(err)
	}
	if reserved.Size() != 0x40 {
		t.Fatalf("reserved size = %d, want 64", reserved.Size())
	}

	allocated, err := manager.Allocate(0x01, 0x40, "allocated")
	if err != nil {
		t.Fatal(err)
	}
	if allocated.StartAddress != 0x012000 {
		t.Fatalf("allocated start = %#x, want 0x012000", allocated.StartAddress)
	}
}

func TestManagerRejectsOverlappingSpaces(t *testing.T) {
	manager := NewManager(New(nil))
	if _, err := manager.Reserve(0x100, 0x1ff, "first"); err != nil {
		t.Fatal(err)
	}

	_, err := manager.Reserve(0x180, 0x200, "second")
	if !errors.Is(err, ErrSpaceConflict) {
		t.Fatalf("error = %v, want ErrSpaceConflict", err)
	}
}

func TestManagerOrdersSpaces(t *testing.T) {
	manager := NewManager(New(nil))
	second, err := manager.Reserve(0x200, 0x20f, "second")
	if err != nil {
		t.Fatal(err)
	}
	first, err := manager.Reserve(0x100, 0x10f, "first")
	if err != nil {
		t.Fatal(err)
	}

	if got := manager.Spaces(); !reflect.DeepEqual(got, []*Space{first, second}) {
		t.Fatalf("spaces are not ordered: %v", got)
	}
}

func TestSpaceWriteFlattensValues(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "nested values")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Write(1, []byte{2, 3}, []any{4, []int{5, 6}}); err != nil {
		t.Fatal(err)
	}

	got, err := manager.ROM.GetBytes(0x100, 6)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("bytes = %v, want [1 2 3 4 5 6]", got)
	}
}

func TestSpaceWriteRejectsOverflowWithoutWriting(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x101, "small")
	if err != nil {
		t.Fatal(err)
	}

	err = space.Write([]byte{1, 2, 3})
	if !errors.Is(err, ErrSpaceOverflow) {
		t.Fatalf("error = %v, want ErrSpaceOverflow", err)
	}

	got, err := manager.ROM.GetBytes(0x100, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0xff, 0xff}) {
		t.Fatalf("overflow write changed ROM: %x", got)
	}
}

func TestSpaceClearAndCopyFrom(t *testing.T) {
	manager := NewManager(New([]byte{1, 2, 3, 4}))
	space, err := manager.Reserve(0x100, 0x107, "clear")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Clear([]byte{0xaa, 0xbb}); err != nil {
		t.Fatal(err)
	}
	got, err := manager.ROM.GetBytes(0x100, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0xaa, 0xbb, 0xaa, 0xbb, 0xaa, 0xbb, 0xaa, 0xbb}) {
		t.Fatalf("clear bytes = %x", got)
	}

	if err := space.CopyFrom(0, 3); err != nil {
		t.Fatal(err)
	}
	got, err = manager.ROM.GetBytes(0x100, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{1, 2, 3, 4}) {
		t.Fatalf("copied bytes = %x", got)
	}
}

func TestSpaceResolvesForwardAbsolutePointer(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "labels")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Write(space.LabelAddress16("TARGET"), 0xaa, "TARGET", 0xbb); err != nil {
		t.Fatal(err)
	}

	got, err := manager.ROM.GetBytes(0x100, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0x03, 0x01, 0xaa, 0xbb}) {
		t.Fatalf("bytes = %x, want 0301aabb", got)
	}
}

func TestSpaceResolvesForwardBranchPointer(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "branch")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Write(space.BranchDistance("TARGET"), 0xaa, 0xbb, "TARGET", 0xcc); err != nil {
		t.Fatal(err)
	}

	got, err := manager.ROM.GetBytes(0x100, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{0x02, 0xaa, 0xbb, 0xcc}) {
		t.Fatalf("bytes = %x, want 02aabbcc", got)
	}
}

func TestSpaceUsesExistingLabel(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "existing label")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.AddLabel("TARGET", 0x108); err != nil {
		t.Fatal(err)
	}
	if err := space.Write(space.LabelDistance("TARGET").Sub(1)); err != nil {
		t.Fatal(err)
	}

	got, err := manager.ROM.GetByte(0x100)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("byte = %d, want 7", got)
	}
}

func TestSpaceRejectsDuplicateLabel(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "duplicate")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Write("TARGET"); err != nil {
		t.Fatal(err)
	}
	if err := space.Write("TARGET"); err == nil {
		t.Fatal("expected duplicate label error")
	}
}

func TestSpaceSNESAddresses(t *testing.T) {
	manager := NewManager(New(nil))
	space, err := manager.Reserve(0x1234, 0x123f, "snes")
	if err != nil {
		t.Fatal(err)
	}

	if space.StartAddressSNES() != 0xc01234 ||
		space.NextAddressSNES() != 0xc01234 ||
		space.EndAddressSNES() != 0xc0123f {
		t.Fatal("unexpected SNES address conversion")
	}
}
