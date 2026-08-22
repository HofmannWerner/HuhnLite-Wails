package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("backend/db/postgres_wrapper.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, "?") || strings.Contains(line, "`") {
			fmt.Printf("Line %4d: %s\n", lineNum, strings.TrimSpace(line))
		}
	}
}
