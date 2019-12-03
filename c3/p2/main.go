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
	line1 := scanner.Text()
	if !scanner.Scan() {
		return fmt.Errorf("failed to read inupt")
	}
	line2 := scanner.Text()
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	moves1 := parseMoves(line1)
	moves2 := parseMoves(line2)
	lines1 := movesToLines(moves1)
	lines2 := movesToLines(moves2)
	is := allIntersections(lines1, lines2)
	fmt.Println(closestIntersection(is).DistanceFromStart)
	return nil
}

func closestIntersection(is []IntersectionPoint) IntersectionPoint {
	closest := is[0]
	for _, i := range is {
		if i.DistanceFromStart < closest.DistanceFromStart {
			closest = i
		}
	}
	return closest
}

func allIntersections(lines1, lines2 []Line) []IntersectionPoint {
	var intersections []IntersectionPoint
	for _, line1 := range lines1 {
		for _, line2 := range lines2 {
			intr, ok := intersection(line1, line2)
			if ok {
				intersections = append(intersections, *intr)
			}
		}
	}
	return intersections
}

type Point struct {
	X int
	Y int
}

type IntersectionPoint struct {
	X                 int
	Y                 int
	DistanceFromStart int
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func intersection(line1, line2 Line) (*IntersectionPoint, bool) {
	if line1.StartX == 0 && line1.StartY == 0 && line2.StartX == 0 && line2.StartY == 0 {
		return nil, false
	}
	if line1.Horizontal == line2.Horizontal {
		if line1.Horizontal && line1.StartY != line2.StartY {
			return nil, false
		}
		if !line1.Horizontal && line1.StartX != line2.StartX {
			return nil, false
		}
		panic(fmt.Sprintf("not implemented %#v %#v", line1, line2))
	}
	horizontal := line1
	vertical := line2
	if !line1.Horizontal {
		horizontal = line2
		vertical = line1
	}
	intersectionPoint := Point{Y: horizontal.StartY, X: vertical.StartX}
	if pointIntersection(horizontal, intersectionPoint) && pointIntersection(vertical, intersectionPoint) {
		distanceToLastLinePoints := line1.DistanceFromStart + line2.DistanceFromStart
		// only one of these is non-0
		distanceToIntersectionX := abs(abs(line1.StartX)-abs(intersectionPoint.X)) + abs(abs(line2.StartX)-abs(intersectionPoint.X))
		distanceToIntersectionY := abs(abs(line1.StartY)-abs(intersectionPoint.Y)) + abs(abs(line2.StartY)-abs(intersectionPoint.Y))
		return &IntersectionPoint{
			X:                 intersectionPoint.X,
			Y:                 intersectionPoint.Y,
			DistanceFromStart: distanceToLastLinePoints + distanceToIntersectionX + distanceToIntersectionY,
		}, true
	}
	return nil, false
}

func pointIntersection(line Line, point Point) bool {
	if !line.Horizontal {
		if line.StartX != point.X {
			return false
		}
		bottomY := line.StartY
		if line.Length < 0 {
			bottomY = line.StartY + line.Length
		}
		topY := line.StartY
		if line.Length > 0 {
			topY = line.StartY + line.Length
		}
		if bottomY > point.Y {
			return false
		}
		if topY < point.Y {
			return false
		}
		return true
	}
	if line.StartY != point.Y {
		return false
	}
	leftX := line.StartX
	if line.Length < 0 {
		leftX = line.StartX + line.Length
	}
	rightX := line.StartX
	if line.Length > 0 {
		rightX = line.StartX + line.Length
	}
	if leftX > point.X {
		return false
	}
	if rightX < point.X {
		return false
	}
	return true
}

func movesToLines(moves []Move) []Line {
	currentX, currentY := 0, 0
	distanceFromStart := 0
	lines := []Line{}
	for _, move := range moves {
		line := Line{
			StartX:            currentX,
			StartY:            currentY,
			DistanceFromStart: distanceFromStart,
		}
		endX := currentX
		endY := currentY
		if move.Horizontal {
			line.Horizontal = true
			endX = line.StartX + move.Value
		} else {
			endY = line.StartY + move.Value
		}
		line.Length = move.Value
		distanceFromStart += abs(line.Length)
		currentX, currentY = endX, endY
		lines = append(lines, line)
	}
	return lines
}

type Line struct {
	StartX            int
	StartY            int
	Horizontal        bool
	Length            int
	DistanceFromStart int
}

type Move struct {
	Horizontal bool
	// vaule is positive up and right
	Value int
}

func parseMoves(line string) []Move {
	movesStr := strings.Split(line, ",")
	moves := []Move{}
	for _, moveStr := range movesStr {
		move := Move{}
		if moveStr[0] == 'R' || moveStr[0] == 'L' {
			move.Horizontal = true
		}
		var err error
		move.Value, err = strconv.Atoi(moveStr[1:])
		if err != nil {
			panic(err)
		}
		if moveStr[0] == 'D' || moveStr[0] == 'L' {
			move.Value = -move.Value
		}
		moves = append(moves, move)
	}
	return moves
}
