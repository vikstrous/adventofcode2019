package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read inupt")
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	line := scanner.Text()
	cellsStr := strings.Split(line, ",")
	cells := []int{}
	for _, cellStr := range cellsStr {
		cell, err := strconv.Atoi(cellStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", line, err)
		}
		cells = append(cells, cell)
	}
	err = runProgram(cells, 12, 2)
	if err != nil {
		return fmt.Errorf("error in program %w", err)
	}
	output := cells[0]
	fmt.Println(output)
	return nil
}

var errorHalt = fmt.Errorf("HALT")

type opcode struct {
	code      int
	numParams int
	run       func(cells []int, opcodeAddress int) error
}

var opcodes = map[int]opcode{
	1: opcode{
		code:      1,
		numParams: 3,
		run: func(cells []int, opcodeAddress int) error {
			input1Address := cells[opcodeAddress+1]
			input2Address := cells[opcodeAddress+2]
			outputAddress := cells[opcodeAddress+3]
			cells[outputAddress] = cells[input1Address] + cells[input2Address]
			return nil
		},
	},
	2: opcode{
		code:      2,
		numParams: 3,
		run: func(cells []int, opcodeAddress int) error {
			input1Address := cells[opcodeAddress+1]
			input2Address := cells[opcodeAddress+2]
			outputAddress := cells[opcodeAddress+3]
			cells[outputAddress] = cells[input1Address] * cells[input2Address]
			return nil
		},
	},
	99: opcode{
		code:      99,
		numParams: 0,
		run: func(cells []int, opcodeAddress int) error {
			return errorHalt
		},
	},
}

func runProgram(cells []int, nown, verb int) error {
	cells[1] = nown
	cells[2] = verb
	for ip := 0; ip < len(cells); {
		instructionCode := cells[ip]
		op, ok := opcodes[instructionCode]
		if !ok {
			return fmt.Errorf("unknown insturction code %d at %d", instructionCode, ip)
		}
		err := op.run(cells, ip)
		if err == errorHalt {
			return nil
		}
		if err != nil {
			return err
		}
		ip += op.numParams + 1
	}
	return fmt.Errorf("no HALT found")
}
