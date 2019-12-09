package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	line := scanner.Text()
	height := 6
	width := 25
	layers := lineToLayers(line, width, height)
	result := flatten(layers)
	draw(result)
	return nil
}

func flatten(layers [][][]int) [][]int {
	resultLayer := layers[0]
	for _, layer := range layers[1:] {
		resultLayer = combine(resultLayer, layer)
	}
	return resultLayer
}

func combine(top, bottom [][]int) [][]int {
	result := [][]int{}
	for i := range top {
		resultRow := []int{}
		for j := range top[i] {
			if top[i][j] == 0 || top[i][j] == 1 {
				resultRow = append(resultRow, top[i][j])
			} else {
				resultRow = append(resultRow, bottom[i][j])
			}
		}
		result = append(result, resultRow)
	}
	return result
}

func lineToLayers(line string, width int, height int) [][][]int {
	layers := [][][]int{}
	layer := [][]int{}
	row := []int{}
	for i, c := range line {
		num, err := strconv.Atoi(string(c))
		if err != nil {
			panic(err)
		}
		row = append(row, num)
		if (i+1)%width == 0 {
			layer = append(layer, row)
			row = []int{}
		}
		if (i+1)%(width*height) == 0 {
			layers = append(layers, layer)
			layer = [][]int{}
		}
	}
	return layers
}

func draw(layer [][]int) {
	for _, row := range layer {
		for _, pixel := range row {
			if pixel == 1 {
				fmt.Printf("%d", pixel)
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Println()
	}
}

func count(layer [][]int, search int) int {
	c := 0
	for _, row := range layer {
		for _, pixel := range row {
			if pixel == search {
				c++
			}
		}
	}
	return c
}
