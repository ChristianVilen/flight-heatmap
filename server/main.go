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

	centerLat := 60.3172
	centerLon := 24.9633
	radiusKm := 50.0

	bbox := opensky.GetBoundingBox(centerLat, centerLon, radiusKm)

	params := url.Values{}
	params.Set("lamin", fmt.Sprintf("%.4f", bbox.LatMin))
	params.Set("lamax", fmt.Sprintf("%.4f", bbox.LatMax))
	params.Set("lomin", fmt.Sprintf("%.4f", bbox.LonMin))
	params.Set("lomax", fmt.Sprintf("%.4f", bbox.LonMax))

	fetcher := opensky.Fetcher{
		Client:       http.DefaultClient,
		TokenFetcher: opensky.GetOpenSkyToken,
		Inserter:     db.New(dbConn),
		Config:       cfg,
		APIURL:       baseURL.String(),
	}

	PollInterval := 10 * time.Second

	go func() {
		ticker := time.NewTicker(PollInterval)
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
		Addr:    "127.0.0.1:8080",
		Handler: stack(router),
	}

	fmt.Println("Server listening on port :8080")
	server.ListenAndServe()
}
