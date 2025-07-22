package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ChristianVilen/flight-heatmap/internal/api"
	"github.com/ChristianVilen/flight-heatmap/internal/config"
	db "github.com/ChristianVilen/flight-heatmap/internal/db"
	"github.com/ChristianVilen/flight-heatmap/internal/middleware"
	"github.com/ChristianVilen/flight-heatmap/internal/opensky"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
}

func connectToDB(cfg config.Config) *sql.DB {
	dbConn, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}

	return dbConn
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	dbConn := connectToDB(cfg)
	defer dbConn.Close()

	queries := db.New(dbConn)

	fetcher := opensky.Fetcher{
		Client:       http.DefaultClient,
		TokenFetcher: opensky.GetOpenSkyToken,
		Inserter:     db.New(dbConn),
		Config:       cfg,
		APIURL:       "https://opensky-network.org/api/states/all?lamin=59.0&lamax=62.0&lomin=23.0&lomax=27.0",
	}

	if err := fetcher.FetchAndStore(ctx); err != nil {
		log.Fatal(err.Error())
	}

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
