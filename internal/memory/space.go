package memory

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

const StartAddressSNES = 0xc00000

var (
	ErrSpaceConflict = errors.New("memory space conflicts with existing space")
	ErrSpaceOverflow = errors.New("not enough room in memory space")
)

// Instruction is implemented by values that encode themselves for a Space.
type Instruction interface {
	Encode(space *Space) ([]any, error)
	Size() int
}

// Manager owns a ROM, its free-space heaps, and all reserved spaces.
type Manager struct {
	ROM    *ROM
	Heaps  *Heaps
	spaces []*Space
}

// NewManager returns a memory manager initialized with the default free map.
func NewManager(rom *ROM) *Manager {
	return &Manager{
		ROM:   rom,
		Heaps: NewDefaultHeaps(),
	}
}

// Spaces returns reserved spaces ordered by start address.
func (m *Manager) Spaces() []*Space {
	return append([]*Space(nil), m.spaces...)
}

// Reserve removes [start, end] from free space and returns it for writing.
func (m *Manager) Reserve(start, end int, description string) (*Space, error) {
	block := NewBlock(start, end)
	if err := m.validateSpace(block); err != nil {
		return nil, err
	}
	if err := m.Heaps.Reserve(block.Start, block.End); err != nil {
		return nil, err
	}
	return m.addSpace(block, description), nil
}

// Allocate reserves size bytes from bank using best-fit allocation.
func (m *Manager) Allocate(bank, size int, description string) (*Space, error) {
	start, err := m.Heaps.Allocate(bank, size)
	if err != nil {
		return nil, err
	}
	block := NewBlock(start, start+size-1)
	if err := m.validateSpace(block); err != nil {
		m.Heaps.banks[bank].Free(block.Start, block.End)
		return nil, err
	}
	return m.addSpace(block, description), nil
}

// Free adds an inclusive range to its bank heap.
func (m *Manager) Free(start, end int) error {
	return m.Heaps.Free(start, end)
}

// Read returns the inclusive ROM range [start, end].
func (m *Manager) Read(start, end int) ([]byte, error) {
	block := NewBlock(start, end)
	return m.ROM.GetBytes(block.Start, block.Size())
}

func (m *Manager) validateSpace(block Block) error {
	if m == nil || m.ROM == nil || m.Heaps == nil {
		return errors.New("memory manager is not initialized")
	}
	if _, err := bankForRange(block); err != nil {
		return err
	}

	index := sort.Search(len(m.spaces), func(i int) bool {
		return m.spaces[i].EndAddress >= block.Start
	})
	if index < len(m.spaces) && m.spaces[index].StartAddress <= block.End {
		return fmt.Errorf(
			"%w: [%#06x, %#06x] %q overlaps %s",
			ErrSpaceConflict,
			block.Start,
			block.End,
			"",
			m.spaces[index],
		)
	}
	return nil
}

func (m *Manager) addSpace(block Block, description string) *Space {
	space := &Space{
		manager:      m,
		StartAddress: block.Start,
		EndAddress:   block.End,
		NextAddress:  block.Start,
		Description:  description,
		labels:       make(map[string]*Label),
	}
	index := sort.Search(len(m.spaces), func(i int) bool {
		return m.spaces[i].StartAddress >= space.StartAddress
	})
	m.spaces = append(m.spaces, nil)
	copy(m.spaces[index+1:], m.spaces[index:])
	m.spaces[index] = space
	return space
}

// Space is an inclusive ROM range reserved for a patch or data block.
type Space struct {
	manager *Manager

	StartAddress int
	EndAddress   int
	NextAddress  int
	Description  string

	labels   map[string]*Label
	pointers []*LabelPointer
}

// Size returns the number of bytes in the space.
func (s *Space) Size() int {
	return s.EndAddress - s.StartAddress + 1
}

// StartAddressSNES returns the SNES address corresponding to StartAddress.
func (s *Space) StartAddressSNES() int {
	return s.StartAddress + StartAddressSNES
}

// NextAddressSNES returns the SNES address corresponding to NextAddress.
func (s *Space) NextAddressSNES() int {
	return s.NextAddress + StartAddressSNES
}

// EndAddressSNES returns the SNES address corresponding to EndAddress.
func (s *Space) EndAddressSNES() int {
	return s.EndAddress + StartAddressSNES
}

// Write flattens and writes supported values at the next address.
func (s *Space) Write(values ...any) error {
	data := make([]byte, 0)
	if err := s.appendValues(&data, values); err != nil {
		return err
	}
	if s.NextAddress+len(data)-1 > s.EndAddress {
		return fmt.Errorf(
			"%w %q: next %#x > end %#x",
			ErrSpaceOverflow,
			s.Description,
			s.NextAddress+len(data)-1,
			s.EndAddress,
		)
	}

	if _, err := s.manager.ROM.SetBytes(s.NextAddress, data); err != nil {
		return err
	}
	s.NextAddress += len(data)
	return s.updateLabelPointers()
}

// Clear fills the complete space with a repeating byte pattern and resets the
// next write address.
func (s *Space) Clear(pattern []byte) error {
	if len(pattern) == 0 {
		return errors.New("clear pattern must not be empty")
	}
	if s.Size()%len(pattern) != 0 {
		return fmt.Errorf(
			"clear pattern of size %d does not evenly fill space of size %d",
			len(pattern),
			s.Size(),
		)
	}

	data := make([]byte, s.Size())
	for i := range data {
		data[i] = pattern[i%len(pattern)]
	}
	if _, err := s.manager.ROM.SetBytes(s.StartAddress, data); err != nil {
		return err
	}
	s.NextAddress = s.StartAddress
	return nil
}

