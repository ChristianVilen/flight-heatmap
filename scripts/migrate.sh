#!/bin/bash

cd server

DB_HOST=localhost
DB_PORT=5433
DB_UP=false
STARTED_DB=false

# Check if DB is up
echo "Checking if database is up on $DB_HOST:$DB_PORT..."

if nc -z "$DB_HOST" "$DB_PORT"; then
  echo "Database is already running."
  DB_UP=true
else
  echo "Database is not running. Starting it with Docker Compose..."
  docker compose up -d db
  STARTED_DB=true
fi

# Wait for DB to be fully ready (responding to SQL)
echo "Waiting for DB to become ready..."
for i in {1..15}; do
  if docker exec flight_heatmap_db pg_isready -U postgres > /dev/null 2>&1; then
    echo "PostgreSQL is ready!"
    break
  fi
  echo "Still initializing... ($i)"
  sleep 2
done

# Run Atlas migrations
echo "Running migrations..."
cd ../server
atlas migrate apply --env local

# Optionally stop DB if we started it
if [ "$STARTED_DB" = true ]; then
  echo "Stopping the database..."
  docker compose stop db
fi
