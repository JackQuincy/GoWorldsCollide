package battleevent

import (
	"bytes"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

func TestInstructionsGoldenValues(t *testing.T) {
	ExecuteAnimations()
	tests := []struct {
		name        string
		instruction memory.Instruction
		want        []byte
	}{
		{name: "nop", instruction: NOP(), want: []byte{0x02}},
		{name: "end", instruction: End(), want: []byte{0xff}},
		{name: "add target masks high bit", instruction: AddTarget(0x9a), want: []byte{0x13, 0x1a}},
		{name: "remove target sets high bit", instruction: RemoveTarget(0x1a), want: []byte{0x13, 0x9a}},
		{name: "clear animations", instruction: ClearAnimations(), want: []byte{0x0e}},
		{name: "execute animations", instruction: ExecuteAnimations(), want: []byte{0x0f}},
		{name: "open dialog", instruction: OpenMultiLineDialogWindow(), want: []byte{0x11}},
		{name: "close dialog", instruction: CloseMultiLineDialogWindow(), want: []byte{0x10}},
		{name: "single line dialog", instruction: DisplaySingleLineDialog(0x34), want: []byte{0x00, 0x34}},
		{name: "multi line dialog", instruction: DisplayMultiLineDialog(0x56), want: []byte{0x01, 0x56}},
		{name: "checks complete opcode", instruction: IncrementChecksComplete(), want: []byte{0x15}},
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

func TestInstallHandlers(t *testing.T) {
	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	source := []byte{0x11, 0x22, 0x33, 0x44, 0x55}
	if _, err := rom.SetBytes(addRemoveEntitySourceStart, source); err != nil {
		t.Fatal(err)
	}

	addresses, err := InstallHandlers(manager, 0x1fca)
	if err != nil {
		t.Fatal(err)
	}

	addRemove, err := rom.GetBytes(addresses.AddRemoveEntity, len(source))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(addRemove, source) {
		t.Fatalf("add/remove handler = % x, want % x", addRemove, source)
	}

	increment, err := rom.GetBytes(addresses.IncrementChecksComplete, 4)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0xee, 0xca, 0x1f, 0x60}; !bytes.Equal(increment, want) {
		t.Fatalf("increment handler = % x, want % x", increment, want)
	}

	addRemovePointer, err := rom.GetBytes(opcodeTableStart+0x13*2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{byte(addresses.AddRemoveEntity), byte(addresses.AddRemoveEntity >> 8)}; !bytes.Equal(addRemovePointer, want) {
		t.Fatalf("add/remove pointer = % x, want % x", addRemovePointer, want)
	}

	incrementPointer, err := rom.GetBytes(opcodeTableStart+0x15*2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{byte(addresses.IncrementChecksComplete), byte(addresses.IncrementChecksComplete >> 8)}; !bytes.Equal(incrementPointer, want) {
		t.Fatalf("increment pointer = % x, want % x", incrementPointer, want)
	}
}

func TestCharacterAnimationSlotsResetAfterExecute(t *testing.T) {
	ExecuteAnimations()

	first := writeInstruction(t, AddCharacterAnimation(0x05, 0xd0a9fe))
	second := writeInstruction(t, AddCharacterAnimation(0x06, 0xc01234))
	ExecuteAnimations()
	reset := writeInstruction(t, AddCharacterAnimation(0x07, 0xab5678))
	ExecuteAnimations()

	if want := []byte{0x03, 0x05, 0xfe, 0xa9}; !bytes.Equal(first, want) {
		t.Fatalf("first = % x, want % x", first, want)
	}
	if want := []byte{0x04, 0x06, 0x34, 0x12}; !bytes.Equal(second, want) {
		t.Fatalf("second = % x, want % x", second, want)
	}
	if want := []byte{0x03, 0x07, 0x78, 0x56}; !bytes.Equal(reset, want) {
		t.Fatalf("reset = % x, want % x", reset, want)
	}
}

func TestCharacterAnimationSlotLimit(t *testing.T) {
	ExecuteAnimations()
	AddCharacterAnimation(0, 0)
	AddCharacterAnimation(0, 0)
	AddCharacterAnimation(0, 0)
	defer func() {
		ExecuteAnimations()
		if recover() == nil {
			t.Fatal("expected slot limit panic")
		}
	}()
	AddCharacterAnimation(0, 0)
}

func writeInstruction(t *testing.T, instruction memory.Instruction) []byte {
	t.Helper()

	rom := memory.New(nil)
	manager := memory.NewManager(rom)
	space, err := manager.Reserve(0x100, 0x100+instruction.Size()-1, "battle event")
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
