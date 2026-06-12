// Package event provides the shared encoding foundations for event
// instructions.
package event

import (
	"fmt"
	"reflect"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

const CodeStart = 0x0a0000

// Direction is the direction the party faces after loading a map.
type Direction uint8

const (
	Up Direction = iota
	Right
	Down
	Left
)

// Instruction is an event opcode followed by recursively flattened arguments.
type Instruction struct {
	opcode byte
	args   []any
}

// NewInstruction returns an event instruction with flattened arguments.
func NewInstruction(opcode byte, args ...any) Instruction {
	return Instruction{
		opcode: opcode,
		args:   flatten(args),
	}
}

// Encode implements memory.Instruction.
func (i Instruction) Encode(_ *memory.Space) ([]any, error) {
	result := make([]any, 0, 1+len(i.args))
	result = append(result, i.opcode)
	result = append(result, i.args...)
	return result, nil
}

// Size implements memory.Instruction.
func (i Instruction) Size() int {
	return 1 + len(i.args)
}

// Branch is an event instruction whose destinations are stored as offsets from
// the beginning of event code.
type Branch struct {
	opcode       byte
	args         []any
	destinations []any
}

// NewBranch returns a branch instruction. Each destination must be an absolute
// ROM address or a label name.
func NewBranch(opcode byte, args []any, destinations ...any) Branch {
	return Branch{
		opcode:       opcode,
		args:         flatten(args),
		destinations: append([]any(nil), destinations...),
	}
}

// Encode implements memory.Instruction.
func (b Branch) Encode(space *memory.Space) ([]any, error) {
	result := make([]any, 0, 1+len(b.args)+len(b.destinations))
	result = append(result, b.opcode)
	for _, argument := range b.args {
		if label, ok := argument.(string); ok {
			pointer, err := eventLabel(space, label)
			if err != nil {
				return nil, err
			}
			result = append(result, pointer)
			continue
		}
		result = append(result, argument)
	}

	for _, destination := range b.destinations {
		switch value := destination.(type) {
		case string:
			pointer, err := eventLabel(space, value)
			if err != nil {
				return nil, err
			}
			result = append(result, pointer)
		case int:
			encoded, err := eventAddress(value)
			if err != nil {
				return nil, err
			}
			result = append(result, encoded)
		default:
			return nil, fmt.Errorf(
				"event branch destination must be an address or label, got %T",
				destination,
			)
		}
	}
	return result, nil
}

// Size implements memory.Instruction.
func (b Branch) Size() int {
	size := 1 + 3*len(b.destinations)
	for _, argument := range b.args {
		if _, ok := argument.(string); ok {
			size += 3
		} else {
			size++
		}
	}
	return size
}

// LoadMapOptions controls the flags packed by a map-loading instruction.
type LoadMapOptions struct {
	Direction       Direction
	DefaultMusic    bool
	FadeIn          bool
	EntranceEvent   bool
	Airship         bool
	Chocobo         bool
	UpdateParentMap bool
	Unknown         bool
}

// LoadMap is the shared representation used by field, world, and vehicle map
// loading instructions.
type LoadMap struct {
	Instruction
	MapID int
	X     int
	Y     int
}

// NewLoadMap packs the map, direction, music, position, and transport flags.
func NewLoadMap(opcode byte, mapID, x, y int, options LoadMapOptions) LoadMap {
	mapDirectionMusic := mapID
	switch options.Direction {
	case Right:
		mapDirectionMusic |= 0x1000
	case Down:
		mapDirectionMusic |= 0x2000
	case Left:
		mapDirectionMusic |= 0x3000
	}
	if options.Unknown {
		mapDirectionMusic |= 0x0800
	}
	if !options.DefaultMusic {
		mapDirectionMusic |= 0x0400
	}
	if options.UpdateParentMap {
		mapDirectionMusic |= 0x0200
	}

	flags := 0
	if !options.FadeIn {
		flags |= 0x40
	}
	if options.EntranceEvent {
		flags |= 0x80
	}
	if options.Airship {
		flags |= 0x01
	}
	if options.Chocobo {
		flags |= 0x02
	}

	return LoadMap{
		Instruction: NewInstruction(
			opcode,
			mapDirectionMusic&0xff,
			(mapDirectionMusic>>8)&0xff,
			x,
			y,
			flags,
		),
		MapID: mapID,
		X:     x,
		Y:     y,
	}
}

func eventAddress(address int) ([]byte, error) {
	offset := address - CodeStart
	if offset < 0 || offset > 0xffffff {
		return nil, fmt.Errorf(
			"event address %#x is outside encodable range [%#x, %#x]",
			address,
			CodeStart,
			CodeStart+0xffffff,
		)
	}
	return []byte{byte(offset), byte(offset >> 8), byte(offset >> 16)}, nil
}

func eventLabel(space *memory.Space, name string) (*memory.LabelPointer, error) {
	if space == nil {
		return nil, fmt.Errorf("event branch label %q requires a memory space", name)
	}
	return space.LabelAddress24(name).Sub(CodeStart), nil
}

func flatten(values any) []any {
	result := make([]any, 0)
	appendFlattened(&result, values)
	return result
}

func appendFlattened(destination *[]any, values any) {
	if values == nil {
		*destination = append(*destination, nil)
		return
	}

	value := reflect.ValueOf(values)
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			appendFlattened(destination, value.Index(index).Interface())
		}
	default:
		*destination = append(*destination, values)
	}
}
