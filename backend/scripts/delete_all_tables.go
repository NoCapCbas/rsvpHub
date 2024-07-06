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

	// Get list of tables
	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Error getting tables: %v\n", err)
	}

	// Drop each table
	for _, table := range tables {
		if err := dropTable(db, table); err != nil {
			log.Fatalf("Error dropping table %s: %v\n", table, err)
		}
	}

	fmt.Println("All tables dropped successfully.")
}

// getTables retrieves the list of all tables in the database.
func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// dropTable drops a single table.
func dropTable(db *sql.DB, tableName string) error {
	_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName))
	return err
}

