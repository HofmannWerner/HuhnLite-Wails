package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err != nil {
		log.Fatalf("Error opening MySQL: %v", err)
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

	for _, r := range reports {
		modified := false
		newSum := r.Sum

		if strings.Contains(newSum, ")\nUNION ALL") && !strings.Contains(newSum, ") AS sub_tbl\nUNION ALL") && !strings.Contains(newSum, ") sub_tbl\nUNION ALL") {
			newSum = strings.ReplaceAll(newSum, ")\nUNION ALL", ") AS sub_tbl\nUNION ALL")
			modified = true
		}
		if strings.Contains(newSum, ")\r\nUNION ALL") && !strings.Contains(newSum, ") AS sub_tbl\r\nUNION ALL") && !strings.Contains(newSum, ") sub_tbl\r\nUNION ALL") {
			newSum = strings.ReplaceAll(newSum, ")\r\nUNION ALL", ") AS sub_tbl\r\nUNION ALL")
			modified = true
		}
		
		trimmed := strings.TrimSpace(newSum)
		if strings.HasSuffix(trimmed, ")") && !strings.HasSuffix(trimmed, "sub_tbl") && !strings.HasSuffix(trimmed, "sub") {
			newSum = trimmed + " AS sub_tbl"
			modified = true
		}

		if modified {
			fmt.Printf("Updating MySQL Report %d (%s):\nOLD:\n%s\nNEW:\n%s\n\n", r.ID, r.Desc, r.Sum, newSum)
			_, err := db.Exec("UPDATE DYNAMISCHE_SQL SET SUMMENZEILE = ? WHERE ID = ?", newSum, r.ID)
			if err != nil {
				log.Printf("Failed to update report %d: %v", r.ID, err)
			}
		}
	}
}
