package openapi

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/daveroberts0321/cloudpact/parser/grammar"
)

// APIConfig holds configuration for API generation
type APIConfig struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	ServerURL   string `yaml:"server_url"`
}

// DefaultAPIConfig provides sensible defaults
func DefaultAPIConfig() *APIConfig {
	return &APIConfig{
		Title:       "CloudPact API",
		Version:     "1.0.0",
		Description: "Generated API from CloudPact models and services",
		ServerURL:   "http://localhost:8080",
	}
}

// LoadAPIConfig attempts to load API configuration from cloudpact.yaml
func LoadAPIConfig(configPath string) (*APIConfig, error) {
	config := DefaultAPIConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil // Use defaults if no config file
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	var projectConfig struct {
		API *APIConfig `yaml:"api"`
	}

	if err := yaml.Unmarshal(data, &projectConfig); err != nil {
		return config, err
	}

	if projectConfig.API != nil {
		// Merge with defaults
		if projectConfig.API.Title != "" {
			config.Title = projectConfig.API.Title
		}
		if projectConfig.API.Version != "" {
			config.Version = projectConfig.API.Version
		}
		if projectConfig.API.Description != "" {
			config.Description = projectConfig.API.Description
		}
		if projectConfig.API.ServerURL != "" {
			config.ServerURL = projectConfig.API.ServerURL
		}
	}

	return config, nil
}

// Generate converts a parsed CloudPact AST into an OpenAPI document
// represented as a YAML string with support for semantic types and relationships
func Generate(file *grammar.File) (string, error) {
	return GenerateWithConfig(file, DefaultAPIConfig())
}

// GenerateWithConfig allows custom API configuration
func GenerateWithConfig(file *grammar.File, config *APIConfig) (string, error) {
	if file == nil {
		return "", fmt.Errorf("nil file")
	}

	doc := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       config.Title,
			"version":     config.Version,
			"description": config.Description,
		},
		"servers": []interface{}{
			map[string]interface{}{
				"url":         config.ServerURL,
				"description": "Development server",
			},
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{},
		},
		"paths": map[string]interface{}{},
	}

	schemas := doc["components"].(map[string]interface{})["schemas"].(map[string]interface{})
	paths := doc["paths"].(map[string]interface{})

	// Generate schemas for models
	for _, m := range file.Models {
		schema := generateModelSchema(m)
		schemas[m.Name] = schema

		// Generate basic CRUD paths for each model
		generateModelPaths(paths, m)
	}

	// TODO: Generate paths for functions when function parsing is implemented
	// for _, f := range file.Functions {
	//     generateFunctionPath(paths, f)
	// }

	return toYAML(doc, 0), nil
}

// generateModelSchema creates an OpenAPI schema for a CloudPact model
func generateModelSchema(model *grammar.Model) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []interface{}{},
	}

	props := schema["properties"].(map[string]interface{})
	required := []interface{}{}

	// Add an ID field by default for all models
	props["id"] = map[string]interface{}{
		"type":        "string",
		"format":      "uuid",
		"description": "Unique identifier",
		"example":     "123e4567-e89b-12d3-a456-426614174000",
	}
	required = append(required, "id")

	for _, field := range model.Fields {
		fieldSchema := generateFieldSchema(field)
		props[field.Name] = fieldSchema

		// For now, mark all fields as required
		// TODO: Add optional field support to CloudPact syntax
		required = append(required, field.Name)
	}

	schema["required"] = required
	return schema
}

// generateFieldSchema creates an OpenAPI schema for a model field with semantic type support
func generateFieldSchema(field *grammar.Field) map[string]interface{} {
	baseType, format, description, example, constraints := mapSemanticType(field.Type.Name)

	fieldSchema := map[string]interface{}{
		"type": baseType,
	}

	if format != "" {
		fieldSchema["format"] = format
	}

	if description != "" {
		fieldSchema["description"] = description
	}

	if example != nil {
		fieldSchema["example"] = example
	}

	// Add validation constraints
	for key, value := range constraints {
		fieldSchema[key] = value
	}

	return fieldSchema
}

