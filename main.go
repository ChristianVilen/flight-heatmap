package main

import (
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

	dbConn := connectToDB(cfg)
	defer dbConn.Close()

	queries := db.New(dbConn)

	r := chi.NewRouter()
	r.Get("/api/heatmap", api.HeatmapHandler(queries))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
