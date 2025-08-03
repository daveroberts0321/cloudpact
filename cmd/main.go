// File: cmd/cloudpact/main.go
package main

import (
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

	"github.com/fsnotify/fsnotify"
)

//go:embed templates/*
var templates embed.FS

// ProjectConfig holds the configuration for a CloudPact project
type ProjectConfig struct {
	Name       string   `yaml:"name"`
	Version    string   `yaml:"version"`
	GoModule   string   `yaml:"go_module"`
	Port       int      `yaml:"port"`
	WatchPaths []string `yaml:"watch_paths"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact init <project-name>")
			return
		}
		projectName := os.Args[2]
		if err := initProject(projectName); err != nil {
			fmt.Printf("Error initializing project: %v\n", err)
			return
		}
		fmt.Printf("Project '%s' initialized successfully!\n", projectName)
		fmt.Printf("   cd %s\n", projectName)
		fmt.Printf("   cloudpact start http\n")

	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact start <http|build>")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "http":
			if err := startDevServer(); err != nil {
				fmt.Printf("Error starting dev server: %v\n", err)
			}
		case "build":
			if err := buildProject(); err != nil {
				fmt.Printf("Error building project: %v\n", err)
			} else {
				fmt.Println("Project built successfully!")
			}
		default:
			fmt.Printf("Unknown start command: %s\n", subCmd)
		}

	case "gen":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact gen <record|function|openapi> [args...]")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "record":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen record <RecordName>")
				return
			}
			recordName := os.Args[3]
			generateRecord(recordName)
		case "function":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen function <FunctionName>")
				return
			}
			functionName := os.Args[3]
			generateFunction(functionName)
		case "model": // Legacy support
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen model <ModelName>")
				return
			}
			modelName := os.Args[3]
			generateModel(modelName)
		case "openapi":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact gen openapi <file.cp>")
				return
			}
			cfFile := os.Args[3]
			if err := generateOpenAPI(cfFile); err != nil {
				fmt.Printf("Error generating OpenAPI: %v\n", err)
			}
		default:
			fmt.Printf("Unknown gen command: %s\n", subCmd)
		}

	case "ai":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact ai <review|feedback|status|accept> [args...]")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "review":
			if len(os.Args) < 4 {
				fmt.Println("Usage: cloudpact ai review <file.cp>")
				return
			}
			fileName := os.Args[3]
			fmt.Printf("AI review for %s (not yet implemented)\n", fileName)
		case "feedback":
			fmt.Println("AI feedback session (not yet implemented)")
		case "status":
			fmt.Println("AI suggestions status (not yet implemented)")
		case "accept":
			fmt.Println("Accept AI suggestion (not yet implemented)")
		default:
			fmt.Printf("Unknown ai command: %s\n", subCmd)
		}

	case "watch":
		if err := watchAndBuild(); err != nil {
			fmt.Printf("Error watching files: %v\n", err)
		}

	case "version":
		fmt.Println("CloudPact v0.2.0 - Human/AI collaborative programming language")

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

// printUsage displays the help information for CloudPact CLI
func printUsage() {
	fmt.Println(`CloudPact - Human/AI collaborative programming language

USAGE:
    cloudpact <command> [arguments]

COMMANDS:
    init <name>           Initialize a new CloudPact project
    start http            Start development server with hot reload
    start build           Build the project once
    gen record <name>     Generate a record template
    gen function <name>   Generate a function template
    gen model <name>      Generate a model template (legacy)
    gen openapi <file>    Generate OpenAPI spec from .cp file
    ai review <file>      AI reviews a specific file
    ai feedback           Interactive AI feedback session
    ai status             Show pending AI suggestions
    ai accept <id>        Accept a specific AI suggestion
    watch                 Watch files and rebuild on changes
    version               Show version information
    help                  Show this help message

EXAMPLES:
    cloudpact init myapp
    cloudpact start http
    cloudpact gen record User
    cloudpact gen function validateUser
    cloudpact ai review models/user.cp
    cloudpact gen openapi models/user.cp`)
}

// initProject creates a new CloudPact project with scaffolding
func initProject(name string) error {
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
		"go.mod":                   "templates/go.mod",
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

// writeTemplateFile writes a template file with variable substitution
func writeTemplateFile(projectDir, filePath, templatePath, projectName string) error {
	content, err := templates.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Simple template variable substitution
	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, "{{.ProjectName}}", projectName)
	contentStr = strings.ReplaceAll(contentStr, "{{.ModuleName}}", strings.ToLower(projectName))

	fullPath := filepath.Join(projectDir, filePath)
	return os.WriteFile(fullPath, []byte(contentStr), 0644)
}

// startDevServer starts the development server with file watching and hot reload
func startDevServer() error {
	fmt.Println("Starting CloudPact development server...")

	// Initial build
	if err := buildProject(); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	// Start file watcher in a goroutine
	go func() {
		if err := watchAndBuild(); err != nil {
			log.Printf("File watcher error: %v", err)
		}
	}()

	// Start HTTP server for frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.Handle("/generated/", http.StripPrefix("/generated/", http.FileServer(http.Dir("./generated"))))

	// Add a simple API endpoint for testing
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

// buildProject compiles all .cp files in the project
func buildProject() error {
	fmt.Println("Building CloudPact project...")

	// Find all .cp files
	cpFiles, err := findCloudPactFiles(".")
	if err != nil {
		return err
	}

	if len(cpFiles) == 0 {
		fmt.Println("   No .cp files found")
		return nil
	}

	// Process each .cp file
	for _, file := range cpFiles {
		fmt.Printf("   Processing %s...\n", file)

		// Parse the file with filename for better error reporting
		parsedFile, err := parseCloudPactFile(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		// Generate Go code
		if err := generateGoCode(parsedFile, file); err != nil {
			return fmt.Errorf("failed to generate Go code for %s: %w", file, err)
		}

		// Generate TypeScript code
		if err := generateTSCode(parsedFile, file); err != nil {
			return fmt.Errorf("failed to generate TypeScript code for %s: %w", file, err)
		}

		// Generate OpenAPI spec
		specPath := filepath.Join("generated", "openapi", strings.TrimSuffix(filepath.Base(file), ".cp")+".yaml")
		if err := openapi.WriteFile(parsedFile, specPath); err != nil {
			return fmt.Errorf("failed to generate OpenAPI spec for %s: %w", file, err)
		}
	}

	fmt.Printf("Built %d CloudPact files\n", len(cpFiles))
	return nil
}

// parseCloudPactFile parses a CloudPact file with filename context
func parseCloudPactFile(filename string) (*grammar.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return grammar.ParseWithFilename(f, filename)
}

// watchAndBuild watches for file changes and rebuilds
func watchAndBuild() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add directories to watch
	watchDirs := []string{"models", "services"}
	for _, dir := range watchDirs {
		if _, err := os.Stat(dir); err == nil {
			if err := watcher.Add(dir); err != nil {
				return err
			}
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only rebuild on .cp file changes
			if strings.HasSuffix(event.Name, ".cp") && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				fmt.Printf("File changed: %s\n", event.Name)
				if err := buildProject(); err != nil {
					fmt.Printf("Build failed: %v\n", err)
				} else {
					fmt.Println("Rebuild complete")
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// findCloudPactFiles recursively finds all .cp files
func findCloudPactFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip generated directory and AI integration cache
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

// Generation helpers for new CloudPact syntax
func generateRecord(name string) {
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

func generateFunction(name string) {
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

	// Check if file exists, if so append to it
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

// Legacy function for backward compatibility
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

	fmt.Printf("Legacy model %s generated in Go and TypeScript.\n", model)
}

func generateOpenAPI(path string) error {
	parsedFile, err := parseCloudPactFile(path)
	if err != nil {
		return err
	}

	if err := openapi.WriteFile(parsedFile, "generated/openapi/spec.yaml"); err != nil {
		return err
	}

	fmt.Println("OpenAPI spec written to generated/openapi/spec.yaml")
	return nil
}
