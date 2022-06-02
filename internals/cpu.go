package internals

import (
	"fmt"
)

const (
	_ = iota
	INTERRUPTS_NONE
	INTERRUPTS_NMI
	INTERRUPTS_IRQ
)

type CPU struct {
	A uint8    // Accumulator
	X uint8    // X index
	Y uint8    // Y index
	P struct { // Status register
		C uint8 // Carry flag
		Z uint8 // Zero flag
		I uint8 // Interrupt (if set, the interrupts are disabled)
		D uint8 // Decimal mode
		B uint8 // Software interrupt
		N uint8 // Not used, always 1
		V uint8 // Overflow flag
		S uint8 // Sign flag
	}
	PC         uint16 // Program counter
	SP         uint8  // Stack pointer
	CycleCount uint64
	CycleDelay uint64
	Interrupt  uint32
	Bus        IBus
}

const (
	_ = iota
	Implied
	Immediate
	Accumulator
	Absolute
	AbsoluteX
	AbsoluteY
	ZeroPage
	ZeroPageX
	ZeroPageY
	Indirect
	IndirectX
	IndirectY
	Relative
)

type opcode struct {
	ID             uint8
	AddressingMode uint8
	Size           uint8
	Cycles         uint8
	PageCycles     uint8
	run            func(*CPU, uint8, uint16, bool)
	Name           string
}

//

