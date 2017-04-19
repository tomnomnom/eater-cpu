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

var flagNames = map[int]string{
	HLT: "HLT",
	MI:  "MI",
	RI:  "RI",
	RO:  "RO",
	IO:  "IO",
	II:  "II",
	AI:  "AI",
	AO:  "AO",
	ZO:  "ZO",
	SU:  "SU",
	BI:  "BI",
	OI:  "OI",
	CE:  "CE",
	CO:  "CO",
	J:   "J",
}

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
	{RO, func(c *cpu) { c.bus = c.ram[c.addr] }},
	{IO, func(c *cpu) { c.bus = c.ir & 0x0F }},
	{AO, func(c *cpu) { c.bus = c.a }},
	{CO, func(c *cpu) { c.bus = c.pc }},
	{ZO, func(c *cpu) {
		if c.flags&SU == 0 {
			c.bus = c.a + c.b
		} else {
			c.bus = c.a - c.b
		}
	}},

	// Inputs
	{BI, func(c *cpu) { c.b = c.bus }},
	{OI, func(c *cpu) { c.out = c.bus }},
	{MI, func(c *cpu) { c.addr = c.bus & 0x0F }},
	{RI, func(c *cpu) { c.ram[c.addr] = c.bus }},
	{II, func(c *cpu) { c.ir = c.bus }},
	{AI, func(c *cpu) { c.a = c.bus }},

	// Misc control
	{HLT, func(c *cpu) {}},
	{CE, func(c *cpu) { c.pc++ }},
	{J, func(c *cpu) { c.pc = c.bus }},
}