// mapSemanticType maps CloudPact semantic types to OpenAPI types with validation and examples
func mapSemanticType(cpType string) (baseType, format, description string, example interface{}, constraints map[string]interface{}) {
	constraints = make(map[string]interface{})

	switch strings.ToLower(cpType) {
	// Basic types
	case "int", "integer":
		return "integer", "int32", "Integer value", 42, constraints
	case "long", "bigint":
		return "integer", "int64", "Long integer value", 1234567890, constraints
	case "float", "double", "number":
		return "number", "float", "Floating point number", 123.45, constraints
	case "bool", "boolean":
		return "boolean", "", "Boolean value", true, constraints
	case "text", "string":
		return "string", "", "Text string", "Sample text", constraints

	// Semantic string types
	case "email":
		constraints["pattern"] = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
		return "string", "email", "Email address", "user@example.com", constraints

	case "url", "uri":
		constraints["format"] = "uri"
		return "string", "uri", "URL/URI", "https://example.com", constraints

	case "uuid", "id":
		constraints["pattern"] = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
		return "string", "uuid", "UUID identifier", "123e4567-e89b-12d3-a456-426614174000", constraints

	case "phone", "phone_number":
		constraints["pattern"] = "^\\+?[1-9]\\d{1,14}$"
		return "string", "", "Phone number", "+1-555-0123", constraints

	// Address types
	case "street_address", "address":
		constraints["minLength"] = 5
		constraints["maxLength"] = 200
		return "string", "", "Street address", "123 Main St, Anytown, ST 12345", constraints

	case "zip_code", "postal_code":
		constraints["pattern"] = "^\\d{5}(-\\d{4})?$"
		return "string", "", "ZIP/Postal code", "12345", constraints

	case "country_code":
		constraints["pattern"] = "^[A-Z]{2}$"
		constraints["minLength"] = 2
		constraints["maxLength"] = 2
		return "string", "", "ISO country code", "US", constraints

	case "state_code":
		constraints["pattern"] = "^[A-Z]{2}$"
		constraints["minLength"] = 2
		constraints["maxLength"] = 2
		return "string", "", "State/province code", "CA", constraints

	// Currency and financial types
	case "usd_currency", "currency_usd":
		constraints["minimum"] = 0
		constraints["multipleOf"] = 0.01
		return "number", "currency", "USD currency amount", 99.99, constraints

	case "eur_currency", "currency_eur":
		constraints["minimum"] = 0
		constraints["multipleOf"] = 0.01
		return "number", "currency", "EUR currency amount", 85.50, constraints

	case "percentage":
		constraints["minimum"] = 0
		constraints["maximum"] = 100
		return "number", "float", "Percentage value (0-100)", 75.5, constraints

	// Date and time types
	case "date":
		return "string", "date", "Date (YYYY-MM-DD)", "2023-12-25", constraints

	case "datetime", "timestamp":
		return "string", "date-time", "Date and time (ISO 8601)", "2023-12-25T10:30:00Z", constraints

	case "time":
		constraints["pattern"] = "^([0-1]?[0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9])?$"
		return "string", "time", "Time (HH:MM or HH:MM:SS)", "14:30:00", constraints

	case "duration":
		constraints["pattern"] = "^P(?:([0-9]+)D)?(?:T(?:([0-9]+)H)?(?:([0-9]+)M)?(?:([0-9]+(?:\\.[0-9]+)?)S)?)?$"
		return "string", "duration", "ISO 8601 duration", "P1DT2H30M", constraints

	// Security and authentication types
	case "password":
		constraints["minLength"] = 8
		constraints["maxLength"] = 128
		return "string", "password", "Password (masked in examples)", "********", constraints

	case "token", "access_token":
		constraints["pattern"] = "^[A-Za-z0-9_-]+$"
		return "string", "", "Authentication token", "eyJhbGciOiJIUzI1NiIs...", constraints

	case "api_key":
		constraints["pattern"] = "^[A-Za-z0-9_-]{32,}$"
		return "string", "", "API key", "ak_1234567890abcdef", constraints

	// Content types
	case "html":
		return "string", "", "HTML content", "<p>Hello world</p>", constraints

	case "markdown":
		return "string", "", "Markdown content", "# Hello\n\nWorld", constraints

	case "json":
		return "string", "", "JSON string", "{\"key\": \"value\"}", constraints

	// Default fallback
	default:
		return "string", "", fmt.Sprintf("String value (%s)", cpType), "sample value", constraints
	}
}

