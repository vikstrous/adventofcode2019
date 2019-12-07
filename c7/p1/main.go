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
	name  string
	code  int
	arity int
	run   func(vm *VM, modes []paramMode) error
}

type paramMode int

const (
	paramModePosition paramMode = iota
	paramModeImmediate
)

func read(cells []int, param int, mode paramMode) int {
	switch mode {
	case paramModePosition:
		return cells[param]
	case paramModeImmediate:
		return param
	}
	panic(mode)
}

type VM struct {
	memory   []int
	ip       int
	trace    bool
	inputer  Inputer
	outputer Outputer
}

type Inputer func() int

type Outputer func(int)

func NewVM(memory []int, inputer Inputer, outputer Outputer) *VM {
	memoryCopy := make([]int, len(memory))
	copy(memoryCopy, memory)
	return &VM{memory: memoryCopy, inputer: inputer, outputer: outputer}
}

func (v *VM) run() error {
	for v.ip < len(v.memory) {
		op, modes := v.decodeOpCode()
		if v.trace {
			fmt.Println("executing:", op.name)
		}
		err := op.run(v, modes)
		if err == errorHalt {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("no HALT found")
}

func (v *VM) decodeOpCode() (opcode, []paramMode) {
	code := v.memory[v.ip]
	op, ok := opcodes[code%100]
	if !ok {
		panic(fmt.Sprintf("failed to parse %d at %d", code, v.ip))
	}
	modeInt := code / 100
	modes := []paramMode{}
	for i := 0; i < op.arity; i++ {
		modes = append(modes, paramMode(modeInt%10))
		modeInt = modeInt / 10
	}
	return op, modes
}

var opcodes = map[int]opcode{
	1: opcode{
		name:  "add",
		code:  1,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			input1 := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			input2 := read(vm.memory, vm.memory[vm.ip+2], modes[1])
			outputAddress := vm.memory[vm.ip+3]
			vm.memory[outputAddress] = input1 + input2
			vm.ip += 4
			return nil
		},
	},
	2: opcode{
		name:  "multiply",
		code:  2,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			input1 := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			input2 := read(vm.memory, vm.memory[vm.ip+2], modes[1])
			outputAddress := vm.memory[vm.ip+3]
			vm.memory[outputAddress] = input1 * input2
			vm.ip += 4
			return nil
		},
	},
	3: opcode{
		name:  "input",
		code:  3,
		arity: 1,
		run: func(vm *VM, modes []paramMode) error {
			outputAddress := vm.memory[vm.ip+1]
			input := vm.inputer()
			vm.memory[outputAddress] = input
			vm.ip += 2
			return nil
		},
	},
	4: opcode{
		name:  "output",
		code:  4,
		arity: 1,
		run: func(vm *VM, modes []paramMode) error {
			output := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			vm.outputer(output)
			vm.ip += 2
			return nil
		},
	},
	5: opcode{
		name:  "jump-if-true",
		code:  5,
		arity: 2,
		run: func(vm *VM, modes []paramMode) error {
			input := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			if input != 0 {
				vm.ip = read(vm.memory, vm.memory[vm.ip+2], modes[1])
				return nil
			}
			vm.ip += 3
			return nil
		},
	},
	6: opcode{
		name:  "jump-if-false",
		code:  6,
		arity: 2,
		run: func(vm *VM, modes []paramMode) error {
			input := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			if input == 0 {
				vm.ip = read(vm.memory, vm.memory[vm.ip+2], modes[1])
				return nil
			}
			vm.ip += 3
			return nil
		},
	},
	7: opcode{
		name:  "less-than",
		code:  7,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			arg1 := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			arg2 := read(vm.memory, vm.memory[vm.ip+2], modes[1])
			if arg1 < arg2 {
				vm.memory[vm.memory[vm.ip+3]] = 1
			} else {
				vm.memory[vm.memory[vm.ip+3]] = 0
			}
			vm.ip += 4
			return nil
		},
	},
	8: opcode{
		name:  "equals",
		code:  8,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			arg1 := read(vm.memory, vm.memory[vm.ip+1], modes[0])
			arg2 := read(vm.memory, vm.memory[vm.ip+2], modes[1])
			if arg1 == arg2 {
				vm.memory[vm.memory[vm.ip+3]] = 1
			} else {
				vm.memory[vm.memory[vm.ip+3]] = 0
			}
			vm.ip += 4
			return nil
		},
	},
	99: opcode{
		name:  "halt",
		code:  99,
		arity: 0,
		run: func(vm *VM, modes []paramMode) error {
			return errorHalt
		},
	},
}

func stdinInputer() int {
	fmt.Printf("> ")
	inStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	input, err := strconv.Atoi(inStr[:len(inStr)-1])
	if err != nil {
		panic(err)
	}
	return input
}
func stdoutOutputer(out int) {
	fmt.Println("OUT:", out)
}

type combinator struct {
	current   []int
	iteration int
}

func newCombinator() *combinator {
	return &combinator{}
}

// gets the next unused number in positions from the first to `pos`
// if mustBeGreater is true, it must also be greather than the number in position `pos`
func (c *combinator) nextUnused(pos int, mustBeGreater bool) int {
	useds := c.current[:pos+1]
	try := 0
	if mustBeGreater {
		try = c.current[pos] + 1
	}
	for ; try < 5; try++ {
		isUsed := false
		for _, used := range useds {
			if try == used {
				isUsed = true
				break
			}
		}
		if !isUsed {
			return try
		}
	}
	panic("impossible: all used")
}

var factorials = []int{1, 1, 2, 6, 24, 120}

// iteration order:
// 0 1 2 3 4
// 0 1 2 4 3
// 0 1 3 2 4
// ...
func (c *combinator) next() ([]int, bool) {
	if c.iteration >= factorials[5]-1 {
		return nil, false
	}
	// initialize iteration 0
	if c.current == nil {
		c.current = []int{0, 1, 2, 3, 4}
		return c.current, true
	}
	// every other iteration increment numbers from the left without repeating any
	c.iteration++

	// numbers at firstChangePos are incremented when the iteration reaches (4-i)!
	firstChangePos := -1
	for i := 0; i < 5; i++ {
		shouldChange := c.iteration%factorials[4-i] == 0
		if shouldChange {
			firstChangePos = i
			break
		}
	}
	if firstChangePos == -1 {
		panic("impossible")
	}
	c.current[firstChangePos] = c.nextUnused(firstChangePos, true)
	// all numbers after that are filled in based on the min unused number so far (from previous positions)
	for i := firstChangePos + 1; i < 5; i++ {
		c.current[i] = c.nextUnused(i-1, false)
	}

	return c.current, true
}

func runProgram(cells []int) error {
	c := newCombinator()
	for {
		next, ok := c.next()
		if !ok {
			break
		}
		fmt.Println(next)
	}
	return nil

	vm := NewVM(cells, stdinInputer, stdoutOutputer)
	return vm.run()
}
