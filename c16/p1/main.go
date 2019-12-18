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
		panic("no input")
	}
	line := scanner.Text()
	input := []int{}
	for _, c := range line {
		n, err := strconv.Atoi(string(c))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", line, err)
		}
		input = append(input, n)
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	fmt.Println(fftTimes(input, 100))
	return nil
}

func fftTimes(input []int, times int) []int {
	for i := 0; i < times; i++ {
		input = fft(input)
	}
	return input
}

func fft(input []int) []int {
	output := []int{}
	for offset := range input {
		pattern := genPattern(len(input), offset)
		output = append(output, applyFFTPattern(input, pattern))
	}
	return output
}

func applyFFTPattern(input, pattern []int) int {
	accum := 0
	for i := range input {
		accum += input[i] * pattern[i]
		//fmt.Printf("%d*%d + ", input[i], pattern[i])
	}
	accum = abs(accum) % 10
	//fmt.Println(accum)
	return accum
}

func genPattern(repeatTargetLen, inputOffset int) []int {
	basePattern := []int{}
	for _, n := range []int{0, 1, 0, -1} {
		for r := 0; r < inputOffset+1; r++ {
			basePattern = append(basePattern, n)
		}
	}
	repeatedPattern := []int{}
	for len(repeatedPattern) < repeatTargetLen+1 {
		repeatedPattern = append(repeatedPattern, basePattern...)
	}
	return repeatedPattern[1:]
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
