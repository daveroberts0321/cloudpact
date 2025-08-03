package grammar

import (
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

// Grammar:
//   File   := { Model }
//   Model  := 'model' IDENT '{' { Field } '}'
//   Field  := IDENT ':' IDENT
//   Type   := IDENT

// File represents a .cf file consisting of one or more model declarations.
type File struct {
	Models []*Model
}

// Model declares a new model with a name and a list of fields.
type Model struct {
	Name   string
	Fields []*Field
}

// Field represents a single field within a model.
type Field struct {
	Name string
	Type *Type
}

// Type of a field. For now it is simply an identifier like Int or String.
type Type struct {
	Name string
}

// Parse reads .cf content from r and returns the parsed AST.
func Parse(r io.Reader) (*File, error) {
	p := &parser{}
	p.scanner.Init(r)
	p.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.SkipComments
	p.next()
	return p.parseFile()
}

// ParseString parses a string containing .cf grammar into an AST.
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

func (p *parser) parseFile() (*File, error) {
	file := &File{}
	for p.tok != scanner.EOF {
		// Skip any stray semicolons or newlines handled automatically.
		if p.tok == scanner.EOF {
			break
		}
		m, err := p.parseModel()
		if err != nil {
			return nil, err
		}
		file.Models = append(file.Models, m)
	}
	return file, nil
}

func (p *parser) expect(tok rune, expected string) error {
	if p.tok != tok {
		return fmt.Errorf("expected %s, got %q", expected, p.scanner.TokenText())
	}
	p.next()
	return nil
}

func (p *parser) parseModel() (*Model, error) {
	if p.tok != scanner.Ident || p.scanner.TokenText() != "model" {
		return nil, fmt.Errorf("expected 'model', got %q", p.scanner.TokenText())
	}
	p.next()
	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected model name, got %q", p.scanner.TokenText())
	}
	name := p.scanner.TokenText()
	p.next()
	if err := p.expect('{', "'{'"); err != nil {
		return nil, err
	}
	model := &Model{Name: name}
	for p.tok != '}' && p.tok != scanner.EOF {
		fld, err := p.parseField()
		if err != nil {
			return nil, err
		}
		model.Fields = append(model.Fields, fld)
	}
	if err := p.expect('}', "'}'"); err != nil {
		return nil, err
	}
	return model, nil
}

func (p *parser) parseField() (*Field, error) {
	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected field name, got %q", p.scanner.TokenText())
	}
	name := p.scanner.TokenText()
	p.next()
	if err := p.expect(':', "':'"); err != nil {
		return nil, err
	}
	if p.tok != scanner.Ident {
		return nil, fmt.Errorf("expected type name, got %q", p.scanner.TokenText())
	}
	typ := &Type{Name: p.scanner.TokenText()}
	p.next()
	return &Field{Name: name, Type: typ}, nil
}
