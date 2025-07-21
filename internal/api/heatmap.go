package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ChristianVilen/flight-heatmap/internal/db"
)

type HeatmapQuerier interface {
	GetHeatmapData(context.Context) ([]db.GetHeatmapDataRow, error)
}

func HeatmapHandler(queries HeatmapQuerier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		heatmap, err := queries.GetHeatmapData(r.Context())
		if err != nil {
			http.Error(w, "error fetching heatmap", 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(heatmap)
	}
}
