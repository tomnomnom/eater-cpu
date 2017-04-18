package main

// bitfields to represent all of the different control
// flags that can be set
const (
	HLT = 1 << iota // Halt
	MI              // Memory address register in
	RI              // RAM In
	RO              // RAM Out
	IO              // Instruction Register Out
	II              // Instruction Register In
	AI              // A Register In
	AO              // A Register Out
	ZO              // Sum Out
	SU              // Subtract
	BI              // B Register In
	OI              // Output Register In
	CE              // Counter Enable
	CO              // Counter Out
	J               // Jump
)

// a flagHandler associates a function with a flag so that
// the function can be called when that flag is set
type flagHandler struct {
	typ int
	fn  func(*cpu)
}

// flagHandlers is an ordered list of functions that
// define what happens in a cycle when particular
// control flags are set. The outputs must come first
// so that they can write to the bus before the inputs
// try to read from it.
var flagHandlers = []flagHandler{

	// Outputs
	{RO, func(v *cpu) { v.bus = v.ram[v.addr] }},
	{IO, func(v *cpu) { v.bus = v.ir & 0x0F }},
	{AO, func(v *cpu) { v.bus = v.a }},
	{CO, func(v *cpu) { v.bus = v.pc }},
	{ZO, func(v *cpu) {
		if v.flags&SU == 0 {
			v.bus = v.a + v.b
		} else {
			v.bus = v.a - v.b
		}
	}},

	// Inputs
	{BI, func(v *cpu) { v.b = v.bus }},
	{OI, func(v *cpu) { v.out = v.bus }},
	{MI, func(v *cpu) { v.addr = v.bus }},
	{RI, func(v *cpu) { v.ram[v.addr] = v.bus }},
	{II, func(v *cpu) { v.ir = v.bus }},
	{AI, func(v *cpu) { v.a = v.bus }},

	// Misc control
	{HLT, func(v *cpu) { close(v.pipeline) }},
	{CE, func(v *cpu) { v.pc++ }},
	{J, func(v *cpu) {}},
}
