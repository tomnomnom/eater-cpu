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

type vm struct {
	ram  map[uint8]uint8
	pc   uint8
	addr uint8
	ir   uint8
	a    uint8
	b    uint8
	bus  uint8

	// subtract flag
	sub bool

	pipeline chan cycle
}

type signal func(*vm)
type cycle []signal

type instr []cycle

func pcOut(v *vm) {
	v.bus = v.pc
}

func addrIn(v *vm) {
	v.addr = v.bus
}

func ramOut(v *vm) {
	v.bus = v.ram[v.addr]
}

func irIn(v *vm) {
	v.ir = v.bus
}

func irOut(v *vm) {
	v.bus = v.ir & 0x0F
}

func pcInc(v *vm) {
	v.pc++
}

func aIn(v *vm) {
	v.a = v.bus
}

func bIn(v *vm) {
	v.b = v.bus
}

func sumOut(v *vm) {

	if v.sub {
		v.bus = v.a - v.b
	} else {
		v.bus = v.a + v.b
	}

	// calling sumOut resets the subtract flag
	v.sub = false

}

func sub(v *vm) {
	v.sub = true
}

var instrs = map[uint8]instr{
	LDA: {
		{irOut, addrIn},
		{ramOut, aIn},
	},
	ADD: {
		{irOut, addrIn},
		{ramOut, bIn},
		{sumOut, aIn},
	},
	HALT: {
		{halt},
	},
}

func (v *vm) run() {
	for {
		// fetch
		v.pipeline <- cycle{pcOut, addrIn}
		v.pipeline <- cycle{ramOut, irIn}
		v.pipeline <- cycle{pcInc}

		// decode
		if instr, ok := instrs[v.ir>>4]; ok {
			for _, cycle := range instr {

				// execute
				v.pipeline <- cycle
			}
		}
	}
}

func halt(v *vm) {
	close(v.pipeline)
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
		fmt.Printf("%s\n", v)

		// there is often multiple signals to set
		// for a given cycle
		for _, sig := range cy {
			sig(v)
		}
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
	buf.WriteString("RAM:\n")

	var i uint8
	for i = 0; i < ramSize; i++ {
		fmt.Fprintf(buf, "  %02d: %#02x\n", i, v.ram[i])
	}
	fmt.Fprintf(buf, "PC:\n  %02d\n", v.pc)
	fmt.Fprintf(buf, "ADDR:\n  %02d\n", v.addr)
	fmt.Fprintf(buf, "IR:\n  %#02x\n", v.ir)
	fmt.Fprintf(buf, "A:\n  %#02x\n", v.a)

	return buf.String()
}

func op(instr uint8, arg uint8) uint8 {
	return (instr << 4) | (arg & 0x0F)
}
