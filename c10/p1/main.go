package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
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
	findBestAsteroid(convert(aMap))
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
		score := uniqueSlopes(candidate, asteroids)
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
	//fmt.Println(uniqueSlopes(Asteroid{5, 8}, asteroids, true))
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

func uniqueSlopes(from Asteroid, asteroids []Asteroid, shouldPrint ...bool) int {
	ratsPositive := []*big.Rat{}
	ratsNegative := []*big.Rat{}
	hasVerticalPositive := false
	hasVerticalNegative := false
	for _, asteroid := range asteroids {
		if asteroid == from {
			continue
		}
		if asteroid.X-from.X == 0 {
			if asteroid.Y-from.Y > 0 {
				hasVerticalPositive = true
			} else {
				hasVerticalNegative = true
			}
			continue
		}
		ratY := int64(asteroid.Y - from.Y)
		ratX := int64(asteroid.X - from.X)
		r := big.NewRat(ratY, ratX)
		rats := &ratsNegative
		if ratX > 0 {
			rats = &ratsPositive
		}
		if !inRats(*rats, r) {
			*rats = append(*rats, r)
		}
	}
	if len(shouldPrint) > 0 {
		fmt.Println(ratsPositive)
		fmt.Println(ratsNegative)
	}
	numRats := len(ratsPositive) + len(ratsNegative)
	if hasVerticalPositive {
		numRats++
	}
	if hasVerticalNegative {
		numRats++
	}
	return numRats
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
