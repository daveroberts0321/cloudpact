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

	fields := file.Models[0].Fields
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Name != "id" || fields[0].Type.Name != "Int" {
		t.Fatalf("unexpected first field %#v", fields[0])
	}
	if fields[1].Name != "name" || fields[1].Type.Name != "String" {
		t.Fatalf("unexpected second field %#v", fields[1])
	}
}

// Test parsing of a record definition with fields.
func TestParseRecordDefinition(t *testing.T) {
	src := `define record User
id: Int
name: String
`
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(file.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(file.Models))
	}
	model := file.Models[0]
	if len(model.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(model.Fields))
	}
	if model.Fields[0].Name != "id" || model.Fields[0].Type.Name != "Int" {
		t.Fatalf("unexpected first field: %#v", model.Fields[0])
	}
	if model.Fields[1].Name != "name" || model.Fields[1].Type.Name != "String" {
		t.Fatalf("unexpected second field: %#v", model.Fields[1])

	}
}
