package asm

import (
	"fmt"
	"strings"

	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

// Mode is a 65816 addressing mode.
type Mode uint8

const (
	IMM8 Mode = iota
	DIR
	DIR_X
	DIR_Y
	DIR_16
	DIR_X_16
	DIR_16_Y
	DIR_24
	DIR_24_Y
	S
	S_16_Y
	IMM16
	ABS
	ABS_X
	ABS_Y
	ABS_16
	ABS_X_16
	ABS_24
	LNG
	LNG_X
)

// Instruction is an encoded 65816 instruction.
type Instruction struct {
	name     string
	opcode   byte
	argument int
	mode     Mode
	hasMode  bool
	hasArg   bool
	branchTo string
}

// Encode implements memory.Instruction.
func (i Instruction) Encode(space *memory.Space) ([]any, error) {
	result := []any{i.opcode}
	if i.branchTo != "" {
		return append(result, space.BranchDistance(i.branchTo)), nil
	}
	if !i.hasArg {
		return result, nil
	}

	argument, err := encodeArgument(i.argument, i.mode, i.hasMode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i.name, err)
	}
	return append(result, argument), nil
}

// Size implements memory.Instruction.
func (i Instruction) Size() int {
	if i.branchTo != "" {
		return 2
	}
	if !i.hasArg {
		return 1
	}
	if !i.hasMode {
		return 2
	}
	return 1 + modeWidth(i.mode)
}

func (i Instruction) String() string {
	if i.branchTo != "" {
		return fmt.Sprintf("%s %q", i.name, i.branchTo)
	}
	if !i.hasArg {
		return i.name
	}

	width := 2
	if i.hasMode {
		width = modeWidth(i.mode) * 2
	}
	value := i.argument & widthMask(width/2)
	hex := fmt.Sprintf("%0*x", width, value)
	switch i.mode {
	case IMM8, IMM16:
		return fmt.Sprintf("%s #$%s", i.name, hex)
	case DIR_X, ABS_X, LNG_X:
		return fmt.Sprintf("%s $%s, X", i.name, hex)
	case DIR_Y:
		return fmt.Sprintf("%s $%s, Y", i.name, hex)
	case DIR_16, ABS_16:
		return fmt.Sprintf("%s ($%s)", i.name, hex)
	case DIR_X_16, ABS_X_16:
		return fmt.Sprintf("%s ($%s, X)", i.name, hex)
	case DIR_16_Y:
		return fmt.Sprintf("%s ($%s), Y", i.name, hex)
	case DIR_24, ABS_24:
		return fmt.Sprintf("%s [$%s]", i.name, hex)
	case DIR_24_Y:
		return fmt.Sprintf("%s [$%s], Y", i.name, hex)
	case S:
		return fmt.Sprintf("%s $%s, S", i.name, hex)
	case S_16_Y:
		return fmt.Sprintf("%s ($%s, S), Y", i.name, hex)
	default:
		return fmt.Sprintf("%s $%s", i.name, hex)
	}
}

func instruction(name string, opcodes map[Mode]byte, argument int, mode Mode) Instruction {
	opcode, ok := opcodes[mode]
	if !ok {
		panic(fmt.Sprintf("asm: %s does not support addressing mode %d", name, mode))
	}
	return Instruction{
		name:     name,
		opcode:   opcode,
		argument: argument,
		mode:     mode,
		hasMode:  true,
		hasArg:   true,
	}
}

func implied(name string, opcode byte) Instruction {
	return Instruction{name: name, opcode: opcode}
}

func immediate8(name string, opcode byte, argument int) Instruction {
	return Instruction{name: name, opcode: opcode, argument: argument, hasArg: true}
}

func accumulator(name string, opcodes map[Mode]byte, arguments ...any) Instruction {
	if len(arguments) == 0 {
		opcode, ok := opcodes[0xff]
		if !ok {
			panic(fmt.Sprintf("asm: %s requires an argument and mode", name))
		}
		return implied(name, opcode)
	}
	if len(arguments) != 2 {
		panic(fmt.Sprintf("asm: %s expects zero arguments or argument and mode", name))
	}
	argument, ok := arguments[0].(int)
	if !ok {
		panic(fmt.Sprintf("asm: %s argument must be int", name))
	}
	mode, ok := arguments[1].(Mode)
	if !ok {
		panic(fmt.Sprintf("asm: %s mode must be asm.Mode", name))
	}
	return instruction(name, opcodes, argument, mode)
}

