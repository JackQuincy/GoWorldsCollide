package fieldentity

import (
	"bytes"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

func TestInstructionsGoldenValues(t *testing.T) {
	tests := []struct {
		name        string
		instruction memory.Instruction
		want        []byte
	}{
		{name: "diagonal normal order", instruction: MoveDiagonal(Right, 1, Up, 2), want: []byte{0xa4}},
		{name: "diagonal swapped order", instruction: MoveDiagonal(Up, 2, Right, 1), want: []byte{0xa4}},
		{name: "walking on", instruction: EnableWalkingAnimation(), want: []byte{0xc6}},
		{name: "walking off", instruction: DisableWalkingAnimation(), want: []byte{0xc7}},
		{name: "sprite layer", instruction: SetSpriteLayer(2), want: []byte{0xc8, 0x02}},
		{name: "hide", instruction: Hide(), want: []byte{0xd1}},
		{name: "position", instruction: SetPosition(0x12, 0x34), want: []byte{0xd5, 0x12, 0x34}},
		{name: "center", instruction: CenterScreen(), want: []byte{0xd7}},
		{name: "standing front", instruction: AnimateStandingFront(), want: []byte{0x01}},
		{name: "kneeling", instruction: AnimateKneeling(), want: []byte{0x09}},
		{name: "close eyes", instruction: AnimateCloseEyes(), want: []byte{0x13}},
		{name: "attack", instruction: AnimateAttack(), want: []byte{0x0a}},
		{name: "attacked", instruction: AnimateAttacked(), want: []byte{0x0b}},
		{name: "hands up", instruction: AnimateHandsUp(), want: []byte{0x0f}},
		{name: "front hands up", instruction: AnimateFrontHandsUp(), want: []byte{0x16}},
		{name: "front right hand up", instruction: AnimateFrontRightHandUp(), want: []byte{0x19}},
		{name: "surprised", instruction: AnimateSurprised(), want: []byte{0x1f}},
		{name: "head down", instruction: AnimateStandingHeadDown(), want: []byte{0x20}},
		{name: "knocked out", instruction: AnimateKnockedOut(), want: []byte{0x28}},
		{name: "knocked out 2", instruction: AnimateKnockedOut2(), want: []byte{0x29}},
		{name: "low jump", instruction: AnimateLowJump(), want: []byte{0xdc}},
		{name: "high jump", instruction: AnimateHighJump(), want: []byte{0xdd}},
		{name: "finger up", instruction: AnimateFingerUp(), want: []byte{0x24}},
		{name: "finger wag", instruction: AnimateFingerWag(), want: []byte{0x25}},
		{name: "random backward numeric", instruction: RandomlyBranchBackwards(0x12), want: []byte{0xfa, 0x12}},
		{name: "random forward numeric", instruction: RandomlyBranchForwards(0x13), want: []byte{0xfb, 0x13}},
		{name: "backward numeric", instruction: BranchBackwards(0x14), want: []byte{0xfc, 0x14}},
		{name: "forward numeric", instruction: BranchForwards(0x15), want: []byte{0xfd, 0x15}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := writeInstruction(t, 0x100, test.instruction)
			if !bytes.Equal(got, test.want) {
				t.Fatalf("bytes = % x, want % x", got, test.want)
			}
		})
	}
}

func TestBranchDistanceLabels(t *testing.T) {
	tests := []struct {
		name        string
		instruction memory.Instruction
		label       int
		want        []byte
	}{
		{name: "backward", instruction: BranchBackwards("loop"), label: 0x100, want: []byte{0xfc, 0x02}},
		{name: "forward", instruction: BranchForwards("next"), label: 0x106, want: []byte{0xfd, 0x02}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rom := memory.New(nil)
			manager := memory.NewManager(rom)
			space, err := manager.Reserve(0x102, 0x103, "entity branch")
			if err != nil {
				t.Fatal(err)
			}
			if err := space.Write(test.instruction); err != nil {
				t.Fatal(err)
			}
			name := "loop"
			if test.label > 0x102 {
				name = "next"
			}
			if err := space.AddLabel(name, test.label); err != nil {
				t.Fatal(err)
			}
			got, err := rom.GetBytes(0x102, 2)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("bytes = % x, want % x", got, test.want)
			}
		})
	}
}

func TestEntityConstants(t *testing.T) {
	if Camera != 0x30 || Party0 != 0x31 || Party3 != 0x34 {
		t.Fatalf("entity constants = %#x %#x %#x", Camera, Party0, Party3)
	}
}

func writeInstruction(t *testing.T, address int, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(address, address+instruction.Size()-1, "field entity")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}
	result, err := rom.GetBytes(address, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	return result
}
