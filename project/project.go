package project

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daveroberts0321/cloudpact/parser/grammar"
	"github.com/daveroberts0321/cloudpact/spec/openapi"
	"github.com/daveroberts0321/cloudpact/watch"
)

//go:embed templates/*
var templates embed.FS

// Init creates a new CloudPact project with scaffolding
func Init(name string) error {
	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create directory structure
	dirs := []string{
		filepath.Join(name, "models"),
		filepath.Join(name, "services"),
		filepath.Join(name, "generated", "go"),
		filepath.Join(name, "generated", "ts"),
		filepath.Join(name, "generated", "openapi"),
		filepath.Join(name, "web"),
		filepath.Join(name, "cmd", "ai-integration", "context"),
		filepath.Join(name, "cmd", "ai-integration", "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write template files
	templateFiles := map[string]string{
		"cloudpact.yaml":           "templates/cloudpact.yaml",
		"models/user.cp":           "templates/user.cp",
		"services/user_service.cp": "templates/user_service.cp",
		"web/index.html":           "templates/index.html",
		"web/main.ts":              "templates/main.ts",
		"go.mod":                   "templates/go.mod.tmpl",
		"README.md":                "templates/README.md",
		".gitignore":               "templates/gitignore",
	}

	for filePath, templatePath := range templateFiles {
		if err := writeTemplateFile(name, filePath, templatePath, name); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}

	return nil
}

func writeTemplateFile(projectDir, filePath, templatePath, projectName string) error {
	content, err := templates.ReadFile(templatePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, "{{.ProjectName}}", projectName)
	contentStr = strings.ReplaceAll(contentStr, "{{.ModuleName}}", strings.ToLower(projectName))

	fullPath := filepath.Join(projectDir, filePath)
	return os.WriteFile(fullPath, []byte(contentStr), 0644)
}

// StartDevServer starts the development server with file watching and hot reload
func StartDevServer() error {
	fmt.Println("Starting CloudPact development server...")

	if err := Build(); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	go func() {
		if err := watch.Watch(context.Background(), Build); err != nil {
			log.Printf("File watcher error: %v", err)
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.Handle("/generated/", http.StripPrefix("/generated/", http.FileServer(http.Dir("./generated"))))

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	port := 8080
	fmt.Printf("Server running at http://localhost:%d\n", port)
	fmt.Println("   Frontend: http://localhost:8080")
	fmt.Println("   API: http://localhost:8080/api/health")
	fmt.Println("   Generated files: http://localhost:8080/generated/")
	fmt.Println("\nWatching for file changes...")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// Build compiles all .cp files in the project
func Build() error {
	fmt.Println("Building CloudPact project...")

	cpFiles, err := FindCloudPactFiles(".")
	if err != nil {
		return err
	}

	if len(cpFiles) == 0 {
		fmt.Println("   No .cp files found")
		return nil
	}

	for _, file := range cpFiles {
		fmt.Printf("   Processing %s...\n", file)

		parsedFile, err := ParseCloudPactFile(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		if err := generateGoCode(parsedFile, file); err != nil {
			return fmt.Errorf("failed to generate Go code for %s: %w", file, err)
		}

		if err := generateTSCode(parsedFile, file); err != nil {
			return fmt.Errorf("failed to generate TypeScript code for %s: %w", file, err)
		}

		specPath := filepath.Join("generated", "openapi", strings.TrimSuffix(filepath.Base(file), ".cp")+".yaml")
		if err := openapi.WriteFile(parsedFile, specPath); err != nil {
			return fmt.Errorf("failed to generate OpenAPI spec for %s: %w", file, err)
		}
	}

	fmt.Printf("Built %d CloudPact files\n", len(cpFiles))
	return nil
}

func ParseCloudPactFile(filename string) (*grammar.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return grammar.ParseWithFilename(f, filename)
}

func FindCloudPactFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, "generated") || strings.Contains(path, "cmd/ai-integration/cache") {
			return nil
		}
		if strings.HasSuffix(path, ".cp") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// --- helper functions for code generation (generateGoCode, generateTSCode, etc.) will be placed here ---

// generateGoCode generates Go code from parsed CloudPact file with enhanced syntax support
func generateGoCode(file *grammar.File, sourcePath string) error {
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ".cp")
	outputPath := filepath.Join("generated", "go", baseName+".go")

	var goCode strings.Builder

	// Package and imports
	packageName := "main"
	if file.Module != nil {
		packageName = strings.ToLower(file.Module.Name)
	}

	goCode.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	goCode.WriteString("import (\n")
	goCode.WriteString("\t\"encoding/json\"\n")
	goCode.WriteString("\t\"fmt\"\n")
	goCode.WriteString("\t\"time\"\n")
	goCode.WriteString("\t\"errors\"\n")
	goCode.WriteString(")\n\n")

	// Generate module comment if present
	if file.Module != nil {
		goCode.WriteString(fmt.Sprintf("// %s module generated from CloudPact\n", file.Module.Name))
		goCode.WriteString("// This module contains business logic with embedded context\n\n")
	}

	// Generate Records (new syntax)
	for _, record := range file.Records {
		goCode.WriteString(generateGoRecord(record))
	}

	// Generate Models (legacy syntax - for backward compatibility)
	for _, model := range file.Models {
		goCode.WriteString(generateGoModel(model))
	}

	// Generate Functions with business logic
	for _, function := range file.Functions {
		goCode.WriteString(generateGoFunction(function))
	}

	return os.WriteFile(outputPath, []byte(goCode.String()), 0644)
}

// generateGoRecord creates Go struct from CloudPact record
func generateGoRecord(record *grammar.Record) string {
	var code strings.Builder

	code.WriteString(fmt.Sprintf("// %s represents a %s entity\n", record.Name, strings.ToLower(record.Name)))
	code.WriteString(fmt.Sprintf("type %s struct {\n", record.Name))

	// Add ID field by default
	code.WriteString("\tID string `json:\"id\" validate:\"required,uuid\"`\n")

	for _, field := range record.Fields {
		goType := mapCloudPactTypeToGo(field.Type.Name)
		jsonTag := strings.ToLower(field.Name)
		validateTag := getValidationTag(field.Type.Name)

		tag := fmt.Sprintf("`json:\"%s\"", jsonTag)
		if validateTag != "" {
			tag += fmt.Sprintf(" validate:\"%s\"", validateTag)
		}
		tag += "`"

		code.WriteString(fmt.Sprintf("\t%s %s %s\n", field.Name, goType, tag))
	}

	code.WriteString("}\n\n")
	return code.String()
}

// generateGoModel creates Go struct from legacy CloudPact model
func generateGoModel(model *grammar.Model) string {
	var code strings.Builder

	code.WriteString(fmt.Sprintf("// %s represents a %s entity (legacy model)\n", model.Name, strings.ToLower(model.Name)))
	code.WriteString(fmt.Sprintf("type %s struct {\n", model.Name))

	for _, field := range model.Fields {
		goType := mapCloudPactTypeToGo(field.Type.Name)
		jsonTag := fmt.Sprintf("`json:\"%s\"`", strings.ToLower(field.Name))
		code.WriteString(fmt.Sprintf("\t%s %s %s\n", field.Name, goType, jsonTag))
	}

	code.WriteString("}\n\n")
	return code.String()
}

// generateGoFunction creates Go function from CloudPact function with business context
func generateGoFunction(function *grammar.Function) string {
	var code strings.Builder

	// Function signature
	code.WriteString(fmt.Sprintf("// %s %s\n", function.Name, function.Why))

	// Add AI annotations as comments
	for _, annotation := range function.AIAnnotations {
		code.WriteString(fmt.Sprintf("// AI %s: %s\n", annotation.Type, annotation.Content))
	}

	code.WriteString(fmt.Sprintf("func %s(", function.Name))

	// Parameters
	for i, param := range function.Parameters {
		if i > 0 {
			code.WriteString(", ")
		}
		goType := mapCloudPactTypeToGo(param.Type.Name)
		code.WriteString(fmt.Sprintf("%s %s", param.Name, goType))
	}

	code.WriteString(")")

	// Return type
	if function.ReturnType != nil {
		goType := mapCloudPactTypeToGo(function.ReturnType.Name)
		code.WriteString(fmt.Sprintf(" %s", goType))
	}

	code.WriteString(" {\n")

	// Function body - convert CloudPact statements to Go
	if function.Body != nil {
		bodyCode := generateGoFunctionBody(function.Body)
		code.WriteString(bodyCode)
	}

	code.WriteString("}\n\n")
	return code.String()
}

// generateGoFunctionBody converts CloudPact function body to Go code
func generateGoFunctionBody(body *grammar.FunctionBody) string {
	var code strings.Builder

	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *grammar.IfStatement:
			code.WriteString(generateGoIfStatement(s))
		case *grammar.ReturnStatement:
			code.WriteString(generateGoReturnStatement(s))
		case *grammar.AssignStatement:
			code.WriteString(generateGoAssignStatement(s))
		case *grammar.CreateStatement:
			code.WriteString(generateGoCreateStatement(s))
		case *grammar.FailStatement:
			code.WriteString(generateGoFailStatement(s))
		}
	}

	// Add native Go blocks
	for _, nativeBlock := range body.NativeBlocks {
		if nativeBlock.Language == "go" {
			code.WriteString("\t// Native Go code block\n")
			// Split code by lines and indent each line
			lines := strings.Split(nativeBlock.Code, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					code.WriteString(fmt.Sprintf("\t%s\n", line))
				}
			}
		}
	}

	return code.String()
}

// generateGoIfStatement converts CloudPact if statement to Go
func generateGoIfStatement(stmt *grammar.IfStatement) string {
	var code strings.Builder

	condition := generateGoExpression(stmt.Condition)
	code.WriteString(fmt.Sprintf("\tif %s {\n", condition))

	// Then body
	if stmt.ThenStmt != nil {
		thenCode := generateGoStatement(stmt.ThenStmt)
		code.WriteString(fmt.Sprintf("\t\t%s\n", thenCode))
	}

	code.WriteString("\t}")

	// Else body
	if stmt.ElseStmt != nil {
		code.WriteString(" else {\n")
		elseCode := generateGoStatement(stmt.ElseStmt)
		code.WriteString(fmt.Sprintf("\t\t%s\n", elseCode))
		code.WriteString("\t}")
	}

	code.WriteString("\n")
	return code.String()
}

// generateGoReturnStatement converts CloudPact return to Go
func generateGoReturnStatement(stmt *grammar.ReturnStatement) string {
	if stmt.Value != nil {
		value := generateGoExpression(stmt.Value)
		return fmt.Sprintf("\treturn %s\n", value)
	}
	return "\treturn\n"
}

// generateGoAssignStatement converts CloudPact assignment to Go
func generateGoAssignStatement(stmt *grammar.AssignStatement) string {
	value := generateGoExpression(stmt.Value)
	return fmt.Sprintf("\t%s := %s\n", stmt.Variable, value)
}

// generateGoCreateStatement converts CloudPact create statement to Go
func generateGoCreateStatement(stmt *grammar.CreateStatement) string {
	var code strings.Builder

	code.WriteString(fmt.Sprintf("\t%s := &%s{\n", strings.ToLower(stmt.TypeName), stmt.TypeName))

	for _, assignment := range stmt.Assignments {
		value := generateGoExpression(assignment.Value)
		code.WriteString(fmt.Sprintf("\t\t%s: %s,\n", assignment.Field, value))
	}

	code.WriteString("\t}\n")
	return code.String()
}

// generateGoFailStatement converts CloudPact fail to Go error
func generateGoFailStatement(stmt *grammar.FailStatement) string {
	return fmt.Sprintf("\treturn errors.New(\"%s\")\n", stmt.Message)
}

// generateGoStatement converts any CloudPact statement to Go
func generateGoStatement(stmt grammar.Statement) string {
	switch s := stmt.(type) {
	case *grammar.IfStatement:
		return strings.TrimSpace(generateGoIfStatement(s))
	case *grammar.ReturnStatement:
		return strings.TrimSpace(generateGoReturnStatement(s))
	case *grammar.AssignStatement:
		return strings.TrimSpace(generateGoAssignStatement(s))
	case *grammar.CreateStatement:
		return strings.TrimSpace(generateGoCreateStatement(s))
	case *grammar.FailStatement:
		return strings.TrimSpace(generateGoFailStatement(s))
	default:
		return "// Unknown statement type"
	}
}

// generateGoExpression converts CloudPact expressions to Go
func generateGoExpression(expr grammar.Expression) string {
	switch e := expr.(type) {
	case *grammar.IdentifierExpression:
		return e.Name
	case *grammar.LiteralExpression:
		return fmt.Sprintf("%v", e.Value)
	case *grammar.BinaryExpression:
		left := generateGoExpression(e.Left)
		right := generateGoExpression(e.Right)

		// Map CloudPact operators to Go
		switch e.Operator {
		case "contains":
			return fmt.Sprintf("strings.Contains(%s, %s)", left, right)
		case "not contains":
			return fmt.Sprintf("!strings.Contains(%s, %s)", left, right)
		default:
			return fmt.Sprintf("%s %s %s", left, e.Operator, right)
		}
	case *grammar.MemberExpression:
		object := generateGoExpression(e.Object)
		return fmt.Sprintf("%s.%s", object, e.Property)
	case *grammar.CallExpression:
		var args []string
		for _, arg := range e.Arguments {
			args = append(args, generateGoExpression(arg))
		}
		return fmt.Sprintf("%s(%s)", e.Function, strings.Join(args, ", "))
	default:
		return "/* unknown expression */"
	}
}

// generateTSCode generates TypeScript code from parsed CloudPact file
func generateTSCode(file *grammar.File, sourcePath string) error {
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ".cp")
	outputPath := filepath.Join("generated", "ts", baseName+".ts")

	var tsCode strings.Builder
	tsCode.WriteString("// Generated TypeScript interfaces and functions from CloudPact\n")

	if file.Module != nil {
		tsCode.WriteString(fmt.Sprintf("// Module: %s\n", file.Module.Name))
	}

	tsCode.WriteString("// This code contains business logic with embedded context\n\n")

	// Generate Records (new syntax)
	for _, record := range file.Records {
		tsCode.WriteString(generateTSRecord(record))
	}

	// Generate Models (legacy syntax)
	for _, model := range file.Models {
		tsCode.WriteString(generateTSModel(model))
	}

	// Generate Functions
	for _, function := range file.Functions {
		tsCode.WriteString(generateTSFunction(function))
	}

	return os.WriteFile(outputPath, []byte(tsCode.String()), 0644)
}

