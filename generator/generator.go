package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/daveroberts0321/cloudpact/project"
	"github.com/daveroberts0321/cloudpact/spec/openapi"
)

func GenerateRecord(name string) {
	recordTemplate := fmt.Sprintf(`module %sModule

define record %s
    name: text
    email: email
    createdAt: datetime

function validate%s(record: %s) returns boolean
    why: "Validates %s data meets business requirements"
    do:
        if record.name is empty
            then return false
        if not record.email contains "@"
            then return false
        return true
`, name, name, name, name, strings.ToLower(name))

	os.MkdirAll("models", 0755)
	filename := fmt.Sprintf("models/%s.cp", strings.ToLower(name))
	err := os.WriteFile(filename, []byte(recordTemplate), 0644)
	if err != nil {
		fmt.Printf("Error writing record template: %v\n", err)
		return
	}

	fmt.Printf("Record %s generated at %s\n", name, filename)
}

func GenerateFunction(name string) {
	functionTemplate := fmt.Sprintf(`function %s(input: text) returns boolean
    ai-feedback: "Consider adding input validation"
    why: "Performs %s operation with business context"
    do:
        if input is empty
            then return false
        // Add your business logic here
        return true
`, name, strings.ToLower(name))

	os.MkdirAll("services", 0755)
	filename := fmt.Sprintf("services/%s_service.cp", strings.ToLower(name))

	var content string
	if existingContent, err := os.ReadFile(filename); err == nil {
		content = string(existingContent) + "\n\n" + functionTemplate
	} else {
		content = fmt.Sprintf("module %sService\n\n", name) + functionTemplate
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing function template: %v\n", err)
		return
	}

	fmt.Printf("Function %s generated at %s\n", name, filename)
}

func GenerateModel(name string) {
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

	fmt.Printf("Legacy model %s generated in Go and TypeScript.\n", model)
}

func GenerateOpenAPI(path string) error {
	parsedFile, err := project.ParseCloudPactFile(path)
	if err != nil {
		return err
	}

	if err := openapi.WriteFile(parsedFile, "generated/openapi/spec.yaml"); err != nil {
		return err
	}

	fmt.Println("OpenAPI spec written to generated/openapi/spec.yaml")
	return nil
}
