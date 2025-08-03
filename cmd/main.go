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

	"cloudpact/parser/grammar"
	"cloudpact/spec/openapi"

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
		fmt.Printf("✅ Project '%s' initialized successfully!\n", projectName)
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
				fmt.Println("✅ Project built successfully!")
			}
		default:
			fmt.Printf("Unknown start command: %s\n", subCmd)
		}

	case "gen":
		if len(os.Args) < 3 {
			fmt.Println("Usage: cloudpact gen <model|openapi> [args...]")
			return
		}
		subCmd := os.Args[2]
		switch subCmd {
		case "model":
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

	case "watch":
		if err := watchAndBuild(); err != nil {
			fmt.Printf("Error watching files: %v\n", err)
		}

	case "version":
		fmt.Println("CloudPact v0.1.0 - Human/AI collaborative programming language")

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
    init <name>        Initialize a new CloudPact project
    start http         Start development server with hot reload
    start build        Build the project once
    gen model <name>   Generate a model template
    gen openapi <file> Generate OpenAPI spec from .cp file
    watch              Watch files and rebuild on changes
    version            Show version information
    help               Show this help message

EXAMPLES:
    cloudpact init myapp
    cloudpact start http
    cloudpact gen model User
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
	fmt.Println("🚀 Starting CloudPact development server...")

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
	fmt.Printf("📡 Server running at http://localhost:%d\n", port)
	fmt.Println("   Frontend: http://localhost:8080")
	fmt.Println("   API: http://localhost:8080/api/health")
	fmt.Println("   Generated files: http://localhost:8080/generated/")
	fmt.Println("\n👀 Watching for file changes...")

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// buildProject compiles all .cp files in the project
func buildProject() error {
	fmt.Println("🔨 Building CloudPact project...")

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

		// Parse the file
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", file, err)
		}

		parsedFile, err := grammar.Parse(f)
		f.Close()
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

	fmt.Printf("✅ Built %d CloudPact files\n", len(cpFiles))
	return nil
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
				fmt.Printf("🔄 File changed: %s\n", event.Name)
				if err := buildProject(); err != nil {
					fmt.Printf("❌ Build failed: %v\n", err)
				} else {
					fmt.Println("✅ Rebuild complete")
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

		// Skip generated directory
		if strings.Contains(path, "generated") {
			return nil
		}

		if strings.HasSuffix(path, ".cp") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// generateGoCode generates Go code from parsed CloudPact file
func generateGoCode(file *grammar.File, sourcePath string) error {
	// For now, use the existing model generation logic
	// This would be expanded to handle full CloudPact syntax

	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ".cp")
	outputPath := filepath.Join("generated", "go", baseName+".go")

	// Simple Go code generation - extend this for full CloudPact support
	var goCode strings.Builder
	goCode.WriteString("package main\n\n")
	goCode.WriteString("import (\n\t\"encoding/json\"\n\t\"fmt\"\n)\n\n")

	for _, model := range file.Models {
		goCode.WriteString(fmt.Sprintf("// %s represents a %s entity\n", model.Name, strings.ToLower(model.Name)))
		goCode.WriteString(fmt.Sprintf("type %s struct {\n", model.Name))

		for _, field := range model.Fields {
			goType := mapCloudPactTypeToGo(field.Type.Name)
			jsonTag := fmt.Sprintf("`json:\"%s\"`", strings.ToLower(field.Name))
			goCode.WriteString(fmt.Sprintf("\t%s %s %s\n", field.Name, goType, jsonTag))
		}

		goCode.WriteString("}\n\n")
	}

	return os.WriteFile(outputPath, []byte(goCode.String()), 0644)
}

// generateTSCode generates TypeScript code from parsed CloudPact file
func generateTSCode(file *grammar.File, sourcePath string) error {
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), ".cp")
	outputPath := filepath.Join("generated", "ts", baseName+".ts")

	var tsCode strings.Builder
	tsCode.WriteString("// Generated TypeScript interfaces from CloudPact\n\n")

	for _, model := range file.Models {
		tsCode.WriteString(fmt.Sprintf("export interface %s {\n", model.Name))

		for _, field := range model.Fields {
			tsType := mapCloudPactTypeToTS(field.Type.Name)
			tsCode.WriteString(fmt.Sprintf("  %s: %s;\n", strings.ToLower(field.Name), tsType))
		}

		tsCode.WriteString("}\n\n")
	}

	return os.WriteFile(outputPath, []byte(tsCode.String()), 0644)
}

// mapCloudPactTypeToGo maps CloudPact types to Go types
func mapCloudPactTypeToGo(cpType string) string {
	switch strings.ToLower(cpType) {
	case "int", "integer":
		return "int"
	case "float", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "text", "string", "email", "url", "uuid":
		return "string"
	default:
		return "string"
	}
}

// mapCloudPactTypeToTS maps CloudPact types to TypeScript types
func mapCloudPactTypeToTS(cpType string) string {
	switch strings.ToLower(cpType) {
	case "int", "integer", "float", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "text", "string", "email", "url", "uuid":
		return "string"
	default:
		return "string"
	}
}

// Legacy functions for backward compatibility
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
