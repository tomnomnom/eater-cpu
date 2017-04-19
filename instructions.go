package main

// the most significant nibble is all that's
// used to identify an instruction
const (
	NOP  uint8 = 0x00
	LDA  uint8 = 0x10
	ADD  uint8 = 0x20
	JMP  uint8 = 0x30
	OUT  uint8 = 0xE0
	HALT uint8 = 0xF0
)

var instructionNames = map[uint8]string{
	NOP:  "NOP",
	LDA:  "LDA",
	ADD:  "ADD",
	JMP:  "ADD",
	OUT:  "OUT",
	HALT: "HALT",
}

// instructionMap is a lookup table that defines
// how the control flags should be set for each
// cycle. Instructions can take a variable number
// of cycles.
var instructionMap = map[uint8][]int{

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
//   op(LDA, 15) -> 0x1E
func op(instr uint8, arg uint8) uint8 {
	return instr | (arg & 0x0F)
}
