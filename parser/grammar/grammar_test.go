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
	if len(file.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(file.Records))
	}
	rec := file.Records[0]
	if rec.Name != "User" {
		t.Fatalf("unexpected record name %q", rec.Name)
	}
	if len(rec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(rec.Fields))
	}
	if rec.Fields[0].Name != "id" || rec.Fields[0].Type.Name != "Int" {
		t.Fatalf("unexpected first field %#v", rec.Fields[0])
	}
	if rec.Fields[1].Name != "name" || rec.Fields[1].Type.Name != "String" {
		t.Fatalf("unexpected second field %#v", rec.Fields[1])
	}
}

// Test parsing of function definitions with control-flow statements.
func TestParseFunctionDefinitionWithControlFlow(t *testing.T) {
	src := `function check(age: Int) returns Int why: "age check" do:
if age > 18 then set result = 1 else set result = 0
return result`
	file, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(file.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(file.Functions))
	}
	fn := file.Functions[0]
	if fn.Name != "check" {
		t.Fatalf("expected function name 'check', got %q", fn.Name)
	}
	if fn.Body == nil || len(fn.Body.Statements) != 2 {
		t.Fatalf("expected 2 statements in body, got %#v", fn.Body)
	}
	ifStmt, ok := fn.Body.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("expected first statement to be IfStatement, got %T", fn.Body.Statements[0])
	}
	cond, ok := ifStmt.Condition.(*BinaryExpression)
	if !ok || cond.Operator != ">" {
		t.Fatalf("unexpected condition %#v", ifStmt.Condition)
	}
	left, ok := cond.Left.(*IdentifierExpression)
	if !ok || left.Name != "age" {
		t.Fatalf("unexpected left side %#v", cond.Left)
	}
	right, ok := cond.Right.(*LiteralExpression)
	if !ok || right.Value != "18" {
		t.Fatalf("unexpected right side %#v", cond.Right)
	}
	if _, ok := ifStmt.ThenStmt.(*AssignStatement); !ok {
		t.Fatalf("expected then branch to be assignment, got %T", ifStmt.ThenStmt)
	}
	if _, ok := ifStmt.ElseStmt.(*AssignStatement); !ok {
		t.Fatalf("expected else branch to be assignment, got %T", ifStmt.ElseStmt)
	}
	if _, ok := fn.Body.Statements[1].(*ReturnStatement); !ok {
		t.Fatalf("expected second statement to be ReturnStatement, got %T", fn.Body.Statements[1])
	}
}

// Test parsing of assign-use declarations (currently expected to fail).
func TestParseAssignUse(t *testing.T) {
	src := `assign-use Email as String`
	if _, err := ParseString(src); err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

// Error case: record field missing type should fail.
func TestParseRecordDefinitionError(t *testing.T) {
	src := `define record User
id Int`
	if _, err := ParseString(src); err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

// Error case: function missing why clause should fail.
func TestParseFunctionMissingWhy(t *testing.T) {
	src := `function bad() returns Int do:
return 0`
	if _, err := ParseString(src); err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

// Error case: if statement missing 'then' should fail.
func TestParseIfStatementMissingThen(t *testing.T) {
	src := `function f() returns Int why: "test" do:
if x > 0 set y = 1`
	if _, err := ParseString(src); err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}
