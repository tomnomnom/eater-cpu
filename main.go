package main

import (
	"bytes"
	"fmt"
)

const ramSize = 16

const (
	LDA  uint8 = 0x1
	ADD  uint8 = 0x2
	OUT  uint8 = 0xE
	HALT uint8 = 0xF
)

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

type ctrlHandler struct {
	typ int
	fn  func(*vm)
}

// ctrlHandlers is an ordered list of control signal handlers
var ctrlHandlers = []ctrlHandler{
	// Outputs
	{RO, func(v *vm) { v.bus = v.ram[v.addr] }},
	{IO, func(v *vm) { v.bus = v.ir & 0x0F }},
	{AO, func(v *vm) { v.bus = v.a }},
	{CO, func(v *vm) { v.bus = v.pc }},
	{ZO, func(v *vm) {
		if v.ctrl&SU == 0 {
			v.bus = v.a + v.b
		} else {
			v.bus = v.a - v.b
		}
	}},

	// Inputs
	{BI, func(v *vm) { v.b = v.bus }},
	{OI, func(v *vm) { v.o = v.bus }},
	{MI, func(v *vm) { v.addr = v.bus }},
	{RI, func(v *vm) { v.ram[v.addr] = v.bus }},
	{II, func(v *vm) { v.ir = v.bus }},
	{AI, func(v *vm) { v.a = v.bus }},

	// Misc control
	{HLT, func(v *vm) { close(v.pipeline) }},
	{CE, func(v *vm) { v.pc++ }},
	{J, func(v *vm) {}},
}

type cycle int

type vm struct {
	ram  map[uint8]uint8
	pc   uint8
	addr uint8
	ir   uint8
	a    uint8
	b    uint8
	bus  uint8
	o    uint8

	ctrl int

	pipeline chan cycle
}

// each instruction is a slice of cycles
type instruction []cycle

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

func (v *vm) run() {
	for {
		// fetch
		v.pipeline <- cycle(CO | MI)
		v.pipeline <- cycle(RO | II)
		v.pipeline <- cycle(CE)

		// decode
		if instr, ok := instructionMap[v.ir>>4]; ok {
			for _, cycle := range instr {

				// execute
				v.pipeline <- cycle
			}
		}
	}
}

func irOut(v *vm) {
	v.bus = v.ir & 0x0F
}
func halt(v *vm) {
	close(v.pipeline)
}

func (v *vm) update() {
	for _, h := range ctrlHandlers {
		if v.ctrl&h.typ != 0 {
			h.fn(v)
		}
	}
}

func main() {

	v := &vm{
		ram:  initRAM(),
		pc:   0,
		addr: 0,
		a:    0,
		b:    0,

		pipeline: make(chan cycle),
	}

	// code
	v.ram[0] = op(LDA, 14)
	v.ram[1] = op(ADD, 15)
	v.ram[2] = op(OUT, 0)
	v.ram[3] = op(HALT, 0)

	// data
	v.ram[14] = 28
	v.ram[15] = 14

	go v.run()

	for cy := range v.pipeline {

		v.ctrl = int(cy)
		v.update()
		fmt.Printf("%s\n", v)
	}

}

func initRAM() map[uint8]uint8 {
	ram := make(map[uint8]uint8, ramSize)
	var i uint8
	for i = 0; i < ramSize; i++ {
		ram[i] = 0
	}

	return ram
}

func (v vm) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("-----------------\n")
	//buf.WriteString("RAM:\n")

	//var i uint8
	//for i = 0; i < ramSize; i++ {
	//fmt.Fprintf(buf, "  %02d: %#02x\n", i, v.ram[i])
	//}
	fmt.Fprintf(buf, "  PC: %02d\n", v.pc)
	fmt.Fprintf(buf, "ADDR: %02d\n", v.addr)
	fmt.Fprintf(buf, "  IR: %#02x\n", v.ir)
	fmt.Fprintf(buf, "   A: %#02x\n", v.a)
	fmt.Fprintf(buf, "   B: %#02x\n", v.b)
	fmt.Fprintf(buf, " OUT: %#02x\n", v.o)

	return buf.String()
}

func op(instr uint8, arg uint8) uint8 {
	return (instr << 4) | (arg & 0x0F)
}