func branch(name string, opcode byte, argument any) Instruction {
	switch value := argument.(type) {
	case string:
		return Instruction{name: name, opcode: opcode, branchTo: value}
	case int:
		return immediate8(name, opcode, value)
	default:
		panic(fmt.Sprintf("asm: %s branch argument must be int or string", name))
	}
}

func encodeArgument(argument int, mode Mode, hasMode bool) ([]byte, error) {
	width := 1
	if !hasMode {
		if argument < 0 || argument > 0xff {
			return nil, fmt.Errorf("one-byte argument %#x does not fit in 1 byte", argument)
		}
	} else {
		width = modeWidth(mode)
	}
	if width == 3 && (argument < 0 || argument > 0xffffff) {
		return nil, fmt.Errorf("long argument %#x does not fit in 3 bytes", argument)
	}

	value := argument & widthMask(width)
	result := make([]byte, width)
	for index := range result {
		result[index] = byte(value >> (8 * index))
	}
	return result, nil
}

func modeWidth(mode Mode) int {
	switch {
	case mode >= LNG:
		return 3
	case mode >= IMM16:
		return 2
	default:
		return 1
	}
}

func widthMask(width int) int {
	if width >= 4 {
		return int(^uint(0) >> 1)
	}
	return 1<<(width*8) - 1
}

