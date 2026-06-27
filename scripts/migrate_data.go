package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		// Coba merakit dari env vars Laravel
		user := os.Getenv("DB_USERNAME")
		pass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_DATABASE")
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "host.docker.internal"
		}
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "3306"
		}

		if user != "" && dbName != "" {
			if pass != "" {
				mysqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s", user, pass, dbHost, dbPort, dbName)
			} else {
				mysqlDSN = fmt.Sprintf("%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s", user, dbHost, dbPort, dbName)
			}
		} else {
			log.Fatal("MYSQL_DSN (atau DB_USERNAME & DB_DATABASE dari file .env Laravel) is required in env")
		}
	}

	pgDSN := os.Getenv("DB_SOURCE") // Gunakan DB_SOURCE dari docker-compose
	if pgDSN == "" {
		pgDSN = os.Getenv("POSTGRES_DSN")
		if pgDSN == "" {
			log.Fatal("DB_SOURCE or POSTGRES_DSN is required")
		}
	}

	log.Println("Connecting to MySQL...")
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer mysqlDB.Close()

	log.Println("Connecting to PostgreSQL...")
	pgDB, err := sql.Open("pgx", pgDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pgDB.Close()

	// Truncate tables (CASCADE will handle foreign keys automatically)
	// We skip users and profiles as requested.
	log.Println("Truncating tables in Postgres...")
	truncateQuery := `TRUNCATE TABLE categories, tech_stacks, tags, messages, projects, work_experiences CASCADE;`
	_, err = pgDB.Exec(truncateQuery)
	if err != nil {
		log.Fatalf("Failed to truncate: %v", err)
	}

	// Map of MySQL tables to PostgreSQL tables
	tables := []struct {
		mysql string
		pg    string
	}{
		{"categories", "categories"},
		{"tech_stacks", "tech_stacks"},
		{"tags", "tags"},
		{"messages", "messages"},
		{"projects", "projects"},
		{"work_experiences", "work_experiences"},
		// Pivot tables with different names
		{"project_tech_stack", "project_tech_stacks"},
		{"project_tag", "project_tags"},
		{"tech_stack_work_experience", "work_experience_tech_stacks"},
		{"tag_work_experience", "work_experience_tags"},
	}

	for _, t := range tables {
		err := migrateTable(mysqlDB, pgDB, t.mysql, t.pg)
		if err != nil {
			log.Fatalf("Failed migrating %s -> %s: %v", t.mysql, t.pg, err)
		}
	}

	log.Println("✅ Data migration completed successfully!")
}

func migrateTable(mysqlDB, pgDB *sql.DB, mysqlTable, pgTable string) error {
	// Get PostgreSQL columns to avoid inserting columns that don't exist
	pgCols := make(map[string]bool)
	pgColsRows, err := pgDB.Query(fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name='%s'", pgTable))
	if err != nil {
		return fmt.Errorf("failed to get pg columns: %w", err)
	}
	for pgColsRows.Next() {
		var colName string
		if err := pgColsRows.Scan(&colName); err == nil {
			pgCols[colName] = true
		}
	}
	pgColsRows.Close()

	log.Printf("Migrating %s -> %s...", mysqlTable, pgTable)
	rows, err := mysqlDB.Query(fmt.Sprintf("SELECT * FROM %s", mysqlTable))
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		var validCols []string
		var validValues []interface{}
		for i, colName := range cols {
			if !pgCols[colName] {
				continue // Skip column if it doesn't exist in Postgres
			}
			validCols = append(validCols, colName)
			
			// Convert []byte to string for Postgres compatibility (driver quirks)
			val := columns[i]
			if b, ok := val.([]byte); ok {
				validValues = append(validValues, string(b))
			} else {
				validValues = append(validValues, val)
			}
		}

		// Generate $1, $2, $3 placeholders
		placeholders := make([]string, len(validCols))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", pgTable, strings.Join(validCols, ", "), strings.Join(placeholders, ", "))
		_, err := pgDB.Exec(query, validValues...)
		if err != nil {
			return fmt.Errorf("insert failed: %w (query: %s)", err, query)
		}
	}

	// Fix PostgreSQL sequences after explicit ID inserts
	var hasId int
	err = pgDB.QueryRow(fmt.Sprintf("SELECT 1 FROM information_schema.columns WHERE table_name='%s' AND column_name='id'", pgTable)).Scan(&hasId)
	if err == nil && hasId == 1 {
		seqQuery := fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 0) + 1, false)", pgTable, pgTable)
		_, err = pgDB.Exec(seqQuery)
		if err != nil {
			log.Printf("Warning: failed to reset sequence for %s: %v", pgTable, err)
		}
	}

	return nil
}
