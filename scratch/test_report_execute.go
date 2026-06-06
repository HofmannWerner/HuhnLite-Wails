package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"huhnlite-wails/backend/api"
	appconfig "huhnlite-wails/backend/config"
	wailsdb "huhnlite-wails/backend/db"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Connect to the user's MariaDB/MySQL database
	cfg := appconfig.Config{
		DBEngine:           "mysql",
		DBConnectionString: "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true",
	}

	database, err := wailsdb.Connect(cfg)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.SQL.Close()

	engine := api.StartServer(database)

	// Load all report IDs
	rows, err := database.SQL.Query("SELECT ID, BESCHREIBUNG FROM DYNAMISCHE_SQL WHERE TYP_KZ != 'H'")
	if err != nil {
		log.Fatalf("Failed to query report IDs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var desc string
		if err := rows.Scan(&id, &desc); err == nil {
			fmt.Printf("\n========================================\n")
			fmt.Printf("REPORT %d: %s\n", id, desc)
			testReport(engine, id, "de")
			testReport(engine, id, "it")
			testReport(engine, id, "en")
		}
	}
}

func testReport(engine *gin.Engine, id int, lang string) {
	fmt.Printf("--- LANG: %s ---\n", lang)
	
	// Prepare empty JSON body or default parameters if needed
	body := []byte(`{"PARAMS": {
		"Mindestmenge": 0,
		"Ab_Datum": "2000-01-01",
		"VON_BUCHUNGSDATUM": "2000-01-01",
		"BIS_BUCHUNGSDATUM": "2030-01-01",
		"AKTIV": true
	}}`)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/reports/execute/%d?lang=%s", id, lang), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("  HTTP Status: %d\n", w.Code)
	respBody := w.Body.String()
	if w.Code != 200 {
		fmt.Printf("  Error response: %s\n", respBody)
	} else {
		// Print length and checks
		if len(respBody) > 300 {
			fmt.Printf("  Response length: %d | Starts with: %s\n", len(respBody), respBody[:150])
		} else {
			fmt.Printf("  Response: %s\n", respBody)
		}
	}
}
