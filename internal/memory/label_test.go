package memory

import (
	"bytes"
	"errors"
	"testing"
)

func TestAbsoluteLabelPointer(t *testing.T) {
	label := NewLabel("TARGET")
	label.Resolve(0x123456)
	pointer := NewLabelPointer(label, Absolute24).Add(2)

	value, err := pointer.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != 0x123458 {
		t.Fatalf("value = %#x, want 0x123458", value)
	}

	data, err := pointer.Bytes(3)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte{0x58, 0x34, 0x12}) {
		t.Fatalf("bytes = %x, want 583412", data)
	}
}

func TestRelativeLabelPointers(t *testing.T) {
	label := NewLabel("TARGET")
	label.Resolve(120)

	relative := NewLabelPointer(label, Relative).At(100)
	value, err := relative.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != 20 {
		t.Fatalf("relative value = %d, want 20", value)
	}

	absoluteRelative := NewLabelPointer(label, AbsoluteRelative).At(150)
	value, err = absoluteRelative.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != 30 {
		t.Fatalf("absolute-relative value = %d, want 30", value)
	}
}

func TestBranchRelativeForwardAndBackward(t *testing.T) {
	tests := []struct {
		name           string
		target, source int
		want           int
	}{
		{name: "forward", target: 110, source: 100, want: 9},
		{name: "backward", target: 90, source: 100, want: 245},
		{name: "same address", target: 100, source: 100, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := NewLabel("TARGET")
			label.Resolve(tt.target)
			pointer := NewLabelPointer(label, BranchRelative).At(tt.source)

			got, err := pointer.Value()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("value = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBranchRelativeRejectsOutOfRange(t *testing.T) {
	label := NewLabel("TARGET")
	label.Resolve(228)
	pointer := NewLabelPointer(label, BranchRelative).At(100)

	if _, err := pointer.Value(); err == nil {
		t.Fatal("expected range error")
	}
}

func TestLabelPointerReportsUnresolvedValues(t *testing.T) {
	pointer := NewLabelPointer(NewLabel("TARGET"), Absolute)

	_, err := pointer.Value()
	if !errors.Is(err, ErrUnresolvedLabel) {
		t.Fatalf("error = %v, want ErrUnresolvedLabel", err)
	}
}

func TestLabelPointerBytesRejectsValuesOutsideWidth(t *testing.T) {
	tests := []struct {
		name   string
		target int
		source int
		mode   PointerMode
		length int
	}{
		{name: "negative", target: 90, source: 100, mode: Relative, length: 1},
		{name: "too large", target: 0x100, mode: Absolute, length: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := NewLabel("TARGET")
			label.Resolve(tt.target)
			pointer := NewLabelPointer(label, tt.mode)
			if tt.mode == Relative {
				pointer.At(tt.source)
			}

			if _, err := pointer.Bytes(tt.length); err == nil {
				t.Fatal("expected encoding error")
			}
		})
	}
}

func TestLabelPointerOffsetOperationsMutatePointer(t *testing.T) {
	label := NewLabel("TARGET")
	label.Resolve(100)
	pointer := NewLabelPointer(label, Absolute)

	if pointer.Add(10) != pointer || pointer.Sub(3) != pointer {
		t.Fatal("offset operation did not return the same pointer")
	}

	value, err := pointer.Value()
	if err != nil {
		t.Fatal(err)
	}
	if value != 107 {
		t.Fatalf("value = %d, want 107", value)
	}
}