// generateTSRecord creates TypeScript interface from CloudPact record
func generateTSRecord(record *grammar.Record) string {
	var code strings.Builder

	code.WriteString(fmt.Sprintf("// %s interface\n", record.Name))
	code.WriteString(fmt.Sprintf("export interface %s {\n", record.Name))

	// Add ID field by default
	code.WriteString("  id: string; // UUID\n")

	for _, field := range record.Fields {
		tsType := mapCloudPactTypeToTS(field.Type.Name)
		comment := getTypeComment(field.Type.Name)

		if comment != "" {
			code.WriteString(fmt.Sprintf("  %s: %s; // %s\n", strings.ToLower(field.Name), tsType, comment))
		} else {
			code.WriteString(fmt.Sprintf("  %s: %s;\n", strings.ToLower(field.Name), tsType))
		}
	}

	code.WriteString("}\n\n")
	return code.String()
}

// generateTSModel creates TypeScript interface from legacy CloudPact model
func generateTSModel(model *grammar.Model) string {
	var code strings.Builder

	code.WriteString(fmt.Sprintf("// %s interface (legacy model)\n", model.Name))
	code.WriteString(fmt.Sprintf("export interface %s {\n", model.Name))

	for _, field := range model.Fields {
		tsType := mapCloudPactTypeToTS(field.Type.Name)
		code.WriteString(fmt.Sprintf("  %s: %s;\n", strings.ToLower(field.Name), tsType))
	}

	code.WriteString("}\n\n")
	return code.String()
}

