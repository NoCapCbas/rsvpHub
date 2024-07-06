package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connection string
  connStr := ""

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v\n", err)
	}
	defer db.Close()

  // users
  createUsersTable := `
  CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  )
  `
  _, err = db.Exec(createUsersTable)
  if err != nil {
    fmt.Errorf("error creating users table: %v", err)
  }
  createJWTClaimsTable := `
  CREATE TABLE IF NOT EXISTS jwt_claims (
    user_id SERIAL PRIMARY KEY,
    expires TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  )
  `
  _, err = db.Exec(createJWTClaimsTable)
  if err != nil {
    fmt.Errorf("error creating jwt claims table: %v", err)
  }

  // events
  createEventsTable := `
  CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_by_user TEXT NOT NULL, 
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  )
  `
  _, err = db.Exec(createEventsTable)
  if err != nil {
    fmt.Errorf("error creating events table: %v", err)
  }
  // rsvp
  createRsvpTable := `
  CREATE TABLE IF NOT EXISTS rsvps (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    will_attend BOOL NOT NULL,
    event_rsvped_to TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
  )
  `
  _, err = db.Exec(createRsvpTable)
  if err != nil {
    fmt.Errorf("error creating rsvps table: %v", err)
  }
  fmt.Println("All tables created succesfully.")
}

