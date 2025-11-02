package runtime

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlDB is nil when MySQL is not configured.
var mysqlDB *sql.DB

// InitMySQLFromEnv tries to connect to MySQL using MYSQL_DSN env var.
// Example DSN: root:password@tcp(127.0.0.1:3306)/osun_db
func InitMySQLFromEnv() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("MySQL open error:", err)
		return
	}
	if err := db.Ping(); err != nil {
		fmt.Println("MySQL ping error:", err)
		_ = db.Close()
		return
	}
	mysqlDB = db
	fmt.Println("MySQL connected")
}

// DBInsertMySQL inserts jsonStr (JSON object) into `table`.
// It unmarshals JSON into map[string]interface{} and maps keys -> columns.
// NOTE: column names must exist in the table. This function uses `?` placeholders.
func DBInsertMySQL(table, jsonStr string) error {
	if mysqlDB == nil {
		return fmt.Errorf("mysql not configured")
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

	for k, v := range payload {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ","),
		strings.Join(placeholders, ","))

	_, err := mysqlDB.Exec(query, vals...)
	if err != nil {
		return fmt.Errorf("mysql exec error: %w", err)
	}
	return nil
}
