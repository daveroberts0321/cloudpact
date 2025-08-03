package tsgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0.0"
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        age:
          type: integer
      required: [id, age]
`
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(specPath, []byte(spec), 0644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := Generate(specPath); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	iface, err := os.ReadFile(filepath.Join(dir, "generated/ts/User.ts"))
	if err != nil {
		t.Fatalf("read interface: %v", err)
	}
	if !strings.Contains(string(iface), "export interface User") {
		t.Fatalf("interface not generated: %s", string(iface))
	}
	client, err := os.ReadFile(filepath.Join(dir, "generated/ts/client.ts"))
	if err != nil {
		t.Fatalf("read client: %v", err)
	}
	if !strings.Contains(string(client), "class APIClient") {
		t.Fatalf("client not generated: %s", string(client))
	}
}
