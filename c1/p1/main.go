package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	reader := bufio.NewReader(os.Stdin)

	modulesFuel := uint64(0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		line = strings.TrimSuffix(line, "\n")
		moduleMass, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", line, err)
		}
		moduleFuel := moduleMass/3 - 2
		modulesFuel += moduleFuel
	}
	fmt.Println(modulesFuel)
	return nil
}
