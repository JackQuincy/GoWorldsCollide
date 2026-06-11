package memory

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrUnresolvedLabel = errors.New("unresolved label")

// PointerMode controls how a label pointer is resolved.
type PointerMode uint8

const (
	Absolute PointerMode = iota
	Absolute16
	Absolute24
	Relative
	AbsoluteRelative
	BranchRelative
)

// Label is a named address that may be resolved after a pointer is created.
type Label struct {
	Name    string
	Address *int
}

// NewLabel returns an unresolved label.
func NewLabel(name string) *Label {
	return &Label{Name: name}
}

// Resolve sets the label address.
func (l *Label) Resolve(address int) {
	l.Address = intPointer(address)
}

// Resolved reports whether the label has an address.
func (l *Label) Resolved() bool {
	return l != nil && l.Address != nil
}

// LabelPointer references a label from an optional pointer address.
type LabelPointer struct {
	Label   *Label
	Address *int
	Mode    PointerMode
	Offset  int
}

// NewLabelPointer returns an unresolved pointer to label.
func NewLabelPointer(label *Label, mode PointerMode) *LabelPointer {
	return &LabelPointer{Label: label, Mode: mode}
}

// At sets the address of the pointer itself.
func (p *LabelPointer) At(address int) *LabelPointer {
	p.Address = intPointer(address)
	return p
}

// Add adds offset to the pointed-to label address.
func (p *LabelPointer) Add(offset int) *LabelPointer {
	p.Offset += offset
	return p
}

// Sub subtracts offset from the pointed-to label address.
func (p *LabelPointer) Sub(offset int) *LabelPointer {
	p.Offset -= offset
	return p
}

// Value resolves the pointer to its encoded integer value.
func (p *LabelPointer) Value() (int, error) {
	if p == nil || p.Label == nil || !p.Label.Resolved() {
		name := ""
		if p != nil && p.Label != nil {
			name = p.Label.Name
		}
		return 0, fmt.Errorf("%w %q", ErrUnresolvedLabel, name)
	}

	value := *p.Label.Address + p.Offset
	switch p.Mode {
	case Absolute, Absolute16, Absolute24:
		return value, nil
	case Relative:
		address, err := p.pointerAddress()
		if err != nil {
			return 0, err
		}
		return value - address, nil
	case AbsoluteRelative:
		address, err := p.pointerAddress()
		if err != nil {
			return 0, err
		}
		distance := value - address
		if distance < 0 {
			distance = -distance
		}
		return distance, nil
	case BranchRelative:
		address, err := p.pointerAddress()
		if err != nil {
			return 0, err
		}
		distance := value - address
		if distance > 127 || distance < -128 {
			return 0, fmt.Errorf(
				"branch to label %q out of range: %d",
				p.Label.Name,
				distance-1,
			)
		}
		if distance > 0 {
			return distance - 1, nil
		}
		if distance < 0 {
			return distance + 0xff, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown label pointer mode: %d", p.Mode)
	}
}

// Bytes returns the little-endian representation of the pointer value.
func (p *LabelPointer) Bytes(length int) ([]byte, error) {
	value, err := p.Value()
	if err != nil {
		return nil, err
	}
	if length <= 0 || length > 8 {
		return nil, fmt.Errorf("invalid label pointer length: %d", length)
	}
	if value < 0 {
		return nil, fmt.Errorf("negative label pointer value cannot be encoded unsigned: %d", value)
	}
	if length < 8 && uint64(value) >= uint64(1)<<(length*8) {
		return nil, fmt.Errorf(
			"label pointer value %#x does not fit in %d bytes",
			value,
			length,
		)
	}

	result := make([]byte, length)
	var encoded [8]byte
	binary.LittleEndian.PutUint64(encoded[:], uint64(value))
	copy(result, encoded[:length])
	return result, nil
}

func (p *LabelPointer) pointerAddress() (int, error) {
	if p.Address == nil {
		return 0, errors.New("label pointer address is unresolved")
	}
	return *p.Address, nil
}

func intPointer(value int) *int {
	return &value
}
