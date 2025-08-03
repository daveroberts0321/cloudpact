package grammar

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

// Extended Grammar:
//   File        := { Declaration }
//   Declaration := Model | Function | Assignment
//   Model       := 'model' IDENT '{' { Field } '}'
//   Field       := IDENT ':' Type [ RelationshipDecl ]
//   Type        := IDENT [ TypeConstraints ]
//   Function    := 'function' IDENT '(' ParamList ')' [ 'returns' Type ] FunctionBody
//   FunctionBody:= WhyClause 'do:' StatementList
//   WhyClause   := 'why:' STRING
//   Assignment  := 'assign-use' IDENT 'as' Type [ 'why:' STRING ] [ 'validate:' ValidationRule ]
//   RelationshipDecl := 'belongs_to' | 'has_one' | 'has_many' | 'references'
//   NativeBlock := ('go-native:' | 'ts-native:') '```' LanguageCode '```'

// Position represents a source code position
type Position struct {
	Line   int
	Column int
	Offset int
}

// String returns a human-readable position
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// File represents a .cp file consisting of declarations
type File struct {
	Models      []*Model      `json:"models"`
	Functions   []*Function   `json:"functions"`
	Assignments []*Assignment `json:"assignments"`
}

// Model declares a new model with a name and a list of fields
type Model struct {
	Name     string    `json:"name"`
	Fields   []*Field  `json:"fields"`
	Position *Position `json:"position,omitempty"`
}

// Field represents a single field within a model with optional relationships
type Field struct {
	Name         string        `json:"name"`
	Type         *Type         `json:"type"`
	Relationship *Relationship `json:"relationship,omitempty"`
	Position     *Position     `json:"position,omitempty"`
}

// Type of a field with semantic information and constraints
type Type struct {
	Name        string                 `json:"name"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Position    *Position              `json:"position,omitempty"`
}

// Relationship defines model relationships (Django-style)
type Relationship struct {
	Kind     string    `json:"kind"`   // "belongs_to", "has_one", "has_many", "references"
	Target   string    `json:"target"` // Target model name
	Position *Position `json:"position,omitempty"`
}

// Function represents a CloudPact function with business context
type Function struct {
	Name       string        `json:"name"`
	Parameters []*Parameter  `json:"parameters"`
	ReturnType *Type         `json:"return_type,omitempty"`
	Why        string        `json:"why"` // Business explanation
	Body       *FunctionBody `json:"body"`
	Position   *Position     `json:"position,omitempty"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name     string    `json:"name"`
	Type     *Type     `json:"type"`
	Position *Position `json:"position,omitempty"`
}

// FunctionBody contains the function implementation
type FunctionBody struct {
	Statements   []Statement    `json:"statements"`
	NativeBlocks []*NativeBlock `json:"native_blocks,omitempty"`
	Position     *Position      `json:"position,omitempty"`
}

// Statement represents different types of statements in function bodies
type Statement interface {
	StatementType() string
	GetPosition() *Position
}

// IfStatement represents conditional logic
type IfStatement struct {
	Condition Expression  `json:"condition"`
	ThenBody  []Statement `json:"then_body"`
	ElseBody  []Statement `json:"else_body,omitempty"`
	Position  *Position   `json:"position,omitempty"`
}

func (s *IfStatement) StatementType() string  { return "if" }
func (s *IfStatement) GetPosition() *Position { return s.Position }

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Value    Expression `json:"value,omitempty"`
	Position *Position  `json:"position,omitempty"`
}

func (s *ReturnStatement) StatementType() string  { return "return" }
func (s *ReturnStatement) GetPosition() *Position { return s.Position }

// AssignStatement represents variable assignment
type AssignStatement struct {
	Variable string     `json:"variable"`
	Value    Expression `json:"value"`
	Position *Position  `json:"position,omitempty"`
}

func (s *AssignStatement) StatementType() string  { return "assign" }
func (s *AssignStatement) GetPosition() *Position { return s.Position }

// Expression represents different types of expressions
type Expression interface {
	ExpressionType() string
	GetPosition() *Position
}

// IdentifierExpression represents a variable or function reference
type IdentifierExpression struct {
	Name     string    `json:"name"`
	Position *Position `json:"position,omitempty"`
}

func (e *IdentifierExpression) ExpressionType() string { return "identifier" }
func (e *IdentifierExpression) GetPosition() *Position { return e.Position }

// LiteralExpression represents literal values
type LiteralExpression struct {
	Value    interface{} `json:"value"`
	Position *Position   `json:"position,omitempty"`
}

func (e *LiteralExpression) ExpressionType() string { return "literal" }
func (e *LiteralExpression) GetPosition() *Position { return e.Position }

// NativeBlock represents go-native: or ts-native: code blocks
type NativeBlock struct {
	Language string    `json:"language"` // "go" or "ts"
	Code     string    `json:"code"`
	Position *Position `json:"position,omitempty"`
}

