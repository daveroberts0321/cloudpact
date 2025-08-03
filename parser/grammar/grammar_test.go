package grammar

import (
	"testing"
)

// Test parsing of a model declaration without fields.
func TestParseModelDeclaration(t *testing.T) {
	src := `model User {}`
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(file.Models))
	}
	if file.Models[0].Name != "User" {
		t.Fatalf("expected model name 'User', got %q", file.Models[0].Name)
	}
}

// Test parsing of fields and types within a model.
func TestParseFieldsAndTypes(t *testing.T) {
	src := `
model User {
    id: Int
    name: String
}
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(file.Models))
	}

	m := file.Models[0]
	if m.Name != "User" {
		t.Fatalf("expected model name 'User', got %q", m.Name)
	}
	if len(m.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(m.Fields))
	}
	if m.Fields[0].Name != "id" || m.Fields[0].Type.Name != "Int" {
		t.Fatalf("unexpected first field: %#v", m.Fields[0])
	}
	if m.Fields[1].Name != "name" || m.Fields[1].Type.Name != "String" {
		t.Fatalf("unexpected second field: %#v", m.Fields[1])
	}
}
