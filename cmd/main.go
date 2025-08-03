// File: cmd/cloudpact/main.go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cloudpact gen model <ModelName>")
		return
	}

	cmd := os.Args[1]
	sub := os.Args[2]

	if cmd == "gen" && sub == "model" {
		modelName := os.Args[3]
		generateModel(modelName)
	} else {
		fmt.Println("Unknown command")
	}
}

func generateModel(name string) {
	model := strings.Title(name)
	goCode := fmt.Sprintf("package models\n\ntype %s struct {\n\tID   string `json:\"id\"`\n\tName string `json:\"name\"`\n}\n", model)
	tsCode := fmt.Sprintf("export interface %s {\n\tid: string;\n\tname: string;\n}\n", model)

	os.MkdirAll("generated/go", 0755)
	os.MkdirAll("generated/ts", 0755)
	_ = os.WriteFile(fmt.Sprintf("generated/go/%s.go", strings.ToLower(name)), []byte(goCode), 0644)
	_ = os.WriteFile(fmt.Sprintf("generated/ts/%s.ts", strings.ToLower(name)), []byte(tsCode), 0644)

	fmt.Printf("Model %s generated in Go and TypeScript.\n", model)
}