// Assignment represents assign-use declarations for semantic types
type Assignment struct {
	TypeName   string                 `json:"type_name"`
	BaseType   *Type                  `json:"base_type"`
	Why        string                 `json:"why,omitempty"`
	Validation map[string]interface{} `json:"validation,omitempty"`
	Position   *Position              `json:"position,omitempty"`
}

// Parse reads CloudPact content from r and returns the parsed AST
func Parse(r io.Reader) (*File, error) {
	p := &parser{}
	p.scanner.Init(r)
	p.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats |
		scanner.ScanChars | scanner.ScanStrings | scanner.ScanComments
	p.next()
	return p.parseFile()
}

// ParseString parses a string containing CloudPact grammar into an AST
func ParseString(s string) (*File, error) {
	return Parse(strings.NewReader(s))
}

type parser struct {
	scanner scanner.Scanner
	tok     rune
}

func (p *parser) next() {
	p.tok = p.scanner.Scan()
}

func (p *parser) position() *Position {
	pos := p.scanner.Position
	return &Position{
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
	}
}

func (p *parser) parseFile() (*File, error) {
	file := &File{
		Models:      []*Model{},
		Functions:   []*Function{},
		Assignments: []*Assignment{},
	}

	for p.tok != scanner.EOF {
		if p.tok == scanner.EOF {
			break
		}

		switch {
		case p.tok == scanner.Ident && p.scanner.TokenText() == "model":
			model, err := p.parseModel()
			if err != nil {
				return nil, err
			}
			file.Models = append(file.Models, model)

		case p.tok == scanner.Ident && p.scanner.TokenText() == "function":
			function, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			file.Functions = append(file.Functions, function)

		case p.tok == scanner.Ident && p.scanner.TokenText() == "assign-use":
			assignment, err := p.parseAssignment()
			if err != nil {
				return nil, err
			}
			file.Assignments = append(file.Assignments, assignment)

		default:
			return nil, fmt.Errorf("unexpected token %q at %s", p.scanner.TokenText(), p.position())
		}
	}

	return file, nil
}

func (p *parser) expect(tok rune, expected string) error {
	if p.tok != tok {
		return fmt.Errorf("expected %s, got %q at %s", expected, p.scanner.TokenText(), p.position())
	}
	p.next()
	return nil
}

func (p *parser) expectKeyword(keyword string) error {
	if p.tok != scanner.Ident || p.scanner.TokenText() != keyword {
		return fmt.Errorf("expected '%s', got %q at %s", keyword, p.scanner.TokenText(), p.position())
	}
	p.next()
	return nil
}

func (p *parser) parseModel() (*Model, error) {
	pos := p.position()

	if err := p.expectKeyword("model"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected model name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	if err := p.expect('{', "'{'"); err != nil {
		return nil, err
	}

	model := &Model{
		Name:     name,
		Position: pos,
		Fields:   []*Field{},
	}

	for p.tok != '}' && p.tok != scanner.EOF {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		model.Fields = append(model.Fields, field)
	}

	if err := p.expect('}', "'}'"); err != nil {
		return nil, err
	}

	return model, nil
}

func (p *parser) parseField() (*Field, error) {
	pos := p.position()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected field name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	fieldType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	field := &Field{
		Name:     name,
		Type:     fieldType,
		Position: pos,
	}

	// Check for relationship declaration
	if p.tok == scanner.Ident {
		relationshipKind := p.scanner.TokenText()
		if isRelationshipKeyword(relationshipKind) {
			relationship, err := p.parseRelationship()
			if err != nil {
				return nil, err
			}
			field.Relationship = relationship
		}
	}

	return field, nil
}

func (p *parser) parseType() (*Type, error) {
	pos := p.position()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected type name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	typeName := p.scanner.TokenText()
	p.next()

	return &Type{
		Name:        typeName,
		Position:    pos,
		Constraints: make(map[string]interface{}),
	}, nil
}

func (p *parser) parseRelationship() (*Relationship, error) {
	pos := p.position()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected relationship keyword at %s", p.position())
	}

	kind := p.scanner.TokenText()
	if !isRelationshipKeyword(kind) {
		return nil, fmt.Errorf("invalid relationship type %q at %s", kind, p.position())
	}
	p.next()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected target model name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	target := p.scanner.TokenText()
	p.next()

	return &Relationship{
		Kind:     kind,
		Target:   target,
		Position: pos,
	}, nil
}