var (
	ldaOpcodes = map[Mode]byte{
		DIR_X_16: 0xa1, S: 0xa3, DIR: 0xa5, DIR_24: 0xa7,
		IMM8: 0xa9, IMM16: 0xa9, ABS: 0xad, LNG: 0xaf,
		DIR_16_Y: 0xb1, DIR_16: 0xb2, S_16_Y: 0xb3, DIR_X: 0xb5,
		DIR_24_Y: 0xb7, ABS_Y: 0xb9, ABS_X: 0xbd, LNG_X: 0xbf,
	}
	ldxOpcodes = map[Mode]byte{
		IMM8: 0xa2, IMM16: 0xa2, DIR: 0xa6, ABS: 0xae,
		DIR_Y: 0xb6, ABS_Y: 0xbe,
	}
	ldyOpcodes = map[Mode]byte{
		IMM8: 0xa0, IMM16: 0xa0, DIR: 0xa4, ABS: 0xac,
		DIR_X: 0xb4, ABS_X: 0xbc,
	}
	staOpcodes = map[Mode]byte{
		DIR_X_16: 0x81, S: 0x83, DIR: 0x85, DIR_24: 0x87,
		ABS: 0x8d, LNG: 0x8f, DIR_16_Y: 0x91, DIR_16: 0x92,
		S_16_Y: 0x93, DIR_X: 0x95, DIR_24_Y: 0x97, ABS_Y: 0x99,
		ABS_X: 0x9d, LNG_X: 0x9f,
	}
	stxOpcodes = map[Mode]byte{DIR: 0x86, ABS: 0x8e, DIR_Y: 0x96}
	styOpcodes = map[Mode]byte{DIR: 0x84, ABS: 0x8c, DIR_X: 0x94}
	stzOpcodes = map[Mode]byte{DIR: 0x64, DIR_X: 0x74, ABS: 0x9c, ABS_X: 0x9e}
	incOpcodes = map[Mode]byte{0xff: 0x1a, DIR: 0xe6, ABS: 0xee, DIR_X: 0xf6, ABS_X: 0xfe}
	decOpcodes = map[Mode]byte{0xff: 0x3a, DIR: 0xc6, ABS: 0xce, DIR_X: 0xd6, ABS_X: 0xde}
	andOpcodes = map[Mode]byte{
		DIR_X_16: 0x21, S: 0x23, DIR: 0x25, DIR_24: 0x27,
		IMM8: 0x29, IMM16: 0x29, ABS: 0x2d, LNG: 0x2f,
		DIR_16_Y: 0x31, DIR_16: 0x32, S_16_Y: 0x33, DIR_X: 0x35,
		DIR_24_Y: 0x37, ABS_Y: 0x39, ABS_X: 0x3d, LNG_X: 0x3f,
	}
	xorOpcodes = map[Mode]byte{
		DIR_X_16: 0x41, S: 0x43, DIR: 0x45, DIR_24: 0x47,
		IMM8: 0x49, IMM16: 0x49, ABS: 0x4d, LNG: 0x4f,
		DIR_16_Y: 0x51, DIR_16: 0x52, S_16_Y: 0x53, DIR_X: 0x55,
		DIR_24_Y: 0x57, ABS_Y: 0x59, ABS_X: 0x5d, LNG_X: 0x5f,
	}
	oraOpcodes = map[Mode]byte{
		DIR_X_16: 0x01, S: 0x03, DIR: 0x05, DIR_24: 0x07,
		IMM8: 0x09, IMM16: 0x09, ABS: 0x0d, LNG: 0x0f,
		DIR_16_Y: 0x11, DIR_16: 0x12, S_16_Y: 0x13, DIR_X: 0x15,
		DIR_24_Y: 0x17, ABS_Y: 0x19, ABS_X: 0x1d, LNG_X: 0x1f,
	}
	aslOpcodes = map[Mode]byte{DIR: 0x06, 0xff: 0x0a, ABS: 0x0e, DIR_X: 0x16, ABS_X: 0x1e}
	lsrOpcodes = map[Mode]byte{DIR: 0x46, 0xff: 0x4a, ABS: 0x4e, DIR_X: 0x56, ABS_X: 0x5e}
	rolOpcodes = map[Mode]byte{DIR: 0x26, 0xff: 0x2a, ABS: 0x2e, DIR_X: 0x36, ABS_X: 0x3e}
	rorOpcodes = map[Mode]byte{DIR: 0x66, 0xff: 0x6a, ABS: 0x6e, DIR_X: 0x76, ABS_X: 0x7e}
	adcOpcodes = map[Mode]byte{
		DIR_X_16: 0x61, S: 0x63, DIR: 0x65, DIR_24: 0x67,
		IMM8: 0x69, IMM16: 0x69, ABS: 0x6d, LNG: 0x6f,
		DIR_16_Y: 0x71, DIR_16: 0x72, S_16_Y: 0x73, DIR_X: 0x75,
		DIR_24_Y: 0x77, ABS_Y: 0x79, ABS_X: 0x7d, LNG_X: 0x7f,
	}
	sbcOpcodes = map[Mode]byte{
		DIR_X_16: 0xe1, S: 0xe3, DIR: 0xe5, DIR_24: 0xe7,
		IMM8: 0xe9, IMM16: 0xe9, ABS: 0xed, LNG: 0xef,
		DIR_16_Y: 0xf1, DIR_16: 0xf2, S_16_Y: 0xf3, DIR_X: 0xf5,
		DIR_24_Y: 0xf7, ABS_Y: 0xf9, ABS_X: 0xfd, LNG_X: 0xff,
	}
	bitOpcodes = map[Mode]byte{
		DIR: 0x24, ABS: 0x2c, DIR_X: 0x34, ABS_X: 0x3c,
		IMM8: 0x89, IMM16: 0x89,
	}
	tsbOpcodes = map[Mode]byte{DIR: 0x04, ABS: 0x0c}
	trbOpcodes = map[Mode]byte{DIR: 0x14, ABS: 0x1c}
	cmpOpcodes = map[Mode]byte{
		DIR_X_16: 0xc1, S: 0xc3, DIR: 0xc5, DIR_24: 0xc7,
		IMM8: 0xc9, IMM16: 0xc9, ABS: 0xcd, LNG: 0xcf,
		DIR_16_Y: 0xd1, DIR_16: 0xd2, S_16_Y: 0xd3, DIR_X: 0xd5,
		DIR_24_Y: 0xd7, ABS_Y: 0xd9, ABS_X: 0xdd, LNG_X: 0xdf,
	}
	cpxOpcodes = map[Mode]byte{IMM8: 0xe0, IMM16: 0xe0, DIR: 0xe4, ABS: 0xec}
	cpyOpcodes = map[Mode]byte{IMM8: 0xc0, IMM16: 0xc0, DIR: 0xc4, ABS: 0xcc}
	jmpOpcodes = map[Mode]byte{ABS: 0x4c, LNG: 0x5c, ABS_16: 0x6c, ABS_X_16: 0x7c, ABS_24: 0xdc}
	jsrOpcodes = map[Mode]byte{ABS: 0x20, ABS_X_16: 0xfc}
)

