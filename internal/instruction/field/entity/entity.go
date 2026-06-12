// Package fieldentity provides entity action instructions used by field events.
package fieldentity

import (
	"fmt"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/entity"
	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

const (
	Camera = 0x30 + iota
	Party0
	Party1
	Party2
	Party3
)

type (
	Direction = entity.Direction
	Speed     = entity.Speed
)

const (
	Up    = entity.Up
	Right = entity.Right
	Down  = entity.Down
	Left  = entity.Left

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

func MoveDiagonal(dir1 Direction, dist1 int, dir2 Direction, dist2 int) event.Instruction {
	if dir1 == Up || dir1 == Down {
		dir1, dir2 = dir2, dir1
		dist1, dist2 = dist2, dist1
	}

	key := fmt.Sprintf("%c%c%d%d", directionLetter(dir1), directionLetter(dir2), dist1, dist2)
	opcodes := map[string]byte{
		"ru11": 0xa0, "rd11": 0xa1, "ld11": 0xa2, "lu11": 0xa3,
		"ru12": 0xa4, "ru21": 0xa5, "rd21": 0xa6, "rd12": 0xa7,
		"ld12": 0xa8, "ld21": 0xa9, "lu21": 0xaa, "lu12": 0xab,
	}
	opcode, ok := opcodes[key]
	if !ok {
		panic(fmt.Sprintf("fieldentity: unsupported diagonal movement %q", key))
	}
	return event.NewInstruction(opcode)
}

func EnableWalkingAnimation() event.Instruction  { return event.NewInstruction(0xc6) }
func DisableWalkingAnimation() event.Instruction { return event.NewInstruction(0xc7) }
func SetSpriteLayer(layer int) event.Instruction { return event.NewInstruction(0xc8, layer) }
func Hide() event.Instruction                    { return event.NewInstruction(0xd1) }
func SetPosition(x, y int) event.Instruction     { return event.NewInstruction(0xd5, x, y) }
func CenterScreen() event.Instruction            { return event.NewInstruction(0xd7) }

func AnimateStandingFront() event.Instruction    { return event.NewInstruction(0x01) }
func AnimateKneeling() event.Instruction         { return event.NewInstruction(0x09) }
func AnimateCloseEyes() event.Instruction        { return event.NewInstruction(0x13) }
func AnimateAttack() event.Instruction           { return event.NewInstruction(0x0a) }
func AnimateAttacked() event.Instruction         { return event.NewInstruction(0x0b) }
func AnimateHandsUp() event.Instruction          { return event.NewInstruction(0x0f) }
func AnimateFrontHandsUp() event.Instruction     { return event.NewInstruction(0x16) }
func AnimateFrontRightHandUp() event.Instruction { return event.NewInstruction(0x19) }
func AnimateSurprised() event.Instruction        { return event.NewInstruction(0x1f) }
func AnimateStandingHeadDown() event.Instruction { return event.NewInstruction(0x20) }
func AnimateKnockedOut() event.Instruction       { return event.NewInstruction(0x28) }
func AnimateKnockedOut2() event.Instruction      { return event.NewInstruction(0x29) }
func AnimateLowJump() event.Instruction          { return event.NewInstruction(0xdc) }
func AnimateHighJump() event.Instruction         { return event.NewInstruction(0xdd) }
func AnimateFingerUp() event.Instruction         { return event.NewInstruction(0x24) }
func AnimateFingerWag() event.Instruction        { return event.NewInstruction(0x25) }
func RandomlyBranchBackwards(distance any) BranchDistance {
	return newBranchDistance(0xfa, distance, -1)
}
func RandomlyBranchForwards(distance any) BranchDistance {
	return newBranchDistance(0xfb, distance, 1)
}
func BranchBackwards(distance any) BranchDistance {
	return newBranchDistance(0xfc, distance, -1)
}
func BranchForwards(distance any) BranchDistance {
	return newBranchDistance(0xfd, distance, 1)
}

// BranchDistance is a one-byte absolute-distance entity branch.
type BranchDistance struct {
	opcode   byte
	distance any
	offset   int
}

func newBranchDistance(opcode byte, distance any, offset int) BranchDistance {
	switch distance.(type) {
	case int, string:
	default:
		panic(fmt.Sprintf("fieldentity: branch distance must be int or label, got %T", distance))
	}
	return BranchDistance{opcode: opcode, distance: distance, offset: offset}
}

func (b BranchDistance) Encode(space *memory.Space) ([]any, error) {
	switch distance := b.distance.(type) {
	case int:
		return []any{b.opcode, distance}, nil
	case string:
		if space == nil {
			return nil, fmt.Errorf("fieldentity: label %q requires a memory space", distance)
		}
		return []any{b.opcode, space.AbsoluteDistance(distance).Sub(b.offset)}, nil
	default:
		return nil, fmt.Errorf("fieldentity: invalid branch distance %T", b.distance)
	}
}

func (b BranchDistance) Size() int { return 2 }

func directionLetter(direction Direction) byte {
	switch direction {
	case Up:
		return 'u'
	case Right:
		return 'r'
	case Down:
		return 'd'
	case Left:
		return 'l'
	default:
		panic(fmt.Sprintf("fieldentity: invalid direction %d", direction))
	}
}
