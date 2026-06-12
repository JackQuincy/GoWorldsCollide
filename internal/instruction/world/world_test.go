package world

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
		{name: "submerge", instruction: SubmergeFigaroCastle(), want: []byte{0xfd}},
		{name: "emerge", instruction: EmergeFigaroCastle(), want: []byte{0xfe}},
		{name: "entity end", instruction: End(), want: []byte{0xff}},
		{name: "entity turn", instruction: Turn(Down), want: []byte{0xce}},
		{
			name:        "branch if set",
			instruction: BranchIfEventBitSet(0x234, event.CodeStart+0x563412),
			want:        []byte{0xb0, 0x34, 0x82, 0x12, 0x34, 0x56},
		},
		{
			name:        "branch if clear",
			instruction: BranchIfEventBitClear(0x234, event.CodeStart+0x563412),
			want:        []byte{0xb0, 0x34, 0x02, 0x12, 0x34, 0x56},
		},
		{
			name:        "end if set",
			instruction: EndIfEventBitSet(0x234),
			want:        []byte{0xb0, 0x34, 0x82, 0xb4, 0x5e, 0x00},
		},
		{
			name:        "end if clear",
			instruction: EndIfEventBitClear(0x234),
			want:        []byte{0xb0, 0x34, 0x02, 0xb4, 0x5e, 0x00},
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

func TestLoadMapGoldenValues(t *testing.T) {
	options := LoadMapOptions{
		Direction:    Left,
		DefaultMusic: true,
		FadeIn:       true,
		Airship:      true,
	}

	tests := []struct {
		name        string
		instruction memory.Instruction
		opcode      byte
	}{
		{name: "fade", instruction: FadeLoadMap(0x075, 36, 2, options), opcode: 0xd2},
		{name: "load", instruction: LoadMap(0x075, 36, 2, options), opcode: 0xd3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := writeInstruction(t, 0x100, test.instruction)
			want := []byte{test.opcode, 0x75, 0x30, 0x24, 0x02, 0x01}
			if !bytes.Equal(got, want) {
				t.Fatalf("bytes = % x, want % x", got, want)
			}
		})
	}
}

func TestBranchResolvesLabel(t *testing.T) {
	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x0a1000, 0x0a1005, "world branch")
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

func writeInstruction(t *testing.T, address int, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(address, address+instruction.Size()-1, "world instruction")
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
