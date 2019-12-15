package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	moons := []Moon{}
	for scanner.Scan() {
		line := scanner.Text()
		moons = append(moons, newMoon(line))
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	printMoons(moons, 0)
	for i := 1; i <= 1000; i++ {
		moons = applyGravity(moons)
		moons = applyVelocity(moons)
		printMoons(moons, i)
	}
	fmt.Println(totalEnergy(moons))
	return nil
}

func totalEnergy(moons []Moon) int {
	totals := 0
	for _, moon := range moons {
		potential := 0
		for _, value := range moon.position {
			potential += abs(value)
		}
		kinetic := 0
		for _, value := range moon.velocity {
			kinetic += abs(value)
		}
		totals += potential * kinetic
	}
	return totals
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func printMoons(moons []Moon, step int) {
	fmt.Printf("After %d steps:\n", step)
	for _, moon := range moons {
		fmt.Printf("%s\n", moon)
	}
	fmt.Println()
}

func newMoon(line string) Moon {
	line = line[1 : len(line)-1]
	pairs := strings.Split(line, ", ")
	position := map[string]int{}
	velocity := map[string]int{}
	for _, p := range pairs {
		sides := strings.Split(p, "=")
		axis := sides[0]
		value, err := strconv.Atoi(sides[1])
		if err != nil {
			panic(err)
		}
		position[axis] = value
		velocity[axis] = 0
	}
	return Moon{position: position, velocity: velocity}
}

type Moon struct {
	// axis -> value
	position map[string]int
	// axis -> value
	velocity map[string]int
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func (m Moon) String() string {
	s := "pos=<"
	positionParts := []string{}
	for _, axis := range sortedKeys(m.position) {
		positionParts = append(positionParts, fmt.Sprintf("%s=%3d", axis, m.position[axis]))
	}
	velocityParts := []string{}
	for _, axis := range sortedKeys(m.velocity) {
		velocityParts = append(velocityParts, fmt.Sprintf("%s=%3d", axis, m.velocity[axis]))
	}
	s += strings.Join(positionParts, ", ")
	s += "> vel=<"
	s += strings.Join(velocityParts, ", ")
	s += ">"
	return s
}

func applyGravity(original []Moon) []Moon {
	newMoons := make([]Moon, len(original))
	copy(newMoons, original)
	for m1I, m1 := range original {
		// iterate through pairs withuot duplicates
		for m2I, m2 := range original[:m1I] {
			for axis, m1Value := range m1.position {
				m2Value := m2.position[axis]
				if m2Value > m1Value {
					newMoons[m2I].velocity[axis] = newMoons[m2I].velocity[axis] - 1
					newMoons[m1I].velocity[axis] = newMoons[m1I].velocity[axis] + 1
				} else if m2Value < m1Value {
					newMoons[m2I].velocity[axis] = newMoons[m2I].velocity[axis] + 1
					newMoons[m1I].velocity[axis] = newMoons[m1I].velocity[axis] - 1
				}
			}
		}
	}
	return newMoons
}
func applyVelocity(original []Moon) []Moon {
	newMoons := make([]Moon, len(original))
	copy(newMoons, original)
	for m1I, m1 := range original {
		for axis, position := range m1.position {
			newMoons[m1I].position[axis] = position + newMoons[m1I].velocity[axis]
		}
	}
	return newMoons
}
