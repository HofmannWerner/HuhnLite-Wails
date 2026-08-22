package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("backend/api/server.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lower := strings.ToLower(line)
		if strings.Contains(lower, "years") || strings.Contains(lower, "eggstats") {
			fmt.Printf("Line %4d: %s\n", lineNum, strings.TrimSpace(line))
		}
	}
}