// generateTSFunction creates TypeScript function from CloudPact function
func generateTSFunction(function *grammar.Function) string {
	var code strings.Builder

	// Function comment with business context
	code.WriteString(fmt.Sprintf("/**\n * %s\n", function.Why))

	// Add AI annotations as JSDoc comments
	for _, annotation := range function.AIAnnotations {
		code.WriteString(fmt.Sprintf(" * @%s %s\n", annotation.Type, annotation.Content))
	}

	code.WriteString(" */\n")

	// Function signature
	code.WriteString(fmt.Sprintf("export function %s(", function.Name))

	// Parameters
	for i, param := range function.Parameters {
		if i > 0 {
			code.WriteString(", ")
		}
		tsType := mapCloudPactTypeToTS(param.Type.Name)
		code.WriteString(fmt.Sprintf("%s: %s", param.Name, tsType))
	}

	code.WriteString(")")

	// Return type
	if function.ReturnType != nil {
		tsType := mapCloudPactTypeToTS(function.ReturnType.Name)
		code.WriteString(fmt.Sprintf(": %s", tsType))
	}

	code.WriteString(" {\n")

	// Function body - simplified TypeScript generation
	if function.Body != nil {
		code.WriteString("  // Business logic implementation\n")
		code.WriteString("  // TODO: Implement CloudPact function body translation\n")

		// Add native TypeScript blocks
		for _, nativeBlock := range function.Body.NativeBlocks {
			if nativeBlock.Language == "ts" {
				code.WriteString("  // Native TypeScript code block\n")
				lines := strings.Split(nativeBlock.Code, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						code.WriteString(fmt.Sprintf("  %s\n", line))
					}
				}
			}
		}

		// Placeholder return for now
		if function.ReturnType != nil {
			switch mapCloudPactTypeToTS(function.ReturnType.Name) {
			case "boolean":
				code.WriteString("  return false;\n")
			case "number":
				code.WriteString("  return 0;\n")
			case "string":
				code.WriteString("  return '';\n")
			default:
				code.WriteString("  return null as any;\n")
			}
		}
	}

	code.WriteString("}\n\n")
	return code.String()
}

