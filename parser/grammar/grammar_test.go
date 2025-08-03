package grammar

import (
	"reflect"
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
`
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	expected := &File{Models: []*Model{
		{Name: "User", Fields: []*Field{
			{Name: "id", Type: &Type{Name: "Int"}},
			{Name: "name", Type: &Type{Name: "String"}},
		}},
	}}
	if !reflect.DeepEqual(file, expected) {
		t.Fatalf("expected %#v, got %#v", expected, file)
	}
}