// Missing illegal opcodes (undocumented) - https://www.nesdev.com/undocumented_opcodes.txt
// http://www.6502.org/tutorials/6502opcodes.html#BRA
// https://www.nesdev.com/6502.txt
var instructions = [256]opcode{
	{ID: 0x00, AddressingMode: Implied, Size: 1, Cycles: 7, PageCycles: 0, Name: "BRK", run: _BRK},
	{ID: 0x01, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "ORA", run: _ORA},
	{}, // 0x02
	{}, // 0x03
	{ID: 0x04, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x05, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "ORA", run: _ORA},
	{ID: 0x06, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "ASL", run: _ASL},
	{}, // 0x07
	{ID: 0x08, AddressingMode: Implied, Size: 1, Cycles: 3, PageCycles: 0, Name: "PHP", run: _PHP},
	{ID: 0x09, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "ORA", run: _ORA},
	{ID: 0x0A, AddressingMode: Accumulator, Size: 1, Cycles: 2, PageCycles: 0, Name: "ASL", run: _ASL},
	{}, // 0x0B
	{ID: 0x0C, AddressingMode: Implied, Size: 3, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x0D, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "ORA", run: _ORA},
	{ID: 0x0E, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "ASL", run: _ASL},
	{}, // 0x0F
	{ID: 0x10, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BPL", run: _BPL},
	{ID: 0x11, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "ORA", run: _ORA},
	{}, // 0x12
	{}, // 0x13
	{ID: 0x14, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x15, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "ORA", run: _ORA},
	{ID: 0x16, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "ASL", run: _ASL},
	{}, // 0x17
	{ID: 0x18, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "CLC", run: _CLC},
	{ID: 0x19, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "ORA", run: _ORA},
	{ID: 0x1A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0x1B
	{ID: 0x1C, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0x1D, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "ORA", run: _ORA},
	{ID: 0x1E, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "ASL", run: _ASL},
	{}, // 0x1F
	{ID: 0x20, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "JSR", run: _JSR},
	{ID: 0x21, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "AND", run: _AND},
	{}, // 0x22
	{}, // 0x23
	{ID: 0x24, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "BIT", run: _BIT},
	{ID: 0x25, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "AND", run: _AND},
	{ID: 0x26, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "ROL", run: _ROL},
	{}, // 0x27
	{ID: 0x28, AddressingMode: Implied, Size: 1, Cycles: 4, PageCycles: 0, Name: "PLP", run: _PLP},
	{ID: 0x29, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "AND", run: _AND},
	{ID: 0x2A, AddressingMode: Accumulator, Size: 1, Cycles: 2, PageCycles: 0, Name: "ROL", run: _ROL},
	{}, // 0x2B
	{ID: 0x2C, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "BIT", run: _BIT},
	{ID: 0x2D, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "AND", run: _AND},
	{ID: 0x2E, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "ROL", run: _ROL},
	{}, // 0x2F
	{ID: 0x30, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BMI", run: _BMI},
	{ID: 0x31, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "AND", run: _AND},
	{}, // 0x32
	{}, // 0x33
	{ID: 0x34, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x35, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "AND", run: _AND},
	{ID: 0x36, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "ROL", run: _ROL},
	{}, // 0x37
	{ID: 0x38, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "SEC", run: _SEC},
	{ID: 0x39, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "AND", run: _AND},
	{ID: 0x3A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0x3B
	{ID: 0x3C, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0x3D, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "AND", run: _AND},
	{ID: 0x3E, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "ROL", run: _ROL},
	{}, // 0x3F
	{ID: 0x40, AddressingMode: Implied, Size: 1, Cycles: 6, PageCycles: 0, Name: "RTI", run: _RTI},
	{ID: 0x41, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "EOR", run: _EOR},
	{}, // 0x42
	{}, // 0x43
	{ID: 0x44, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x45, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "EOR", run: _EOR},
	{ID: 0x46, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "LSR", run: _LSR},
	{}, // 0x47
	{ID: 0x48, AddressingMode: Implied, Size: 1, Cycles: 3, PageCycles: 0, Name: "PHA", run: _PHA},
	{ID: 0x49, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "EOR", run: _EOR},
	{ID: 0x4A, AddressingMode: Accumulator, Size: 1, Cycles: 2, PageCycles: 0, Name: "LSR", run: _LSR},
	{}, // 0x4B
	{ID: 0x4C, AddressingMode: Absolute, Size: 3, Cycles: 3, PageCycles: 0, Name: "JMP", run: _JMP},
	{ID: 0x4D, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "EOR", run: _EOR},
	{ID: 0x4E, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "LSR", run: _LSR},
	{}, // 0x4F
	{ID: 0x50, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BVC", run: _BVC},
	{ID: 0x51, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "EOR", run: _EOR},
	{}, // 0x52
	{}, // 0x53
	{ID: 0x54, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x55, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "EOR", run: _EOR},
	{ID: 0x56, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "LSR", run: _LSR},
	{}, // 0x57
	{ID: 0x58, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "CLI", run: _CLI},
	{ID: 0x59, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "EOR", run: _EOR},
	{ID: 0x5A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0x5B
	{ID: 0x5C, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0x5D, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "EOR", run: _EOR},
	{ID: 0x5E, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "LSR", run: _LSR},
	{}, // 0x5F
	{ID: 0x60, AddressingMode: Implied, Size: 1, Cycles: 6, PageCycles: 0, Name: "RTS", run: _RTS},
	{ID: 0x61, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "ADC", run: _ADC},
	{}, // 0x62
	{}, // 0x63
	{ID: 0x64, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x65, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "ADC", run: _ADC},
	{ID: 0x66, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "ROR", run: _ROR},
	{}, // 0x67
	{ID: 0x68, AddressingMode: Implied, Size: 1, Cycles: 4, PageCycles: 0, Name: "PLA", run: _PLA},
	{ID: 0x69, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "ADC", run: _ADC},
	{ID: 0x6A, AddressingMode: Accumulator, Size: 1, Cycles: 2, PageCycles: 0, Name: "ROR", run: _ROR},
	{}, // 0x6B
	{ID: 0x6C, AddressingMode: Indirect, Size: 3, Cycles: 5, PageCycles: 0, Name: "JMP", run: _JMP},
	{ID: 0x6D, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "ADC", run: _ADC},
	{ID: 0x6E, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "ROR", run: _ROR},
	{}, // 0x6F
	{ID: 0x70, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BVS", run: _BVS},
	{ID: 0x71, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "ADC", run: _ADC},
	{}, // 0x72
	{}, // 0x73
	{ID: 0x74, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x75, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "ADC", run: _ADC},
	{ID: 0x76, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "ROR", run: _ROR},
	{}, // 0x77
	{ID: 0x78, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "SEI", run: _SEI},
	{ID: 0x79, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "ADC", run: _ADC},
	{ID: 0x7A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0x7B
	{ID: 0x7C, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0x7D, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "ADC", run: _ADC},
	{ID: 0x7E, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "ROR", run: _ROR},
	{}, // 0x7F
	{ID: 0x80, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x81, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "STA", run: _STA},
	{ID: 0x82, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0x83
	{ID: 0x84, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "STY", run: _STY},
	{ID: 0x85, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "STA", run: _STA},
	{ID: 0x86, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "STX", run: _STX},
	{}, // 0x87
	{ID: 0x88, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "DEY", run: _DEY},
	{ID: 0x89, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0x8A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TXA", run: _TXA},
	{}, // 0x8B
	{ID: 0x8C, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "STY", run: _STY},
	{ID: 0x8D, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "STA", run: _STA},
	{ID: 0x8E, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "STX", run: _STX},
	{}, // 0x8F
	{ID: 0x90, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BCC", run: _BCC},
	{ID: 0x91, AddressingMode: IndirectY, Size: 2, Cycles: 6, PageCycles: 0, Name: "STA", run: _STA},
	{}, // 0x92
	{}, // 0x93
	{ID: 0x94, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "STY", run: _STY},
	{ID: 0x95, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "STA", run: _STA},
	{ID: 0x96, AddressingMode: ZeroPageY, Size: 2, Cycles: 4, PageCycles: 0, Name: "STX", run: _STX},
	{}, // 0x97
	{ID: 0x98, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TYA", run: _TYA},
	{ID: 0x99, AddressingMode: AbsoluteY, Size: 3, Cycles: 5, PageCycles: 0, Name: "STA", run: _STA},
	{ID: 0x9A, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TXS", run: _TXS},
	{}, // 0x9B
	{}, // 0x9C
	{ID: 0x9D, AddressingMode: AbsoluteX, Size: 3, Cycles: 5, PageCycles: 0, Name: "STA", run: _STA},
	{}, // 0x9E
	{}, // 0x9F
	{ID: 0xA0, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "LDY", run: _LDY},
	{ID: 0xA1, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "LDA", run: _LDA},
	{ID: 0xA2, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "LDX", run: _LDX},
	{}, // 0xA3
	{ID: 0xA4, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "LDY", run: _LDY},
	{ID: 0xA5, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "LDA", run: _LDA},
	{ID: 0xA6, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "LDX", run: _LDX},
	{}, // 0xA7
	{ID: 0xA8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TAY", run: _TAY},
	{ID: 0xA9, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "LDA", run: _LDA},
	{ID: 0xAA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TAX", run: _TAX},
	{}, // 0xAB
	{ID: 0xAC, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "LDY", run: _LDY},
	{ID: 0xAD, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "LDA", run: _LDA},
	{ID: 0xAE, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "LDX", run: _LDX},
	{}, // 0xAF
	{ID: 0xB0, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BCS", run: _BCS},
	{ID: 0xB1, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "LDA", run: _LDA},
	{}, // 0xB2
	{}, // 0xB3
	{ID: 0xB4, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "LDY", run: _LDY},
	{ID: 0xB5, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "LDA", run: _LDA},
	{ID: 0xB6, AddressingMode: ZeroPageY, Size: 2, Cycles: 4, PageCycles: 0, Name: "LDX", run: _LDX},
	{}, // 0xB7
	{ID: 0xB8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "CLV", run: _CLV},
	{ID: 0xB9, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "LDA", run: _LDA},
	{ID: 0xBA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "TSX", run: _TSX},
	{}, // 0xBB
	{ID: 0xBC, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "LDY", run: _LDY},
	{ID: 0xBD, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "LDA", run: _LDA},
	{ID: 0xBE, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "LDX", run: _LDX},
	{}, // 0xBF
	{ID: 0xC0, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "CPY", run: _CPY},
	{ID: 0xC1, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "CMP", run: _CMP},
	{ID: 0xC2, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0xC3
	{ID: 0xC4, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "CPY", run: _CPY},
	{ID: 0xC5, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "CMP", run: _CMP},
	{ID: 0xC6, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "DEC", run: _DEC},
	{}, // 0xC7
	{ID: 0xC8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "INY", run: _INY},
	{ID: 0xC9, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "CMP", run: _CMP},
	{ID: 0xCA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "DEX", run: _DEX},
	{}, // 0xCB
	{ID: 0xCC, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "CPY", run: _CPY},
	{ID: 0xCD, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "CMP", run: _CMP},
	{ID: 0xCE, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "DEC", run: _DEC},
	{}, // 0xCF
	{ID: 0xD0, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BNE", run: _BNE},
	{ID: 0xD1, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "CMP", run: _CMP},
	{}, // 0xD2
	{}, // 0xD3
	{ID: 0xD4, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0xD5, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "CMP", run: _CMP},
	{ID: 0xD6, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "DEC", run: _DEC},
	{}, // 0xD7
	{ID: 0xD8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "CLD", run: _CLD},
	{ID: 0xD9, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "CMP", run: _CMP},
	{ID: 0xDA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0xDB
	{ID: 0xDC, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0xDD, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "CMP", run: _CMP},
	{ID: 0xDE, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "DEC", run: _DEC},
	{}, // 0xDF
	{ID: 0xE0, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "CPX", run: _CPX},
	{ID: 0xE1, AddressingMode: IndirectX, Size: 2, Cycles: 6, PageCycles: 0, Name: "SBC", run: _SBC},
	{ID: 0xE2, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0xE3
	{ID: 0xE4, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "CPX", run: _CPX},
	{ID: 0xE5, AddressingMode: ZeroPage, Size: 2, Cycles: 3, PageCycles: 0, Name: "SBC", run: _SBC},
	{ID: 0xE6, AddressingMode: ZeroPage, Size: 2, Cycles: 5, PageCycles: 0, Name: "INC", run: _INC},
	{}, // 0xE7
	{ID: 0xE8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "INX", run: _INX},
	{ID: 0xE9, AddressingMode: Immediate, Size: 2, Cycles: 2, PageCycles: 0, Name: "SBC", run: _SBC},
	{ID: 0xEA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0xEB
	{ID: 0xEC, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "CPX", run: _CPX},
	{ID: 0xED, AddressingMode: Absolute, Size: 3, Cycles: 4, PageCycles: 0, Name: "SBC", run: _SBC},
	{ID: 0xE6, AddressingMode: Absolute, Size: 3, Cycles: 6, PageCycles: 0, Name: "INC", run: _INC},
	{}, // 0xEF
	{ID: 0xF0, AddressingMode: Relative, Size: 2, Cycles: 2, PageCycles: 0, Name: "BEQ", run: _BEQ},
	{ID: 0xF1, AddressingMode: IndirectY, Size: 2, Cycles: 5, PageCycles: 1, Name: "SBC", run: _SBC},
	{}, // 0xF2
	{}, // 0xF3
	{ID: 0xF4, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "NOP", run: _NOP},
	{ID: 0xF5, AddressingMode: ZeroPageX, Size: 2, Cycles: 4, PageCycles: 0, Name: "SBC", run: _SBC},
	{ID: 0xF6, AddressingMode: ZeroPageX, Size: 2, Cycles: 6, PageCycles: 0, Name: "INC", run: _INC},
	{}, // 0xF7
	{ID: 0xF8, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "SED", run: _SED},
	{ID: 0xF9, AddressingMode: AbsoluteY, Size: 3, Cycles: 4, PageCycles: 1, Name: "SBC", run: _SBC},
	{ID: 0xFA, AddressingMode: Implied, Size: 1, Cycles: 2, PageCycles: 0, Name: "NOP", run: _NOP},
	{}, // 0xFB
	{ID: 0xFC, AddressingMode: Implied, Size: 3, Cycles: 5, PageCycles: 0, Name: "NOP", run: _NOP}, // Something with page crossing; Might need to read the arguments. Ignore for now
	{ID: 0xFD, AddressingMode: AbsoluteX, Size: 3, Cycles: 4, PageCycles: 1, Name: "SBC", run: _SBC},
	{ID: 0xFE, AddressingMode: AbsoluteX, Size: 3, Cycles: 7, PageCycles: 0, Name: "INC", run: _INC},
	{}, // 0xFF
}

// Returns the address asociated with an opcode and if a memory page was crossed (where applicable, default is false)
func (cpu *CPU) getAddress(instruction opcode) (uint16, bool) {
	switch instruction.AddressingMode {
	case Immediate:
		return cpu.PC + 1, false
	case Absolute:
		return cpu.Bus.ReadAddress(cpu.PC + 1), false
	case AbsoluteX:
		readAddress := cpu.Bus.ReadAddress(cpu.PC + 1)
		address := readAddress + uint16(cpu.X)
		return address, ((address&0xFF00)-(readAddress&0xFF00) != 0)
	case AbsoluteY:
		readAddress := cpu.Bus.ReadAddress(cpu.PC + 1)
		address := readAddress + uint16(cpu.Y)
		return address, ((address&0xFF00)-(readAddress&0xFF00) != 0)
	case ZeroPage:
		return uint16(cpu.Bus.Read(cpu.PC + 1)), false
	case ZeroPageX:
		return uint16(cpu.Bus.Read(cpu.PC+1) + cpu.X), false
	case ZeroPageY:
		return uint16(cpu.Bus.Read(cpu.PC+1) + cpu.Y), false
	case Indirect:
		return cpu.Bus.ReadAddressBug(cpu.Bus.ReadAddress(cpu.PC + 1)), false
	case IndirectX:
		return cpu.Bus.ReadAddressBug(uint16(cpu.Bus.Read(cpu.PC+1) + cpu.X)), false
	case IndirectY:
		readAddress := cpu.Bus.ReadAddressBug(uint16(cpu.Bus.Read(cpu.PC + 1)))
		address := readAddress + uint16(cpu.Y)
		return address, ((address&0xFF00)-(readAddress&0xFF00) != 0)
	case Relative:
		displacement := uint16(cpu.Bus.Read(cpu.PC + 1))
		address := cpu.PC + 2 + displacement
		if displacement >= 0x80 {
			address -= 0x100
		}
		return address, ((address&0xFF00)-((cpu.PC+2)&0xFF00) != 0)
	default:
		return 0x0, false
	}
}

var DEBUG bool = false

func (cpu *CPU) Step() (uint64, opcode) {
	var startingCycles uint64 = cpu.CycleCount

	op := cpu.Bus.Read(cpu.PC)
	instruction := instructions[op]
	if instruction.run == nil {
		instruction = instructions[0x1A] // NOP
	}

	address, pageCycle := cpu.getAddress(instruction)

	if DEBUG {
		instructionBytes := fmt.Sprintf("%2x", cpu.Bus.Read(cpu.PC))
		for i := 1; i < int(instruction.Size); i++ {
			instructionBytes += fmt.Sprintf(" %2x", cpu.Bus.Read(cpu.PC+uint16(i)))
		}
		if instruction.Size < 3 {
			instructionBytes += "\t"
		}

		stack := ""
		for i := 1; i < 10 && i+int(cpu.SP) <= 0xFF; i++ {
			stack += fmt.Sprintf("%2x ", cpu.Bus.Read(uint16(i)+uint16(cpu.SP)+0x100))
		}

		fmt.Printf("%4x\t%v\t%v\tA:%2x X:%2x Y:%2x P:%x SP:%2x ADDR:%4x CYC:%d\tSTK:%v\n", cpu.PC, instructionBytes, instruction.Name, cpu.A, cpu.X, cpu.Y, cpu.GetFlags(), cpu.SP, address, cpu.CycleCount, stack)

	}

	cpu.PC += uint16(instruction.Size)
	cpu.CycleCount += uint64(instruction.Cycles)
	if pageCycle {
		cpu.CycleCount += uint64(instruction.PageCycles)
	}

	instruction.run(cpu, instruction.AddressingMode, address, pageCycle)

	return cpu.CycleCount - startingCycles, instruction
}

func (cpu *CPU) Cycle() {
	if cpu.CycleDelay == 0 {
		switch cpu.Interrupt {
		case INTERRUPTS_NONE:
			cpu.CycleDelay, _ = cpu.Step()
		case INTERRUPTS_NMI:
			cpu._NMI()
			cpu.Interrupt = INTERRUPTS_NONE
			cpu.CycleDelay = 7
		case INTERRUPTS_IRQ:
			cpu._IRQ()
			cpu.Interrupt = INTERRUPTS_NONE
			cpu.CycleDelay = 7
		}
	}
	cpu.CycleDelay--
}

func (cpu *CPU) SetFlags(flags uint8) {
	cpu.P.C = (flags >> 0) & 1
	cpu.P.Z = (flags >> 1) & 1
	cpu.P.I = (flags >> 2) & 1
	cpu.P.D = (flags >> 3) & 1
	cpu.P.B = (flags >> 4) & 1
	cpu.P.N = 1 // Always 1
	cpu.P.V = (flags >> 6) & 1
	cpu.P.S = (flags >> 7) & 1
}

func (cpu *CPU) GetFlags() uint8 {
	var flags uint8 = 0
	flags = flags | (cpu.P.C << 0)
	flags = flags | (cpu.P.Z << 1)
	flags = flags | (cpu.P.I << 2)
	flags = flags | (cpu.P.D << 3)
	flags = flags | (cpu.P.B << 4)
	flags = flags | (1 << 5) // Always 1
	flags = flags | (cpu.P.V << 6)
	flags = flags | (cpu.P.S << 7)
	return flags
}

// https://wiki.nesdev.org/w/index.php?title=CPU_power_up_state
func (cpu *CPU) PowerUp() {
	cpu.SetFlags(0x34)
	cpu.PC = cpu.Bus.ReadAddress(0xFFFC)
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0
	cpu.SP = 0xFD
	cpu.CycleCount = 7 // Warming up
	cpu.P.I = 1

	cpu.Interrupt = INTERRUPTS_NONE
}

// https://wiki.nesdev.org/w/index.php?title=CPU_power_up_state
func (cpu *CPU) Reset() {
	cpu.PC = cpu.Bus.ReadAddress(0xFFFC)
	cpu.SP = 0xFD
	cpu.P.I = 1

	cpu.Interrupt = INTERRUPTS_NONE
}

func (cpu *CPU) setZero(value uint8) {
	if value == 0 {
		cpu.P.Z = 1
	} else {
		cpu.P.Z = 0
	}
}

func (cpu *CPU) setSign(value uint8) {
	if value&0x80 == 0 {
		cpu.P.S = 0
	} else {
		cpu.P.S = 1
	}
}

func (cpu *CPU) Push(value uint8) {
	cpu.Bus.Write(uint16(cpu.SP)+0x100, value)
	cpu.SP--
}

func (cpu *CPU) PushAddress(value uint16) {
	cpu.SP--
	cpu.Bus.WriteAddress(uint16(cpu.SP)+0x100, value)
	cpu.SP--
}

func (cpu *CPU) Pop() uint8 {
	cpu.SP++
	return cpu.Bus.Read(uint16(cpu.SP) + 0x100)
}

func (cpu *CPU) PopAddress() uint16 {
	cpu.SP += 2
	if cpu.SP > 0xFF {
		panic("Stack underflow")
	}
	return cpu.Bus.ReadAddress(uint16(cpu.SP-1) + 0x100)
}

func (cpu *CPU) InterruptNMI() {
	cpu.Interrupt = INTERRUPTS_NMI
}

func (cpu *CPU) InterruptIRQ() {
	if cpu.P.I == 0 {
		cpu.Interrupt = INTERRUPTS_IRQ
	}
}

func (cpu *CPU) _NMI() {
	cpu.PushAddress(cpu.PC)
	cpu.Push(cpu.GetFlags())
	cpu.PC = cpu.Bus.ReadAddress(0xFFFA)
	cpu.P.I = 1
	cpu.CycleCount += 7
}

func (cpu *CPU) _IRQ() {
	cpu.PushAddress(cpu.PC)
	cpu.Push(cpu.GetFlags())
	cpu.PC = cpu.Bus.ReadAddress(0xFFFE)
	cpu.P.I = 1
	cpu.CycleCount += 7
}

func _ADC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	var temp uint16 = uint16(src) + uint16(cpu.A) + uint16(cpu.P.C)
	cpu.setSign(uint8(temp))
	cpu.setZero(uint8(temp))
	// Set carry
	if temp > 0xff {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}

	// Set overflow
	if ((cpu.A^src)&0x80) == 0 && ((cpu.A^uint8(temp))&0x80) != 0 {
		cpu.P.V = 1
	} else {
		cpu.P.V = 0
	}

	cpu.A = uint8(temp)
}

func _AND(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	cpu.A &= src
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _ASL(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8
	if addressingMode == Accumulator {
		src = cpu.A
	} else {
		src = cpu.Bus.Read(address)
	}

	if src&0x80 != 0 {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}

	src <<= 1
	cpu.setSign(src)
	cpu.setZero(src)

	if addressingMode == Accumulator {
		cpu.A = src
	} else {
		cpu.Bus.Write(address, src)
	}
}

func _BCC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.C == 0 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BCS(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.C == 1 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BEQ(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.Z == 1 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BIT(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	cpu.setSign(src)
	cpu.setZero(src & cpu.A)

	if src&0x40 != 0 {
		cpu.P.V = 1
	} else {
		cpu.P.V = 0
	}
}

func _BMI(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.S == 1 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BNE(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.Z == 0 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BPL(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.S == 0 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BRK(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.PushAddress(cpu.PC)
	cpu.P.B = 1
	cpu.Push(cpu.GetFlags())
	cpu.P.I = 1
	cpu.PC = cpu.Bus.ReadAddress(0xFFFE)
}

func _BVC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.V == 0 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _BVS(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	if cpu.P.V == 1 {
		cpu.PC = address
		cpu.CycleCount++
		if pageCycle {
			cpu.CycleCount++
		}
	}
}

func _CLC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.C = 0
}

func _CLD(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.D = 0
}

func _CLI(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.I = 0
}

func _CLV(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.V = 0
}

func _CMP(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint16 = uint16(cpu.Bus.Read(address))
	src = uint16(cpu.A) - src
	if uint16(src) < 0x100 {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}
	cpu.setSign(uint8(src))
	cpu.setZero(uint8(src))
}

func _CPX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)

	src = cpu.X - src
	if cpu.X >= src {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}
	cpu.setSign(uint8(src))
	cpu.setZero(uint8(src))
}

func _CPY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint16 = uint16(cpu.Bus.Read(address))
	src = uint16(cpu.Y) - src
	if uint16(src) < 0x100 {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}
	cpu.setSign(uint8(src))
	cpu.setZero(uint8(src))
}

func _DEC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address) - 1
	cpu.setSign(src)
	cpu.setZero(src)
	cpu.Bus.Write(address, src)
}

func _DEX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.X--
	cpu.setSign(cpu.X)
	cpu.setZero(cpu.X)
}

func _DEY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Y--
	cpu.setSign(cpu.Y)
	cpu.setZero(cpu.Y)
}

func _EOR(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	cpu.A ^= src
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _INC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	src++
	cpu.setSign(src)
	cpu.setZero(src)
	cpu.Bus.Write(address, src)
}

func _INX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.X++
	cpu.setSign(cpu.X)
	cpu.setZero(cpu.X)
}

func _INY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Y++
	cpu.setSign(cpu.Y)
	cpu.setZero(cpu.Y)
}

func _JMP(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.PC = address
}

func _JSR(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.PC--
	cpu.PushAddress(cpu.PC)
	cpu.PC = address
}

func _LDA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.A = cpu.Bus.Read(address)
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _LDX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.X = cpu.Bus.Read(address)
	cpu.setSign(cpu.X)
	cpu.setZero(cpu.X)
}

func _LDY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Y = cpu.Bus.Read(address)
	cpu.setSign(cpu.Y)
	cpu.setZero(cpu.Y)
}

func _LSR(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8
	if addressingMode == Accumulator {
		src = cpu.A
	} else {
		src = cpu.Bus.Read(address)
	}

	cpu.P.C = src & 0x01
	src >>= 1
	cpu.setSign(src)
	cpu.setZero(src)

	if addressingMode == Accumulator {
		cpu.A = src
	} else {
		cpu.Bus.Write(address, src)
	}
}

func _NOP(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {

}

func _ORA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.A |= cpu.Bus.Read(address)
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _PHA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Push(cpu.A)
}

func _PHP(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	// TODO: Something?
	cpu.Push(cpu.GetFlags())
}

func _PLA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.A = cpu.Pop()
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _PLP(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.SetFlags(cpu.Pop()&0xEF | 0x20)
}

func _ROL(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8
	if addressingMode == Accumulator {
		src = cpu.A
	} else {
		src = cpu.Bus.Read(address)
	}

	var carry uint8 = (src >> 7) & 1

	src <<= 1
	src |= cpu.P.C

	cpu.P.C = carry
	cpu.setSign(src)
	cpu.setZero(src)

	if addressingMode == Accumulator {
		cpu.A = src
	} else {
		cpu.Bus.Write(address, src)
	}
}

func _ROR(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8
	if addressingMode == Accumulator {
		src = cpu.A
	} else {
		src = cpu.Bus.Read(address)
	}

	var carry uint8 = src & 1

	src >>= 1
	src |= (cpu.P.C << 7)

	cpu.P.C = carry
	cpu.setSign(src)
	cpu.setZero(src)

	if addressingMode == Accumulator {
		cpu.A = src
	} else {
		cpu.Bus.Write(address, src)
	}
}

func _RTI(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.SetFlags(cpu.Pop())
	cpu.PC = cpu.PopAddress()
}

func _RTS(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.PC = cpu.PopAddress() + 1
}

func _SBC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	var src uint8 = cpu.Bus.Read(address)
	var temp uint16 = uint16(cpu.A) - uint16(src) - 1 + uint16(cpu.P.C)
	cpu.setSign(uint8(temp))
	cpu.setZero(uint8(temp))

	if temp < 0x100 {
		cpu.P.C = 1
	} else {
		cpu.P.C = 0
	}

	if (cpu.A^uint8(temp))&0x80 != 0 && (cpu.A^src)&0x80 != 0 {
		cpu.P.V = 1
	} else {
		cpu.P.V = 0
	}

	cpu.A = uint8(temp)
}

func _SEC(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.C = 1
}

func _SED(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.D = 1
}

func _SEI(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.P.I = 1
}

func _STA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Bus.Write(address, cpu.A)
}

func _STX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Bus.Write(address, cpu.X)
}

func _STY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Bus.Write(address, cpu.Y)
}

func _TAX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.X = cpu.A
	cpu.setSign(cpu.X)
	cpu.setZero(cpu.X)
}

func _TAY(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.Y = cpu.A
	cpu.setSign(cpu.Y)
	cpu.setZero(cpu.Y)
}

func _TSX(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.X = cpu.SP
	cpu.setSign(cpu.X)
	cpu.setZero(cpu.X)
}

func _TXA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.A = cpu.X
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}

func _TXS(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.SP = cpu.X
}

func _TYA(cpu *CPU, addressingMode uint8, address uint16, pageCycle bool) {
	cpu.A = cpu.Y
	cpu.setSign(cpu.A)
	cpu.setZero(cpu.A)
}
