package main

import (
	"bytes"
	"fmt"
)

// 16 bytes of ram
const ramSize = 16

type cpu struct {
	ram  map[uint8]uint8 // ram
	pc   uint8           // program counter
	addr uint8           // address register
	ir   uint8           // instruction register
	a    uint8           // a register
	b    uint8           // b register
	bus  uint8           // bus
	out  uint8           // output register

	// flags holds a bitfield representing which
	// control flags are set. E.g. to set the
	// instruction register out and the memory
	// address register in, it would be set to:
	//
	//   IO | MI
	flags int

	// pipeline is a channel of ints, where
	// each int is actually just a bitfield
	// of which control flags should be set
	// for that cycle.
	pipeline chan int
}

// run should launched as a goroutine. It runs the fetch,
// decode and load to pipeline for the cpu
func (v *cpu) run() {
	for {
		// load the program counter into the memory address register
		v.pipeline <- CO | MI

		// load ram into the intruction register
		v.pipeline <- RO | II

		// increment the counter
		v.pipeline <- CE

		// look up the the cycles needed to execute the
		// instruction now loaded into the instruction
		// register. Only the most significant nibble
		// is used to ifentify the instruction
		if cycles, ok := instructionMap[v.ir&0xF0]; ok {
			for _, cycle := range cycles {
				v.pipeline <- cycle
			}
		}
	}
}

// clockPulse takes the next set of flags off the
// pipeline, and then runs the flag handler for
// each control flag that is set. It returns true
// if the cpu is not yet halted.
func (v *cpu) clockPulse() bool {
	flags, ok := <-v.pipeline
	if !ok {
		return false
	}

	v.flags = flags

	for _, h := range flagHandlers {
		if v.flags&h.typ != 0 {
			h.fn(v)
		}
	}
	return true
}

// initRAM initialises a map for use as the RAM
func initRAM() map[uint8]uint8 {
	ram := make(map[uint8]uint8, ramSize)
	var i uint8
	for i = 0; i < ramSize; i++ {
		ram[i] = 0
	}

	return ram
}

// String returns the current state of the cpu as
// a string.
func (v cpu) String() string {
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
	fmt.Fprintf(buf, "   A: %#d\n", v.a)
	fmt.Fprintf(buf, "   B: %#d\n", v.b)
	fmt.Fprintf(buf, " OUT: %#d\n", v.out)

	return buf.String()
}

func main() {

	// initialise the cycle pipeline and the RAM
	v := &cpu{
		ram:      initRAM(),
		pipeline: make(chan int),
	}

	// code
	// load from address 14 to the A register
	v.ram[0] = op(LDA, 14)

	// add from address 15 to the A register
	v.ram[1] = op(ADD, 15)

	// load from the A register to the output register
	v.ram[2] = op(OUT, 0)

	// halt the CPU
	v.ram[3] = op(HALT, 0)

	// data
	v.ram[14] = 28
	v.ram[15] = 14

	// start filling the cycle pipeline
	go v.run()

	for {
		// v.clockPulse() returns false when
		// the CPU is halted
		if v.clockPulse() {
			fmt.Printf("%s\n", v)
		} else {
			break
		}
	}

}
