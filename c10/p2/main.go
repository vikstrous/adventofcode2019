package main

import (
	"bufio"
	"fmt"
	"math"
	"math/big"
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
	aMap := [][]bool{}
	for scanner.Scan() {
		line := scanner.Text()
		row := []bool{}
		for _, c := range line {
			row = append(row, c == '#')
		}
		aMap = append(aMap, row)
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	// station is at 20, 18
	station := Asteroid{X: 20, Y: 18}
	vapourize(asteroidsWithInfo(station, convert(aMap)), 200)
	return nil
}

func inRats(list []*big.Rat, search *big.Rat) bool {
	for _, cmp := range list {
		if cmp.Cmp(search) == 0 {
			return true
		}
	}
	return false
}

func findBestAsteroid(asteroids []Asteroid) {
	scored := []ScoredAsteroid{}
	for _, candidate := range asteroids {
		score := len(uniqueDirections(candidate, asteroids))
		scored = append(scored, ScoredAsteroid{asteroid: candidate, score: score})
	}
	best := scored[0]
	for _, candidate := range scored {
		if best.score < candidate.score {
			best = candidate
		}
	}
	printScoredMap(toMap(scored))
	//fmt.Println("unique slopes", uniqueSlopes(best, asteroids, true))
	fmt.Println("asteroids total", len(asteroids))
	fmt.Println(uniqueDirections(Asteroid{4, 3}, asteroids))
	fmt.Println(best)
}

type ScoredAsteroid struct {
	asteroid Asteroid
	score    int
}

func printScoredMap(scoredMap [][]int) {
	for _, row := range scoredMap {
		for _, score := range row {
			fmt.Printf(" %2d ", score)
		}
		fmt.Println()
	}
}

func toMap(scored []ScoredAsteroid) [][]int {
	maxX := 0
	maxY := 0
	for _, scoredAsteroid := range scored {
		if maxX < scoredAsteroid.asteroid.X {
			maxX = scoredAsteroid.asteroid.X
		}
		if maxY < scoredAsteroid.asteroid.Y {
			maxY = scoredAsteroid.asteroid.Y
		}
	}
	aMap := [][]int{}
	for i := 0; i < maxY+1; i++ {
		row := make([]int, maxX+1)
		aMap = append(aMap, row)
	}
	for _, a := range scored {
		aMap[a.asteroid.Y][a.asteroid.X] = a.score
	}
	return aMap
}

type Direction struct {
	isStraightUp   bool
	isStraightDown bool
	isLeft         bool
	isRight        bool
	direction      *big.Rat
}

func (d Direction) String() string {
	switch {
	case d.isStraightUp:
		return "<Up>"
	case d.isStraightDown:
		return "<Right>"
	case d.isLeft:
		return fmt.Sprintf("<Left %s>", d.direction)
	case d.isRight:
		return fmt.Sprintf("<Right %s>", d.direction)
	}
	panic("impossible")
}

func (d Direction) Equal(d2 Direction) bool {
	switch {
	case d.isStraightUp:
		if d2.isStraightUp {
			return true
		}
		return false
	case d.isStraightDown:
		if d2.isStraightDown {
			return true
		}
		return false
	case d.isLeft:
		if d2.isLeft {
			if d.direction.Cmp(d2.direction) == 0 {
				return true
			}
		}
		return false
	case d.isRight:
		if d2.isRight {
			if d.direction.Cmp(d2.direction) == 0 {
				return true
			}
		}
		return false
	}
	panic("impossible")
}

func inDirections(list []Direction, search Direction) bool {
	for _, cmp := range list {
		if cmp.Equal(search) {
			return true
		}
	}
	return false
}

type AsteroidWithInfo struct {
	direction Direction
	distance  float64
	asteroid  Asteroid
}

func printWithInfo(scoredMap [][]string) {
	for _, row := range scoredMap {
		for _, score := range row {
			if len(score) > 0 {
				fmt.Printf(" %s ", score)
			} else {
				fmt.Printf("            ")
			}
		}
		fmt.Println()
	}
}

func toMapWithInfo(scored []AsteroidWithInfo) [][]string {
	maxX := 0
	maxY := 0
	for _, scoredAsteroid := range scored {
		if maxX < scoredAsteroid.asteroid.X {
			maxX = scoredAsteroid.asteroid.X
		}
		if maxY < scoredAsteroid.asteroid.Y {
			maxY = scoredAsteroid.asteroid.Y
		}
	}
	aMap := [][]string{}
	for i := 0; i < maxY+1; i++ {
		row := make([]string, maxX+1)
		aMap = append(aMap, row)
	}
	for _, a := range scored {
		str := a.direction.String()
		if len(str) < 10 {
			str = strings.Repeat(" ", 10-len(str)) + str
		}
		aMap[a.asteroid.Y][a.asteroid.X] = str
	}
	return aMap
}
func toMapWithInfo2(scored []AsteroidWithInfo) [][]string {
	maxX := 0
	maxY := 0
	for _, scoredAsteroid := range scored {
		if maxX < scoredAsteroid.asteroid.X {
			maxX = scoredAsteroid.asteroid.X
		}
		if maxY < scoredAsteroid.asteroid.Y {
			maxY = scoredAsteroid.asteroid.Y
		}
	}
	aMap := [][]string{}
	for i := 0; i < maxY+1; i++ {
		row := make([]string, maxX+1)
		aMap = append(aMap, row)
	}
	for _, a := range scored {
		str := fmt.Sprintf("%010.3f", a.direction.toDegrees())
		if len(str) < 10 {
			str = strings.Repeat(" ", 10-len(str)) + str
		}
		aMap[a.asteroid.Y][a.asteroid.X] = str
	}
	return aMap
}

func asteroidsWithInfo(from Asteroid, asteroids []Asteroid) []AsteroidWithInfo {
	infos := []AsteroidWithInfo{}
	for _, asteroid := range asteroids {
		if asteroid == from {
			continue
		}
		dir := calcDirection(from, asteroid)
		infos = append(infos, AsteroidWithInfo{
			asteroid:  asteroid,
			direction: dir,
			distance:  math.Sqrt(float64(sq(asteroid.Y-from.Y)) + float64(sq(asteroid.X-from.X))),
		})
	}
	return infos
}

func sq(i int) int {
	return i * i
}

func vapourize(asteroids []AsteroidWithInfo, n int) {
	vapourizedSoFar := 0
	for len(asteroids) > 0 {
		sort.Slice(asteroids, func(i, j int) bool {
			if asteroids[j].direction.isAfter(asteroids[i].direction) {
				return true
			}
			if asteroids[i].direction.isAfter(asteroids[j].direction) {
				return false
			}
			if asteroids[i].distance < asteroids[j].distance {
				return true
			}
			if asteroids[i].distance > asteroids[j].distance {
				return false
			}
			panic(fmt.Sprintf("duplicate %v %v %f %f", asteroids[i], asteroids[j], asteroids[i].distance, asteroids[j].distance))
		})
		//fmt.Println("asteroids", len(asteroids))
		//fmt.Println("asteroid0", asteroids[0])
		//fmt.Println("asteroid1", asteroids[1])
		//printWithInfo(toMapWithInfo(asteroids))
		//printWithInfo(toMapWithInfo2(asteroids))

		currDir := Direction{isStraightUp: true}
		toDestroy := []AsteroidWithInfo{asteroids[0]}
		for _, asteroid := range asteroids[1:] {
			if asteroid.direction.isAfter(currDir) {
				toDestroy = append(toDestroy, asteroid)
				currDir = asteroid.direction
			}
		}
		fmt.Println("destroying", len(toDestroy))
		if vapourizedSoFar+len(toDestroy) >= n {
			fmt.Println("FOUND", toDestroy[n-vapourizedSoFar-1].asteroid)
			return
		}
		//printWithInfo(toMapWithInfo2(toDestroy))
		asteroids = without(asteroids, toDestroy)
		vapourizedSoFar += len(toDestroy)
	}
}

func without(originals []AsteroidWithInfo, removalSet []AsteroidWithInfo) []AsteroidWithInfo {
	left := []AsteroidWithInfo{}
	for _, original := range originals {
		found := false
		for _, removal := range removalSet {
			if removal.asteroid == original.asteroid {
				found = true
				break
			}
		}
		if !found {
			left = append(left, original)
		}
	}
	return left
}

func (d Direction) toDegrees() float64 {
	switch {
	case d.isStraightUp:
		return 0
	case d.isStraightDown:
		return 180
	case d.isRight:
		dir, _ := d.direction.Float64()
		return math.Atan(dir)*180/math.Pi + 90
	case d.isLeft:
		dir, _ := d.direction.Float64()
		return math.Atan(dir)*180/math.Pi + 270
	}
	panic("impossible")
}

func (d Direction) isAfter(d2 Direction) bool {
	return d.toDegrees() > d2.toDegrees()
	//	switch {
	//	case d2.isStraightUp:
	//		// everything other than stright up is after straight up
	//		if d.isStraightUp {
	//			return false
	//		}
	//		return true
	//	case d2.isRight:
	//		// straight up is not after right
	//		if d.isStraightUp {
	//			return false
	//		}
	//		if d.isRight {
	//			// if both are right compare slopes
	//			// slope is y/x, higher slope means more vertial (unless negative)
	//			// more vertical means before
	//			// we must compare slopes only if they have equal signs. If they have different signs, the positive one is before the negative one
	//
	//			// if d2 is negative, d can be after it only if it has a smaller (more negative) slope
	//			if d2.direction.Sign() < 0 {
	//				if d.direction.Cmp(d2.direction) < 0 {
	//					return true
	//				}
	//				return false
	//			}
	//			// if d2 is flat, d can be after it only if it has a negative slope
	//			if d2.direction.Sign() == 0 {
	//				if d.direction.Sign() == -1 {
	//					return true
	//				}
	//				return false
	//			}
	//
	//			// d is before d2 if d is positive and d2 is negative or flat
	//			if d.direction.Sign() > 0 && d2.direction.Sign() <= 0 {
	//				return false
	//			}
	//			// d is before or equal to d2 if d is flat and d2 is negative or flat
	//			if d.direction.Sign() == 0 && d2.direction.Sign() <= 0 {
	//				return false
	//			}
	//			d2.direction.Cmp(d2.direction)
	//		}
	//		// anything else is after right
	//		return true
	//	}
}

func calcDirection(from Asteroid, to Asteroid) Direction {
	dir := Direction{}
	if to.X-from.X == 0 {
		if to.Y-from.Y > 0 {
			dir = Direction{isStraightDown: true}
		} else {
			dir = Direction{isStraightUp: true}
		}
	} else {
		ratY := int64(to.Y - from.Y)
		ratX := int64(to.X - from.X)
		r := big.NewRat(ratY, ratX)
		isLeft := ratX < 0
		isRight := !isLeft
		dir = Direction{
			isLeft:    isLeft,
			isRight:   isRight,
			direction: r,
		}
	}
	return dir
}

func uniqueDirections(from Asteroid, asteroids []Asteroid) []Direction {
	directions := []Direction{}
	for _, asteroid := range asteroids {
		if asteroid == from {
			continue
		}
		dir := calcDirection(from, asteroid)
		if !inDirections(directions, dir) {
			directions = append(directions, dir)
		}
	}
	return directions
}

type Asteroid struct {
	X int
	Y int
}

func convert(aMap [][]bool) []Asteroid {
	asteroids := []Asteroid{}
	for y, row := range aMap {
		for x, isAstorid := range row {
			if isAstorid {
				asteroids = append(asteroids, Asteroid{Y: y, X: x})
			}
		}
	}
	return asteroids
}
