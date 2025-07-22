env "local" {
  url = "postgres://postgres:postgres@localhost:5433/opensky?sslmode=disable"

  migration {
    dir = "file://sql/migrations"
  }
}
