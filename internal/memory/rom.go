package memory

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

const (
	ShortPointerSize = 2
	LongPointerSize  = 3
	ExpandedROMSize  = 4 * 1024 * 1024

	expectedROMSize   = 3 * 1024 * 1024
	expectedROMSHA256 = "0f51b4fca41b7fd509e4b8f9d543151f68efa5e97b08493e4b2a0c06f5d8d5e2"
)

var ErrInvalidROM = errors.New("invalid ROM file")

// ROM stores the mutable bytes of a Worlds Collide input ROM.
type ROM struct {
	data []byte
}

// Load reads, validates, and expands a supported ROM file.
func Load(path string) (*ROM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ROM %q: %w", path, err)
	}
	if !ValidROM(data) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidROM, path)
	}
	return New(data), nil
}

// New copies data and expands it to the size used by Worlds Collide.
//
// New intentionally does not validate data so packages and tests can construct
// small synthetic ROM images. Use Load for user-provided ROM files.
func New(data []byte) *ROM {
	rom := &ROM{data: append([]byte(nil), data...)}
	rom.Expand()
	return rom
}

// ValidROM reports whether data is the supported unheadered FFIII US ROM.
func ValidROM(data []byte) bool {
	if len(data) != expectedROMSize {
		return false
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum) == expectedROMSHA256
}

// Size returns the current ROM size in bytes.
func (r *ROM) Size() int {
	return len(r.data)
}

// Bytes returns a copy of the complete ROM data.
func (r *ROM) Bytes() []byte {
	return append([]byte(nil), r.data...)
}

// Expand extends the ROM to 4 MiB with 0xff bytes. Larger ROMs are unchanged.
func (r *ROM) Expand() {
	if len(r.data) >= ExpandedROMSize {
		return
	}

	originalSize := len(r.data)
	r.data = append(r.data, make([]byte, ExpandedROMSize-originalSize)...)
	for i := originalSize; i < len(r.data); i++ {
		r.data[i] = 0xff
	}
}

// Write writes the complete ROM data to path.
func (r *ROM) Write(path string) error {
	if err := os.WriteFile(path, r.data, 0o644); err != nil {
		return fmt.Errorf("write ROM %q: %w", path, err)
	}
	return nil
}

// GetBits returns the bits selected by mask at address.
func (r *ROM) GetBits(address int, mask byte) (byte, error) {
	value, err := r.GetByte(address)
	if err != nil {
		return 0, err
	}
	return value & mask, nil
}

// GetByte returns the byte at address.
func (r *ROM) GetByte(address int) (byte, error) {
	if err := r.checkRange(address, 1); err != nil {
		return 0, err
	}
	return r.data[address], nil
}

// GetShort returns a little-endian 16-bit value at address.
func (r *ROM) GetShort(address int) (uint16, error) {
	values, err := r.GetBytes(address, ShortPointerSize)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(values), nil
}

// GetBytes returns a copy of count bytes starting at address.
func (r *ROM) GetBytes(address, count int) ([]byte, error) {
	if err := r.checkRange(address, count); err != nil {
		return nil, err
	}
	return append([]byte(nil), r.data[address:address+count]...), nil
}

// GetBytesEndianSwap returns count bytes in reverse order.
func (r *ROM) GetBytesEndianSwap(address, count int) ([]byte, error) {
	values, err := r.GetBytes(address, count)
	if err != nil {
		return nil, err
	}
	reverse(values)
	return values, nil
}

// SetBits replaces the bits selected by mask at address.
func (r *ROM) SetBits(address int, mask, value byte) error {
	if err := r.checkRange(address, 1); err != nil {
		return err
	}
	r.data[address] = (value & mask) | (r.data[address] &^ mask)
	return nil
}

// SetBitNum sets or clears a bit relative to address.
func (r *ROM) SetBitNum(address, bitNumber int, value bool) error {
	if bitNumber < 0 {
		return fmt.Errorf("ROM bit number must be non-negative: %d", bitNumber)
	}

	byteOffset := bitNumber / 8
	bitOffset := uint(bitNumber % 8)
	if err := r.checkRange(address+byteOffset, 1); err != nil {
		return err
	}

	if value {
		r.data[address+byteOffset] |= 1 << bitOffset
	} else {
		r.data[address+byteOffset] &^= 1 << bitOffset
	}
	return nil
}

// SetByte writes value at address.
func (r *ROM) SetByte(address int, value byte) error {
	if err := r.checkRange(address, 1); err != nil {
		return err
	}
	r.data[address] = value
	return nil
}

// SetShort writes a little-endian 16-bit value at address.
func (r *ROM) SetShort(address int, value uint16) error {
	var values [ShortPointerSize]byte
	binary.LittleEndian.PutUint16(values[:], value)
	_, err := r.SetBytes(address, values[:])
	return err
}

// SetBytes writes values at address and returns the address after the write.
func (r *ROM) SetBytes(address int, values []byte) (int, error) {
	if err := r.checkRange(address, len(values)); err != nil {
		return address, err
	}
	copy(r.data[address:], values)
	return address + len(values), nil
}

// SetBytesEndianSwap reverses values before writing them at address.
func (r *ROM) SetBytesEndianSwap(address int, values []byte) (int, error) {
	reversed := append([]byte(nil), values...)
	reverse(reversed)
	return r.SetBytes(address, reversed)
}

func (r *ROM) checkRange(address, count int) error {
	if address < 0 {
		return fmt.Errorf("ROM address must be non-negative: %d", address)
	}
	if count < 0 {
		return fmt.Errorf("ROM byte count must be non-negative: %d", count)
	}
	if address > len(r.data) || count > len(r.data)-address {
		return fmt.Errorf(
			"ROM range [%#x, %#x) exceeds size %#x",
			address,
			address+count,
			len(r.data),
		)
	}
	return nil
}

func reverse(values []byte) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
