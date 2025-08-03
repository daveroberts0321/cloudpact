// Package grammar implements the CloudPact language parser.
// parser.go contains the parsing logic and helper routines.
package grammar

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

// Enhanced CloudPact Grammar:
//   File            := ModuleDecl { Declaration }
//   ModuleDecl      := 'module' IDENT
//   Declaration     := RecordDef | FunctionDef | TypeDef | Model | Assignment
//   RecordDef       := 'define' 'record' IDENT { FieldDef }
//   FieldDef        := IDENT ':' Type
//   FunctionDef     := 'function' IDENT '(' ParamList ')' [ 'returns' Type ] AIAnnotations WhyClause DoBlock
//   DoBlock         := 'do:' { Statement }
//   Statement       := IfStatement | Assignment | Return | CreateStatement | Expression
//   IfStatement     := 'if' Expression 'then' Statement [ 'else' Statement ]
//   CreateStatement := 'create' IDENT 'with:' { FieldAssignment }
//   AIAnnotation    := ('ai-feedback:' | 'ai-suggests:' | 'ai-security:' | 'ai-performance:') STRING
//
//   // Legacy support for existing models
//   Model           := 'model' IDENT '{' { Field } '}'
//   Field           := IDENT ':' Type [ RelationshipDecl ]

// Enhanced Parser
type parser struct {
	scanner  scanner.Scanner
	tok      rune
	filename string
}

// Parse reads CloudPact content from r and returns the parsed AST
func Parse(r io.Reader) (*File, error) {
	return ParseWithFilename(r, "")
}

