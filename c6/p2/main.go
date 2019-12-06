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
	pathYOU := pathToCOM(orbits, "YOU")
	pathSAN := pathToCOM(orbits, "SAN")
	distanceYOU, distanceSAN, _ := intersection(pathYOU, pathSAN)
	fmt.Println(distanceYOU + distanceSAN)
	return nil
}

func intersection(aPath, bPath []string) (int, int, string) {
	for i, a := range aPath {
		for j, b := range bPath {
			if a == b {
				return i, j, a
			}
		}
	}
	panic("not intersection")
}

func pathToCOM(orbits map[string]string, start string) []string {
	path := []string{}
	orbiter := start
	orbited := orbits[start]
	for orbited != "COM" {
		orbiter = orbited
		orbited = orbits[orbiter]
		path = append(path, orbiter)
	}
	return path
}
