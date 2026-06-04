package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	appconfig "huhnlite-wails/backend/config"
	wailsdb "huhnlite-wails/backend/db"
	"huhnlite-wails/backend/api"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Load config
	cfg := appconfig.Config{
		DBEngine:           "sqlite",
		DBConnectionString: "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db",
	}

	// Connect to DB
	database, err := wailsdb.Connect(cfg)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.SQL.Close()

	// Start server (setup routes)
	engine := api.StartServer(database)

	// Mock request
	req, _ := http.NewRequest("GET", "/api/futter", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	fmt.Printf("HTTP Status: %d\n", w.Code)
	body := w.Body.String()
	if len(body) > 200 {
		fmt.Printf("Body length: %d, start of body: %s\n", len(body), body[:200])
	} else {
		fmt.Printf("Body: %s\n", body)
	}
}
