package grammar

import "testing"

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
`
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(file.Models))
	}
	model := file.Models[0]
	if model.Name != "User" {
		t.Fatalf("expected model name 'User', got %q", model.Name)
	}
	expectedFields := []struct {
		name string
		typ  string
	}{
		{"id", "Int"},
		{"name", "String"},
	}
	if len(model.Fields) != len(expectedFields) {
		t.Fatalf("expected %d fields, got %d", len(expectedFields), len(model.Fields))
	}
	for i, ef := range expectedFields {
		f := model.Fields[i]
		if f.Name != ef.name || f.Type == nil || f.Type.Name != ef.typ {
			t.Fatalf("expected field %q of type %q, got %#v", ef.name, ef.typ, f)
		}
	}
}
