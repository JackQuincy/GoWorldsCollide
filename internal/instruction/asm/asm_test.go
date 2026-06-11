package asm

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

func TestPythonGoldenEncodings(t *testing.T) {
	tests := []struct {
		name        string
		instruction Instruction
		want        []byte
	}{
		{name: "NOP", instruction: NOP(), want: []byte{0xea}},
		{name: "REP", instruction: REP(0x23), want: []byte{0xc2, 0x23}},
		{name: "PEI", instruction: PEI(0x123), want: []byte{0xd4, 0x23}},
		{name: "PER", instruction: PER(0x12345), want: []byte{0x62, 0x45, 0x23}},
		{name: "LDA IMM8", instruction: LDA(0x123, IMM8), want: []byte{0xa9, 0x23}},
		{name: "LDA IMM16", instruction: LDA(0x12345, IMM16), want: []byte{0xa9, 0x45, 0x23}},
		{name: "LDA LNG", instruction: LDA(0x123456, LNG), want: []byte{0xaf, 0x56, 0x34, 0x12}},
		{name: "STA DIR_X_16", instruction: STA(0x123, DIR_X_16), want: []byte{0x81, 0x23}},
		{name: "LDX ABS_Y", instruction: LDX(0x12345, ABS_Y), want: []byte{0xbe, 0x45, 0x23}},
		{name: "INC accumulator", instruction: INC(), want: []byte{0x1a}},
		{name: "INC ABS_X", instruction: INC(0x12345, ABS_X), want: []byte{0xfe, 0x45, 0x23}},
		{name: "AND DIR_24_Y", instruction: AND(0x123, DIR_24_Y), want: []byte{0x37, 0x23}},
		{name: "ASL accumulator", instruction: ASL(), want: []byte{0x0a}},
		{name: "ADC S_16_Y", instruction: ADC(0x123, S_16_Y), want: []byte{0x73, 0x23}},
		{name: "BIT IMM16", instruction: BIT(0x12345, IMM16), want: []byte{0x89, 0x45, 0x23}},
		{name: "CMP LNG_X", instruction: CMP(0x123456, LNG_X), want: []byte{0xdf, 0x56, 0x34, 0x12}},
		{name: "BRA numeric", instruction: BRA(0xfe), want: []byte{0x80, 0xfe}},
		{name: "JMP ABS_24", instruction: JMP(0x12345, ABS_24), want: []byte{0xdc, 0x45, 0x23}},
		{name: "JSL", instruction: JSL(0x123456), want: []byte{0x22, 0x56, 0x34, 0x12}},
		{name: "RTL", instruction: RTL(), want: []byte{0x6b}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encode(t, tt.instruction); !bytes.Equal(got, tt.want) {
				t.Fatalf("encoding = %x, want %x", got, tt.want)
			}
			if tt.instruction.Size() != len(tt.want) {
				t.Fatalf("size = %d, want %d", tt.instruction.Size(), len(tt.want))
			}
		})
	}
}

func TestOpcodeFamilies(t *testing.T) {
	tests := []struct {
		name        string
		instruction Instruction
		opcode      byte
	}{
		{name: "LDY", instruction: LDY(1, ABS_X), opcode: 0xbc},
		{name: "STX", instruction: STX(1, DIR_Y), opcode: 0x96},
		{name: "STY", instruction: STY(1, DIR_X), opcode: 0x94},
		{name: "STZ", instruction: STZ(1, ABS_X), opcode: 0x9e},
		{name: "DEC", instruction: DEC(1, DIR_X), opcode: 0xd6},
		{name: "XOR", instruction: XOR(1, LNG_X), opcode: 0x5f},
		{name: "EOR alias", instruction: EOR(1, ABS), opcode: 0x4d},
		{name: "ORA", instruction: ORA(1, DIR_16_Y), opcode: 0x11},
		{name: "LSR", instruction: LSR(1, ABS_X), opcode: 0x5e},
		{name: "ROL", instruction: ROL(1, DIR), opcode: 0x26},
		{name: "ROR", instruction: ROR(1, ABS), opcode: 0x6e},
		{name: "SBC", instruction: SBC(1, DIR_24), opcode: 0xe7},
		{name: "TSB", instruction: TSB(1, ABS), opcode: 0x0c},
		{name: "TRB", instruction: TRB(1, DIR), opcode: 0x14},
		{name: "CPX", instruction: CPX(1, ABS), opcode: 0xec},
		{name: "CPY", instruction: CPY(1, IMM16), opcode: 0xc0},
		{name: "JMP", instruction: JMP(1, ABS_X_16), opcode: 0x7c},
		{name: "JSR", instruction: JSR(1, ABS_X_16), opcode: 0xfc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encode(t, tt.instruction)
			if got[0] != tt.opcode {
				t.Fatalf("opcode = %#x, want %#x", got[0], tt.opcode)
			}
		})
	}
}

