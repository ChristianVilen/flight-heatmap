package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"

	"github.com/ChristianVilen/flight-heatmap/internal/api"
	db "github.com/ChristianVilen/flight-heatmap/internal/db"
)

func connectToDB() *sql.DB {
	dbConn, err := sql.Open("postgres", "postgresql://postgres:postgres@localhost:5433/opensky?sslmode=disable")
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	if err := dbConn.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}

	return dbConn
}

func main() {
	dbConn := connectToDB()
	defer dbConn.Close()

	queries := db.New(dbConn)

	r := chi.NewRouter()
	r.Get("/api/heatmap", api.HeatmapHandler(queries))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
