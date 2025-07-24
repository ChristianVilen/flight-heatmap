CREATE TABLE aircraft_positions (
    id SERIAL PRIMARY KEY,
    icao24 TEXT,
    callsign TEXT,
    origin_country TEXT,
    time_position TIMESTAMP,
    longitude DOUBLE PRECISION,
    latitude DOUBLE PRECISION,
    baro_altitude DOUBLE PRECISION,
    on_ground BOOLEAN,
    velocity DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    vertical_rate DOUBLE PRECISION
);
