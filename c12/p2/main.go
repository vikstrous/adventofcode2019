package main

import (
	"bufio"
	"fmt"
	"os"
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
	axes := []Axis{AxisZ}
	initial := copyMoons(moons)
	//tortoise := copyMoons(moons)
	//hare := copyMoons(moons)
	for i := 1; i <= 500000000; i++ {
		//for _, axis := range axes {
		//	applyStep(hare, axis)
		//}
		//for _, axis := range axes {
		//	applyStep(hare, axis)
		//}
		//for _, axis := range axes {
		//	applyStep(tortoise, axis)
		//}
		for _, axis := range axes {
			applyStep(moons, axis)
		}
		if eq(moons, initial) {
			fmt.Println("loop at", i)
			printMoons(moons, i)
			break
		}
		//printMoons(tortoise, i)
		//if eq(hare, tortoise) {
		//	fmt.Println("loop at", i)
		//	printMoons(hare, i)
		//	printMoons(tortoise, i)
		//	break
		//}
		if i%100000 == 0 {
			fmt.Printf(".")
		}
	}
	return nil
}

func copyMoons(moons []Moon) []Moon {
	newMoons := []Moon{}
	for _, moon := range moons {
		newMoons = append(newMoons, moon.copy())
	}
	return newMoons
}

func eq(moons1, moons2 []Moon) bool {
	for i := range moons1 {
		m1 := moons1[i]
		m2 := moons2[i]
		for axis := range m1.position {
			if *m1.position[axis] != *m2.position[axis] {
				return false
			}
			if *m1.velocity[axis] != *m2.velocity[axis] {
				return false
			}
		}
	}
	return true
}

func moonStr(moons []Moon) string {
	s := ""
	for _, moon := range moons {
		s += moon.String() + "\n"
	}
	return s
}

func totalEnergy(moons []Moon) int {
	totals := 0
	for _, moon := range moons {
		potential := 0
		for _, value := range moon.position {
			potential += abs(*value)
		}
		kinetic := 0
		for _, value := range moon.velocity {
			kinetic += abs(*value)
		}
		totals += potential * kinetic
	}
	return totals
}

func abs(i int) int {
	// 32 bit ints
	mask := i >> 31
	return (i + mask) ^ mask
}

func printMoons(moons []Moon, step int) {
	fmt.Printf("After %d steps:\n", step)
	for _, moon := range moons {
		fmt.Printf("%s\n", moon)
	}
	fmt.Println()
}

func strToAxis(s string) Axis {
	switch s {
	case "x":
		return AxisX
	case "y":
		return AxisY
	case "z":
		return AxisZ
	}
	panic(s)
}

type Axis int

const (
	AxisX Axis = iota
	AxisY
	AxisZ
)

func newMoon(line string) Moon {
	line = line[1 : len(line)-1]
	pairs := strings.Split(line, ", ")
	position := map[Axis]*int{}
	velocity := map[Axis]*int{}
	for _, p := range pairs {
		sides := strings.Split(p, "=")
		axis := sides[0]
		value, err := strconv.Atoi(sides[1])
		if err != nil {
			panic(err)
		}
		a := strToAxis(axis)
		position[a] = &value
		zero := 0
		velocity[a] = &zero
	}
	return Moon{position: position, velocity: velocity}
}

type Moon struct {
	// axis -> value
	position map[Axis]*int
	// axis -> value
	velocity map[Axis]*int
}

func (m Moon) copy() Moon {
	m2 := Moon{position: map[Axis]*int{}, velocity: map[Axis]*int{}}
	for k, v := range m.position {
		v2 := *v
		m2.position[k] = &v2
	}
	for k, v := range m.velocity {
		v2 := *v
		m2.velocity[k] = &v2
	}
	return m2
}

func sortedKeys(m map[Axis]*int) []Axis {
	return []Axis{AxisX, AxisY, AxisZ}
}

func (m Moon) String() string {
	s := "pos=<"
	positionParts := []string{}
	for _, axis := range sortedKeys(m.position) {
		positionParts = append(positionParts, fmt.Sprintf("%d=%3d", axis, *m.position[axis]))
	}
	velocityParts := []string{}
	for _, axis := range sortedKeys(m.velocity) {
		velocityParts = append(velocityParts, fmt.Sprintf("%d=%3d", axis, *m.velocity[axis]))
	}
	s += strings.Join(positionParts, ", ")
	s += "> vel=<"
	s += strings.Join(velocityParts, ", ")
	s += ">"
	return s
}

func applyStep(moons []Moon, axis Axis) {
	for m1I, m1 := range moons {
		// iterate through pairs without duplicates
		for _, m2 := range moons[:m1I] {
			diff := *m2.position[axis] - *m1.position[axis]
			if diff == 0 {
				continue
			}
			change := diff / abs(diff)
			*m2.velocity[axis] -= change
			*m1.velocity[axis] += change
		}
	}
	for _, m1 := range moons {
		*m1.position[axis] += *m1.velocity[axis]
	}
}