// ParseWithFilename allows tracking source file for better error messages
func ParseWithFilename(r io.Reader, filename string) (*File, error) {
	p := &parser{filename: filename}
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

func (p *parser) next() {
	p.tok = p.scanner.Scan()
}

func (p *parser) position() *Position {
	pos := p.scanner.Position
	return &Position{
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
		File:   p.filename,
	}
}

func (p *parser) parseFile() (*File, error) {
	file := &File{
		Records:     []*Record{},
		Models:      []*Model{},
		Functions:   []*Function{},
		TypeDefs:    []*TypeDef{},
		Assignments: []*Assignment{},
		Position:    p.position(),
	}

	// Parse optional module declaration
	if p.tok == scanner.Ident && p.scanner.TokenText() == "module" {
		module, err := p.parseModule()
		if err != nil {
			return nil, err
		}
		file.Module = module
	}

	// Parse declarations
	for p.tok != scanner.EOF {
		switch {
		case p.tok == scanner.Ident && p.scanner.TokenText() == "define":
			if err := p.parseDefine(file); err != nil {
				return nil, err
			}

		case p.tok == scanner.Ident && p.scanner.TokenText() == "function":
			function, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			file.Functions = append(file.Functions, function)

		case p.tok == scanner.Ident && p.scanner.TokenText() == "model":
			model, err := p.parseModel()
			if err != nil {
				return nil, err
			}
			file.Models = append(file.Models, model)

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

func (p *parser) parseModule() (*Module, error) {
	pos := p.position()

	if err := p.expectKeyword("module"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected module name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	return &Module{
		Name:     name,
		Position: pos,
	}, nil
}

func (p *parser) parseDefine(file *File) error {
	if err := p.expectKeyword("define"); err != nil {
		return err
	}

	if p.tok != scanner.Ident {
		return fmt.Errorf("expected 'record' or 'type' after 'define', got %q at %s", p.scanner.TokenText(), p.position())
	}

	switch p.scanner.TokenText() {
	case "record":
		record, err := p.parseRecord()
		if err != nil {
			return err
		}
		file.Records = append(file.Records, record)
	case "type":
		typeDef, err := p.parseTypeDef()
		if err != nil {
			return err
		}
		file.TypeDefs = append(file.TypeDefs, typeDef)
	default:
		return fmt.Errorf("expected 'record' or 'type' after 'define', got %q at %s", p.scanner.TokenText(), p.position())
	}

	return nil
}

func (p *parser) parseRecord() (*Record, error) {
	pos := p.position()

	if err := p.expectKeyword("record"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected record name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	record := &Record{
		Name:     name,
		Position: pos,
		Fields:   []*FieldDef{},
	}

	// Parse fields until we hit a keyword that starts a new declaration
	for p.tok == scanner.Ident && !isTopLevelKeyword(p.scanner.TokenText()) {
		field, err := p.parseFieldDef()
		if err != nil {
			return nil, err
		}
		record.Fields = append(record.Fields, field)
	}

	return record, nil
}

func (p *parser) parseFieldDef() (*FieldDef, error) {
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

	return &FieldDef{
		Name:     name,
		Type:     fieldType,
		Position: pos,
	}, nil
}

func (p *parser) parseTypeDef() (*TypeDef, error) {
	pos := p.position()

	if err := p.expectKeyword("type"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected type name, got %q at %s", p.scanner.TokenText(), p.position())
	}

	name := p.scanner.TokenText()
	p.next()

	if err := p.expectKeyword("as"); err != nil {
		return nil, err
	}

	baseType, err := p.parseType()
	if err != nil {
		return nil, err
	}

	typeDef := &TypeDef{
		Name:       name,
		BaseType:   baseType,
		Position:   pos,
		Validation: make(map[string]interface{}),
	}

	// Parse optional why and validation clauses
	for p.tok == scanner.Ident {
		switch p.scanner.TokenText() {
		case "why":
			p.next()
			if err := p.expect(':', "':'"); err != nil {
				return nil, err
			}
			if p.tok != scanner.String {
				return nil, fmt.Errorf("expected string after 'why:', got %q at %s", p.scanner.TokenText(), p.position())
			}
			typeDef.Why = strings.Trim(p.scanner.TokenText(), `"`)
			p.next()
		case "validate":
			p.next()
			if err := p.expect(':', "':'"); err != nil {
				return nil, err
			}
			if p.tok == scanner.String {
				typeDef.Validation["rule"] = strings.Trim(p.scanner.TokenText(), `"`)
				p.next()
			}
		default:
			// Not a type definition clause, break out
			return typeDef, nil
		}
	}

	return typeDef, nil
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
		Name:          name,
		Parameters:    parameters,
		Position:      pos,
		AIAnnotations: []*AIAnnotation{},
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

	// Parse AI annotations
	for p.tok == scanner.Ident && isAIAnnotation(p.scanner.TokenText()) {
		annotation, err := p.parseAIAnnotation()
		if err != nil {
			return nil, err
		}
		function.AIAnnotations = append(function.AIAnnotations, annotation)
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

	function.Why = strings.Trim(p.scanner.TokenText(), `"`)
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

func (p *parser) parseAIAnnotation() (*AIAnnotation, error) {
	pos := p.position()

	if !isAIAnnotation(p.scanner.TokenText()) {
		return nil, fmt.Errorf("expected AI annotation, got %q at %s", p.scanner.TokenText(), p.position())
	}

	annotationType := strings.TrimPrefix(p.scanner.TokenText(), "ai-")
	annotationType = strings.TrimSuffix(annotationType, ":")
	p.next()

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	if p.tok != scanner.String {
		return nil, fmt.Errorf("expected string after AI annotation, got %q at %s", p.scanner.TokenText(), p.position())
	}

	content := strings.Trim(p.scanner.TokenText(), `"`)
	p.next()

	return &AIAnnotation{
		Type:     annotationType,
		Content:  content,
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

	// Parse statements until we hit EOF or a top-level keyword
	for p.tok != scanner.EOF && !(p.tok == scanner.Ident && isTopLevelKeyword(p.scanner.TokenText())) {
		// Check for native blocks
		if p.tok == scanner.Ident && (p.scanner.TokenText() == "go-native" || p.scanner.TokenText() == "ts-native") {
			nativeBlock, err := p.parseNativeBlock()
			if err != nil {
				return nil, err
			}
			body.NativeBlocks = append(body.NativeBlocks, nativeBlock)
			continue
		}

		// Parse regular statements
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body.Statements = append(body.Statements, stmt)
		}
	}

	return body, nil
}

func (p *parser) parseStatement() (Statement, error) {
	switch {
	case p.tok == scanner.Ident && p.scanner.TokenText() == "if":
		return p.parseIfStatement()
	case p.tok == scanner.Ident && p.scanner.TokenText() == "return":
		return p.parseReturnStatement()
	case p.tok == scanner.Ident && p.scanner.TokenText() == "set":
		return p.parseSetStatement()
	case p.tok == scanner.Ident && p.scanner.TokenText() == "create":
		return p.parseCreateStatement()
	case p.tok == scanner.Ident && p.scanner.TokenText() == "fail":
		return p.parseFailStatement()
	case p.tok == scanner.Ident && p.scanner.TokenText() == "use":
		// Handle "use SHA256 algorithm" style statements
		return p.parseUseStatement()
	default:
		// Skip unknown tokens for now - could be expression statements
		p.next()
		return nil, nil
	}
}

func (p *parser) parseIfStatement() (*IfStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("if"); err != nil {
		return nil, err
	}

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.expectKeyword("then"); err != nil {
		return nil, err
	}

	thenStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	ifStmt := &IfStatement{
		Condition: condition,
		ThenStmt:  thenStmt,
		Position:  pos,
	}

	// Optional else clause
	if p.tok == scanner.Ident && p.scanner.TokenText() == "else" {
		p.next()
		elseStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		ifStmt.ElseStmt = elseStmt
	}

	return ifStmt, nil
}

func (p *parser) parseReturnStatement() (*ReturnStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("return"); err != nil {
		return nil, err
	}

	// Optional return value
	var value Expression
	if p.tok != scanner.EOF && !(p.tok == scanner.Ident && isStatementKeyword(p.scanner.TokenText())) {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return &ReturnStatement{
		Value:    value,
		Position: pos,
	}, nil
}

func (p *parser) parseSetStatement() (*AssignStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("set"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected variable name after 'set', got %q at %s", p.scanner.TokenText(), p.position())
	}

	variable := p.scanner.TokenText()
	p.next()

	if err := p.expect('=', "'='"); err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &AssignStatement{
		Variable: variable,
		Value:    value,
		Position: pos,
	}, nil
}

func (p *parser) parseCreateStatement() (*CreateStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("create"); err != nil {
		return nil, err
	}

	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected type name after 'create', got %q at %s", p.scanner.TokenText(), p.position())
	}

	typeName := p.scanner.TokenText()
	p.next()

	if err := p.expectKeyword("with"); err != nil {
		return nil, err
	}

	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}

	var assignments []*FieldAssignment

	// Parse field assignments
	for p.tok == scanner.Ident && !isStatementKeyword(p.scanner.TokenText()) && !isTopLevelKeyword(p.scanner.TokenText()) {
		fieldPos := p.position()
		field := p.scanner.TokenText()
		p.next()

		if err := p.expect('=', "'='"); err != nil {
			return nil, err
		}

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, &FieldAssignment{
			Field:    field,
			Value:    value,
			Position: fieldPos,
		})
	}

	return &CreateStatement{
		TypeName:    typeName,
		Assignments: assignments,
		Position:    pos,
	}, nil
}

func (p *parser) parseFailStatement() (*FailStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("fail"); err != nil {
		return nil, err
	}

	if p.tok != scanner.String {
		return nil, fmt.Errorf("expected error message string after 'fail', got %q at %s", p.scanner.TokenText(), p.position())
	}

	message := strings.Trim(p.scanner.TokenText(), `"`)
	p.next()

	return &FailStatement{
		Message:  message,
		Position: pos,
	}, nil
}

func (p *parser) parseUseStatement() (*AssignStatement, error) {
	pos := p.position()

	if err := p.expectKeyword("use"); err != nil {
		return nil, err
	}

	// Parse the rest as a simple expression for now
	// "use SHA256 algorithm" becomes an assignment
	var parts []string
	for p.tok == scanner.Ident {
		parts = append(parts, p.scanner.TokenText())
		p.next()
	}

	useExpr := strings.Join(parts, " ")

	return &AssignStatement{
		Variable: "__use__",
		Value: &LiteralExpression{
			Value:    useExpr,
			Position: pos,
		},
		Position: pos,
	}, nil
}

func (p *parser) parseExpression() (Expression, error) {
	return p.parseComparison()
}

func (p *parser) parseComparison() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Handle comparison operators
	for p.tok == '<' || p.tok == '>' || p.tok == '=' ||
		(p.tok == scanner.Ident && (p.scanner.TokenText() == "contains" || p.scanner.TokenText() == "not")) {

		var operator string
		if p.tok == scanner.Ident {
			if p.scanner.TokenText() == "not" {
				p.next()
				if p.tok == scanner.Ident && p.scanner.TokenText() == "contains" {
					operator = "not contains"
					p.next()
				} else {
					operator = "not"
				}
			} else {
				operator = p.scanner.TokenText()
				p.next()
			}
		} else {
			operator = string(rune(p.tok))
			p.next()
		}

		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
			Position: left.GetPosition(),
		}
	}

	return left, nil
}

func (p *parser) parsePrimary() (Expression, error) {
	pos := p.position()

	switch p.tok {
	case scanner.Ident:
		name := p.scanner.TokenText()
		p.next()

		// Check for member access (user.email)
		if p.tok == '.' {
			p.next()
			if p.tok != scanner.Ident {
				return nil, fmt.Errorf("expected property name after '.', got %q at %s", p.scanner.TokenText(), p.position())
			}
			property := p.scanner.TokenText()
			p.next()
			return &MemberExpression{
				Object: &IdentifierExpression{
					Name:     name,
					Position: pos,
				},
				Property: property,
				Position: pos,
			}, nil
		}

		// Check for function call (functionName())
		if p.tok == '(' {
			p.next()
			var args []Expression

			// Parse arguments
			if p.tok != ')' {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)

					if p.tok != ',' {
						break
					}
					p.next() // consume comma
				}
			}

			if err := p.expect(')', "')'"); err != nil {
				return nil, err
			}

			return &CallExpression{
				Function:  name,
				Arguments: args,
				Position:  pos,
			}, nil
		}

		// Simple identifier
		return &IdentifierExpression{
			Name:     name,
			Position: pos,
		}, nil

	case scanner.String:
		value := strings.Trim(p.scanner.TokenText(), `"`)
		p.next()
		return &LiteralExpression{
			Value:    value,
			Position: pos,
		}, nil

	case scanner.Int:
		value := p.scanner.TokenText()
		p.next()
		// Convert to int - simplified for now
		return &LiteralExpression{
			Value:    value,
			Position: pos,
		}, nil

	case scanner.Float:
		value := p.scanner.TokenText()
		p.next()
		return &LiteralExpression{
			Value:    value,
			Position: pos,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression: %q at %s", p.scanner.TokenText(), p.position())
	}
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

// Legacy parser methods for backward compatibility
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

// Utility functions
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

// Helper functions for keyword recognition
func isTopLevelKeyword(keyword string) bool {
	topLevel := []string{"module", "define", "function", "model", "assign-use"}
	for _, kw := range topLevel {
		if keyword == kw {
			return true
		}
	}
	return false
}

func isStatementKeyword(keyword string) bool {
	statements := []string{"if", "return", "set", "create", "fail", "use", "for", "while"}
	for _, kw := range statements {
		if keyword == kw {
			return true
		}
	}
	return false
}

func isAIAnnotation(keyword string) bool {
	annotations := []string{"ai-feedback", "ai-suggests", "ai-security", "ai-performance", "ai-decision-accepted", "ai-decision-rejected"}
	for _, ann := range annotations {
		if keyword == ann || keyword == ann+":" {
			return true
		}
	}
	return false
}

func isRelationshipKeyword(keyword string) bool {
	relationships := []string{"belongs_to", "has_one", "has_many", "references"}
	for _, rel := range relationships {
		if keyword == rel {
			return true
		}
	}
	return false
}
