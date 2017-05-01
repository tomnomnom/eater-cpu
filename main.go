package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
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

	fmt.Fprintf(buf, "  BUS: %s (%d)\n", ledString(c.bus), c.bus)
	fmt.Fprintf(buf, "   PC: %s (%d)\n", ledString(c.pc), c.pc)
	fmt.Fprintf(buf, " ADDR: %s (%d)\n", ledString(c.addr), c.addr)
	fmt.Fprintf(buf, "  RAM: %s (%d)\n", ledString(c.ram[c.addr]), c.ram[c.addr])
	fmt.Fprintf(buf, "   IR: %s (%s %d)\n", ledString(c.ir), instructionNames[c.ir&0xF0], c.ir&0x0F)
	fmt.Fprintf(buf, "    A: %s (%d)\n", ledString(c.a), c.a)
	fmt.Fprintf(buf, "    B: %s (%d)\n", ledString(c.b), c.b)
	fmt.Fprintf(buf, "  OUT: %s (%d)\n", ledString(c.out), c.out)

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

func ledString(n uint8) string {
	red := color.New(color.FgRed).FprintfFunc()

	buf := &bytes.Buffer{}
	i := uint8(8)
	for {
		i--

		if n>>i&1 == 0 {
			red(buf, "○")
		} else {
			red(buf, "●")
		}

		if i == 0 {
			break
		}
	}
	return buf.String()
}

func main() {

	flag.Parse()

	ramfile := flag.Arg(0)
	if ramfile == "" {
		fmt.Println("usage: eater-cpu <ramfile>")
		os.Exit(1)
	}

	f, err := os.Open(ramfile)
	if err != nil {
		fmt.Printf("failed to open ramfile (%s)\n", err)
		os.Exit(2)
	}

	ram, err := parseRAMFile(f)
	if err != nil {
		fmt.Printf("failed to parse ramfile (%s)\n", err)
		os.Exit(3)
	}

	// initialise the RAM and the clock
	c := &cpu{
		ram:       ram,
		clock:     make(chan interface{}),
		cycleDone: make(chan interface{}),
	}

	// start the CPU
	go c.run()

	// open stdin so the user can hit return to
	// pulse the clock
	//in := bufio.NewReader(os.Stdin)

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

		// wait for user input
		//_, _ = in.ReadString('\n')
	}
}
