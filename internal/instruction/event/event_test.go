package event

import (
	"bytes"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

func TestInstructionFlattensArguments(t *testing.T) {
	instruction := NewInstruction(
		0x12,
		1,
		[]byte{2, 3},
		[]any{4, []int{5, 6}},
	)

	if instruction.Size() != 7 {
		t.Fatalf("size = %d, want 7", instruction.Size())
	}

	got := writeInstruction(t, 0x100, instruction)
	want := []byte{0x12, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = % x, want % x", got, want)
	}
}

func TestBranchEncodesAbsoluteDestinations(t *testing.T) {
	instruction := NewBranch(
		0xb0,
		[]any{[]byte{0x34, 0x12}},
		CodeStart+0x563412,
	)

	if instruction.Size() != 6 {
		t.Fatalf("size = %d, want 6", instruction.Size())
	}

	got := writeInstruction(t, 0x100, instruction)
	want := []byte{0xb0, 0x34, 0x12, 0x12, 0x34, 0x56}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = % x, want % x", got, want)
	}
}

func TestBranchResolvesLabelRelativeToEventCodeStart(t *testing.T) {
	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x0a1000, 0x0a1005, "event branch")
	if err != nil {
		t.Fatal(err)
	}

	instruction := NewBranch(0xb2, nil, "destination")
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}
	if err := space.AddLabel("destination", 0x0a3456); err != nil {
		t.Fatal(err)
	}

	got, err := rom.GetBytes(0x0a1000, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xb2, 0x56, 0x34, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = % x, want % x", got, want)
	}
}

func TestBranchResolvesLabelInBaseArguments(t *testing.T) {
	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x0a1000, 0x0a1003, "event branch")
	if err != nil {
		t.Fatal(err)
	}

	instruction := NewBranch(0xb2, []any{"destination"})
	if instruction.Size() != 4 {
		t.Fatalf("size = %d, want 4", instruction.Size())
	}
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}
	if err := space.AddLabel("destination", 0x0a3456); err != nil {
		t.Fatal(err)
	}

	got, err := rom.GetBytes(0x0a1000, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xb2, 0x56, 0x34, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes = % x, want % x", got, want)
	}
}

func TestLoadMapGoldenValues(t *testing.T) {
	tests := []struct {
		name    string
		mapID   int
		x       int
		y       int
		options LoadMapOptions
		want    []byte
	}{
		{
			name:  "defaults",
			mapID: 0x123,
			x:     0x45,
			y:     0x67,
			options: LoadMapOptions{
				Direction:    Right,
				DefaultMusic: true,
				FadeIn:       true,
			},
			want: []byte{0xd3, 0x23, 0x11, 0x45, 0x67, 0x00},
		},
		{
			name:  "all optional flags",
			mapID: 0x0ab,
			x:     0x12,
			y:     0x34,
			options: LoadMapOptions{
				Direction:       Left,
				EntranceEvent:   true,
				Airship:         true,
				Chocobo:         true,
				UpdateParentMap: true,
				Unknown:         true,
			},
			want: []byte{0xd3, 0xab, 0x3e, 0x12, 0x34, 0xc3},
		},
		{
			name:  "unknown direction leaves direction clear",
			mapID: 0x001,
			x:     0x02,
			y:     0x03,
			options: LoadMapOptions{
				Direction:    Direction(0xff),
				DefaultMusic: true,
				FadeIn:       true,
			},
			want: []byte{0xd3, 0x01, 0x00, 0x02, 0x03, 0x00},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			instruction := NewLoadMap(0xd3, test.mapID, test.x, test.y, test.options)
			got := writeInstruction(t, 0x100, instruction)
			if !bytes.Equal(got, test.want) {
				t.Fatalf("bytes = % x, want % x", got, test.want)
			}
		})
	}
}

func TestBranchRejectsAddressBeforeEventCode(t *testing.T) {
	instruction := NewBranch(0xb0, nil, CodeStart-1)
	if _, err := instruction.Encode(nil); err == nil {
		t.Fatal("expected invalid event address error")
	}
}

func writeInstruction(t *testing.T, address int, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(
		address,
		address+instruction.Size()-1,
		"event instruction",
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
