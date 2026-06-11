package memory

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCopiesAndExpandsROM(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03}
	rom := New(input)
	input[0] = 0xff

	if rom.Size() != ExpandedROMSize {
		t.Fatalf("size = %d, want %d", rom.Size(), ExpandedROMSize)
	}

	got, err := rom.GetBytes(0, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x01, 0x02, 0x03, 0xff, 0xff}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = %x, want %x", got, want)
	}
}

func TestNewDoesNotShrinkLargeROM(t *testing.T) {
	input := make([]byte, ExpandedROMSize+1)
	rom := New(input)

	if rom.Size() != len(input) {
		t.Fatalf("size = %d, want %d", rom.Size(), len(input))
	}
}

func TestByteAndBitAccess(t *testing.T) {
	rom := New([]byte{0b10101010, 0, 0})

	bits, err := rom.GetBits(0, 0b11110000)
	if err != nil {
		t.Fatal(err)
	}
	if bits != 0b10100000 {
		t.Fatalf("bits = %08b, want 10100000", bits)
	}

	if err := rom.SetBits(0, 0b00001111, 0b00000101); err != nil {
		t.Fatal(err)
	}
	if err := rom.SetBitNum(1, 12, true); err != nil {
		t.Fatal(err)
	}

	got, err := rom.GetBytes(0, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0b10100101, 0, 0b00010000}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = %08b, want %08b", got, want)
	}

	if err := rom.SetBitNum(1, 12, false); err != nil {
		t.Fatal(err)
	}
	value, err := rom.GetByte(2)
	if err != nil {
		t.Fatal(err)
	}
	if value != 0 {
		t.Fatalf("byte = %#x, want 0", value)
	}
}

func TestShortAccessIsLittleEndian(t *testing.T) {
	rom := New(nil)

	if err := rom.SetShort(10, 0x1234); err != nil {
		t.Fatal(err)
	}

	values, err := rom.GetBytes(10, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(values, []byte{0x34, 0x12}) {
		t.Fatalf("bytes = %x, want 3412", values)
	}

	value, err := rom.GetShort(10)
	if err != nil {
		t.Fatal(err)
	}
	if value != 0x1234 {
		t.Fatalf("short = %#x, want 0x1234", value)
	}
}

func TestEndianSwapAccess(t *testing.T) {
	rom := New([]byte{1, 2, 3, 4})

	values, err := rom.GetBytesEndianSwap(0, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(values, []byte{4, 3, 2, 1}) {
		t.Fatalf("bytes = %v, want [4 3 2 1]", values)
	}

	next, err := rom.SetBytesEndianSwap(4, []byte{5, 6, 7})
	if err != nil {
		t.Fatal(err)
	}
	if next != 7 {
		t.Fatalf("next address = %d, want 7", next)
	}

	written, err := rom.GetBytes(4, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, []byte{7, 6, 5}) {
		t.Fatalf("bytes = %v, want [7 6 5]", written)
	}
}

func TestGetBytesReturnsCopy(t *testing.T) {
	rom := New([]byte{1, 2, 3})

	values, err := rom.GetBytes(0, 3)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = 9

	value, err := rom.GetByte(0)
	if err != nil {
		t.Fatal(err)
	}
	if value != 1 {
		t.Fatalf("ROM changed through returned slice: got %d, want 1", value)
	}
}

func TestSetBytesReturnsNextAddress(t *testing.T) {
	rom := New(nil)

	next, err := rom.SetBytes(20, []byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if next != 23 {
		t.Fatalf("next address = %d, want 23", next)
	}
}

func TestAccessRejectsOutOfBoundsRanges(t *testing.T) {
	rom := New(nil)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "negative address",
			run: func() error {
				_, err := rom.GetByte(-1)
				return err
			},
		},
		{
			name: "read past end",
			run: func() error {
				_, err := rom.GetBytes(ExpandedROMSize-1, 2)
				return err
			},
		},
		{
			name: "write past end",
			run: func() error {
				_, err := rom.SetBytes(ExpandedROMSize, []byte{1})
				return err
			},
		},
		{
			name: "negative bit",
			run: func() error {
				return rom.SetBitNum(0, -1, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestWriteWritesCompleteROM(t *testing.T) {
	rom := New([]byte{1, 2, 3})
	path := filepath.Join(t.TempDir(), "output.smc")

	if err := rom.Write(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, rom.Bytes()) {
		t.Fatal("written file differs from ROM data")
	}
}

func TestLoadRejectsInvalidROM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.smc")
	if err := os.WriteFile(path, []byte("not a ROM"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if !errors.Is(err, ErrInvalidROM) {
		t.Fatalf("error = %v, want ErrInvalidROM", err)
	}
}

func TestValidROMRejectsWrongSize(t *testing.T) {
	if ValidROM(make([]byte, expectedROMSize-1)) {
		t.Fatal("wrong-sized ROM reported valid")
	}
}
