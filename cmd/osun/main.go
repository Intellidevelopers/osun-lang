package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/intellidevelopers/osun-lang/internal/interpreter"
	"github.com/intellidevelopers/osun-lang/internal/runtime"
)

func main() {
	// Initialize built-in functions and objects
	runtime.InitBuiltins()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file.os>")
		return
	}

	file := os.Args[1]
	runAndMaybeStartServer(file)
}

func runAndMaybeStartServer(file string) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return
	}

	// Create a server and store it in interpreter variables
	server := runtime.NewOsunServer(8080) // default port, change if needed
	interpreter.SetVariable("server", server)

	// Run the .os code
	interpreter.Run(string(data))

	// If the code created any handlers and called server.Listen(), it will run
	fmt.Println("Server started on port 8080. Press Ctrl+C to stop.")
	select {} // block forever to keep the server running
}