func (p *parser) parseFunction() (*Function, error) {
	pos := p.position()

	if err := p.expectKeyword("function"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected function name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	if err := p.expect('(', "'('"); err != nil {
		return nil, err
	}

	parameters, err := p.parseParameterList()
	if err != nil {
		return nil, err
	}

	if err := p.expect(')', "')'"); err != nil {
		return nil, err
	}

	function := &Function{
		Name:       name,
		Parameters: parameters,
		Position:   pos,
	}

	// Optional return type
	if p.tok == scanner.Ident && p.scanner.TokenText() == "returns" {
		p.next()
		returnType, err := p.parseType()
		if err != nil {
			return nil, err
		}
		function.ReturnType = returnType
	}

	// Parse why clause
	if err := p.expectKeyword("why"); err != nil {
		return nil, err
	}

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	if p.tok != scanner.String {
		return nil, fmt.Errorf("expected string after 'why:', got %q at %s", p.scanner.TokenText(), p.position())
	}

	why := strings.Trim(p.scanner.TokenText(), `"`)
	function.Why = why
	p.next()

	// Parse function body
	if err := p.expectKeyword("do"); err != nil {
		return nil, err
	}

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	body, err := p.parseFunctionBody()
	if err != nil {
		return nil, err
	}

	function.Body = body

	return function, nil
}

func (p *parser) parseParameterList() ([]*Parameter, error) {
	var parameters []*Parameter

	if p.tok == ')' {
		return parameters, nil // Empty parameter list
	}

	for {
		param, err := p.parseParameter()
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, param)

		if p.tok != ',' {
			break
		}
		p.next() // consume comma
	}

	return parameters, nil
}

func (p *parser) parseParameter() (*Parameter, error) {
	pos := p.position()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected parameter name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	paramType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	return &Parameter{
		Name:     name,
		Type:     paramType,
		Position: pos,
	}, nil
}

func (p *parser) parseFunctionBody() (*FunctionBody, error) {
	pos := p.position()

	body := &FunctionBody{
		Position:     pos,
		Statements:   []Statement{},
		NativeBlocks: []*NativeBlock{},
	}

	// For now, we'll implement basic statement parsing
	// This is a simplified version - you'd expand this for full CloudPact syntax
	for p.tok != scanner.EOF {
		// Check for native blocks
		if p.tok == scanner.Ident && (p.scanner.TokenText() == "go-native" || p.scanner.TokenText() == "ts-native") {
			nativeBlock, err := p.parseNativeBlock()
			if err != nil {
				return nil, err
			}
			body.NativeBlocks = append(body.NativeBlocks, nativeBlock)
			continue
		}

		// For now, skip to next model/function declaration
		if p.tok == scanner.Ident && (p.scanner.TokenText() == "model" || p.scanner.TokenText() == "function" || p.scanner.TokenText() == "assign-use") {
			break
		}

		// Simple statement parsing - extend this for full syntax
		p.next()
	}

	return body, nil
}

func (p *parser) parseNativeBlock() (*NativeBlock, error) {
	pos := p.position()

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected native block type at %s", p.position())
	}

	blockType := p.scanner.TokenText()
	var language string

	switch blockType {
	case "go-native":
		language = "go"
	case "ts-native":
		language = "ts"
	default:
		return nil, fmt.Errorf("invalid native block type %q at %s", blockType, p.position())
	}
	p.next()

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	// For now, we'll expect the native code as a string
	// In a full implementation, you'd parse the ``` delimited code blocks
	if p.tok != scanner.String {
		return nil, fmt.Errorf("expected native code string at %s", p.position())
	}

	code := strings.Trim(p.scanner.TokenText(), `"`)
	p.next()

	return &NativeBlock{
		Language: language,
		Code:     code,
		Position: pos,
	}, nil
}

func (p *parser) parseAssignment() (*Assignment, error) {
	pos := p.position()

	if err := p.expectKeyword("assign-use"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected type name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	typeName := p.scanner.TokenText()
	p.next()

	if err := p.expectKeyword("as"); err != nil {
		return nil, err
	}

	baseType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	assignment := &Assignment{
		TypeName:   typeName,
		BaseType:   baseType,
		Position:   pos,
		Validation: make(map[string]interface{}),
	}

	// Optional why clause
	if p.tok == scanner.Ident && p.scanner.TokenText() == "why" {
		p.next()
		if err := p.expect(':', "':'"); err != nil {
			return nil, err
		}
		if p.tok != scanner.String {
			return nil, fmt.Errorf("expected string after 'why:', got %q at %s", p.scanner.TokenText(), p.position())
		}
		assignment.Why = strings.Trim(p.scanner.TokenText(), `"`)
		p.next()
	}

	// Optional validate clause (simplified)
	if p.tok == scanner.Ident && p.scanner.TokenText() == "validate" {
		p.next()
		if err := p.expect(':', "':'"); err != nil {
			return nil, err
		}
		// For now, we'll store validation as a string
		// You'd extend this to parse actual validation rules
		if p.tok == scanner.String {
			assignment.Validation["rule"] = strings.Trim(p.scanner.TokenText(), `"`)
			p.next()
		}
	}

	return assignment, nil
}

// Helper functions
func isRelationshipKeyword(keyword string) bool {
	relationships := []string{"belongs_to", "has_one", "has_many", "references"}
	for _, rel := range relationships {
		if keyword == rel {
			return true
		}
	}
	return false
}
