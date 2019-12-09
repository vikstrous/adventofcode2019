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
	max0Layer := layers[0]
	for i, layer := range layers {
		if count(layer, 0) < count(max0Layer, 0) {
			fmt.Println(i)
			max0Layer = layer
		}
	}
	for _, layer := range layers {
		fmt.Println(count(layer, 0))
	}
	draw(max0Layer)
	ones := count(max0Layer, 1)
	twos := count(max0Layer, 2)
	fmt.Println(ones * twos)
	return nil
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
			fmt.Printf("%d", pixel)
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
