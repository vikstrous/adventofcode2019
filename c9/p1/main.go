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
	cells := []int64{}
	for _, cellStr := range cellsStr {
		cell, err := strconv.ParseInt(cellStr, 10, 64)
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

type paramMode int64

const (
	paramModePosition paramMode = iota
	paramModeImmediate
	paramModeRelative
)

type VM struct {
	memory   []int64
	ip       int64
	trace    bool
	inputer  Inputer
	outputer Outputer
	relbase  int64
}

func (v *VM) read(arg int64, modes []paramMode) (read int64) {
	defer func() {
		if v.trace {
			fmt.Println("read:", read)
		}
	}()
	param := v.memory[v.ip+arg]
	mode := modes[arg-1]
	switch mode {
	case paramModePosition:
		if int64(len(v.memory)) < param {
			return 0
		}
		return v.memory[param]
	case paramModeImmediate:
		return param
	case paramModeRelative:
		if int64(len(v.memory)) < param+v.relbase {
			return 0
		}
		return v.memory[param+v.relbase]
	}
	panic(mode)
}
func (v *VM) outputAddress(arg int64, modes []paramMode) (read int64) {
	param := v.memory[v.ip+arg]
	mode := modes[arg-1]
	switch mode {
	case paramModePosition:
		return param
	case paramModeRelative:
		return param + v.relbase
	}
	panic(mode)
}
func (v *VM) write(address int64, value int64) {
	if int64(len(v.memory)) < (address + 1) {
		v.memory = append(v.memory, make([]int64, int(address+1)-len(v.memory))...)
	}
	v.memory[address] = value
	if v.trace {
		fmt.Println("write", value, "to", address)
	}
}

type Inputer func() int64

type Outputer func(int64)

func NewVM(memory []int64, inputer Inputer, outputer Outputer) *VM {
	memoryCopy := make([]int64, len(memory))
	copy(memoryCopy, memory)
	return &VM{memory: memoryCopy, inputer: inputer, outputer: outputer, trace: false}
}

func (v *VM) runToOutput() error {
	for v.ip < int64(len(v.memory)) {
		op, modes := v.decodeOpCode()
		if v.trace {
			fmt.Println("status: relbase", v.relbase, "ip", v.ip, "memory", len(v.memory))
			args := v.memory[v.ip+1 : int(v.ip)+op.arity+1]
			fmt.Println("executing:", op.name, args, modes)
		}
		err := op.run(v, modes)
		if err != nil {
			return err
		}
		if op.name == "output" {
			return nil
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
	modeint64 := code / 100
	modes := []paramMode{}
	for i := 0; i < op.arity; i++ {
		modes = append(modes, paramMode(modeint64%10))
		modeint64 = modeint64 / 10
	}
	return op, modes
}

var opcodes = map[int64]opcode{
	1: opcode{
		name:  "add",
		code:  1,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			input1 := vm.read(1, modes)
			input2 := vm.read(2, modes)
			outputAddress := vm.outputAddress(3, modes)
			vm.write(outputAddress, input1+input2)
			vm.ip += 4
			return nil
		},
	},
	2: opcode{
		name:  "multiply",
		code:  2,
		arity: 3,
		run: func(vm *VM, modes []paramMode) error {
			input1 := vm.read(1, modes)
			input2 := vm.read(2, modes)
			outputAddress := vm.outputAddress(3, modes)
			vm.write(outputAddress, input1*input2)
			vm.ip += 4
			return nil
		},
	},
	3: opcode{
		name:  "input",
		code:  3,
		arity: 1,
		run: func(vm *VM, modes []paramMode) error {
			outputAddress := vm.outputAddress(1, modes)
			input := vm.inputer()
			vm.write(outputAddress, input)
			vm.ip += 2
			return nil
		},
	},
	4: opcode{
		name:  "output",
		code:  4,
		arity: 1,
		run: func(vm *VM, modes []paramMode) error {
			output := vm.read(1, modes)
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
			input := vm.read(1, modes)
			if input != 0 {
				vm.ip = vm.read(2, modes)
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
			input := vm.read(1, modes)
			if input == 0 {
				vm.ip = vm.read(2, modes)
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
			arg1 := vm.read(1, modes)
			arg2 := vm.read(2, modes)
			if arg1 < arg2 {
				vm.write(vm.outputAddress(3, modes), 1)
			} else {
				vm.write(vm.outputAddress(3, modes), 0)
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
			arg1 := vm.read(1, modes)
			arg2 := vm.read(2, modes)
			if arg1 == arg2 {
				vm.write(vm.outputAddress(3, modes), 1)
			} else {
				vm.write(vm.outputAddress(3, modes), 0)
			}
			vm.ip += 4
			return nil
		},
	},
	9: opcode{
		name:  "add-relbase",
		code:  9,
		arity: 1,
		run: func(vm *VM, modes []paramMode) error {
			arg1 := vm.read(1, modes)
			vm.relbase += arg1
			vm.ip += 2
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

func stdinInputter() int64 {
	fmt.Printf("> ")
	inStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	input, err := strconv.ParseInt(inStr[:len(inStr)-1], 10, 64)
	if err != nil {
		panic(err)
	}
	return input
}
func stdoutOutputter(out int64) {
	fmt.Println("OUT:", out)
}

func makeReferenceInputter(input *int64) func() int64 {
	return func() int64 {
		return *input
	}
}
func makeSingleOutputter(target *int64) func(int64) {
	return func(output int64) {
		*target = output
	}
}

func makePrefixedInputter(firstInput int64, f func() int64) func() int64 {
	first := true
	return func() int64 {
		if first {
			first = false
			return firstInput
		}
		return f()
	}
}

func runProgram(cells []int64) error {
	vm := NewVM(cells, stdinInputter, stdoutOutputter)
	for {
		err := vm.runToOutput()
		if err == errorHalt {
			break
		}
	}
	return nil
}