// generateModelPaths creates basic CRUD paths for a model
func generateModelPaths(paths map[string]interface{}, model *grammar.Model) {
	modelName := model.Name
	modelNameLower := strings.ToLower(modelName)
	modelNamePlural := modelNameLower + "s" // Simple pluralization

	// List endpoint: GET /users
	paths[fmt.Sprintf("/%s", modelNamePlural)] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     fmt.Sprintf("List all %s", modelNamePlural),
			"description": fmt.Sprintf("Retrieve a list of all %s records", modelNameLower),
			"tags":        []string{modelName},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
								},
							},
						},
					},
				},
			},
		},
		"post": map[string]interface{}{
			"summary":     fmt.Sprintf("Create a new %s", modelNameLower),
			"description": fmt.Sprintf("Create a new %s record", modelNameLower),
			"tags":        []string{modelName},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Created successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
							},
						},
					},
				},
			},
		},
	}

	// Individual resource endpoints: GET/PUT/DELETE /users/{id}
	paths[fmt.Sprintf("/%s/{id}", modelNamePlural)] = map[string]interface{}{
		"parameters": []interface{}{
			map[string]interface{}{
				"name":        "id",
				"in":          "path",
				"required":    true,
				"description": fmt.Sprintf("%s ID", modelName),
				"schema": map[string]interface{}{
					"type":   "string",
					"format": "uuid",
				},
			},
		},
		"get": map[string]interface{}{
			"summary":     fmt.Sprintf("Get a %s by ID", modelNameLower),
			"description": fmt.Sprintf("Retrieve a specific %s record", modelNameLower),
			"tags":        []string{modelName},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Record not found",
				},
			},
		},
		"put": map[string]interface{}{
			"summary":     fmt.Sprintf("Update a %s", modelNameLower),
			"description": fmt.Sprintf("Update an existing %s record", modelNameLower),
			"tags":        []string{modelName},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Updated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": fmt.Sprintf("#/components/schemas/%s", modelName),
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Record not found",
				},
			},
		},
		"delete": map[string]interface{}{
			"summary":     fmt.Sprintf("Delete a %s", modelNameLower),
			"description": fmt.Sprintf("Delete a %s record", modelNameLower),
			"tags":        []string{modelName},
			"responses": map[string]interface{}{
				"204": map[string]interface{}{
					"description": "Deleted successfully",
				},
				"404": map[string]interface{}{
					"description": "Record not found",
				},
			},
		},
	}
}

// WriteFile renders doc as YAML and writes it to the provided path with configuration
func WriteFile(file *grammar.File, path string) error {
	return WriteFileWithConfig(file, path, "cloudpact.yaml")
}

// WriteFileWithConfig allows specifying a custom config file path
func WriteFileWithConfig(file *grammar.File, path, configPath string) error {
	config, err := LoadAPIConfig(configPath)
	if err != nil {
		// Use defaults if config loading fails
		config = DefaultAPIConfig()
	}

	yaml, err := GenerateWithConfig(file, config)
	if err != nil {
		return err
	}

	if err := os.MkdirAll("generated/openapi", 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(yaml), 0644)
}

// toYAML encodes v into YAML with deterministic output
func toYAML(v interface{}, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var lines []string
		for _, k := range keys {
			vv := val[k]
			if isScalar(vv) {
				lines = append(lines, fmt.Sprintf("%s%s: %s", indentStr, k, formatScalar(vv)))
			} else {
				lines = append(lines, fmt.Sprintf("%s%s:", indentStr, k))
				lines = append(lines, toYAML(vv, indent+2))
			}
		}
		return strings.Join(lines, "\n")
	case []interface{}:
		var lines []string
		for _, item := range val {
			if isScalar(item) {
				lines = append(lines, fmt.Sprintf("%s- %s", indentStr, formatScalar(item)))
			} else {
				lines = append(lines, fmt.Sprintf("%s-", indentStr))
				lines = append(lines, toYAML(item, indent+2))
			}
		}
		return strings.Join(lines, "\n")
	default:
		return fmt.Sprintf("%s%s", indentStr, formatScalar(val))
	}
}

func isScalar(v interface{}) bool {
	switch v.(type) {
	case string, int, int64, float64, bool, nil:
		return true
	default:
		return false
	}
}

func formatScalar(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case nil:
		return "null"
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(val)
	}
}
