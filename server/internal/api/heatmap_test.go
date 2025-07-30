package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ChristianVilen/flight-heatmap/server/internal/repository"
)

type mockQueries struct{}

func (m *mockQueries) GetHeatmapDataDynamic(ctx context.Context, args repository.GetHeatmapDataDynamicParams) ([]repository.GetHeatmapDataDynamicRow, error) {
	return []repository.GetHeatmapDataDynamicRow{
		{
			LatBin: sql.NullFloat64{Float64: 60.25, Valid: true},
			LonBin: sql.NullFloat64{Float64: 24.75, Valid: true},
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

	var data []repository.GetHeatmapDataDynamicRow
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal("invalid JSON response")
	}

	if len(data) != 1 || data[0].Count != 12 {
		t.Errorf("unexpected data: %+v", data)
	}
}