// CopyFrom copies the inclusive ROM range [start, end] into the space.
func (s *Space) CopyFrom(start, end int) error {
	data, err := s.manager.Read(start, end)
	if err != nil {
		return err
	}
	return s.Write(data)
}

// AddLabel resolves name to address.
func (s *Space) AddLabel(name string, address int) error {
	if _, exists := s.labels[name]; exists {
		return fmt.Errorf("label %q already exists in %s", name, s)
	}
	label := NewLabel(name)
	label.Resolve(address)
	s.labels[name] = label
	return s.updateLabelPointers()
}

// LabelAddress returns an absolute one-byte label pointer.
func (s *Space) LabelAddress(name string) *LabelPointer {
	return s.newLabelPointer(name, Absolute)
}

// LabelAddress16 returns a 16-bit absolute label pointer.
func (s *Space) LabelAddress16(name string) *LabelPointer {
	return s.newLabelPointer(name, Absolute16)
}

// LabelAddress24 returns a 24-bit absolute label pointer.
func (s *Space) LabelAddress24(name string) *LabelPointer {
	return s.newLabelPointer(name, Absolute24)
}

// LabelDistance returns a relative label pointer.
func (s *Space) LabelDistance(name string) *LabelPointer {
	return s.newLabelPointer(name, Relative)
}

// AbsoluteDistance returns an absolute-distance label pointer.
func (s *Space) AbsoluteDistance(name string) *LabelPointer {
	return s.newLabelPointer(name, AbsoluteRelative)
}

// BranchDistance returns an 8-bit branch-relative label pointer.
func (s *Space) BranchDistance(name string) *LabelPointer {
	return s.newLabelPointer(name, BranchRelative)
}

func (s *Space) newLabelPointer(name string, mode PointerMode) *LabelPointer {
	pointer := NewLabelPointer(NewLabel(name), mode)
	s.pointers = append(s.pointers, pointer)
	return pointer
}

func (s *Space) appendValues(destination *[]byte, values any) error {
	switch value := values.(type) {
	case nil:
		return errors.New("cannot write nil value")
	case string:
		return s.addInlineLabel(value, len(*destination))
	case byte:
		*destination = append(*destination, value)
		return nil
	case int:
		if value < 0 || value > 0xff {
			return fmt.Errorf("byte value out of range: %d", value)
		}
		*destination = append(*destination, byte(value))
		return nil
	case []byte:
		*destination = append(*destination, value...)
		return nil
	case *LabelPointer:
		return s.appendLabelPointer(destination, value)
	case Instruction:
		encoded, err := value.Encode(s)
		if err != nil {
			return err
		}
		return s.appendValues(destination, encoded)
	}

	reflected := reflect.ValueOf(values)
	switch reflected.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflected.Len(); i++ {
			if err := s.appendValues(destination, reflected.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := reflected.Int()
		if integer < 0 || integer > 0xff {
			return fmt.Errorf("byte value out of range: %d", integer)
		}
		*destination = append(*destination, byte(integer))
		return nil
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		integer := reflected.Uint()
		if integer > 0xff {
			return fmt.Errorf("byte value out of range: %d", integer)
		}
		*destination = append(*destination, byte(integer))
		return nil
	default:
		return fmt.Errorf("unsupported memory write value %T", values)
	}
}

func (s *Space) addInlineLabel(name string, offset int) error {
	if _, exists := s.labels[name]; exists {
		return fmt.Errorf("label %q already exists in %s", name, s)
	}
	label := NewLabel(name)
	label.Resolve(s.NextAddress + offset)
	s.labels[name] = label
	return nil
}

func (s *Space) appendLabelPointer(destination *[]byte, pointer *LabelPointer) error {
	pointer.At(s.NextAddress + len(*destination))
	size := pointerSize(pointer.Mode)
	if label, exists := s.labels[pointer.Label.Name]; exists {
		pointer.Label.Resolve(*label.Address)
		data, err := pointer.Bytes(size)
		if err != nil {
			return err
		}
		*destination = append(*destination, data...)
		return nil
	}

	*destination = append(*destination, make([]byte, size)...)
	return nil
}

func (s *Space) updateLabelPointers() error {
	for _, pointer := range s.pointers {
		label, exists := s.labels[pointer.Label.Name]
		if !exists || pointer.Address == nil {
			continue
		}
		pointer.Label.Resolve(*label.Address)
		data, err := pointer.Bytes(pointerSize(pointer.Mode))
		if err != nil {
			return err
		}
		if _, err := s.manager.ROM.SetBytes(*pointer.Address, data); err != nil {
			return err
		}
	}
	return nil
}

func pointerSize(mode PointerMode) int {
	switch mode {
	case Absolute16:
		return 2
	case Absolute24:
		return 3
	default:
		return 1
	}
}

func (s *Space) String() string {
	return fmt.Sprintf(
		"[%#06x - %#06x] %q",
		s.StartAddress,
		s.EndAddress,
		s.Description,
	)
}
