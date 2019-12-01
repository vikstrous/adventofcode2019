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
	modulesFuel := uint64(0)
	for scanner.Scan() {
		line := scanner.Text()
		moduleMass, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", line, err)
		}
		moduleFuel := fuelForMassLoop(moduleMass)
		modulesFuel += moduleFuel
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	fmt.Println(modulesFuel)
	return nil
}

func fuelForMassLoop(massInitial uint64) uint64 {
	fuelTotal := uint64(0)
	massRemaining := massInitial
	for massRemaining != 0 {
		fuelNeeded := fuelForMass(massRemaining)
		fuelTotal += fuelNeeded
		massRemaining = fuelNeeded
	}
	return fuelTotal
}

func fuelForMass(m uint64) uint64 {
	f := m/3 - 2
	// detect underflow
	if f > m {
		return 0
	}
	return f
}
