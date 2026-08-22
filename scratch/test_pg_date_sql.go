package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsnPG := "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable"
	dbPG, err := sql.Open("postgres", dsnPG)
	if err != nil {
		log.Fatalf("PG Open error: %v", err)
	}
	defer dbPG.Close()

	// Test 1: Get distinct years using SUBSTRING
	fmt.Println("=== Test 1: Distinct Years ===")
	rowsYears, err := dbPG.Query(`
		SELECT DISTINCT SUBSTRING(BUCHUNGSDATUM FROM 1 FOR 4) as year
		FROM BUCHUNG
		WHERE ($1 = -1 OR ID_HERDEN = $1)
		  AND ($2 = 0 OR EXISTS (SELECT 1 FROM HERDEN H WHERE H.ID = BUCHUNG.ID_HERDEN AND H.AKTIV = 1))
		ORDER BY year DESC
	`, -1, 0)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
	} else {
		for rowsYears.Next() {
			var y sql.NullString
			rowsYears.Scan(&y)
			fmt.Printf("  Year: %s\n", y.String)
		}
		rowsYears.Close()
	}

	// Test 2: Filtered eggstats
	fmt.Println("\n=== Test 2: Filtered Eggstats (year=2026, month=4) ===")
	var sumA, sumS, sumM, sumL, sumXL, sumV sql.NullInt64
	err = dbPG.QueryRow(`
		SELECT COALESCE(SUM(KLASSEA), 0)  AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)    AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)   AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)    AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)       AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE ($1 = -1 OR BUCHUNG.ID_HERDEN = $1)
		  AND ($2 = 0 OR EXISTS (SELECT 1 FROM HERDEN H WHERE H.ID = BUCHUNG.ID_HERDEN AND H.AKTIV = 1))
		  AND ($3 = '' OR SUBSTRING(BUCHUNGSDATUM FROM 1 FOR 4) = $3)
		  AND ($4 = 0 OR (CAST(SUBSTRING(BUCHUNGSDATUM FROM 6 FOR 2) AS INTEGER) + 2) / 3 = $4)
		  AND ($5 = 0 OR CAST(SUBSTRING(BUCHUNGSDATUM FROM 6 FOR 2) AS INTEGER) = $5)
		  AND VERMITTELT IN ('N', 'V')
	`, -1, 0, "2026", 0, 0).Scan(&sumA, &sumS, &sumM, &sumL, &sumXL, &sumV)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
	} else {
		fmt.Printf("  SUM_KLASSE_A: %d | SMALL: %d | MEDIUM: %d | LARGE: %d | XL: %d | VERLUSTE: %d\n",
			sumA.Int64, sumS.Int64, sumM.Int64, sumL.Int64, sumXL.Int64, sumV.Int64)
	}

	// Test 3: GetEggStatsByHerde
	fmt.Println("\n=== Test 3: GetEggStatsByHerde (ID_HERDEN=22) ===")
	err = dbPG.QueryRow(`
		SELECT COALESCE(SUM(KLASSEA), 0)                                                               AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)                                                                 AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)                                                                AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)                                                                 AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)                                                                    AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) +
		       COALESCE((SELECT SUM(T.BEWEGUNGEN)
		                 FROM TIERBEWEGUNGEN T
		                          JOIN HERDEN H ON T.HERDENNUMMER = H.HERDENNUMMER
		                 WHERE H.ID = $1), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE BUCHUNG.ID_HERDEN = $2
		  AND VERMITTELT IN ('N', 'V')
	`, 22, 22).Scan(&sumA, &sumS, &sumM, &sumL, &sumXL, &sumV)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
	} else {
		fmt.Printf("  SUM_KLASSE_A: %d | SMALL: %d | MEDIUM: %d | LARGE: %d | XL: %d | VERLUSTE: %d\n",
			sumA.Int64, sumS.Int64, sumM.Int64, sumL.Int64, sumXL.Int64, sumV.Int64)
	}

	// Test 4: GetEggStatsWeeklyByHerde
	fmt.Println("\n=== Test 4: GetEggStatsWeeklyByHerde (ID_HERDEN=22) ===")
	rowsWeekly, err := dbPG.Query(`
		SELECT LW                         AS LEBENSWOCHE,
		       MAX(BUCHUNGSDATUM)         AS LETZTES_DATUM,
		       COALESCE(SUM(KLASSEA), 0)  AS SUM_KLASSE_A,
		       COALESCE(SUM(SMALL), 0)    AS SUM_SMALL,
		       COALESCE(SUM(MEDIUM), 0)   AS SUM_MEDIUM,
		       COALESCE(SUM(LARGE), 0)    AS SUM_LARGE,
		       COALESCE(SUM(XL), 0)       AS SUM_XL,
		       COALESCE(SUM(VERLUSTE), 0) AS SUM_VERLUSTE
		FROM BUCHUNG
		WHERE ID_HERDEN = $1
		  AND VERMITTELT IN ('N', 'V')
		GROUP BY LW
		ORDER BY LEBENSWOCHE DESC
	`, 22)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
	} else {
		countW := 0
		for rowsWeekly.Next() {
			countW++
			var lw sql.NullInt64
			var datum sql.NullString
			rowsWeekly.Scan(&lw, &datum, &sumA, &sumS, &sumM, &sumL, &sumXL, &sumV)
			if countW <= 5 {
				fmt.Printf("  LW: %d | DATUM: %s | SUM_A: %d\n", lw.Int64, datum.String, sumA.Int64)
			}
		}
		rowsWeekly.Close()
		fmt.Printf("  Total weeks: %d\n", countW)
	}
}
