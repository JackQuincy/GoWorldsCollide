package vehicle

import (
	"bytes"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

func TestInstructionsGoldenValues(t *testing.T) {
	tests := []struct {
		name        string
		instruction memory.Instruction
		want        []byte
	}{
		{name: "end", instruction: End(), want: []byte{0xff}},
		{
			name:        "set position",
			instruction: SetPosition(0xfb, 0xe7),
			want:        []byte{0xc7, 0xfb, 0xe7},
		},
		{
			name:        "set event bit",
			instruction: SetEventBit(0x5ab),
			want:        []byte{0xc8, 0xab, 0x05},
		},
		{
			name:        "clear event bit",
			instruction: ClearEventBit(0x5ab),
			want:        []byte{0xc9, 0xab, 0x05},
		},
		{
			name:        "branch if clear",
			instruction: BranchIfEventBitClear(0x5ab, event.CodeStart+0x563412),
			want:        []byte{0xb0, 0xab, 0x05, 0x12, 0x34, 0x56},
		},
		{
			name:        "unconditional branch",
			instruction: Branch(event.CodeStart + 0x563412),
			want:        []byte{0xb0, 0x76, 0x01, 0x12, 0x34, 0x56},
		},
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

func TestBranchResolvesLabelRelativeToEventCodeStart(t *testing.T) {
	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x0a1000, 0x0a1005, "vehicle branch")
	if err != nil {
		t.Fatal(err)
	}

	instruction := BranchIfEventBitClear(0x234, "destination")
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}
	if err := space.AddLabel("destination", 0x0a5678); err != nil {
		t.Fatal(err)
	}

	got, err := rom.GetBytes(0x0a1000, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xb0, 0x34, 0x02, 0x78, 0x56, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = % x, want % x", got, want)
	}
}

func TestLoadMapGoldenValues(t *testing.T) {
	options := LoadMapOptions{
		Direction:       Down,
		DefaultMusic:    true,
		UpdateParentMap: true,
	}

	tests := []struct {
		name        string
		instruction memory.Instruction
		wantOpcode  byte
	}{
		{
			name:        "fade load map",
			instruction: FadeLoadMap(0x15d, 61, 13, options),
			wantOpcode:  0xd2,
		},
		{
			name:        "load map",
			instruction: LoadMap(0x15d, 61, 13, options),
			wantOpcode:  0xd3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := writeInstruction(t, 0x100, test.instruction)
			want := []byte{test.wantOpcode, 0x5d, 0x23, 0x3d, 0x0d, 0x40}
			if !bytes.Equal(got, want) {
				t.Fatalf("bytes = % x, want % x", got, want)
			}
		})
	}
}

func TestEventBitBoundaries(t *testing.T) {
	for _, eventBit := range []int{0, maxEventBit} {
		t.Run("valid", func(t *testing.T) {
			writeInstruction(t, 0x100, SetEventBit(eventBit))
		})
	}

	for _, eventBit := range []int{-1, maxEventBit + 1} {
		t.Run("invalid", func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("event bit %#x did not panic", eventBit)
				}
			}()
			SetEventBit(eventBit)
		})
	}
}

func writeInstruction(t *testing.T, address int, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(
		address,
		address+instruction.Size()-1,
		"vehicle instruction",
	)
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
