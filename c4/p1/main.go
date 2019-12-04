package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func isValid(current int) bool {
	currentStr := strconv.Itoa(current)
	c := currentStr[0]
	foundMatch := false
	for _, c2 := range currentStr[1:] {
		if byte(c2) < c {
			return false
		}
		if c == byte(c2) {
			foundMatch = true
		}
		c = byte(c2)
	}
	return foundMatch
}

func run() error {
	start := 271973
	end := 785961
	matches := 0
	for current := start; current <= end; current++ {
		if isValid(current) {
			matches++
		}
	}
	fmt.Println(matches)
	return nil
}