func TestImpliedInstructions(t *testing.T) {
	tests := []struct {
		instruction Instruction
		opcode      byte
	}{
		{PHA(), 0x48}, {PLA(), 0x68}, {PHX(), 0xda}, {PLX(), 0xfa},
		{PHY(), 0x5a}, {PLY(), 0x7a}, {PHP(), 0x08}, {PLP(), 0x28},
		{PHB(), 0x8b}, {PLB(), 0xab}, {PHK(), 0x4b}, {TAX(), 0xaa},
		{TXA(), 0x8a}, {TAY(), 0xa8}, {TYA(), 0x98}, {TXY(), 0x9b},
		{TYX(), 0xbb}, {TXS(), 0x9a}, {TSX(), 0xba}, {INX(), 0xe8},
		{DEX(), 0xca}, {INY(), 0xc8}, {DEY(), 0x88}, {CLC(), 0x18},
		{SEC(), 0x38}, {RTS(), 0x60}, {RTL(), 0x6b}, {XBA(), 0xeb},
		{TDC(), 0x7b},
	}

	for _, tt := range tests {
		name := tt.instruction.String()
		t.Run(name, func(t *testing.T) {
			got := encode(t, tt.instruction)
			if !bytes.Equal(got, []byte{tt.opcode}) {
				t.Fatalf("encoding = %x, want %02x", got, tt.opcode)
			}
		})
	}
}

func TestWidthHelpers(t *testing.T) {
	tests := []struct {
		name        string
		instruction Instruction
		want        []byte
	}{
		{name: "A8", instruction: A8(), want: []byte{0xe2, 0x20}},
		{name: "XY8", instruction: XY8(), want: []byte{0xe2, 0x10}},
		{name: "AXY8", instruction: AXY8(), want: []byte{0xe2, 0x30}},
		{name: "A16", instruction: A16(), want: []byte{0xc2, 0x20}},
		{name: "XY16", instruction: XY16(), want: []byte{0xc2, 0x10}},
		{name: "AXY16", instruction: AXY16(), want: []byte{0xc2, 0x30}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encode(t, tt.instruction); !bytes.Equal(got, tt.want) {
				t.Fatalf("encoding = %x, want %x", got, tt.want)
			}
		})
	}
}

func TestNamedBranchesResolveThroughSpace(t *testing.T) {
	manager := memory.NewManager(memory.New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "branches")
	if err != nil {
		t.Fatal(err)
	}

	if err := space.Write(
		BRA("FORWARD"),
		NOP(),
		"FORWARD",
		"BACKWARD",
		BNE("BACKWARD"),
		RTS(),
	); err != nil {
		t.Fatal(err)
	}

	got, err := manager.ROM.GetBytes(0x100, 6)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x80, 0x01, 0xea, 0xd0, 0xfe, 0x60}
	if !bytes.Equal(got, want) {
		t.Fatalf("encoding = %x, want %x", got, want)
	}
}

func TestBranchAliasesMatchUpstreamClasses(t *testing.T) {
	if got, want := encode(t, BGE(1)), encode(t, BCS(1)); !bytes.Equal(got, want) {
		t.Fatalf("BGE = %x, BCS = %x", got, want)
	}
	if got, want := encode(t, BLT(1)), encode(t, BCC(1)); !bytes.Equal(got, want) {
		t.Fatalf("BLT = %x, BCC = %x", got, want)
	}
}

func TestDirectByteArgumentsRejectOverflow(t *testing.T) {
	instruction := REP(0x100)
	manager := memory.NewManager(memory.New(nil))
	space, err := manager.Reserve(0x100, 0x101, "overflow")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.Write(instruction); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestLongArgumentsRejectOverflow(t *testing.T) {
	instruction := JSL(0x1000000)
	manager := memory.NewManager(memory.New(nil))
	space, err := manager.Reserve(0x100, 0x103, "overflow")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.Write(instruction); err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestInstructionNamesCoverUpstreamSurface(t *testing.T) {
	if got, want := len(Names()), 80; got != want {
		t.Fatalf("name count = %d, want %d", got, want)
	}
	if reflect.DeepEqual(Names(), []string{}) {
		t.Fatal("instruction names unexpectedly empty")
	}
}

func TestInstructionFormatting(t *testing.T) {
	tests := []struct {
		instruction Instruction
		want        string
	}{
		{LDA(0x12, IMM8), "LDA #$12"},
		{STA(0x1234, ABS_X), "STA $1234, X"},
		{LDA(0x12, DIR_16_Y), "LDA ($12), Y"},
		{LDA(0x123456, LNG), "LDA $123456"},
		{BRA("TARGET"), `BRA "TARGET"`},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.instruction), func(t *testing.T) {
			if got := tt.instruction.String(); got != tt.want {
				t.Fatalf("string = %q, want %q", got, tt.want)
			}
		})
	}
}

func encode(t *testing.T, instruction Instruction) []byte {
	t.Helper()
	manager := memory.NewManager(memory.New(nil))
	space, err := manager.Reserve(0x100, 0x10f, "instruction")
	if err != nil {
		t.Fatal(err)
	}
	if err := space.Write(instruction); err != nil {
		t.Fatal(err)
	}
	result, err := manager.ROM.GetBytes(0x100, instruction.Size())
	if err != nil {
		t.Fatal(err)
	}
	return result
}
