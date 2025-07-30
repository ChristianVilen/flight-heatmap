// Package api responds to api requests
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ChristianVilen/flight-heatmap/server/internal/repository"
)

type HeatPoint struct {
	ID    int32   `json:"id"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Count int64   `json:"count"`
}

type HeatmapQuerier interface {
	GetHeatmapDataDynamic(ctx context.Context, args repository.GetHeatmapDataDynamicParams) ([]repository.GetHeatmapDataDynamicRow, error)
}

func HeatmapHandler(queries HeatmapQuerier) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		binSize := 80 // default bin granularity
		var interval sql.NullString

		if v := req.URL.Query().Get("bin"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				binSize = parsed
			}
		}

		// Only set interval if it's explicitly passed
		if v := req.URL.Query().Get("minutes"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil {
				interval = sql.NullString{String: fmt.Sprintf("%d", parsed), Valid: true}
			}
		} else {
			interval = sql.NullString{Valid: false} // explicitly invalid = no filtering
		}

		raw, err := queries.GetHeatmapDataDynamic(req.Context(), repository.GetHeatmapDataDynamicParams{
			BinSize:  sql.NullFloat64{Float64: float64(binSize), Valid: true},
			Interval: interval,
		})
		if err != nil {
			http.Error(res, "error fetching heatmap", http.StatusInternalServerError)
			return
		}

		points := make([]HeatPoint, 0, len(raw))
		for _, row := range raw {
			if row.LatBin.Valid && row.LonBin.Valid {
				points = append(points, HeatPoint{
					ID:    row.ID,
					Lat:   row.LatBin.Float64,
					Lon:   row.LonBin.Float64,
					Count: row.Count,
				})
			}
		}

		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(points)
	}
}
