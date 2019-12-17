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
	lastDirection Direction
	droidLocaton  Point
	tiles         map[Point]TileID
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

func NewGame() *Game {
	return &Game{
		tiles: map[Point]TileID{Point{}: TileIDDroid},
	}
}

func (g *Game) AcceptStatus(i int64) {
	status := DroidStatus(i)
	switch status {
	case DroidStatusWall:
		x, y := g.lastDirection.Offsets()
		g.tiles[Point{X: g.droidLocaton.X + x, Y: g.droidLocaton.Y + y}] = TileIDWall
		return
	case DroidStatusMoved:
		g.tiles[Point{X: g.droidLocaton.X, Y: g.droidLocaton.Y}] = TileIDEmpty
		x, y := g.lastDirection.Offsets()
		g.droidLocaton.X += x
		g.droidLocaton.Y += y
		g.tiles[Point{X: g.droidLocaton.X, Y: g.droidLocaton.Y}] = TileIDDroid
		return
	case DroidStatusOxygen:
		g.tiles[Point{X: g.droidLocaton.X, Y: g.droidLocaton.Y}] = TileIDEmpty
		x, y := g.lastDirection.Offsets()
		g.droidLocaton.X += x
		g.droidLocaton.Y += y
		g.tiles[Point{X: g.droidLocaton.X, Y: g.droidLocaton.Y}] = TileIDOxygen
		return
	}
}

type TileID int64

const (
	TileIDEmpty TileID = iota
	TileIDWall
	TileIDDroid
	TileIDOxygen
)

type Tile struct {
	Point
	TileID TileID
}

func (t TileID) Color() termbox.Attribute {
	switch t {
	case TileIDEmpty:
		return termbox.ColorBlack
	case TileIDWall:
		return termbox.ColorBlue
	case TileIDDroid:
		return termbox.ColorGreen
	case TileIDOxygen:
		return termbox.ColorRed
	}
	panic(t)
}

func drawScreen(ts []Tile) {
	gi := dimentions(ts)
	for _, t := range ts {
		x, y := gi.PointToScreen(t.Point)
		termbox.SetCell(x, y, ' ', termbox.ColorWhite, t.TileID.Color())
	}
}

type gridInfo struct {
	minX   int64
	maxX   int64
	minY   int64
	maxY   int64
	width  int64
	height int64
}

func (g gridInfo) PointToScreen(p Point) (int, int) {
	return int(p.X - g.minX + 4), int(p.Y - g.minY + 5)
}

func dimentions(ts []Tile) gridInfo {
	gi := gridInfo{}
	for _, t := range ts {
		if t.X > gi.maxX {
			gi.maxX = t.X
		}
		if t.X < gi.minX {
			gi.minX = t.X
		}
		if t.Y > gi.maxY {
			gi.maxY = t.Y
		}
		if t.Y < gi.minY {
			gi.minY = t.Y
		}
	}
	gi.width = gi.maxX - gi.minX + 1
	gi.height = gi.maxY - gi.minY + 1
	return gi
}

type Direction int64

const (
	DirectionInvalid Direction = iota
	DirectionNorth
	DirectionSouth
	DirectionWest
	DirectionEast
)

func (d Direction) Offsets() (int64, int64) {
	switch d {
	case DirectionNorth:
		return 0, -1
	case DirectionSouth:
		return 0, 1
	case DirectionWest:
		return -1, 0
	case DirectionEast:
		return 1, 0
	}
	panic(d)
}

type DroidStatus int64

const (
	DroidStatusWall DroidStatus = iota
	DroidStatusMoved
	DroidStatusOxygen
)

func runProgram(cells []int64) error {
	draw := true
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
	}

	g := NewGame()
	control := func() int64 {
		for {
			ev := <-eventQueue
			if ev.Type == termbox.EventKey {
				move := DirectionInvalid
				switch {
				case ev.Key == termbox.KeyEnd:
					panic("exit")
				case ev.Key == termbox.KeyArrowUp:
					move = DirectionNorth
				case ev.Key == termbox.KeyArrowLeft:
					move = DirectionWest
				case ev.Key == termbox.KeyArrowRight:
					move = DirectionEast
				case ev.Key == termbox.KeyArrowDown:
					move = DirectionSouth
				}
				if move != 0 {
					g.lastDirection = move
					return int64(move)
				}
			}
			termbox.Clear(termbox.ColorWhite, termbox.ColorWhite)
			for i, c := range fmt.Sprintf("X,Y: %v", g.droidLocaton) {
				termbox.SetCell(i, 0, c, termbox.ColorWhite, termbox.ColorBlack)
			}
			drawScreen(g.getTiles())
			termbox.Flush()
			//panic(fmt.Sprintf("%v %v", ev.Type, ev.Key))
		}
	}
	useAI := false
	if useAI {
		//control = g.AI
	}
	if draw {
		termbox.Clear(termbox.ColorWhite, termbox.ColorWhite)
		for i, c := range fmt.Sprintf("X,Y: %v", g.droidLocaton) {
			termbox.SetCell(i, 0, c, termbox.ColorWhite, termbox.ColorBlack)
		}
		drawScreen(g.getTiles())
		termbox.Flush()
	}

	vm := NewVM(cells, control, g.AcceptStatus)
	for {
		err := vm.runToOutput()
		if err == errorHalt {
			break
		}
		if draw {
			termbox.Clear(termbox.ColorWhite, termbox.ColorWhite)
			for i, c := range fmt.Sprintf("X,Y: %v", g.droidLocaton) {
				termbox.SetCell(i, 0, c, termbox.ColorWhite, termbox.ColorBlack)
			}
			drawScreen(g.getTiles())
			termbox.Flush()
		}
	}
	return nil
}
