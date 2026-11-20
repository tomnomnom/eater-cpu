package main

// the most significant nibble is all that's
// used to identify an instruction
const (
	NOP  uint8 = 0x00 // no-op
	LDA  uint8 = 0x10 // load memory into A register
	ADD  uint8 = 0x20 // add value at address to A register
	SUB  uint8 = 0x30 // subtract value at address from A register
	STA  uint8 = 0x40 // store A register at address
	LDI  uint8 = 0x50 // load immediate into A register
	JMP  uint8 = 0x60 // jump to address
	OUT  uint8 = 0xE0 // A register to output reguster
	HALT uint8 = 0xF0 // halt :)
)

var instructionNames = map[uint8]string{
	NOP:  "NOP",
	LDA:  "LDA",
	ADD:  "ADD",
	SUB:  "SUB",
	STA:  "STA",
	LDI:  "LDI",
	JMP:  "JMP",
	OUT:  "OUT",
	HALT: "HALT",
}

// instructionMap is a lookup table that defines
// how the control flags should be set for each
// cycle. Instructions can take a variable number
// of cycles.
var instructionMap = map[uint8][]int{

	// no-op
	NOP: {},

	// load from an address RAM into the A register
	LDA: {
		IO | MI, // instruction register to memory address register
		RO | AI, // RAM to A register
	},

	// add the number from an address in RAM to the A register,
	// via the B register
	ADD: {
		IO | MI, // instruction register to memory address register
		RO | BI, // RAM to B register
		ZO | AI, // sum to A register
	},

	// subtract the number from an address in RAM
	SUB: {
		IO | MI,      // instruction to memory
		RO | BI,      // RAM to B
		ZO | AI | SU, // sum to A, with subtraction
	},

	// store the value in the A register in RAM
	STA: {
		IO | MI,
		AO | RI,
	},

	// store literal value in the A register
	LDI: {
		IO | AI,
	},

	// set the program counter to the least significant
	// nibble of the instruction register
	JMP: {
		IO | J,
	},

	// load from the A register to the output register
	OUT: {
		AO | OI, // A register to output register
	},

	// halt the CPU
	HALT: {
		HLT, // Halt :)
	},
}

// op is a convenience function for combining an
// instruction and an argument. E.g.
//
//	op(LDA, 15) -> 0x1E
func op(instr uint8, arg uint8) uint8 {
	return instr | (arg & 0x0F)
}
