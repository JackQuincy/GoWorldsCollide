// Package vehicle provides vehicle event instructions.
package vehicle

import (
	"fmt"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
)

const (
	alwaysClearEventBit = 0x176
	maxEventBit         = 0x6ff
)

type (
	// Direction is shared with other event instruction packages.
	Direction = event.Direction
	// LoadMapOptions controls map-loading flags.
	LoadMapOptions = event.LoadMapOptions
)

const (
	Up    = event.Up
	Right = event.Right
	Down  = event.Down
	Left  = event.Left
)

// End terminates a vehicle instruction queue.
func End() event.Instruction {
	return event.NewInstruction(0xff)
}

// SetPosition moves the vehicle to world-map coordinates.
func SetPosition(x, y int) event.Instruction {
	return event.NewInstruction(0xc7, x, y)
}

// SetEventBit sets an event bit.
func SetEventBit(eventBit int) event.Instruction {
	validateEventBit(eventBit)
	return event.NewInstruction(0xc8, littleEndian16(eventBit))
}

// ClearEventBit clears an event bit.
func ClearEventBit(eventBit int) event.Instruction {
	validateEventBit(eventBit)
	return event.NewInstruction(0xc9, littleEndian16(eventBit))
}

// BranchIfEventBitClear branches when eventBit is clear.
func BranchIfEventBitClear(eventBit int, destination any) event.Branch {
	validateEventBit(eventBit)
	return event.NewBranch(
		0xb0,
		[]any{littleEndian16(eventBit)},
		destination,
	)
}

// Branch unconditionally branches using the upstream always-clear event bit.
func Branch(destination any) event.Branch {
	return BranchIfEventBitClear(alwaysClearEventBit, destination)
}

// FadeLoadMap loads a map after fading out the screen.
func FadeLoadMap(mapID, x, y int, options LoadMapOptions) event.LoadMap {
	return event.NewLoadMap(0xd2, mapID, x, y, options)
}

// LoadMap loads a map without first fading out the screen.
func LoadMap(mapID, x, y int, options LoadMapOptions) event.LoadMap {
	return event.NewLoadMap(0xd3, mapID, x, y, options)
}

func validateEventBit(eventBit int) {
	if eventBit < 0 || eventBit > maxEventBit {
		panic(fmt.Sprintf(
			"vehicle: event bit %#x outside range [0, %#x]",
			eventBit,
			maxEventBit,
		))
	}
}

func littleEndian16(value int) []byte {
	return []byte{byte(value), byte(value >> 8)}
}
