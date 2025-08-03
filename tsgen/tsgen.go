package tsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Generate reads an OpenAPI spec in YAML format and emits TypeScript
// interfaces and a simple API client stub under generated/ts/.
// The parser understands the limited YAML subset produced by the openapi
// package in this repository.
func Generate(specPath string) error {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	schemas, err := parseSchemas(string(data))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join("generated", "ts"), 0755); err != nil {
		return err
	}
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := writeInterface(name, schemas[name]); err != nil {
			return err
		}
	}
	return writeClient(names)
}

// parseSchemas extracts schema definitions from the OpenAPI YAML using a YAML parser.
func parseSchemas(data string) (map[string]map[string]string, error) {
	var doc struct {
		Components struct {
			Schemas map[string]*schema `yaml:"schemas"`
		} `yaml:"components"`
	}

	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	for name, s := range doc.Components.Schemas {
		fields := make(map[string]string)
		for fname, f := range s.Properties {
			fields[fname] = resolveType(f)
		}
		result[name] = fields
	}
	return result, nil
}

type schema struct {
	Type       string             `yaml:"type"`
	Ref        string             `yaml:"$ref"`
	Properties map[string]*schema `yaml:"properties"`
	Items      *schema            `yaml:"items"`
}

func resolveType(s *schema) string {
	if s == nil {
		return "any"
	}
	if s.Ref != "" {
		parts := strings.Split(s.Ref, "/")
		return parts[len(parts)-1]
	}
	switch s.Type {
	case "array":
		return resolveType(s.Items) + "[]"
	case "object":
		return "object"
	default:
		return s.Type
	}
}

func writeInterface(name string, fields map[string]string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "export interface %s {\n", name)
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "  %s: %s;\n", k, mapType(fields[k]))
	}
	b.WriteString("}\n")
	file := filepath.Join("generated", "ts", fmt.Sprintf("%s.ts", name))
	return os.WriteFile(file, []byte(b.String()), 0644)
}

func writeClient(names []string) error {
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintf(&b, "import { %s } from \"./%s\";\n", n, n)
	}
	b.WriteString("\nexport class APIClient {\n  constructor(private baseUrl: string) {}\n")
	for _, n := range names {
		lower := strings.ToLower(n)
		fmt.Fprintf(&b, "  async get%s(id: string): Promise<%s> {\n", n, n)
		fmt.Fprintf(&b, "    const res = await fetch(`${this.baseUrl}/%s/${id}`);\n", lower)
		b.WriteString("    if (!res.ok) {\n      throw new Error(res.statusText);\n    }\n")
		b.WriteString("    return res.json();\n  }\n")
	}
	b.WriteString("}\n")
	file := filepath.Join("generated", "ts", "client.ts")
	return os.WriteFile(file, []byte(b.String()), 0644)
}

func mapType(t string) string {
	if strings.HasSuffix(t, "[]") {
		return mapType(strings.TrimSuffix(t, "[]")) + "[]"
	}
	switch t {
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "string":
		return "string"
	case "object", "any":
		return "any"
	default:
		return t
	}
}
