package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ChristianVilen/flight-heatmap/internal/api"
	"github.com/ChristianVilen/flight-heatmap/internal/config"
	db "github.com/ChristianVilen/flight-heatmap/internal/db"
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

	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/api/heatmap", api.HeatmapHandler(queries))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
