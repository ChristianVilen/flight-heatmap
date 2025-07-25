package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/ChristianVilen/flight-heatmap/server/internal/db"
)

type mockQueries struct{}

func (m *mockQueries) GetHeatmapDataDynamic(ctx context.Context, args db.GetHeatmapDataDynamicParams) ([]db.GetHeatmapDataDynamicRow, error) {
	return []db.GetHeatmapDataDynamicRow{
		{
			LatBin: float64(60.25),
			LonBin: float64(24.75),
			Count:  int64(12),
		},
	}, nil
}

func TestHeatmapHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/heatmap?bin=80&minutes=15", nil)
	w := httptest.NewRecorder()

	handler := HeatmapHandler(&mockQueries{})
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data []db.GetHeatmapDataDynamicRow
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal("invalid JSON response")
	}

	if len(data) != 1 || data[0].Count != 12 {
		t.Errorf("unexpected data: %+v", data)
	}
}
