// Package entity provides instructions shared by field entities and world-map
// entities.
package entity

import (
	"fmt"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
)

// Direction is shared with event map-loading instructions.
type Direction = event.Direction

const (
	Up    = event.Up
	Right = event.Right
	Down  = event.Down
	Left  = event.Left
)

// Speed is an entity movement-speed opcode.
type Speed byte

const (
	Slowest Speed = 0xc0
	Slow    Speed = 0xc1
	Normal  Speed = 0xc2
	Fast    Speed = 0xc3
	Fastest Speed = 0xc4
)

// End terminates an entity action queue.
func End() event.Instruction {
	return event.NewInstruction(0xff)
}

// Pause waits for units of four frames.
func Pause(units int) event.Instruction {
	return event.NewInstruction(0xe0, units)
}

// Move moves an entity one to eight tiles in direction.
func Move(direction Direction, distance int) event.Instruction {
	if distance > 8 {
		fmt.Println("Warning: char.move: distance > 8, reducing to 8")
		distance = 8
	}

	opcode := (distance - 1) * 4
	switch direction {
	case Up:
		opcode += 0x80
	case Right:
		opcode += 0x81
	case Down:
		opcode += 0x82
	case Left:
		opcode += 0x83
	}
	return event.NewInstruction(byte(opcode))
}

// Turn changes the direction an entity faces.
func Turn(direction Direction) event.Instruction {
	switch direction {
	case Up:
		return event.NewInstruction(0xcc)
	case Right:
		return event.NewInstruction(0xcd)
	case Down:
		return event.NewInstruction(0xce)
	case Left:
		return event.NewInstruction(0xcf)
	default:
		panic(fmt.Sprintf("entity: invalid turn direction %d", direction))
	}
}

// SetSpeed changes an entity's movement speed.
func SetSpeed(speed Speed) event.Instruction {
	return event.NewInstruction(byte(speed))
}
