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
	input := []int8{}
	for _, c := range line {
		n, err := strconv.Atoi(string(c))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", line, err)
		}
		input = append(input, int8(n))
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	inputExpanded := make([]int8, 0, len(input)*10000)
	for i := 0; i < 10000; i++ {
		inputExpanded = append(inputExpanded, input...)
	}
	queryLocation := 0
	for _, num := range input[:7] {
		queryLocation *= 10
		queryLocation += int(num)
	}
	fmt.Println("query", queryLocation)
	fmt.Println(len(input))
	fmt.Println(len(inputExpanded))
	//output := fftAtLevelAndPosition(inputExpanded, 1, queryLocation)
	inputExpanded = inputExpanded[queryLocation:]
	fmt.Println(len(inputExpanded))
	output := fftTimes(inputExpanded, 100)
	fmt.Printf("%v", output[:8])
	//fmt.Printf("%v", output[queryLocation:queryLocation+7])
	return nil
}

//func fftAtLevelAndPosition(input []int, level, position int) int {
//	if level == 0 {
//		return outputAt(input, position)
//	}
//
//	accum := 0
//	// TODO: loop only through +1 entries and -1 entries
//	for i := range input {
//		accum += fftAtLevelAndPosition(input, level-1, i) * patternAt(position, i)
//	}
//	accum = abs(accum) % 10
//	return accum
//}

func fftTimes(input []int8, times int) []int8 {
	for i := 0; i < times; i++ {
		input = fft(input)
	}
	return input
}

func fft(input []int8) []int8 {
	output := make([]int8, len(input))
	for i := len(input) - 1; i >= 0; i-- {
		prev := int8(0)
		if i+1 < len(input) {
			prev = output[i+1]
		}
		output[i] = (input[i] + prev) % 10
	}
	return output
}

//func outputAt(input []int8, skipOffset, offset int) int8 {
//	accum := int(0)
//	for i := range input {
//		accum += int(input[i]) * int(patternAt(skipOffset+offset, skipOffset+i))
//		//fmt.Printf("%d*%d + ", input[i], pattern[i])
//	}
//	accum = abs(accum) % 10
//	//fmt.Println(accum)
//	return int8(accum)
//}

var pattern = []int8{0, 1, 0, -1}

//func patternAt(inputOffset, location int) int8 {
//	// a abbccddaabbccdd...
//	//          ^
//	// aa is a segment
//	// aabbccdd is a pattern
//	//segmentLength := inputOffset + 1
//	//patternLength := segmentLength * 4
//	//// compensate for the initial offset
//	//location = location + 1
//	//offsetInPattern := location % patternLength
//	//indexInPattern := offsetInPattern / segmentLength
//	//return pattern[indexInPattern]
//	//off := ((location + 1) % ((inputOffset + 1) * 4)) / (inputOffset + 1)
//	//return pattern[off]
//	// TODO: not sure how the above becomes this, but it's cool
//	return pattern[(location+1)/(inputOffset+1)%4]
//}

func abs(i int8) int8 {
	if i < 0 {
		return -i
	}
	return i
}
