package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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
	game := NewGame()
	row := 0
	for scanner.Scan() {
		line := scanner.Text()
		col := 0
		for _, r := range line {
			tileID := ParseTile(r)
			if tileID == TileIDWall {
				col++
				continue
			}
			p := Point{X: col, Y: row}
			game.tiles[p] = tileID
			if tileID == TileIDDoor {
				game.doors[r] = p
				game.doorsR[p] = r
			}
			if tileID == TileIDKey {
				game.keys[r] = p
				game.keysR[p] = r
			}
			if tileID == TileIDEntrance {
				game.entrance = p
			}
			col++
		}
		row++
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	fmt.Println(len(game.tiles))
	fmt.Println(len(game.keys))
	fmt.Println(len(game.keysR))
	fmt.Println(len(game.doors))
	fmt.Println(len(game.doorsR))
	fmt.Println(game.entrance)
	printDrawing(drawTiles(game.getTiles()))
	//fmt.Println(*game.solve(game.entrance, map[rune]struct{}{}, map[Point]struct{}{}, 0))
	fmt.Println(game.nextKeysOptions(game.entrance, map[rune]struct{}{}))
	fmt.Println(*game.distanceToHoldingAllKeys(game.entrance, map[rune]struct{}{}))
	return nil
}

type Game struct {
	tiles    map[Point]TileID
	keys     map[rune]Point
	keysR    map[Point]rune
	doors    map[rune]Point
	doorsR   map[Point]rune
	entrance Point
	mem      map[string]map[rune]int
	mem2     map[string]*int
}

func NewGame() *Game {
	return &Game{
		tiles:  map[Point]TileID{},
		keys:   map[rune]Point{},
		keysR:  map[Point]rune{},
		doors:  map[rune]Point{},
		doorsR: map[Point]rune{},
		mem:    map[string]map[rune]int{},
		mem2:   map[string]*int{},
	}
}

func toMemKey(p Point, keysHeld map[rune]struct{}) string {
	s := fmt.Sprint(p)
	s += string(sortedKeys(keysHeld))
	return s
}

func sortedKeys(keysHeld map[rune]struct{}) []rune {
	slice := []rune{}
	for k := range keysHeld {
		slice = append(slice, k)
	}
	sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
	return slice
}

// TODO: dfs on choice of next key; bfs on finding the key

// returns extra distance, not total distance
func (g *Game) distanceToHoldingAllKeys(from Point, keysHeld map[rune]struct{}) (ret *int) {
	memKey := toMemKey(from, keysHeld)
	if found, ok := g.mem2[memKey]; ok {
		return found
	}
	defer func() {
		g.mem2[memKey] = ret
	}()

	//// TODO: when do we backtrack or not? maybe if we already have a way to get x keys in less time, we don't try another way?
	//fmt.Println(len(keysHeld))
	//for _, k := range sortedKeys(keysHeld) {
	//	fmt.Printf("%c", k)
	//}
	//fmt.Println()
	if len(keysHeld) == len(g.keys) {
		zero := 0
		return &zero
	}

	var shortestDistance *int
	options := g.nextKeysOptions(from, keysHeld)
	for option, distanceToKey := range options {
		if _, ok := keysHeld[option]; ok {
			// we already have this key, no need to collect it again
			continue
		}
		newKeysHeld := copyKeys(keysHeld)
		newKeysHeld[option] = struct{}{}
		// if we picked this option next, would we be able to find a shorter way to get all keys?
		optionDistance := g.distanceToHoldingAllKeys(g.keys[option], newKeysHeld)
		if optionDistance != nil && (shortestDistance == nil || *shortestDistance > *optionDistance+distanceToKey) {
			cp := *optionDistance + distanceToKey
			shortestDistance = &cp
		}
	}
	return shortestDistance
}

func (g *Game) nextKeysOptions(from Point, keysHeld map[rune]struct{}) (ret map[rune]int) {
	memKey := toMemKey(from, keysHeld)
	if found, ok := g.mem[memKey]; ok {
		return found
	}
	defer func() {
		g.mem[memKey] = ret
	}()
	// what keys are reachable from here? and how far away are they?
	neighbours := g.neighbours(from)
	visited := map[Point]struct{}{}
	for _, n := range neighbours {
		visited[n.Point] = struct{}{}
	}
	minOptions := map[rune]int{}
	distance := 1
	for {
		newNeighbours := []Tile{}
	NLoop:
		for _, neighbour := range neighbours {
			switch neighbour.TileID {
			case TileIDKey:
				if _, ok := minOptions[neighbour.Letter]; !ok {
					minOptions[neighbour.Letter] = distance
				}
			case TileIDDoor:
				if _, ok := keysHeld[ToKey(neighbour.Letter)]; !ok {
					continue NLoop
				}
			}
			for _, n := range g.neighbours(neighbour.Point) {
				if _, ok := visited[n.Point]; !ok {
					newNeighbours = append(newNeighbours, n)
					visited[n.Point] = struct{}{}
				}
			}
		}
		if len(newNeighbours) == 0 {
			break
		}
		neighbours = newNeighbours
		distance++
	}
	return minOptions
}

