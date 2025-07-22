ALTER TABLE "aircraft_positions" ADD CONSTRAINT "aircraft_positions_icao24_time_position_key" UNIQUE ("icao24", "time_position");
