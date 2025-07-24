package opensky

import (
	"database/sql"
	"math"
	"time"
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

// Earth radius in km
const earthRadiusKm = 6371.0

type BoundingBox struct {
	LatMin float64
	LonMin float64
	LatMax float64
	LonMax float64
}

// GetBoundingBox calculates a bounding box for a given lat/lon + radius in km
func GetBoundingBox(lat, lon, radiusKm float64) BoundingBox {
	// Diagonal distance from center to corner of square
	halfDiagonal := math.Sqrt2 * radiusKm

	latRad := lat * math.Pi / 180
	lonRad := lon * math.Pi / 180

	bearing1 := 225 * math.Pi / 180 // SW
	bearing2 := 45 * math.Pi / 180  // NE

	latMin := math.Asin(math.Sin(latRad)*math.Cos(halfDiagonal/earthRadiusKm) +
		math.Cos(latRad)*math.Sin(halfDiagonal/earthRadiusKm)*math.Cos(bearing1))
	lonMin := lonRad + math.Atan2(math.Sin(bearing1)*math.Sin(halfDiagonal/earthRadiusKm)*math.Cos(latRad),
		math.Cos(halfDiagonal/earthRadiusKm)-math.Sin(latRad)*math.Sin(latMin))

	latMax := math.Asin(math.Sin(latRad)*math.Cos(halfDiagonal/earthRadiusKm) +
		math.Cos(latRad)*math.Sin(halfDiagonal/earthRadiusKm)*math.Cos(bearing2))
	lonMax := lonRad + math.Atan2(math.Sin(bearing2)*math.Sin(halfDiagonal/earthRadiusKm)*math.Cos(latRad),
		math.Cos(halfDiagonal/earthRadiusKm)-math.Sin(latRad)*math.Sin(latMax))

	return BoundingBox{
		LatMin: latMin * 180 / math.Pi,
		LonMin: lonMin * 180 / math.Pi,
		LatMax: latMax * 180 / math.Pi,
		LonMax: lonMax * 180 / math.Pi,
	}
}

func IsNearEFHK(lat, lon float64, radiusKm float64) bool {
	// Haversine distance
	d := Haversine(lat, lon, 60.3172, 24.9633)
	return d <= radiusKm
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	lat1 = lat1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