// Enhanced type mapping functions with semantic types
func mapCloudPactTypeToGo(cpType string) string {
	switch strings.ToLower(cpType) {
	// Basic types
	case "int", "integer":
		return "int"
	case "float", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "text", "string":
		return "string"

	// Semantic types - all map to string but with validation
	case "email", "url", "uuid", "phone":
		return "string"
	case "address", "zip_code", "country_code", "state_code":
		return "string"
	case "password", "token", "api_key":
		return "string"
	case "html", "markdown", "json":
		return "string"

	// Currency types
	case "usd_currency", "eur_currency", "percentage":
		return "float64"

	// Date/time types
	case "date", "datetime", "timestamp":
		return "time.Time"
	case "time":
		return "string" // Store as string for simplicity
	case "duration":
		return "time.Duration"

	// Default
	default:
		return "string"
	}
}

func mapCloudPactTypeToTS(cpType string) string {
	switch strings.ToLower(cpType) {
	// Basic types
	case "int", "integer", "float", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "text", "string":
		return "string"

	// Semantic types - all become string but with type comments
	case "email", "url", "uuid", "phone":
		return "string"
	case "address", "zip_code", "country_code", "state_code":
		return "string"
	case "password", "token", "api_key":
		return "string"
	case "html", "markdown", "json":
		return "string"

	// Currency and numeric types
	case "usd_currency", "eur_currency", "percentage":
		return "number"

	// Date/time types
	case "date", "datetime", "timestamp", "time":
		return "string" // ISO format strings
	case "duration":
		return "string" // ISO duration format

	// Default
	default:
		return "string"
	}
}

