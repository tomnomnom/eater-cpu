package main

import (
	"bytes"
	"fmt"
)

// 16 bytes of ram
const ramSize = 16

type vm struct {
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
// decode and load to pipeline for the vm
func (v *vm) run() {
	for {
		// load the program counter into the memory address register
		v.pipeline <- CO | MI

		// load ram into the intruction register
		v.pipeline <- RO | II

		// increment the counter
		v.pipeline <- CE

		// look up the the cycles needed to execute the
		// instruction now loaded into the instruction
		// register
		if cycles, ok := instructionMap[v.ir>>4]; ok {
			for _, cycle := range cycles {
				v.pipeline <- cycle
			}
		}
	}
}

// clockPulse ranges over all of the flag handlers, calling
// them if the corresponding control flag is set
func (v *vm) clockPulse() {
	for _, h := range flagHandlers {
		if v.flags&h.typ != 0 {
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

		pipeline: make(chan int),
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

		v.flags = int(cy)
		v.clockPulse()
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
	fmt.Fprintf(buf, "   A: %#d\n", v.a)
	fmt.Fprintf(buf, "   B: %#d\n", v.b)
	fmt.Fprintf(buf, " OUT: %#d\n", v.out)

	return buf.String()
}

func op(instr uint8, arg uint8) uint8 {
	return (instr << 4) | (arg & 0x0F)
}
