package runtime

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var pgDB *sql.DB

// InitPostgresFromEnv tries to connect to Postgres using POSTGRES_DSN env var.
// Example DSN: postgres://user:pass@localhost:5432/osun_db?sslmode=disable
func InitPostgresFromEnv() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("Postgres open error:", err)
		return
	}
	if err := db.Ping(); err != nil {
		fmt.Println("Postgres ping error:", err)
		_ = db.Close()
		return
	}
	pgDB = db
	fmt.Println("Postgres connected")
}

// DBInsertPostgres inserts jsonStr into table using numbered placeholders $1, $2...
func DBInsertPostgres(table, jsonStr string) error {
	if pgDB == nil {
		return fmt.Errorf("postgres not configured")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}

	cols := make([]string, 0, len(payload))
	vals := make([]interface{}, 0, len(payload))
	placeholders := make([]string, 0, len(payload))

	i := 1
	for k, v := range payload {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ","),
		strings.Join(placeholders, ","))

	_, err := pgDB.Exec(query, vals...)
	if err != nil {
		return fmt.Errorf("postgres exec error: %w", err)
	}
	return nil
}
