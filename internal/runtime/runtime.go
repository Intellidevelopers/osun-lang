package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"

	
)

// -------------------- Global Symbols --------------------
var symbols = map[string]interface{}{}

func InitBuiltins() {
	// Print function
	symbols["print"] = func(args ...interface{}) {
		for _, a := range args {
			fmt.Println(a)
		}
	}

	// HTTP server object
	symbols["http"] = map[string]interface{}{
		"createServer": func(port int) *OsunServer {
			server := NewOsunServer(port)
			return server
		},
	}

	// Database manager
	symbols["db"] = map[string]interface{}{
		"connect": func(driver, connStr string) string {
			if err := ConnectDatabase(driver, connStr); err != nil {
				fmt.Println("DB connect error:", err)
				return "failed"
			}
			return "connected"
		},
		"insert": func(table string, data map[string]interface{}) {
			if err := Insert(table, data); err != nil {
				fmt.Println("DB insert error:", err)
			} else {
				fmt.Println("✅ Inserted into", table)
			}
		},
	}

	// Auth middleware
	symbols["auth"] = map[string]interface{}{
		"requireAuth": RequireAuth,
	}

	fmt.Println("✅ Osun builtins initialized.")
}

func GetSymbol(name string) interface{} {
	return symbols[name]
}

// -------------------- Server --------------------


type OsunRequest struct {
	Body   map[string]interface{}
	Header http.Header
	Method string
	URL    string
}

type OsunResponse struct {
	writer http.ResponseWriter
}

func (res *OsunResponse) Send(v interface{}) {
	switch t := v.(type) {
	case string:
		res.writer.Write([]byte(t))
	default:
		json.NewEncoder(res.writer).Encode(t)
	}
}

// -------------------- Database Helpers --------------------
func ConnectDatabase(driver, connStr string) error {
	fmt.Println("[osun] connecting to", driver)
	// Implement real DB connection if needed
	return nil
}

func Insert(table string, data map[string]interface{}) error {
	jsonStr, _ := json.Marshal(data)
	return DBInsert(table, string(jsonStr))
}


