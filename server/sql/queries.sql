-- name: InsertPosition :exec
INSERT INTO aircraft_positions (
    icao24, callsign, origin_country, time_position,
    longitude, latitude, baro_altitude, on_ground,
    velocity, heading, vertical_rate
) VALUES (
    $1, $2, $3, to_timestamp($4),
    $5, $6, $7, $8,
    $9, $10, $11
);

-- name: GetHeatmapDataDynamic :many
SELECT
  (floor(latitude * sqlc.arg(bin_size)) / sqlc.arg(bin_size))::float8 AS lat_bin,
  (floor(longitude * sqlc.arg(bin_size)) / sqlc.arg(bin_size))::float8 AS lon_bin,
  COUNT(*) AS count
FROM aircraft_positions
WHERE 
  (@interval::text IS NULL OR time_position > now() - (@interval || ' minutes')::interval)
GROUP BY lat_bin, lon_bin;

