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
	f, err := os.Open(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", os.Args[1], err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read inupt")
	}
	err = scanner.Err()
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
	err = runProgram(cells)
	if err != nil {
		return fmt.Errorf("error in program %w", err)
	}
	return nil
}

var errorHalt = fmt.Errorf("HALT")

type opcode struct {
	code  int
	arity int
	run   func(cells []int, ip *int, modes []paramMode) error
}

type paramMode int

const (
	paramModePosition paramMode = iota
	paramModeImmediate
)

func decodeOpCode(code, ip int) (opcode, []paramMode) {
	op, ok := opcodes[code%100]
	if !ok {
		panic(fmt.Sprintf("failed to parse %d at %d", code, ip))
	}
	modeInt := code / 100
	modes := []paramMode{}
	for i := 0; i < op.arity; i++ {
		modes = append(modes, paramMode(modeInt%10))
		modeInt = modeInt / 10
	}
	return op, modes
}

func read(cells []int, param int, mode paramMode) int {
	switch mode {
	case paramModePosition:
		return cells[param]
	case paramModeImmediate:
		return param
	}
	panic(mode)
}

var opcodes = map[int]opcode{
	// add
	1: opcode{
		code:  1,
		arity: 3,
		run: func(cells []int, ip *int, modes []paramMode) error {
			input1 := read(cells, cells[*ip+1], modes[0])
			input2 := read(cells, cells[*ip+2], modes[1])
			outputAddress := cells[*ip+3]
			cells[outputAddress] = input1 + input2
			*ip += 4
			return nil
		},
	},
	// multiply
	2: opcode{
		code:  2,
		arity: 3,
		run: func(cells []int, ip *int, modes []paramMode) error {
			input1 := read(cells, cells[*ip+1], modes[0])
			input2 := read(cells, cells[*ip+2], modes[1])
			outputAddress := cells[*ip+3]
			cells[outputAddress] = input1 * input2
			*ip += 4
			return nil
		},
	},
	// input
	3: opcode{
		code:  3,
		arity: 1,
		run: func(cells []int, ip *int, modes []paramMode) error {
			outputAddress := cells[*ip+1]
			fmt.Printf("> ")
			inStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return fmt.Errorf("input error: %w", err)
			}
			input, err := strconv.Atoi(inStr[:len(inStr)-1])
			if err != nil {
				return fmt.Errorf("strconv error: %w", err)
			}
			cells[outputAddress] = input
			*ip += 2
			return nil
		},
	},
	// output
	4: opcode{
		code:  4,
		arity: 1,
		run: func(cells []int, ip *int, modes []paramMode) error {
			output := read(cells, cells[*ip+1], modes[0])
			fmt.Println("OUT:", output)
			*ip += 2
			return nil
		},
	},
	// jump-if-true
	5: opcode{
		code:  5,
		arity: 2,
		run: func(cells []int, ip *int, modes []paramMode) error {
			input := read(cells, cells[*ip+1], modes[0])
			if input != 0 {
				*ip = read(cells, cells[*ip+2], modes[1])
				return nil
			}
			*ip += 3
			return nil
		},
	},
	// jump-if-false
	6: opcode{
		code:  6,
		arity: 2,
		run: func(cells []int, ip *int, modes []paramMode) error {
			input := read(cells, cells[*ip+1], modes[0])
			if input == 0 {
				*ip = read(cells, cells[*ip+2], modes[1])
				return nil
			}
			*ip += 3
			return nil
		},
	},
	// less than
	7: opcode{
		code:  7,
		arity: 3,
		run: func(cells []int, ip *int, modes []paramMode) error {
			arg1 := read(cells, cells[*ip+1], modes[0])
			arg2 := read(cells, cells[*ip+2], modes[1])
			if arg1 < arg2 {
				cells[cells[*ip+3]] = 1
			} else {
				cells[cells[*ip+3]] = 0
			}
			*ip += 4
			return nil
		},
	},
	// equals
	8: opcode{
		code:  8,
		arity: 3,
		run: func(cells []int, ip *int, modes []paramMode) error {
			arg1 := read(cells, cells[*ip+1], modes[0])
			arg2 := read(cells, cells[*ip+2], modes[1])
			if arg1 == arg2 {
				cells[cells[*ip+3]] = 1
			} else {
				cells[cells[*ip+3]] = 0
			}
			*ip += 4
			return nil
		},
	},
	99: opcode{
		code:  99,
		arity: 0,
		run: func(cells []int, ip *int, modes []paramMode) error {
			return errorHalt
		},
	},
}

func runProgram(cells []int) error {
	for ip := 0; ip < len(cells); {
		op, modes := decodeOpCode(cells[ip], ip)
		err := op.run(cells, &ip, modes)
		if err == errorHalt {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("no HALT found")
}
