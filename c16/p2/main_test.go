package main

import (
	"bufio"
	"os"
	"strconv"
	"testing"
)

func BenchmarkMain(b *testing.B) {
	f, err := os.Open("../input.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		panic("no input")
	}
	line := scanner.Text()
	input := []int8{}
	for _, c := range line {
		n, err := strconv.Atoi(string(c))
		if err != nil {
			panic(err)
		}
		input = append(input, int8(n))
	}
	err = scanner.Err()
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fftTimes(input, 1)
	}
}
