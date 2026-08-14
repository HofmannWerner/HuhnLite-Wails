package main

import (
	"bufio"
	"fmt"

	"os"
	"strings"
)

func main() {
	inputFile := "backend/db/queries.sql"
	outputFile := "backend/db/queries_postgres.sql"

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input: %v\n", err)
		return
	}
	defer file.Close()

	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output: %v\n", err)
		return
	}
	defer out.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(out)

	var currentQuery []string
	paramCount := 0

	flushQuery := func() {
		if len(currentQuery) == 0 {
			return
		}
		for _, line := range currentQuery {
			// Replace ? with $1, $2 ...
			var sb strings.Builder
			for i := 0; i < len(line); i++ {
				if line[i] == '?' {
					paramCount++
					sb.WriteString(fmt.Sprintf("$%d", paramCount))
				} else {
					sb.WriteByte(line[i])
				}
			}
			writer.WriteString(sb.String() + "\n")
		}
		currentQuery = nil
		paramCount = 0
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "-- name:") {
			flushQuery()
		}
		currentQuery = append(currentQuery, line)
	}
	flushQuery()

	writer.Flush()
	fmt.Println("Successfully converted queries.sql to queries_postgres.sql with $1, $2 parameters!")
}
