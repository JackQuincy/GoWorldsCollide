package memory

import (
	"errors"
	"fmt"
)

var ErrOutOfMemory = errors.New("unable to allocate memory block")

// Block is an inclusive address range.
type Block struct {
	Start int
	End   int
}

// NewBlock returns a normalized inclusive address range.
func NewBlock(start, end int) Block {
	if start > end {
		start, end = end, start
	}
	return Block{Start: start, End: end}
}

// Size returns the number of bytes in the block.
func (b Block) Size() int {
	return b.End - b.Start + 1
}

// Heap tracks free blocks within one ROM bank.
type Heap struct {
	blocks    []Block
	available int
}

// Blocks returns a copy of the free blocks in allocator order.
func (h *Heap) Blocks() []Block {
	return append([]Block(nil), h.blocks...)
}

// Available returns the total number of free bytes.
func (h *Heap) Available() int {
	return h.available
}

// Allocate reserves size bytes from the smallest block that can satisfy it.
func (h *Heap) Allocate(size int) (int, error) {
	if size <= 0 {
		return 0, fmt.Errorf("allocation size must be positive: %d", size)
	}

	bestIndex := -1
	bestDifference := h.available
	for i, block := range h.blocks {
		difference := block.Size() - size
		if difference == 0 {
			bestIndex = i
			break
		}
		if difference > 0 && difference < bestDifference {
			bestIndex = i
			bestDifference = difference
		}
	}
	if bestIndex < 0 {
		return 0, fmt.Errorf("%w of size %d", ErrOutOfMemory, size)
	}

	start := h.blocks[bestIndex].Start
	h.blocks[bestIndex].Start += size
	if h.blocks[bestIndex].Size() == 0 {
		h.blocks = append(h.blocks[:bestIndex], h.blocks[bestIndex+1:]...)
	}
	h.available -= size
	return start, nil
}

// Free adds the inclusive range [start, end], merging overlaps and adjacent
// blocks.
func (h *Heap) Free(start, end int) {
	newBlock := NewBlock(start, end)
	overlaps := make([]bool, len(h.blocks))

	for i, block := range h.blocks {
		switch {
		case block.Start >= newBlock.Start && block.Start <= newBlock.End+1:
			if block.End > newBlock.End {
				newBlock.End = block.End
			}
			overlaps[i] = true
		case block.End <= newBlock.End && block.End >= newBlock.Start-1:
			if block.Start < newBlock.Start {
				newBlock.Start = block.Start
			}
			overlaps[i] = true
		case block.Start < newBlock.Start && block.End > newBlock.End:
			return
		}
	}

	newBlocks := make([]Block, 0, len(h.blocks)+1)
	newBlocks = append(newBlocks, newBlock)
	available := newBlock.Size()
	for i, block := range h.blocks {
		if !overlaps[i] {
			newBlocks = append(newBlocks, block)
			available += block.Size()
		}
	}
	h.blocks = newBlocks
	h.available = available
}

// Reserve removes the inclusive range [start, end] from the heap.
func (h *Heap) Reserve(start, end int) {
	reserved := NewBlock(start, end)
	overlaps := make([]bool, len(h.blocks))

	for i := 0; i < len(h.blocks); i++ {
		block := h.blocks[i]
		switch {
		case block.Start >= reserved.Start && block.Start <= reserved.End:
			if block.End > reserved.End {
				h.blocks[i].Start = reserved.End + 1
			} else {
				overlaps[i] = true
			}
		case block.End <= reserved.End && block.End >= reserved.Start:
			if block.Start < reserved.Start {
				h.blocks[i].End = reserved.Start - 1
			} else {
				overlaps[i] = true
			}
		case block.Start < reserved.Start && block.End > reserved.End:
			h.blocks = append(h.blocks, Block{
				Start: reserved.End + 1,
				End:   block.End,
			})
			h.blocks[i].End = reserved.Start - 1
			h.available -= reserved.Size()
			return
		}
	}

	newBlocks := make([]Block, 0, len(h.blocks))
	available := 0
	for i, block := range h.blocks {
		if i >= len(overlaps) || !overlaps[i] {
			newBlocks = append(newBlocks, block)
			available += block.Size()
		}
	}
	h.blocks = newBlocks
	h.available = available
}
