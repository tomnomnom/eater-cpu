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

	clock     chan interface{}
	cycleDone chan interface{}
}

// run should launched as a goroutine. It runs the fetch,
// decode, execute cycle
func (v *cpu) run() {
	for {
		// load the program counter into the memory address register
		v.cycle(CO | MI)

		// load ram into the intruction register
		v.cycle(RO | II)

		// increment the counter
		v.cycle(CE)

		// look up the the cycles needed to execute the
		// instruction now loaded into the instruction
		// register. Only the most significant nibble
		// is used to ifentify the instruction
		if cycles, ok := instructionMap[v.ir&0xF0]; ok {
			for _, flags := range cycles {
				v.cycle(flags)
			}
		}
	}
}

// cycle accepts some flags to set and then calls
// all of the corresponding flag handlers
func (v *cpu) cycle(flags int) {
	<-v.clock

	v.flags = flags

	for _, h := range flagHandlers {
		if v.flags&h.typ != 0 {
			h.fn(v)
		}
	}

	// we must signal that the cycle has finished
	v.cycleDone <- struct{}{}
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

func (c *cpu) isHalted() bool {
	return c.flags&HLT != 0
}

// String returns the current state of the cpu as
// a string.
func (v cpu) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("-------------------------\n")

	fmt.Fprintf(buf, "  BUS: %08b (%#02X)\n", v.bus, v.bus)
	fmt.Fprintf(buf, "   PC: %08b (%d)\n", v.pc, v.pc)
	fmt.Fprintf(buf, " ADDR: %08b (%d)\n", v.addr, v.addr)
	fmt.Fprintf(buf, "  RAM: %08b (%d)\n", v.ram[v.addr], v.ram[v.addr])
	fmt.Fprintf(buf, "   IR: %08b (%s %d)\n", v.ir, instructionNames[v.ir&0xF0], v.ir&0x0F)
	fmt.Fprintf(buf, "    A: %08b (%d)\n", v.a, v.a)
	fmt.Fprintf(buf, "    B: %08b (%d)\n", v.b, v.b)
	fmt.Fprintf(buf, "  OUT: %08b (%d)\n", v.out, v.out)

	buf.WriteString("FLAGS: ")
	for flag, name := range flagNames {
		if v.flags&flag != 0 {
			fmt.Fprintf(buf, "%s ", name)
		}
	}
	buf.WriteString("\n")
	buf.WriteString("-------------------------\n")

	return buf.String()
}

func main() {

	// initialise the RAM and the clock
	v := &cpu{
		ram:       initRAM(),
		clock:     make(chan interface{}),
		cycleDone: make(chan interface{}),
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

	// start the CPU
	go v.run()

	for {
		if v.isHalted() {
			break
		}
		v.clock <- struct{}{}
		<-v.cycleDone
		fmt.Printf("%s\n", v)
	}
}
