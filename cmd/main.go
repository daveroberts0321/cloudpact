// File: cmd/cloudpact/main.go
package main

import (
	"fmt"
	"os"
	"strings"

	"cloudpact/parser/grammar"
	"cloudpact/spec/openapi"
	"cloudpact/tsgen"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cloudpact gen model <ModelName>\n       cloudpact gen openapi <file.cf>\n       cloudpact gen ts <spec.yaml>")
		return
	}

	cmd := os.Args[1]
	sub := os.Args[2]

	switch {
	case cmd == "gen" && sub == "model":
		if len(os.Args) < 4 {
			fmt.Println("missing model name")
			return
		}
		modelName := os.Args[3]
		generateModel(modelName)
	case cmd == "gen" && sub == "openapi":
		if len(os.Args) < 4 {
			fmt.Println("missing .cf file")
			return
		}
		cf := os.Args[3]
		if err := generateOpenAPI(cf); err != nil {
			fmt.Println("error:", err)
		}
	case cmd == "gen" && sub == "ts":
		if len(os.Args) < 4 {
			fmt.Println("missing OpenAPI spec file")
			return
		}
		spec := os.Args[3]
		if err := tsgen.Generate(spec); err != nil {
			fmt.Println("error:", err)
		}
	default:
		fmt.Println("Unknown command")
	}
}

func generateModel(name string) {
	model := strings.Title(name)
	goCode := "package models\n\n" +
		fmt.Sprintf("type %s struct {\n", model) +
		"\tID   string `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"}\n"
	tsCode := fmt.Sprintf("export interface %s {\n\tid: string;\n\tname: string;\n}\n", model)

	os.MkdirAll("generated/go", 0755)
	os.MkdirAll("generated/ts", 0755)
	_ = os.WriteFile(fmt.Sprintf("generated/go/%s.go", strings.ToLower(name)), []byte(goCode), 0644)
	_ = os.WriteFile(fmt.Sprintf("generated/ts/%s.ts", strings.ToLower(name)), []byte(tsCode), 0644)

	fmt.Printf("Model %s generated in Go and TypeScript.\n", model)
}

func generateOpenAPI(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	file, err := grammar.Parse(f)
	if err != nil {
		return err
	}
	if err := openapi.WriteFile(file, "generated/openapi/spec.yaml"); err != nil {
		return err
	}
	fmt.Println("OpenAPI spec written to generated/openapi/spec.yaml")
	return nil
}
