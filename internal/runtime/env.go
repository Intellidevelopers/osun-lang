package runtime

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// InitEnv loads .env file (if present) and initializes databases.
func InitEnv() {
	// Try loading .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("[osun] No .env file found — using system environment")
	}

	fmt.Println("[osun] Environment initialized")

	// Initialize all databases (auto-detect)
	InitMongoFromEnv()
	InitMySQLFromEnv()
	InitPostgresFromEnv()
}

// GetEnv returns an environment variable or default value if missing.
func GetEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