func (g *Game) solve(from Point, keysHeld map[rune]struct{}, visited map[Point]struct{}, movesDone int) (ret *int) {
	if len(keysHeld) == len(g.keys) {
		return &movesDone
	}
	neighbours := g.neighbours(from)
	var minMoves *int
	for _, neighbour := range neighbours {
		if _, ok := visited[neighbour.Point]; ok {
			continue
		}
		visitedNext := copyVisited(visited)
		visitedNext[from] = struct{}{}
		keysHeldAfter := copyKeys(keysHeld)
		switch neighbour.TileID {
		case TileIDKey:
			if _, ok := keysHeld[neighbour.Letter]; !ok {
				keysHeldAfter[neighbour.Letter] = struct{}{}
				// reset the visited list now that we have mor keys
				visitedNext = map[Point]struct{}{}
			}
		case TileIDDoor:
			if _, ok := keysHeld[ToKey(neighbour.Letter)]; !ok {
				continue
			}
		}
		//fmt.Println("next", neighbour.Point, keysHeldAfter, movesDone+1)
		moves := g.solve(neighbour.Point, keysHeldAfter, visitedNext, movesDone+1)
		if moves == nil {
			continue
		}
		if minMoves == nil || *moves < *minMoves {
			movesCp := *moves
			minMoves = &movesCp
		}
	}
	return minMoves
}
func copyVisited(vs map[Point]struct{}) map[Point]struct{} {
	vs2 := map[Point]struct{}{}
	for v := range vs {
		vs2[v] = struct{}{}
	}
	return vs2
}
func copyKeys(keys map[rune]struct{}) map[rune]struct{} {
	k2 := map[rune]struct{}{}
	for k := range keys {
		k2[k] = struct{}{}
	}
	return k2
}

func printDrawing(grid [][]string) {
	for _, row := range grid {
		for _, cell := range row {
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

func (g *Game) getTiles() []Tile {
	tiles := []Tile{}
	for p := range g.tiles {
		t := g.tileAt(p)
		tiles = append(tiles, *t)
	}
	return tiles
}

func (g *Game) tileAt(p Point) *Tile {
	t, ok := g.tiles[p]
	if !ok {
		return nil
	}
	var letter rune
	if t == TileIDKey {
		letter = g.keysR[p]
	} else if t == TileIDDoor {
		letter = g.doorsR[p]
	}
	return &Tile{Point: p, TileID: t, Letter: letter}
}

var directions = []Direction{DirectionNorth, DirectionSouth, DirectionEast, DirectionWest}

func (g *Game) neighbours(p Point) []Tile {
	tiles := []Tile{}
	for _, d := range directions {
		p2 := d.Apply(p)
		t := g.tileAt(p2)
		if t != nil {
			tiles = append(tiles, *t)
		}
	}
	return tiles
}

type Tile struct {
	Point
	TileID TileID
	Letter rune
}

func drawTiles(ps []Tile) [][]string {
	minX := int(0)
	maxX := int(0)
	minY := int(0)
	maxY := int(0)
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
	fmt.Println("dimentions", minX, minY, maxX, maxY, width, height)
	rows := [][]string{}
	for i := 0; i < height; i++ {
		row := []string{}
		for j := 0; j < width; j++ {
			row = append(row, " ")
		}
		rows = append(rows, row)
	}
	for _, p := range ps {
		symbol := " "
		switch p.TileID {
		case TileIDEmpty:
			symbol = "."
		case TileIDEntrance:
			symbol = "@"
		case TileIDWall:
			symbol = "#"
		case TileIDKey:
			symbol = string(p.Letter)
		case TileIDDoor:
			symbol = string(p.Letter)
		default:
			panic(p.TileID)
		}
		rows[p.Y-minY][p.X-minX] = symbol
	}
	return rows
}

type Point struct {
	X int
	Y int
}
type Direction int

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
func (d Direction) Offsets() (int, int) {
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

type TileID int

const (
	TileIDWall TileID = iota
	TileIDEmpty
	TileIDEntrance
	TileIDKey
	TileIDDoor
)

func ToKey(r rune) rune {
	for _, k := range strings.ToLower(string(r)) {
		return k
	}
	panic("impossible")
}
func ToDoor(r rune) rune {
	for _, d := range strings.ToUpper(string(r)) {
		return d
	}
	panic("impossible")
}

func ParseTile(r rune) TileID {
	switch r {
	case '#':
		return TileIDWall
	case '.':
		return TileIDEmpty
	case '@':
		return TileIDEntrance
	default:
		if strings.ToUpper(string(r)) == string(r) {
			return TileIDDoor
		}
		return TileIDKey
	}
}
