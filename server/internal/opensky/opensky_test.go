package opensky_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ChristianVilen/flight-heatmap/server/internal/config"
	"github.com/ChristianVilen/flight-heatmap/server/internal/opensky"
	"github.com/ChristianVilen/flight-heatmap/server/internal/repository"
)

// mockDB implements db.Querier for testing
type mockDB struct {
	inserted []repository.InsertPositionParams
}

func (m *mockDB) InsertPosition(ctx context.Context, params repository.InsertPositionParams) error {
	m.inserted = append(m.inserted, params)
	return nil
}

func TestFetchAndInsert(t *testing.T) {
	// Sample API response mimicking OpenSky /states/all
	mockResponse := map[string]any{
		"states": [][]any{
			{
				"abc123",     // icao24
				"TEST123",    // callsign
				"Finland",    // origin_country
				1624281000.0, // time_position
				nil,          // last_contact
				24.75,        // longitude
				60.25,        // latitude
				3000.0,       // baro_altitude
				false,        // on_ground
				250.0,        // velocity
				180.0,        // heading
				5.0,          // vertical_rate
			},
		},
	}

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))

	defer apiServer.Close()

	mock := &mockDB{}
	cfg := config.Config{ClientID: "test", ClientSecret: "test"}

	f := opensky.Fetcher{
		Client:       apiServer.Client(),
		TokenFetcher: func(cfg config.Config) (string, error) { return "mock-token", nil },
		Inserter:     mock,
		Config:       cfg,
		APIURL:       apiServer.URL,
	}

	err := f.FetchAndStore(context.Background())
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

	if !insert.OnGround.Valid || insert.OnGround.Bool {
		t.Errorf("expected OnGround=false, got: %v", insert.OnGround)
	}
}
