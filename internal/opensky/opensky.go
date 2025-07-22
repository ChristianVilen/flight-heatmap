// Package opensky interacts with OpenSky API to get flight data
package opensky

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChristianVilen/flight-heatmap/internal/config"
	db "github.com/ChristianVilen/flight-heatmap/internal/db"
)

func toNullString(v any) sql.NullString {
	s, ok := v.(string)
	return sql.NullString{
		String: s,
		Valid:  ok && s != "",
	}
}

func toNullFloat64(v any) sql.NullFloat64 {
	f, ok := v.(float64)
	return sql.NullFloat64{
		Float64: f,
		Valid:   ok,
	}
}

func toNullBool(v any) sql.NullBool {
	b, ok := v.(bool)
	return sql.NullBool{
		Bool:  b,
		Valid: ok,
	}
}

func ToNullTime(v any) sql.NullTime {
	switch val := v.(type) {
	case float64:
		// OpenSky uses UNIX timestamp in seconds
		t := time.Unix(int64(val), 0)
		return sql.NullTime{Time: t, Valid: true}
	default:
		return sql.NullTime{Valid: false}
	}
}

func ToNullFloat64(v any) sql.NullFloat64 {
	switch val := v.(type) {
	case float64:
		return sql.NullFloat64{Float64: val, Valid: true}
	default:
		return sql.NullFloat64{Valid: false}
	}
}

// OpenSkyResponse maps the full OpenSky state vector API
type OpenSkyResponse struct {
	Time   int64   `json:"time"`
	States [][]any `json:"states"`
}

type positionInserter interface {
	InsertPosition(ctx context.Context, params db.InsertPositionParams) error
}

type Fetcher struct {
	Client       *http.Client
	TokenFetcher func(cfg config.Config) (string, error)
	Inserter     positionInserter
	Config       config.Config
	APIURL       string
}

// FetchAndStore polls OpenSky and writes aircraft data to DB
func (f *Fetcher) FetchAndStore(ctx context.Context) error {
	log.Println("Fetching OpenSky data…")
	token, err := f.TokenFetcher(f.Config)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	req, err := http.NewRequest("GET", f.APIURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result OpenSkyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("json decode failed: %w", err)
	}

	return f.storeStates(ctx, result.States)
}

func (f *Fetcher) storeStates(ctx context.Context, states [][]any) error {
	for _, s := range states {
		if len(s) < 12 || s[5] == nil || s[6] == nil {
			continue
		}

		params := db.InsertPositionParams{
			Icao24:        toNullString(s[0]),
			Callsign:      toNullString(s[1]),
			OriginCountry: toNullString(s[2]),
			ToTimestamp:   toNullFloat64(s[3]).Float64,
			Longitude:     toNullFloat64(s[5]),
			Latitude:      toNullFloat64(s[6]),
			BaroAltitude:  toNullFloat64(s[7]),
			OnGround:      toNullBool(s[8]),
			Velocity:      toNullFloat64(s[9]),
			Heading:       toNullFloat64(s[10]),
			VerticalRate:  toNullFloat64(s[11]),
		}

		err := f.Inserter.InsertPosition(ctx, params)
		if err != nil {
			if isDuplicateError(err) {
				log.Print("duplicate, skipped")
			} else {
				log.Printf("insert failed: %v", err)
			}
		}
	}

	log.Print("Done inserting")

	return nil
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

func GetOpenSkyToken(cfg config.Config) (string, error) {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", cfg.ClientID)
	form.Add("client_secret", cfg.ClientSecret)

	req, _ := http.NewRequest("POST", "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.AccessToken == "" {
		return "", fmt.Errorf("no token in response")
	}

	return body.AccessToken, nil
}
