// Package api's marker details endpoint returns data available of a flight marker
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ChristianVilen/flight-heatmap/server/internal/repository"
)

type MarkerDetailsQuerier interface {
	GetAircraftData(ctx context.Context, args int32) (repository.AircraftPosition, error)
}

func MarkerDetailsHandler(queries MarkerDetailsQuerier) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var id int
		if v := req.URL.Query().Get("id"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				id = parsed
			}
		}

		data, err := queries.GetAircraftData(req.Context(), int32(id))
		if err != nil {
			http.Error(res, "error fetching heatmap", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(data)
	}
}
