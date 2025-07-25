// Package opensky interacts with OpenSky API to get flight data
package opensky

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ChristianVilen/flight-heatmap/server/internal/config"
	db "github.com/ChristianVilen/flight-heatmap/server/internal/db"
)

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

	nextAllowed time.Time
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

	if time.Now().Before(f.nextAllowed) {
		log.Println("Skipping fetch due to backoff.")
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			retrySeconds, _ := strconv.Atoi(resp.Header.Get("X-Rate-Limit-Retry-After-Seconds"))
			f.nextAllowed = time.Now().Add(time.Duration(retrySeconds) * time.Second)

			return nil
		}

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
	const maxDistanceFromEFHK = 50.0 // in km
	const maxAltitude = 10000.0      // meters
	duplicateErrors := 0

	for _, s := range states {
		if len(s) < 12 {
			continue
		}

		lat := toNullFloat64(s[6])
		lon := toNullFloat64(s[5])
		if !lat.Valid || !lon.Valid {
			continue
		}

		if toNullBool(s[8]).Bool { // On ground
			continue
		}

		// Optionally filter by distance from EFHK
		if !IsNearEFHK(lat.Float64, lon.Float64, maxDistanceFromEFHK) {
			continue
		}

		// Optional: skip high altitude cruising aircraft
		alt := toNullFloat64(s[7])
		if alt.Valid && alt.Float64 > maxAltitude {
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
				duplicateErrors++
			} else {
				log.Printf("insert failed: %v", err)
			}
		}
	}

	log.Print("Done inserting")
	log.Printf("duplicates %d", duplicateErrors)

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
