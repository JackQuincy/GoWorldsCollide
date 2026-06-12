package entity

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
		{name: "end", instruction: End(), want: []byte{0xff}},
		{name: "pause", instruction: Pause(6), want: []byte{0xe0, 0x06}},
		{name: "turn up", instruction: Turn(Up), want: []byte{0xcc}},
		{name: "turn right", instruction: Turn(Right), want: []byte{0xcd}},
		{name: "turn down", instruction: Turn(Down), want: []byte{0xce}},
		{name: "turn left", instruction: Turn(Left), want: []byte{0xcf}},
		{name: "slowest", instruction: SetSpeed(Slowest), want: []byte{0xc0}},
		{name: "slow", instruction: SetSpeed(Slow), want: []byte{0xc1}},
		{name: "normal", instruction: SetSpeed(Normal), want: []byte{0xc2}},
		{name: "fast", instruction: SetSpeed(Fast), want: []byte{0xc3}},
		{name: "fastest", instruction: SetSpeed(Fastest), want: []byte{0xc4}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := writeInstruction(t, test.instruction)
			if !bytes.Equal(got, test.want) {
				t.Fatalf("bytes = % x, want % x", got, test.want)
			}
		})
	}
}

func TestMoveGoldenValues(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		distance  int
		want      byte
	}{
		{name: "up one", direction: Up, distance: 1, want: 0x80},
		{name: "right two", direction: Right, distance: 2, want: 0x85},
		{name: "down seven", direction: Down, distance: 7, want: 0x9a},
		{name: "left eight", direction: Left, distance: 8, want: 0x9f},
		{name: "distance clamps to eight", direction: Up, distance: 9, want: 0x9c},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := writeInstruction(t, Move(test.direction, test.distance))
			if !bytes.Equal(got, []byte{test.want}) {
				t.Fatalf("bytes = % x, want %02x", got, test.want)
			}
		})
	}
}

func TestTurnRejectsInvalidDirection(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected invalid direction panic")
		}
	}()
	Turn(Direction(0xff))
}

func writeInstruction(t *testing.T, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x100, 0x100+instruction.Size()-1, "entity instruction")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}

	result, err := rom.GetBytes(0x100, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	return result
}
