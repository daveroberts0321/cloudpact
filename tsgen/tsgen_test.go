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
    Address:
      type: object
      properties:
        street:
          type: string
    User:
      type: object
      properties:
        id:
          type: string
        age:
          type: integer
        address:
          $ref: '#/components/schemas/Address'
        tags:
          type: array
          items:
            type: string
        addresses:
          type: array
          items:
            $ref: '#/components/schemas/Address'
      required: [id, age, address, tags, addresses]
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
	userIface, err := os.ReadFile(filepath.Join(dir, "generated/ts/User.ts"))
	if err != nil {
		t.Fatalf("read user interface: %v", err)
	}
	if !strings.Contains(string(userIface), "address: Address;") {
		t.Fatalf("nested object not generated: %s", string(userIface))
	}
	if !strings.Contains(string(userIface), "tags: string[];") {
		t.Fatalf("array of primitives not generated: %s", string(userIface))
	}
	if !strings.Contains(string(userIface), "addresses: Address[];") {
		t.Fatalf("array of objects not generated: %s", string(userIface))
	}
	addrIface, err := os.ReadFile(filepath.Join(dir, "generated/ts/Address.ts"))
	if err != nil {
		t.Fatalf("read address interface: %v", err)
	}
	if !strings.Contains(string(addrIface), "export interface Address") {
		t.Fatalf("address interface not generated: %s", string(addrIface))
	}
	client, err := os.ReadFile(filepath.Join(dir, "generated/ts/client.ts"))
	if err != nil {
		t.Fatalf("read client: %v", err)
	}
	if !strings.Contains(string(client), "class APIClient") {
		t.Fatalf("client not generated: %s", string(client))
	}
}
