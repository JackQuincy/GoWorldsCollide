package memory

import "fmt"

const (
	BankSize  = 0x10000
	BankCount = ExpandedROMSize / BankSize
)

// Heaps stores one free-space heap for each ROM bank.
type Heaps struct {
	banks [BankCount]Heap
}

// NewDefaultHeaps returns heaps initialized with the free-space map from the
// Worlds Collide worlds-divided branch.
func NewDefaultHeaps() *Heaps {
	heaps := &Heaps{}
	for _, block := range defaultFreeBlocks {
		if err := heaps.Free(block.Start, block.End); err != nil {
			panic(err)
		}
	}
	for bank := 0x30; bank < 0x40; bank++ {
		start := bank * BankSize
		if err := heaps.Free(start, start+BankSize-1); err != nil {
			panic(err)
		}
	}
	return heaps
}

// Heap returns the heap for bank.
func (h *Heaps) Heap(bank int) (*Heap, error) {
	if bank < 0 || bank >= len(h.banks) {
		return nil, fmt.Errorf("ROM bank out of range: %#x", bank)
	}
	return &h.banks[bank], nil
}

// Free adds an inclusive free range. Ranges may not cross bank boundaries.
func (h *Heaps) Free(start, end int) error {
	block := NewBlock(start, end)
	bank, err := bankForRange(block)
	if err != nil {
		return err
	}
	h.banks[bank].Free(block.Start, block.End)
	return nil
}

// Reserve removes an inclusive range from its bank heap.
func (h *Heaps) Reserve(start, end int) error {
	block := NewBlock(start, end)
	bank, err := bankForRange(block)
	if err != nil {
		return err
	}
	h.banks[bank].Reserve(block.Start, block.End)
	return nil
}

// Allocate reserves size bytes from bank and returns the start address.
func (h *Heaps) Allocate(bank, size int) (int, error) {
	heap, err := h.Heap(bank)
	if err != nil {
		return 0, err
	}
	return heap.Allocate(size)
}

// Available returns the total free bytes across all banks.
func (h *Heaps) Available() int {
	total := 0
	for i := range h.banks {
		total += h.banks[i].Available()
	}
	return total
}

// BlockCount returns the total number of free blocks across all banks.
func (h *Heaps) BlockCount() int {
	total := 0
	for i := range h.banks {
		total += len(h.banks[i].blocks)
	}
	return total
}

func bankForRange(block Block) (int, error) {
	if block.Start < 0 || block.End >= ExpandedROMSize {
		return 0, fmt.Errorf(
			"ROM range [%#x, %#x] outside expanded ROM",
			block.Start,
			block.End,
		)
	}
	startBank := block.Start / BankSize
	endBank := block.End / BankSize
	if startBank != endBank {
		return 0, fmt.Errorf(
			"ROM range [%#x, %#x] crosses bank boundary",
			block.Start,
			block.End,
		)
	}
	return startBank, nil
}

