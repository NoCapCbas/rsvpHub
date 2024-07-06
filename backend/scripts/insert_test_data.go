package main

import (
  "log"
  "database/sql"

  _ "github.com/lib/pq"
)


func main() {

  connStr := ""

  // Open connection
  db, err := sql.Open("postgres", connStr)
  if err != nil {
    log.Fatalf("Error connecting to the database: %v\n", err)
  }
  defer db.Close()

  // users
  // insertTestUsers := `
  // INSERT INTO users (email, password)
  // VALUES ()
  // `
}