func NOP() Instruction                   { return implied("NOP", 0xea) }
func REP(arg int) Instruction            { return immediate8("REP", 0xc2, arg) }
func SEP(arg int) Instruction            { return immediate8("SEP", 0xe2, arg) }
func A8() Instruction                    { return SEP(0x20) }
func XY8() Instruction                   { return SEP(0x10) }
func AXY8() Instruction                  { return SEP(0x30) }
func A16() Instruction                   { return REP(0x20) }
func XY16() Instruction                  { return REP(0x10) }
func AXY16() Instruction                 { return REP(0x30) }
func PHA() Instruction                   { return implied("PHA", 0x48) }
func PLA() Instruction                   { return implied("PLA", 0x68) }
func PHX() Instruction                   { return implied("PHX", 0xda) }
func PLX() Instruction                   { return implied("PLX", 0xfa) }
func PHY() Instruction                   { return implied("PHY", 0x5a) }
func PLY() Instruction                   { return implied("PLY", 0x7a) }
func PHP() Instruction                   { return implied("PHP", 0x08) }
func PLP() Instruction                   { return implied("PLP", 0x28) }
func PHB() Instruction                   { return implied("PHB", 0x8b) }
func PLB() Instruction                   { return implied("PLB", 0xab) }
func PHK() Instruction                   { return implied("PHK", 0x4b) }
func PEI(arg int) Instruction            { return instruction("PEI", map[Mode]byte{DIR: 0xd4}, arg, DIR) }
func PER(arg int) Instruction            { return instruction("PER", map[Mode]byte{IMM16: 0x62}, arg, IMM16) }
func PEA(arg int) Instruction            { return instruction("PEA", map[Mode]byte{IMM16: 0xf4}, arg, IMM16) }
func TAX() Instruction                   { return implied("TAX", 0xaa) }
func TXA() Instruction                   { return implied("TXA", 0x8a) }
func TAY() Instruction                   { return implied("TAY", 0xa8) }
func TYA() Instruction                   { return implied("TYA", 0x98) }
func TXY() Instruction                   { return implied("TXY", 0x9b) }
func TYX() Instruction                   { return implied("TYX", 0xbb) }
func TXS() Instruction                   { return implied("TXS", 0x9a) }
func TSX() Instruction                   { return implied("TSX", 0xba) }
func LDA(arg int, mode Mode) Instruction { return instruction("LDA", ldaOpcodes, arg, mode) }
func LDX(arg int, mode Mode) Instruction { return instruction("LDX", ldxOpcodes, arg, mode) }
func LDY(arg int, mode Mode) Instruction { return instruction("LDY", ldyOpcodes, arg, mode) }
func STA(arg int, mode Mode) Instruction { return instruction("STA", staOpcodes, arg, mode) }
func STX(arg int, mode Mode) Instruction { return instruction("STX", stxOpcodes, arg, mode) }
func STY(arg int, mode Mode) Instruction { return instruction("STY", styOpcodes, arg, mode) }
func STZ(arg int, mode Mode) Instruction { return instruction("STZ", stzOpcodes, arg, mode) }
func INC(args ...any) Instruction        { return accumulator("INC", incOpcodes, args...) }
func DEC(args ...any) Instruction        { return accumulator("DEC", decOpcodes, args...) }
func INX() Instruction                   { return implied("INX", 0xe8) }
func DEX() Instruction                   { return implied("DEX", 0xca) }
func INY() Instruction                   { return implied("INY", 0xc8) }
func DEY() Instruction                   { return implied("DEY", 0x88) }
func AND(arg int, mode Mode) Instruction { return instruction("AND", andOpcodes, arg, mode) }
func XOR(arg int, mode Mode) Instruction { return instruction("XOR", xorOpcodes, arg, mode) }
func EOR(arg int, mode Mode) Instruction { return XOR(arg, mode) }
func ORA(arg int, mode Mode) Instruction { return instruction("ORA", oraOpcodes, arg, mode) }
func ASL(args ...any) Instruction        { return accumulator("ASL", aslOpcodes, args...) }
func LSR(args ...any) Instruction        { return accumulator("LSR", lsrOpcodes, args...) }
func ROL(args ...any) Instruction        { return accumulator("ROL", rolOpcodes, args...) }
func ROR(args ...any) Instruction        { return accumulator("ROR", rorOpcodes, args...) }
func ADC(arg int, mode Mode) Instruction { return instruction("ADC", adcOpcodes, arg, mode) }
func SBC(arg int, mode Mode) Instruction { return instruction("SBC", sbcOpcodes, arg, mode) }
func BIT(arg int, mode Mode) Instruction { return instruction("BIT", bitOpcodes, arg, mode) }
func TSB(arg int, mode Mode) Instruction { return instruction("TSB", tsbOpcodes, arg, mode) }
func TRB(arg int, mode Mode) Instruction { return instruction("TRB", trbOpcodes, arg, mode) }
func CMP(arg int, mode Mode) Instruction { return instruction("CMP", cmpOpcodes, arg, mode) }
func CPX(arg int, mode Mode) Instruction { return instruction("CPX", cpxOpcodes, arg, mode) }
func CPY(arg int, mode Mode) Instruction { return instruction("CPY", cpyOpcodes, arg, mode) }
func BRA(arg any) Instruction            { return branch("BRA", 0x80, arg) }
func BEQ(arg any) Instruction            { return branch("BEQ", 0xf0, arg) }
func BNE(arg any) Instruction            { return branch("BNE", 0xd0, arg) }
func BCS(arg any) Instruction            { return branch("BCS", 0xb0, arg) }
func BGE(arg any) Instruction            { return BCS(arg) }
func BCC(arg any) Instruction            { return branch("BCC", 0x90, arg) }
func BLT(arg any) Instruction            { return BCC(arg) }
func BMI(arg any) Instruction            { return branch("BMI", 0x30, arg) }
func BPL(arg any) Instruction            { return branch("BPL", 0x10, arg) }
func BVS(arg any) Instruction            { return branch("BVS", 0x70, arg) }
func BVC(arg any) Instruction            { return branch("BVC", 0x50, arg) }
func CLC() Instruction                   { return implied("CLC", 0x18) }
func SEC() Instruction                   { return implied("SEC", 0x38) }
func JMP(arg int, mode Mode) Instruction { return instruction("JMP", jmpOpcodes, arg, mode) }
func JSR(arg int, mode Mode) Instruction { return instruction("JSR", jsrOpcodes, arg, mode) }
func JSL(arg int) Instruction            { return instruction("JSL", map[Mode]byte{LNG: 0x22}, arg, LNG) }
func RTS() Instruction                   { return implied("RTS", 0x60) }
func RTL() Instruction                   { return implied("RTL", 0x6b) }
func XBA() Instruction                   { return implied("XBA", 0xeb) }
func TDC() Instruction                   { return implied("TDC", 0x7b) }

// Names returns the upstream instruction constructor names.
func Names() []string {
	return strings.Fields(
		"NOP REP SEP A8 XY8 AXY8 A16 XY16 AXY16 PHA PLA PHX PLX PHY PLY PHP PLP PHB PLB PHK " +
			"PEI PER PEA TAX TXA TAY TYA TXY TYX TXS TSX LDA LDX LDY STA STX STY STZ INC DEC INX DEX INY DEY " +
			"AND XOR EOR ORA ASL LSR ROL ROR ADC SBC BIT TSB TRB CMP CPX CPY BRA BEQ BNE BCS BGE BCC BLT BMI BPL BVS BVC " +
			"CLC SEC JMP JSR JSL RTS RTL XBA TDC",
	)
}
