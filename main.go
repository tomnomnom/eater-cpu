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
func (c *cpu) run() {
	for {
		// load the program counter into the memory address register
		c.cycle(CO | MI)

		// load ram into the intruction register
		c.cycle(RO | II)

		// increment the counter
		c.cycle(CE)

		// look up the the cycles needed to execute the
		// instruction now loaded into the instruction
		// register. Only the most significant nibble
		// is used to ifentify the instruction
		if cycles, ok := instructionMap[c.ir&0xF0]; ok {
			for _, flags := range cycles {
				c.cycle(flags)
			}
		}
	}
}

// cycle accepts some flags to set and then calls
// all of the corresponding flag handlers
func (c *cpu) cycle(flags int) {
	<-c.clock

	c.flags = flags

	for _, h := range flagHandlers {
		if c.flags&h.typ != 0 {
			h.fn(c)
		}
	}

	// we must signal that the cycle has finished
	c.cycleDone <- struct{}{}
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

// isHalted reports if the halt flag is set
func (c *cpu) isHalted() bool {
	return c.flags&HLT != 0
}

// String returns the current state of the cpu as a string.
func (c *cpu) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("-------------------------\n")

	fmt.Fprintf(buf, "  BUS: %08b (%#02X)\n", c.bus, c.bus)
	fmt.Fprintf(buf, "   PC: %08b (%d)\n", c.pc, c.pc)
	fmt.Fprintf(buf, " ADDR: %08b (%d)\n", c.addr, c.addr)
	fmt.Fprintf(buf, "  RAM: %08b (%d)\n", c.ram[c.addr], c.ram[c.addr])
	fmt.Fprintf(buf, "   IR: %08b (%s %d)\n", c.ir, instructionNames[c.ir&0xF0], c.ir&0x0F)
	fmt.Fprintf(buf, "    A: %08b (%d)\n", c.a, c.a)
	fmt.Fprintf(buf, "    B: %08b (%d)\n", c.b, c.b)
	fmt.Fprintf(buf, "  OUT: %08b (%d)\n", c.out, c.out)

	buf.WriteString("FLAGS: ")
	for flag, name := range flagNames {
		if c.flags&flag != 0 {
			fmt.Fprintf(buf, "%s ", name)
		}
	}
	buf.WriteString("\n")
	buf.WriteString("-------------------------\n")

	return buf.String()
}

func main() {

	// initialise the RAM and the clock
	c := &cpu{
		ram:       initRAM(),
		clock:     make(chan interface{}),
		cycleDone: make(chan interface{}),
	}

	// code
	// load from address 14 to the A register
	c.ram[0] = op(LDA, 14)

	// add from address 15 to the A register
	c.ram[1] = op(ADD, 15)

	// load from the A register to the output register
	c.ram[2] = op(OUT, 0)

	// halt the CPU
	c.ram[3] = op(HALT, 0)

	// data
	c.ram[14] = 28
	c.ram[15] = 14

	// start the CPU
	go c.run()

	for {
		if c.isHalted() {
			break
		}

		// pulse the clock
		c.clock <- struct{}{}

		// wait for the cycle to finish
		<-c.cycleDone

		// dump the state of the cpu
		fmt.Printf("%s\n", c)
	}
}
