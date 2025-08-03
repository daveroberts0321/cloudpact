// Package grammar implements the CloudPact language parser.
// ast.go defines the core AST structures representing CloudPact programs.
package grammar

import "fmt"

// Enhanced Position with more context
type Position struct {
	Line   int
	Column int
	Offset int
	File   string
}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// Enhanced File with module support
type File struct {
	Module      *Module       `json:"module,omitempty"`
	Records     []*Record     `json:"records"`
	Models      []*Model      `json:"models"` // Legacy support
	Functions   []*Function   `json:"functions"`
	TypeDefs    []*TypeDef    `json:"type_defs"`
	Assignments []*Assignment `json:"assignments"` // Legacy support
	Position    *Position     `json:"position,omitempty"`
}

// Module declaration
type Module struct {
	Name     string    `json:"name"`
	Position *Position `json:"position,omitempty"`
}

// Record definition (new syntax)
type Record struct {
	Name     string      `json:"name"`
	Fields   []*FieldDef `json:"fields"`
	Position *Position   `json:"position,omitempty"`
}

// FieldDef for new record syntax
type FieldDef struct {
	Name     string    `json:"name"`
	Type     *Type     `json:"type"`
	Position *Position `json:"position,omitempty"`
}

// TypeDef for custom type definitions
type TypeDef struct {
	Name       string                 `json:"name"`
	BaseType   *Type                  `json:"base_type"`
	Validation map[string]interface{} `json:"validation,omitempty"`
	Why        string                 `json:"why,omitempty"`
	Position   *Position              `json:"position,omitempty"`
}

// Enhanced Function with AI annotations
type Function struct {
	Name          string          `json:"name"`
	Parameters    []*Parameter    `json:"parameters"`
	ReturnType    *Type           `json:"return_type,omitempty"`
	Why           string          `json:"why"`
	AIAnnotations []*AIAnnotation `json:"ai_annotations,omitempty"`
	Body          *FunctionBody   `json:"body"`
	Position      *Position       `json:"position,omitempty"`
}

// AI Annotations for collaborative programming
type AIAnnotation struct {
	Type     string    `json:"type"` // "feedback", "suggests", "security", "performance"
	Content  string    `json:"content"`
	Position *Position `json:"position,omitempty"`
}

// Enhanced FunctionBody with rich statements
type FunctionBody struct {
	Statements   []Statement    `json:"statements"`
	NativeBlocks []*NativeBlock `json:"native_blocks,omitempty"`
	Position     *Position      `json:"position,omitempty"`
}

// Statement interface for all statement types
type Statement interface {
	StatementType() string
	GetPosition() *Position
}

// IfStatement for conditional logic
type IfStatement struct {
	Condition Expression `json:"condition"`
	ThenStmt  Statement  `json:"then_stmt"`
	ElseStmt  Statement  `json:"else_stmt,omitempty"`
	Position  *Position  `json:"position,omitempty"`
}

func (s *IfStatement) StatementType() string  { return "if" }
func (s *IfStatement) GetPosition() *Position { return s.Position }

// ReturnStatement
type ReturnStatement struct {
	Value    Expression `json:"value,omitempty"`
	Position *Position  `json:"position,omitempty"`
}

func (s *ReturnStatement) StatementType() string  { return "return" }
func (s *ReturnStatement) GetPosition() *Position { return s.Position }

// AssignStatement for variable assignments
type AssignStatement struct {
	Variable string     `json:"variable"`
	Value    Expression `json:"value"`
	Position *Position  `json:"position,omitempty"`
}

func (s *AssignStatement) StatementType() string  { return "assign" }
func (s *AssignStatement) GetPosition() *Position { return s.Position }

// CreateStatement for "create user with:" syntax
type CreateStatement struct {
	TypeName    string             `json:"type_name"`
	Assignments []*FieldAssignment `json:"assignments"`
	Position    *Position          `json:"position,omitempty"`
}

func (s *CreateStatement) StatementType() string  { return "create" }
func (s *CreateStatement) GetPosition() *Position { return s.Position }

// FieldAssignment for create statements
type FieldAssignment struct {
	Field    string     `json:"field"`
	Value    Expression `json:"value"`
	Position *Position  `json:"position,omitempty"`
}

// FailStatement for explicit failures
type FailStatement struct {
	Message  string    `json:"message"`
	Position *Position `json:"position,omitempty"`
}

func (s *FailStatement) StatementType() string  { return "fail" }
func (s *FailStatement) GetPosition() *Position { return s.Position }

// Legacy types for backward compatibility
type Model struct {
	Name     string    `json:"name"`
	Fields   []*Field  `json:"fields"`
	Position *Position `json:"position,omitempty"`
}

type Field struct {
	Name         string        `json:"name"`
	Type         *Type         `json:"type"`
	Relationship *Relationship `json:"relationship,omitempty"`
	Position     *Position     `json:"position,omitempty"`
}

type Type struct {
	Name        string                 `json:"name"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Position    *Position              `json:"position,omitempty"`
}

type Relationship struct {
	Kind     string    `json:"kind"`
	Target   string    `json:"target"`
	Position *Position `json:"position,omitempty"`
}

type Parameter struct {
	Name     string    `json:"name"`
	Type     *Type     `json:"type"`
	Position *Position `json:"position,omitempty"`
}

type NativeBlock struct {
	Language string    `json:"language"`
	Code     string    `json:"code"`
	Position *Position `json:"position,omitempty"`
}

type Assignment struct {
	TypeName   string                 `json:"type_name"`
	BaseType   *Type                  `json:"base_type"`
	Why        string                 `json:"why,omitempty"`
	Validation map[string]interface{} `json:"validation,omitempty"`
	Position   *Position              `json:"position,omitempty"`
}
