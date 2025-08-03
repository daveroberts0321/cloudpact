// Package grammar implements the CloudPact language parser.
// expr.go defines expression-related AST nodes used by the parser.
package grammar

// Expression interface
type Expression interface {
	ExpressionType() string
	GetPosition() *Position
}

// IdentifierExpression
type IdentifierExpression struct {
	Name     string    `json:"name"`
	Position *Position `json:"position,omitempty"`
}

func (e *IdentifierExpression) ExpressionType() string { return "identifier" }
func (e *IdentifierExpression) GetPosition() *Position { return e.Position }

// LiteralExpression
type LiteralExpression struct {
	Value    interface{} `json:"value"`
	Position *Position   `json:"position,omitempty"`
}

func (e *LiteralExpression) ExpressionType() string { return "literal" }
func (e *LiteralExpression) GetPosition() *Position { return e.Position }

// BinaryExpression for operations like "user.age < 18"
type BinaryExpression struct {
	Left     Expression `json:"left"`
	Operator string     `json:"operator"`
	Right    Expression `json:"right"`
	Position *Position  `json:"position,omitempty"`
}

func (e *BinaryExpression) ExpressionType() string { return "binary" }
func (e *BinaryExpression) GetPosition() *Position { return e.Position }

// CallExpression for function calls
type CallExpression struct {
	Function  string       `json:"function"`
	Arguments []Expression `json:"arguments"`
	Position  *Position    `json:"position,omitempty"`
}

func (e *CallExpression) ExpressionType() string { return "call" }
func (e *CallExpression) GetPosition() *Position { return e.Position }

// MemberExpression for "user.email"
type MemberExpression struct {
	Object   Expression `json:"object"`
	Property string     `json:"property"`
	Position *Position  `json:"position,omitempty"`
}

func (e *MemberExpression) ExpressionType() string { return "member" }
func (e *MemberExpression) GetPosition() *Position { return e.Position }
