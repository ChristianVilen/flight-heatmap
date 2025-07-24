package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ChristianVilen/flight-heatmap/server/internal/api"
	"github.com/ChristianVilen/flight-heatmap/server/internal/config"
	db "github.com/ChristianVilen/flight-heatmap/server/internal/db"
	"github.com/ChristianVilen/flight-heatmap/server/internal/middleware"
	"github.com/ChristianVilen/flight-heatmap/server/internal/opensky"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
}

func connectToDBWithRetry(cfg config.Config, maxRetries int, delay time.Duration) *sql.DB {
	retries := 0
	for range time.Tick(delay) {
		if retries >= maxRetries {
			log.Fatalf("❌ Could not connect to DB after %d retries", maxRetries)
		}

		dbConn, err := sql.Open("postgres", cfg.DBURL)
		if err == nil && dbConn.Ping() == nil {
			log.Println("Connected to DB")
			return dbConn
		}

		log.Printf("⏳ DB not ready (attempt %d/%d), retrying in %v...", retries+1, maxRetries, delay)
		retries++
	}
	return nil
}

var APIURLStatesAll = "https://opensky-network.org/api/states/all"

func main() {
	cfg := config.Load()
	ctx := context.Background()

	dbConn := connectToDBWithRetry(cfg, 10, 2*time.Second)
	defer dbConn.Close()

	queries := db.New(dbConn)
	baseURL, err := url.Parse(APIURLStatesAll)
	if err != nil {
		log.Fatal("invalid base URL:", err)
	}

	paramsSouthFin := url.Values{}
	paramsSouthFin.Set("lamin", "59.0")
	paramsSouthFin.Set("lamax", "62.0")
	paramsSouthFin.Set("lomin", "23.0")
	paramsSouthFin.Set("lomax", "27.0")

	baseURL.RawQuery = paramsSouthFin.Encode()

	fetcher := opensky.Fetcher{
		Client:       http.DefaultClient,
		TokenFetcher: opensky.GetOpenSkyToken,
		Inserter:     db.New(dbConn),
		Config:       cfg,
		APIURL:       baseURL.String(),
	}

	// poll every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Polling OpenSky API...")
			if err := fetcher.FetchAndStore(ctx); err != nil {
				log.Printf("fetch error: %v", err)
			}
		}
	}()

	router := http.NewServeMux()

	router.HandleFunc("GET /api/heatmap", api.HeatmapHandler(queries))

	stack := middleware.CreateStack(
		middleware.Logging,
	)

	server := http.Server{
		Addr:    ":8080",
		Handler: stack(router),
	}

	fmt.Println("Server listening on port :8080")
	server.ListenAndServe()
}
