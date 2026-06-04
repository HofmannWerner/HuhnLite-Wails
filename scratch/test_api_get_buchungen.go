package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	appconfig "huhnlite-wails/backend/config"
	wailsdb "huhnlite-wails/backend/db"
	"huhnlite-wails/backend/api"
)

func testPath(dbPath string) {
	fmt.Printf("=== Testing /api/buchung for: %s ===\n", dbPath)
	gin.SetMode(gin.ReleaseMode)

	cfg := appconfig.Config{
		DBEngine:           "sqlite",
		DBConnectionString: dbPath,
	}

	database, err := wailsdb.Connect(cfg)
	if err != nil {
		fmt.Printf("DB connection failed: %v\n\n", err)
		return
	}
	defer database.SQL.Close()

	engine := api.StartServer(database)

	req, _ := http.NewRequest("GET", "/api/buchung", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("HTTP Status: %d\n", w.Code)
	body := w.Body.String()
	if w.Code != 200 {
		fmt.Printf("Error response: %s\n", body)
	} else {
		fmt.Printf("Successfully loaded bookings. Response length: %d\n", len(body))
	}
	fmt.Println()
}

func main() {
	testPath("C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	testPath("C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db")
}
