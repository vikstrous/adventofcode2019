package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
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

type Point struct {
	X int64
	Y int64
}

type Game struct {
	score       int64
	tiles       map[Point]TileID
	drawBufferX *int64
	drawBufferY *int64
}

func (g *Game) getTiles() []Tile {
	tiles := []Tile{}
	for p, t := range g.tiles {
		tiles = append(tiles, Tile{Point: p, TileID: t})
	}
	return tiles
}

func (g *Game) getX(tileID TileID) int64 {
	for p, t := range g.tiles {
		if t == tileID {
			return p.X
		}
	}
	panic("not found")
}

func (g *Game) AI() int64 {
	ballX := g.getX(TileIDBall)
	paddleX := g.getX(TileIDPaddle)
	if ballX > paddleX {
		return 1
	} else if ballX < paddleX {
		return -1
	}
	return 0
}

func NewGame() *Game {
	return &Game{
		tiles: map[Point]TileID{},
	}
}

func (g *Game) AcceptDraw(i int64) {
	if g.drawBufferX == nil {
		g.drawBufferX = &i
		return
	}
	if g.drawBufferY == nil {
		g.drawBufferY = &i
		return
	}
	if *g.drawBufferX == -1 && *g.drawBufferY == 0 {
		g.score = i
	} else {
		g.tiles[Point{X: *g.drawBufferX, Y: *g.drawBufferY}] = TileID(i)
	}
	g.drawBufferX = nil
	g.drawBufferY = nil
}

type TileID int64

const (
	TileIDEmpty TileID = iota
	TileIDWall
	TileIDBlock
	TileIDPaddle
	TileIDBall
)

type Tile struct {
	Point
	TileID TileID
}

func (t TileID) Symbol() rune {
	symbol := ""
	switch t {
	case TileIDEmpty:
		symbol = " "
	case TileIDBall:
		symbol = "o"
	case TileIDWall:
		symbol = "|"
	case TileIDBlock:
		symbol = "#"
	case TileIDPaddle:
		symbol = "X"
	default:
		panic(t)
	}
	return rune(symbol[0])
}
func (t TileID) Color() termbox.Attribute {
	switch t {
	case TileIDEmpty:
		return termbox.ColorBlack
	case TileIDBall:
		return termbox.ColorGreen
	case TileIDWall:
		return termbox.ColorBlack
	case TileIDBlock:
		return termbox.ColorRed
	case TileIDPaddle:
		return termbox.ColorBlue
	}
	panic(t)
}

func drawScreen(ts []Tile) {
	for _, t := range ts {
		termbox.SetCell(int(t.Point.X), int(t.Point.Y), ' ', termbox.ColorWhite, t.TileID.Color())
	}
}
func drawTiles(ps []Tile) [][]string {
	minX := int64(0)
	maxX := int64(0)
	minY := int64(0)
	maxY := int64(0)
	for _, p := range ps {
		if p.X > maxX {
			maxX = p.X
		}
		if p.X < minX {
			minX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
		if p.Y < minY {
			minY = p.Y
		}
	}
	width := maxX - minX + 1
	height := maxY - minY + 1
	rows := [][]string{}
	for i := int64(0); i < height; i++ {
		row := []string{}
		for j := int64(0); j < width; j++ {
			row = append(row, " ")
		}
		rows = append(rows, row)
	}
	for _, p := range ps {
		symbol := p.TileID.Symbol()
		rows[p.Y-minY][p.X-minX] = string(symbol)
	}
	return rows
}

func printDrawing(grid [][]string) {
	for _, row := range grid {
		for _, cell := range row {
			fmt.Printf(cell)
		}
		fmt.Println()
	}
}

func runProgram(cells []int64) error {
	cells[0] = 2
	draw := false
	if draw {
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
		defer termbox.Close()
	}

	eventQueue := make(chan termbox.Event)
	if draw {
		go func() {
			for {
				eventQueue <- termbox.PollEvent()
			}
		}()
		termbox.Clear(termbox.ColorWhite, termbox.ColorWhite)
		termbox.Flush()
	}

	control := func() int64 {
		ev := <-eventQueue
		if ev.Type == termbox.EventKey {
			switch {
			case ev.Key == termbox.KeyArrowLeft:
				return -1
			case ev.Key == termbox.KeyArrowRight:
				return 1
			case ev.Key == termbox.KeyArrowDown:
				return 0
			}
		}
		panic(fmt.Sprintf("%v %v", ev.Type, ev.Key))
	}
	g := NewGame()
	useAI := true
	if useAI {
		control = g.AI
	}

	vm := NewVM(cells, control, g.AcceptDraw)
	frame := 0
	for {
		err := vm.runToOutput()
		if err == errorHalt {
			break
		}
		if draw && frame >= 2393 {
			for i, c := range fmt.Sprint(g.score) {
				termbox.SetCell(i, 30, c, termbox.ColorWhite, termbox.ColorBlack)
			}
			drawScreen(g.getTiles())
			termbox.Flush()
		}
		frame += 1
	}
	fmt.Println(g.score)
	return nil
}
