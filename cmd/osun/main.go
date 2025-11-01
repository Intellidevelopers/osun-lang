package main

import (
	"fmt"
	"os"

	"github.com/intellidevelopers/osun-lang/internal/interpreter"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: osun run <filename.os>")
		return
	}

	cmd := os.Args[1]
	file := os.Args[2]

	if cmd != "run" {
		fmt.Println("Unknown command:", cmd)
		return
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("File not found:", file)
		return
	}

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	interpreter.Run(string(content))
}
