package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func parseRAMFile(r io.Reader) (map[uint8]uint8, error) {
	ram := initRAM()

	sc := bufio.NewScanner(r)

	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		addr, op, err := parseRAMLine(sc.Text())
		if err != nil {
			return ram, err
		}
		ram[addr] = op
	}
	if err := sc.Err(); err != nil {
		return ram, err
	}

	return ram, nil
}

func parseRAMLine(l string) (uint8, uint8, error) {

	p := strings.Split(l, ":")

	if len(p) != 2 {
		return 0, 0, fmt.Errorf("failed to parse ramfile line [%s]", l)
	}

	addr, err := strconv.Atoi(p[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse address as number [%s]", p[0])
	}

	op, err := parseOp(strings.TrimSpace(p[1]))
	if err != nil {
		return 0, 0, err
	}

	return uint8(addr), op, nil
}

func parseOp(o string) (uint8, error) {
	p := strings.Split(o, " ")

	switch p[0] {

	case "NOP":
		return NOP, nil

	case "OUT":
		return OUT, nil

	case "HALT":
		return HALT, nil

	case "LDA":
		return parseOpWithArg(LDA, p)

	case "ADD":
		return parseOpWithArg(ADD, p)

	default:
		// parse as literal number
		n, err := strconv.Atoi(p[0])
		if err != nil {
			return 0, fmt.Errorf("failed to parse literal number [%s]", o)
		}
		return uint8(n), nil
	}

	return 0, nil
}

func parseOpWithArg(instr uint8, p []string) (uint8, error) {
	if len(p) != 2 {
		return 0, fmt.Errorf("%s without address %s", instructionNames[instr], p)
	}
	addr, err := strconv.Atoi(p[1])
	if err != nil {
		return 0, fmt.Errorf("%s address is non-number %s", instructionNames[instr], p)
	}
	return op(instr, uint8(addr)), nil

}
