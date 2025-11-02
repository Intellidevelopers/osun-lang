package runtime

import (
	"encoding/json"
	"fmt"
)

// DBInsert attempts to insert into the first available database (Mongo > MySQL > Postgres)
func DBInsert(target, jsonStr string) error {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if mongoClient != nil {
		if err := DBInsertMongo(target, jsonStr); err == nil {
			fmt.Println("[osun] MongoDB insert success")
			return nil
		}
	}

	if mysqlDB != nil {
		if err := DBInsertMySQL(target, jsonStr); err == nil {
			fmt.Println("[osun] MySQL insert success")
			return nil
		}
	}

	if pgDB != nil {
		if err := DBInsertPostgres(target, jsonStr); err == nil {
			fmt.Println("[osun] Postgres insert success")
			return nil
		}
	}

	return fmt.Errorf("no active database connection")
}

// DBConnected prints which databases are active
func DBConnected() {
	fmt.Println("======== Active Databases ========")
	if mongoClient != nil {
		fmt.Println("- MongoDB ✅")
	} else {
		fmt.Println("- MongoDB ❌")
	}

	if mysqlDB != nil {
		fmt.Println("- MySQL ✅")
	} else {
		fmt.Println("- MySQL ❌")
	}

	if pgDB != nil {
		fmt.Println("- Postgres ✅")
	} else {
		fmt.Println("- Postgres ❌")
	}
	fmt.Println("===================================")
}