// getValidationTag returns validation tag for Go struct fields
func getValidationTag(cpType string) string {
	switch strings.ToLower(cpType) {
	case "email":
		return "required,email"
	case "url":
		return "required,url"
	case "uuid":
		return "required,uuid"
	case "phone":
		return "required,e164" // E.164 phone format
	case "zip_code":
		return "required,len=5"
	case "country_code":
		return "required,len=2,alpha"
	case "state_code":
		return "required,len=2,alpha"
	case "percentage":
		return "required,min=0,max=100"
	case "usd_currency", "eur_currency":
		return "required,min=0"
	case "password":
		return "required,min=8"
	default:
		return "required"
	}
}

// getTypeComment returns helpful comment for TypeScript types
func getTypeComment(cpType string) string {
	switch strings.ToLower(cpType) {
	case "email":
		return "Email address format"
	case "url":
		return "URL format"
	case "uuid":
		return "UUID format"
	case "phone":
		return "Phone number format"
	case "zip_code":
		return "ZIP/Postal code"
	case "country_code":
		return "ISO country code (US, CA, etc.)"
	case "state_code":
		return "State/province code"
	case "usd_currency":
		return "USD currency amount"
	case "eur_currency":
		return "EUR currency amount"
	case "percentage":
		return "Percentage (0-100)"
	case "date":
		return "Date in YYYY-MM-DD format"
	case "datetime", "timestamp":
		return "ISO 8601 datetime"
	case "password":
		return "Password (minimum 8 characters)"
	default:
		return ""
	}
}
