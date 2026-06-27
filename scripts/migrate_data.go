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
	// Get PostgreSQL columns and their types to handle data type conversions
	pgCols := make(map[string]string)
	pgColsRows, err := pgDB.Query(fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s'", pgTable))
	if err != nil {
		return fmt.Errorf("failed to get pg columns: %w", err)
	}
	for pgColsRows.Next() {
		var colName, dataType string
		if err := pgColsRows.Scan(&colName, &dataType); err == nil {
			pgCols[colName] = dataType
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
		skipRow := false

		for i, colName := range cols {
			dataType, exists := pgCols[colName]
			if !exists {
				continue // Skip column if it doesn't exist in Postgres
			}
			
			val := columns[i]
			if val == nil {
				continue
			}

			// Handle boolean conversions because MySQL TINYINT(1) doesn't auto-cast to Postgres BOOLEAN
			if dataType == "boolean" {
				switch v := val.(type) {
				case []byte:
					val = (string(v) == "1" || string(v) == "true" || string(v) == "t")
				case string:
					val = (v == "1" || v == "true" || v == "t")
				case int64:
					val = (v == 1)
				case int32:
					val = (v == 1)
				case int:
					val = (v == 1)
				}
			} else {
				// Convert []byte to string for Postgres compatibility
				if b, ok := val.([]byte); ok {
					val = string(b)
				}
			}

			// Validate Foreign Keys to prevent "violates foreign key constraint" error
			if val != nil && strings.HasSuffix(colName, "_id") {
				refTable := ""
				switch colName {
				case "category_id":
					refTable = "categories"
				case "tech_stack_id":
					refTable = "tech_stacks"
				case "project_id":
					refTable = "projects"
				case "work_experience_id":
					refTable = "work_experiences"
				case "tag_id":
					refTable = "tags"
				case "user_id":
					refTable = "users"
				}

				if refTable != "" {
					var exists int
					errCheck := pgDB.QueryRow(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1", refTable), val).Scan(&exists)
					if errCheck != nil {
						// Record does not exist in the referenced table.
						isPivot := pgTable == "project_tech_stacks" || pgTable == "project_tags" || pgTable == "work_experience_tech_stacks" || pgTable == "work_experience_tags"
						if isPivot {
							log.Printf("Warning: Invalid %s=%v in pivot table %s. Skipping entire row.", colName, val, pgTable)
							skipRow = true
							break
						} else {
							// For normal tables like tech_stacks, just set the relation to NULL
							log.Printf("Warning: Invalid %s=%v in table %s. Setting to NULL.", colName, val, pgTable)
							val = nil
						}
					}
				}
			}

			// Jika nil (NULL di MySQL atau karena foreign key tidak valid),
			// kita lewati saja agar Postgres bisa menggunakan DEFAULT-nya (misal: now() untuk created_at, atau default NULL)
			if val == nil {
				continue
			}

			validCols = append(validCols, colName)
			validValues = append(validValues, val)
		}

		if skipRow {
			continue
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
