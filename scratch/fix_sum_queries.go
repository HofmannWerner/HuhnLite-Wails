package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("Error opening SQLite: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, BESCHREIBUNG, SUMMENZEILE FROM DYNAMISCHE_SQL")
	if err != nil {
		log.Fatalf("Error querying DYNAMISCHE_SQL: %v", err)
	}
	defer rows.Close()

	type Report struct {
		ID   int
		Desc string
		Sum  string
	}
	var reports []Report

	for rows.Next() {
		var r Report
		if err := rows.Scan(&r.ID, &r.Desc, &r.Sum); err == nil && r.Sum != "" && r.Sum != "x" {
			reports = append(reports, r)
		}
	}

	// We want to add " AS sub_tbl" to FROM (...) subqueries that lack aliases.
	// In the generated SQL:
	//   FROM (SELECT ...)
	//   UNION ALL
	//   FROM (SELECT ...)
	//
	// A simple but effective way is to find the closing paren that corresponds to the subquery.
	// Since we know the structure is:
	//   FROM (SELECT ... ) [UNION ALL | end of string]
	// We can replace the specific pattern.
	// Let's use a regex or string replacement.
	
	for _, r := range reports {
		modified := false
		newSum := r.Sum

		// Let's replace the first FROM subquery's end
		// We search for `)\nUNION ALL` and replace with `) AS sub_tbl\nUNION ALL`
		if strings.Contains(newSum, ")\nUNION ALL") && !strings.Contains(newSum, ") AS sub_tbl\nUNION ALL") && !strings.Contains(newSum, ") sub_tbl\nUNION ALL") {
			newSum = strings.ReplaceAll(newSum, ")\nUNION ALL", ") AS sub_tbl\nUNION ALL")
			modified = true
		}
		if strings.Contains(newSum, ")\r\nUNION ALL") && !strings.Contains(newSum, ") AS sub_tbl\r\nUNION ALL") && !strings.Contains(newSum, ") sub_tbl\r\nUNION ALL") {
			newSum = strings.ReplaceAll(newSum, ")\r\nUNION ALL", ") AS sub_tbl\r\nUNION ALL")
			modified = true
		}
		
		// Also handle the very end of the statement (the second subquery)
		// It ends with `)` or `)\n` or `)\r\n`
		trimmed := strings.TrimSpace(newSum)
		if strings.HasSuffix(trimmed, ")") && !strings.HasSuffix(trimmed, "sub_tbl") && !strings.HasSuffix(trimmed, "sub") {
			// Find the last index of ')' and insert ' AS sub_tbl' before it? No, it ends with ')' which is the subquery's closing paren.
			// So we append ' AS sub_tbl' at the end of the subquery.
			newSum = trimmed + " AS sub_tbl"
			modified = true
		}

		if modified {
			fmt.Printf("Updating Report %d (%s):\nOLD:\n%s\nNEW:\n%s\n\n", r.ID, r.Desc, r.Sum, newSum)
			_, err := db.Exec("UPDATE DYNAMISCHE_SQL SET SUMMENZEILE = ? WHERE ID = ?", newSum, r.ID)
			if err != nil {
				log.Printf("Failed to update report %d: %v", r.ID, err)
			}
		}
	}
}
