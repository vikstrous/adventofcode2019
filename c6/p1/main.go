package main

import (
	"bufio"
	"fmt"
	"os"
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
	// orbiter -> orbited
	orbits := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ")")
		orbits[parts[1]] = parts[0]
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	orbitsCount := 0
	for orbiter, orbited := range orbits {
		orbitsCount++
		for orbited != "COM" {
			orbiter = orbited
			orbited = orbits[orbiter]
			orbitsCount++
		}
	}
	fmt.Println(orbitsCount)
	return nil
}