var defaultFreeBlocks = []Block{
	{Start: 0x00d613, End: 0x00df9f},
	{Start: 0x00ec20, End: 0x00ee9f},
	{Start: 0x00fcab, End: 0x00fcff},
	{Start: 0x00ff18, End: 0x00ffaf},
	{Start: 0x01ffe5, End: 0x01ffff},
	{Start: 0x026469, End: 0x0267ff},
	{Start: 0x02a65a, End: 0x02a7ff},
	{Start: 0x02faa4, End: 0x02fc6c},
	{Start: 0x03f091, End: 0x03ffff},
	{Start: 0x04a4c0, End: 0x04b9ff},
	{Start: 0x04bfb9, End: 0x04c007},
	{Start: 0x04f1c2, End: 0x04f476},
	{Start: 0x04ff72, End: 0x04ffff},
	{Start: 0x09fcec, End: 0x09fdff},
	{Start: 0x0eefbb, End: 0x0ef0ff},
	{Start: 0x0ef463, End: 0x0ef5ff},
	{Start: 0x0efee0, End: 0x0effff},
	{Start: 0x0f3bae, End: 0x0f3c3f},
	{Start: 0x0f3c9b, End: 0x0f3cff},
	{Start: 0x0f83c0, End: 0x0f83ff},
	{Start: 0x0fcf50, End: 0x0fd0cf},
	{Start: 0x0ffb29, End: 0x0ffbff},
	{Start: 0x0ffce0, End: 0x0ffcff},
	{Start: 0x0ffda0, End: 0x0ffdff},
	{Start: 0x0fff47, End: 0x0fff9d},
	{Start: 0x0fffbe, End: 0x0fffff},
	{Start: 0x1095e6, End: 0x1097ff},
	{Start: 0x10cf4a, End: 0x10cfff},
	{Start: 0x10fc7a, End: 0x10fcff},
	{Start: 0x11e989, End: 0x11ead7},
	{Start: 0x11f751, End: 0x11f79f},
	{Start: 0x11f9d0, End: 0x11f9ff},
	{Start: 0x126f6f, End: 0x126fff},
	{Start: 0x12b224, End: 0x12b2ff},
	{Start: 0x12eb44, End: 0x12ebff},
	{Start: 0x14c998, End: 0x14c9ff},
	{Start: 0x14cf5b, End: 0x14cfff},
	{Start: 0x14f646, End: 0x14ffff},
	{Start: 0x186f29, End: 0x186fff},
	{Start: 0x18ce51, End: 0x18ce9f},
	{Start: 0x18dcd2, End: 0x18dcff},
	{Start: 0x18e7b1, End: 0x18e7ff},
	{Start: 0x18ee47, End: 0x18efff},
	{Start: 0x199a51, End: 0x199d4a},
	{Start: 0x19a569, End: 0x19a7ff},
	{Start: 0x19cc4b, End: 0x19cd0f},
	{Start: 0x1fb3d4, End: 0x1fb3ff},
	{Start: 0x1fbae4, End: 0x1fbaff},
	{Start: 0x1fd978, End: 0x1fd9ff},
	{Start: 0x26cd3d, End: 0x26cd5f},
	{Start: 0x26f198, End: 0x26f1ff},
	{Start: 0x26f440, End: 0x26f48f},
	{Start: 0x2962c1, End: 0x2962ff},
	{Start: 0x2ce200, End: 0x2ce3bf},
	{Start: 0x2d63e0, End: 0x2d63ff},
	{Start: 0x2d7787, End: 0x2d779f},
	{Start: 0x2d8bca, End: 0x2d8e5a},
	{Start: 0x2d8e9b, End: 0x2d8eff},
	{Start: 0x2dfcaa, End: 0x2dfdff},
	{Start: 0x2eaf01, End: 0x2eb1ff},
	{Start: 0x2ffbc8, End: 0x2ffeef},

	{Start: 0x0a4363, End: 0x0a48bf},
	{Start: 0x0a48e3, End: 0x0a533e},
	{Start: 0x0a6629, End: 0x0a6785},
	{Start: 0x0a75ee, End: 0x0a7673},
	{Start: 0x0a83c0, End: 0x0a8467},
	{Start: 0x0a8842, End: 0x0a89ae},
	{Start: 0x0a8d27, End: 0x0a8ee4},
	{Start: 0x0a9749, End: 0x0a9d0f},
	{Start: 0x0ac3c7, End: 0x0ac5c0},
	{Start: 0x0ade64, End: 0x0ae3d5},
	{Start: 0x0afab8, End: 0x0affef},
	{Start: 0x0b0080, End: 0x0b03f5},
	{Start: 0x0b094e, End: 0x0b0a1b},
	{Start: 0x0b0f2f, End: 0x0b1031},
	{Start: 0x0b1b14, End: 0x0b1e4c},
	{Start: 0x0b22bb, End: 0x0b2378},
	{Start: 0x0b39de, End: 0x0b3dca},
	{Start: 0x0b75d6, End: 0x0b77c7},
	{Start: 0x0ba0ec, End: 0x0ba37d},
	{Start: 0x0bba0c, End: 0x0bbec3},
	{Start: 0x0bbfe9, End: 0x0bc026},
	{Start: 0x0bc228, End: 0x0bc5fa},
	{Start: 0x0bc730, End: 0x0bc84c},
	{Start: 0x0bd982, End: 0x0bdcb2},
	{Start: 0x0bdcce, End: 0x0be5ca},
	{Start: 0x0bea65, End: 0x0bec91},
	{Start: 0x0bf168, End: 0x0bf295},
	{Start: 0x0bf2b5, End: 0x0bff6f},
	{Start: 0x0c0000, End: 0x0c0976},
	{Start: 0x0c1a66, End: 0x0c1e3f},
	{Start: 0x0c1f9f, End: 0x0c2047},
	{Start: 0x0c2bf0, End: 0x0c3296},
	{Start: 0x0c3971, End: 0x0c3af7},
	{Start: 0x0c4ced, End: 0x0c5029},
	{Start: 0x0c6150, End: 0x0c62a5},
	{Start: 0x0c6a2e, End: 0x0c6ce3},
	{Start: 0x0c6f8c, End: 0x0c704e},
	{Start: 0x0c73e1, End: 0x0c7564},
	{Start: 0x0c7a85, End: 0x0c7ec3},
	{Start: 0x0c985b, End: 0x0c9a4e},
	{Start: 0x0c9b1d, End: 0x0c9ef1},
	{Start: 0x0cbd05, End: 0x0cc1b2},
}
