table "atlas_schema_revisions" {
  schema = schema.atlas_schema_revisions
  column "version" {
    null = false
    type = character_varying
  }
  column "description" {
    null = false
    type = character_varying
  }
  column "type" {
    null    = false
    type    = bigint
    default = 2
  }
  column "applied" {
    null    = false
    type    = bigint
    default = 0
  }
  column "total" {
    null    = false
    type    = bigint
    default = 0
  }
  column "executed_at" {
    null = false
    type = timestamptz
  }
  column "execution_time" {
    null = false
    type = bigint
  }
  column "error" {
    null = true
    type = text
  }
  column "error_stmt" {
    null = true
    type = text
  }
  column "hash" {
    null = false
    type = character_varying
  }
  column "partial_hashes" {
    null = true
    type = jsonb
  }
  column "operator_version" {
    null = false
    type = character_varying
  }
  primary_key {
    columns = [column.version]
  }
}
table "aircraft_positions" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "icao24" {
    null = true
    type = text
  }
  column "callsign" {
    null = true
    type = text
  }
  column "origin_country" {
    null = true
    type = text
  }
  column "time_position" {
    null = true
    type = timestamp
  }
  column "longitude" {
    null = true
    type = double_precision
  }
  column "latitude" {
    null = true
    type = double_precision
  }
  column "baro_altitude" {
    null = true
    type = double_precision
  }
  column "on_ground" {
    null = true
    type = boolean
  }
  column "velocity" {
    null = true
    type = double_precision
  }
  column "heading" {
    null = true
    type = double_precision
  }
  column "vertical_rate" {
    null = true
    type = double_precision
  }
  primary_key {
    columns = [column.id]
  }
  unique "aircraft_positions_icao24_time_position_key" {
    columns = [column.icao24, column.time_position]
  }
}
schema "atlas_schema_revisions" {
}
schema "public" {
  comment = "standard public schema"
}
