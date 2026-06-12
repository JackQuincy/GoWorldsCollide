// Package world provides world-map event instructions.
package world

import (
	"fmt"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/entity"
	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
)

const (
	alwaysClearEventBit = 0x176
	fieldEndAddress     = 0x0a5eb4
)

type (
	Direction      = event.Direction
	LoadMapOptions = event.LoadMapOptions
	Speed          = entity.Speed
)

const (
	Up    = event.Up
	Right = event.Right
	Down  = event.Down
	Left  = event.Left

	Slowest = entity.Slowest
	Slow    = entity.Slow
	Normal  = entity.Normal
	Fast    = entity.Fast
	Fastest = entity.Fastest
)

func End() event.Instruction            { return entity.End() }
func Pause(units int) event.Instruction { return entity.Pause(units) }
func Move(direction Direction, distance int) event.Instruction {
	return entity.Move(direction, distance)
}
func Turn(direction Direction) event.Instruction { return entity.Turn(direction) }
func SetSpeed(speed Speed) event.Instruction     { return entity.SetSpeed(speed) }

// SubmergeFigaroCastle runs the world-map castle submerge command.
func SubmergeFigaroCastle() event.Instruction {
	return event.NewInstruction(0xfd)
}

// EmergeFigaroCastle runs the world-map castle emerge command.
func EmergeFigaroCastle() event.Instruction {
	return event.NewInstruction(0xfe)
}

// FadeLoadMap loads a map after fading out the screen.
func FadeLoadMap(mapID, x, y int, options LoadMapOptions) event.LoadMap {
	return event.NewLoadMap(0xd2, mapID, x, y, options)
}

// LoadMap loads a map without first fading out the screen.
func LoadMap(mapID, x, y int, options LoadMapOptions) event.LoadMap {
	return event.NewLoadMap(0xd3, mapID, x, y, options)
}

// BranchIfEventBitSet branches when eventBit is set.
func BranchIfEventBitSet(eventBit int, destination any) event.Branch {
	return event.NewBranch(
		0xb0,
		[]any{littleEndian16(eventBit | 0x8000)},
		destination,
	)
}

// BranchIfEventBitClear branches when eventBit is clear.
func BranchIfEventBitClear(eventBit int, destination any) event.Branch {
	return event.NewBranch(
		0xb0,
		[]any{littleEndian16(eventBit)},
		destination,
	)
}

// EndIfEventBitSet ends the event when eventBit is set.
func EndIfEventBitSet(eventBit int) event.Branch {
	return BranchIfEventBitSet(eventBit, fieldEndAddress)
}

// EndIfEventBitClear ends the event when eventBit is clear.
func EndIfEventBitClear(eventBit int) event.Branch {
	return BranchIfEventBitClear(eventBit, fieldEndAddress)
}

// Branch unconditionally branches using the upstream always-clear event bit.
func Branch(destination any) event.Branch {
	return BranchIfEventBitClear(alwaysClearEventBit, destination)
}

func littleEndian16(value int) []byte {
	if value < 0 || value > 0xffff {
		panic(fmt.Sprintf("world: 16-bit value out of range: %#x", value))
	}
	return []byte{byte(value), byte(value >> 8)}
}
