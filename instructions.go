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
	JZ   uint8 = 0x70 // jump to address if zero flag is set
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
	JZ:   "JZ",
	OUT:  "OUT",
	HALT: "HALT",
}

// getMicrocode returns the steps for each cycle.
// Instructions can take a variable number of cycles.
func (c *cpu) getMicrocode(inst uint8) []int {

	switch inst {

	// no-op
	case NOP:
		return []int{}

	// load from an address RAM into the A register
	case LDA:
		return []int{
			IO | MI, // instruction register to memory address register
			RO | AI, // RAM to A register
		}

	// add the number from an address in RAM to the A register,
	// via the B register
	case ADD:
		return []int{
			IO | MI,      // instruction register to memory address register
			RO | BI,      // RAM to B register
			ZO | AI | FI, // sum to A register and flags register
		}

	// subtract the number from an address in RAM
	case SUB:
		return []int{
			IO | MI,           // instruction to memory
			RO | BI,           // RAM to B
			ZO | AI | FI | SU, // sum to A register and flags register, with subtraction
		}

	// store the value in the A register in RAM
	case STA:
		return []int{
			IO | MI,
			AO | RI,
		}

	// store literal value in the A register
	case LDI:
		return []int{
			IO | AI,
		}

	// set the program counter to the least significant
	// nibble of the instruction register
	case JMP:
		return []int{
			IO | J,
		}

	// jump if the zero flag is set
	case JZ:
		if c.zero {
			return []int{IO | J}
		} else {
			return []int{}
		}

	// load from the A register to the output register
	case OUT:
		return []int{
			AO | OI, // A register to output register
		}

	// halt the CPU
	case HALT:
		return []int{
			HLT, // Halt :)
		}
	}

	return []int{}
}

// op is a convenience function for combining an
// instruction and an argument. E.g.
//
//	op(LDA, 15) -> 0x1E
func op(instr uint8, arg uint8) uint8 {
	return instr | (arg & 0x0F)
}
