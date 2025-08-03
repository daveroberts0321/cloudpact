package openapi

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"cloudpact/parser/grammar"
)

// Generate converts a parsed .cf AST into an OpenAPI document represented
// as a YAML string. The resulting YAML contains schema definitions for all
// models defined in the input file.
func Generate(file *grammar.File) (string, error) {
	if file == nil {
		return "", fmt.Errorf("nil file")
	}

	doc := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Cloudpact API",
			"version": "1.0.0",
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{},
		},
	}

	schemas := doc["components"].(map[string]interface{})["schemas"].(map[string]interface{})

	for _, m := range file.Models {
		modelSchema := map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []interface{}{},
		}
		props := modelSchema["properties"].(map[string]interface{})
		req := modelSchema["required"].([]interface{})
		for _, f := range m.Fields {
			t, format := mapType(f.Type.Name)
			fieldSchema := map[string]interface{}{"type": t}
			if format != "" {
				fieldSchema["format"] = format
			}
			props[f.Name] = fieldSchema
			req = append(req, f.Name)
		}
		modelSchema["required"] = req
		schemas[m.Name] = modelSchema
	}

	return toYAML(doc, 0), nil
}

// mapType maps CF types to OpenAPI types and optional formats.
func mapType(t string) (string, string) {
	switch strings.ToLower(t) {
	case "int", "integer":
		return "integer", ""
	case "float", "double", "number":
		return "number", "float"
	case "bool", "boolean":
		return "boolean", ""
	case "id", "uuid":
		return "string", "uuid"
	default:
		return "string", ""
	}
}

// WriteFile renders doc as YAML and writes it to the provided path.
func WriteFile(file *grammar.File, path string) error {
	yaml, err := Generate(file)
	if err != nil {
		return err
	}
	if err := os.MkdirAll("generated/openapi", 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(yaml), 0644)
}

// toYAML encodes v into YAML. Only the patterns used by the spec generator
// are supported. It sorts map keys for deterministic output.
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
