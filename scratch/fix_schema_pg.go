package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	content, err := os.ReadFile("backend/db/schema_postgres.sql")
	if err != nil {
		fmt.Printf("Error reading schema_postgres.sql: %v\n", err)
		return
	}

	re := regexp.MustCompile(`(?i)NUMERIC\s*\([^)]+\)`)
	updated := re.ReplaceAllString(string(content), "DOUBLE PRECISION")

	re2 := regexp.MustCompile(`(?i)NUMERIC`)
	updated = re2.ReplaceAllString(updated, "DOUBLE PRECISION")

	err = os.WriteFile("backend/db/schema_postgres.sql", []byte(updated), 0644)
	if err != nil {
		fmt.Printf("Error writing schema_postgres.sql: %v\n", err)
		return
	}

	fmt.Println("Successfully updated NUMERIC types to DOUBLE PRECISION in schema_postgres.sql!")
}
