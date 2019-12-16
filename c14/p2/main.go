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
	reactions := []Reaction{}
	for scanner.Scan() {
		line := scanner.Text()
		reactions = append(reactions, parseReaction(line))
	}
	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	index := indexReactions(reactions)
	for _, r := range reactions {
		fmt.Println(r)
	}
	fmt.Println(maxFuel(1000000000000, index))
	return nil
}

func indexReactions(rs []Reaction) map[string]Reaction {
	index := map[string]Reaction{}
	for _, r := range rs {
		index[r.RHS.Name] = r
	}
	return index
}

func maxFuel(ore int, index map[string]Reaction) int {
	lower, upper := 0, ore
	for target := upper / 2; lower != target; {
		needed := costOf("FUEL", target, index)
		if needed > ore {
			upper = target
			target = lower + (target-lower)/2
		} else if needed < ore {
			lower = target
			target = target + (upper-target)/2
		} else {
			return target
		}
	}
	return lower
}

func costOf(name string, targetAmount int, index map[string]Reaction) int {
	remainders := map[string]int{}
	currentCost := make([]Ingredient, len(index[name].LHS))
	copy(currentCost, index[name].LHS)
	for i := range currentCost {
		currentCost[i].Quantity *= targetAmount
	}

	for len(currentCost) != 1 || currentCost[0].Name != "ORE" {
		newCost := []Ingredient{}
		found := false
		for _, i := range currentCost {
			if i.Name != "ORE" && !found {
				lhsIs := index[i.Name].LHS
				newIs := []Ingredient{}
				needed := i.Quantity - remainders[i.Name]
				if needed <= 0 {
					remainders[i.Name] -= i.Quantity
					found = true
					continue
				}
				reactionQuantity := index[i.Name].RHS.Quantity
				reactionApplications := (needed + reactionQuantity - 1) / reactionQuantity
				produced := reactionQuantity * reactionApplications
				remainders[i.Name] += produced - i.Quantity
				for _, lhsI := range lhsIs {
					newIs = append(newIs, Ingredient{
						Name:     lhsI.Name,
						Quantity: lhsI.Quantity * reactionApplications,
					})
				}
				newCost = append(newCost, newIs...)
				found = true
				continue
			}
			newCost = append(newCost, i)
		}
		currentCost = Combine(newCost)
	}
	return currentCost[0].Quantity
}
func Combine(is []Ingredient) []Ingredient {
	qs := map[string]int{}
	for _, i := range is {
		qs[i.Name] += i.Quantity
	}
	flat := []Ingredient{}
	for n, q := range qs {
		flat = append(flat, Ingredient{Name: n, Quantity: q})
	}
	sort.Slice(flat, func(i, j int) bool { return flat[i].Name < flat[j].Name })
	return flat
}

func parseReaction(line string) Reaction {
	sides := strings.Split(line, " => ")
	lhs := sides[0]
	rhs := sides[1]
	return Reaction{
		LHS: parseList(lhs),
		RHS: parseList(rhs)[0],
	}
}
func parseList(list string) []Ingredient {
	ingredientsString := strings.Split(list, ", ")
	ingredients := []Ingredient{}
	for _, ingredientString := range ingredientsString {
		parts := strings.Split(ingredientString, " ")
		q, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		ingredients = append(ingredients, Ingredient{
			Name:     parts[1],
			Quantity: q,
		})
	}
	return ingredients
}

type Reaction struct {
	LHS []Ingredient
	RHS Ingredient
}
type Ingredient struct {
	Name     string
	Quantity int
}
