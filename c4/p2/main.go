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
func nextCOrZero(i int, currentStr string) byte {
	if i == (len(currentStr) - 2) {
		return 0
	}
	return currentStr[i+2]
}

func isValid(current int) bool {
	currentStr := strconv.Itoa(current)
	c := currentStr[0]
	foundMatchLen := 1
	foundMatch := false
	for i, c2 := range currentStr[1:] {
		if byte(c2) < c {
			return false
		}
		if c == byte(c2) {
			foundMatchLen++
		} else {
			foundMatchLen = 1
		}
		if foundMatchLen == 2 && (nextCOrZero(i, currentStr) != byte(c2)) {
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
