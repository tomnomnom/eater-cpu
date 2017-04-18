package main

// each instruction is
const (
	LDA  uint8 = 0x1
	ADD  uint8 = 0x2
	OUT  uint8 = 0xE
	HALT uint8 = 0xF
)

// each instruction is a slice of bitfields
// representing which control lines should
// be set for that cycle
type instruction []int

// instruction definitions
var instructionMap = map[uint8]instruction{
	LDA: {
		IO | MI,
		RO | AI,
	},
	ADD: {
		IO | MI,
		RO | BI,
		ZO | AI,
	},
	OUT: {
		AO | OI,
	},
	HALT: {
		HLT,
	},
}
