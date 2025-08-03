package tsgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// parseSchemas extracts schema definitions from the limited OpenAPI YAML.
func parseSchemas(yaml string) (map[string]map[string]string, error) {
	lines := strings.Split(yaml, "\n")
	schemas := map[string]map[string]string{}
	state := 0
	var currentModel string
	var currentField string
	inProperties := false
	for _, line := range lines {
		line = strings.TrimRight(line, " ")
		switch state {
		case 0:
			if strings.HasPrefix(line, "components:") {
				state = 1
			}
		case 1:
			if strings.HasPrefix(line, "  schemas:") {
				state = 2
			}
		case 2:
			// look for model declarations at indent 4
			if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
				trimmed := strings.TrimSpace(line)
				if strings.HasSuffix(trimmed, ":") {
					currentModel = strings.TrimSuffix(trimmed, ":")
					schemas[currentModel] = map[string]string{}
					inProperties = false
					currentField = ""
					continue
				}
			}
			if currentModel == "" {
				continue
			}
			if !inProperties {
				if strings.HasPrefix(line, "      properties:") {
					inProperties = true
				}
				continue
			}
			// inside properties
			if strings.HasPrefix(line, "        ") && !strings.HasPrefix(line, "          ") {
				trimmed := strings.TrimSpace(line)
				if strings.HasSuffix(trimmed, ":") {
					currentField = strings.TrimSuffix(trimmed, ":")
					continue
				}
			}
			if currentField != "" && strings.Contains(line, "type:") {
				t := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "type:"))
				t = strings.Trim(t, "\"")
				schemas[currentModel][currentField] = t
				currentField = ""
				continue
			}
			// leaving properties block
			if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
				inProperties = false
				currentField = ""
			}
		}
	}
	return schemas, nil
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
	switch t {
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	default:
		return "string"
	}
}
