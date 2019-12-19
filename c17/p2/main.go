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

type Point struct {
	X int64
	Y int64
}
type Direction int64

const (
	DirectionInvalid Direction = iota
	DirectionNorth
	DirectionSouth
	DirectionWest
	DirectionEast
)

func (d Direction) Apply(p Point) Point {
	x, y := d.Offsets()
	return Point{X: p.X + x, Y: p.Y + y}
}
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

type Game struct {
	tiles     map[Point]TileID
	drawPoint Point
}

func (g *Game) getTilesFor(searches []TileID) []Tile {
	tiles := []Tile{}
	for p, t := range g.tiles {
		for _, s := range searches {
			if s == t {
				tiles = append(tiles, Tile{Point: p, TileID: t})
			}
		}
	}
	return tiles
}
func (g *Game) getTiles() []Tile {
	tiles := []Tile{}
	for p, t := range g.tiles {
		tiles = append(tiles, Tile{Point: p, TileID: t})
	}
	return tiles
}
func (g *Game) getIntersections() []Tile {
	intersections := []Tile{}
	for p, t := range g.tiles {
		if t == TileIDScaffold {
			isIntersection := true
			for _, d := range []Direction{DirectionNorth, DirectionWest, DirectionEast, DirectionSouth} {
				if t2 := g.tiles[d.Apply(p)]; t2 != TileIDScaffold {
					isIntersection = false
					break
				}
			}
			if isIntersection {
				intersections = append(intersections, Tile{TileID: t, Point: p})
			}
		}
	}
	return intersections
}

func NewGame() *Game {
	return &Game{
		tiles: map[Point]TileID{},
	}
}

func (g *Game) AcceptDraw(i int64) {
	if rune(i) == '\n' {
		g.drawPoint = Point{X: 0, Y: g.drawPoint.Y + 1}
		return
	}
	tile := TileIDEmpty
	switch rune(i) {
	case '.':
		tile = TileIDEmpty
	case '#':
		tile = TileIDScaffold
	case '^':
		tile = TileIDRobotUp
	case '<':
		tile = TileIDRobotLeft
	case '>':
		tile = TileIDRobotRight
	case 'v':
		tile = TileIDRobotDown
	case 'X':
		tile = TileIDRobotDead
	default:
		panic(i)
	}
	g.tiles[g.drawPoint] = tile
	g.drawPoint.X++
}

type TileID int64

const (
	TileIDEmpty      TileID = iota // .
	TileIDScaffold                 // #
	TileIDRobotUp                  // ^
	TileIDRobotLeft                // <
	TileIDRobotRight               // >
	TileIDRobotDown                // v
	TileIDRobotDead                // X
)

type Tile struct {
	Point
	TileID TileID
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
	fmt.Println(minX, minY, maxX, maxY, width, height)
	rows := [][]string{}
	for i := int64(0); i < height; i++ {
		row := []string{}
		for j := int64(0); j < width; j++ {
			row = append(row, " ")
		}
		rows = append(rows, row)
	}
	for _, p := range ps {
		symbol := ""
		switch p.TileID {
		case TileIDEmpty:
			symbol = "."
		case TileIDScaffold:
			symbol = "#"
		case TileIDRobotUp:
			symbol = "^"
		case TileIDRobotDown:
			symbol = "v"
		case TileIDRobotLeft:
			symbol = "<"
		case TileIDRobotRight:
			symbol = ">"
		case TileIDRobotDead:
			symbol = "X"
		default:
			panic(p.TileID)
		}
		rows[p.Y-minY][p.X-minX] = symbol
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

func (d Direction) isOpositeOf(d2 Direction) bool {
	p := Point{0, 0}
	return d2.Apply(d.Apply(p)) == p
}
func (d Direction) isLeftOf(d2 Direction) bool {
	switch d {
	case DirectionNorth:
		return d2 == DirectionEast
	case DirectionEast:
		return d2 == DirectionSouth
	case DirectionSouth:
		return d2 == DirectionWest
	case DirectionWest:
		return d2 == DirectionNorth
	}
	panic("impossible")
}

type Instruction struct {
	isLeft   bool
	isRight  bool
	isNumber bool
	n        int64
}

func runProgram(cells []int64) error {
	g := NewGame()
	// TODO: feed ascii and newlines at the end
	// protocol:
	// A,B,C movement routines, separated by commas
	// 3 lines, entering the contents of each function
	// L,R,n for movement functions, separated by commas
	// y/n for continuous video feed
	// 20 chars max per line, not counting newline
	// objective: retrieve the single output at the end that shows the number of robots / amount of space dust
	vm := NewVM(cells, stdinInputter, g.AcceptDraw)
	for {
		err := vm.runToOutput()
		if err == errorHalt {
			break
		}
	}
	printDrawing(drawTiles(g.getTiles()))
	//fmt.Println(g.getIntersections())
	robotTiles := g.getTilesFor([]TileID{TileIDRobotDown, TileIDRobotLeft, TileIDRobotRight, TileIDRobotUp})
	robotPoint := robotTiles[0].Point
	robotDirection := DirectionEast
	path := []Instruction{Instruction{isRight: true}}
	for {
		// Never go back!
		// Check if we can go forward first
		nextPoint := robotDirection.Apply(robotPoint)
		if g.tiles[nextPoint] == TileIDScaffold {
			// add or increment a number instruction
			instruction := path[len(path)-1]
			if instruction.isNumber {
				instruction.n++
				path[len(path)-1] = instruction
			} else {
				path = append(path, Instruction{isNumber: true, n: 1})
			}
			robotPoint = robotDirection.Apply(robotPoint)
			continue
		}
		nextFound := false
		for _, d := range []Direction{DirectionWest, DirectionEast, DirectionNorth, DirectionSouth} {
			if d.isOpositeOf(robotDirection) || d == robotDirection {
				continue
			}
			nextPoint := d.Apply(robotPoint)
			if g.tiles[nextPoint] == TileIDScaffold {
				if d.isLeftOf(robotDirection) {
					path = append(path, Instruction{isLeft: true})
				} else {
					path = append(path, Instruction{isRight: true})
				}
				robotDirection = d
				nextFound = true
				break
			}
		}
		if !nextFound {
			break
		}
	}
	s := ""
	for _, p := range path {
		switch {
		case p.isLeft:
			s += "L,"
		case p.isRight:
			s += "R,"
		case p.isNumber:
			s += fmt.Sprintf("%d,", p.n)
		}
	}
	s = s[:len(s)-1]
	fmt.Println(s)

	// first, draw a path
	cells[0] = 2
	//318 squares must be covered...
	inputFeed := `A,A,B,C,B,C,B,C,B,A
R,6,L,12,R,6
L,12,R,6,L,8,L,12
R,12,L,10,L,10
y
	`
	lastChar := int64(0)
	vm = NewVM(cells, func() int64 {
		input := inputFeed[0]
		inputFeed = inputFeed[1:]
		fmt.Printf("%c", input)
		return int64(input)
	}, func(c int64) { fmt.Printf("%c", rune(c)); lastChar = c })
	for {
		err := vm.runToOutput()
		if err == errorHalt {
			break
		}
	}
	fmt.Println()
	fmt.Println(lastChar)
	return nil
}
