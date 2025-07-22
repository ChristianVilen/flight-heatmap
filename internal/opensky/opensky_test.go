package opensky_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ChristianVilen/flight-heatmap/internal/config"
	"github.com/ChristianVilen/flight-heatmap/internal/db"
	"github.com/ChristianVilen/flight-heatmap/internal/opensky"
)

// mockDB implements db.Querier for testing
type mockDB struct {
	inserted []db.InsertPositionParams
}

func (m *mockDB) InsertPosition(ctx context.Context, params db.InsertPositionParams) error {
	m.inserted = append(m.inserted, params)
	return nil
}

func TestFetchAndInsert(t *testing.T) {
	// Sample API response mimicking OpenSky /states/all
	mockResponse := map[string]interface{}{
		"states": [][]interface{}{
			{
				"abc123",     // icao24
				"TEST123",    // callsign
				"Finland",    // origin_country
				1624281000.0, // time_position
				nil,          // last_contact
				24.75,        // longitude
				60.25,        // latitude
				3000.0,       // baro_altitude
				true,         // on_ground
				250.0,        // velocity
				180.0,        // heading
				5.0,          // vertical_rate
			},
		},
	}

	// Spin up fake API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer apiServer.Close()

	mock := &mockDB{}
	cfg := config.Config{ClientID: "test", ClientSecret: "test"}

	err := opensky.FetchAndStore(
		context.Background(),
		apiServer.Client(),
		mock,
		cfg,
		apiServer.URL,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.inserted) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(mock.inserted))
	}

	insert := mock.inserted[0]
	if !insert.Icao24.Valid || insert.Icao24.String != "abc123" {
		t.Errorf("expected Icao24=abc123, got: %v", insert.Icao24)
	}
	if !insert.OnGround.Valid || !insert.OnGround.Bool {
		t.Errorf("expected OnGround=true, got: %v", insert.OnGround)
	}
}

