package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

func (p *Parser) parseType() ast.Type {
	r := p.Next()
	switch r.Token {
	case lex.INTEGER:
		return ast.Integer
	case lex.REAL:
		return ast.Real
	case lex.STRING:
		return ast.String
	case lex.CHARACTER:
		return ast.Character
	case lex.BOOLEAN:
		return ast.Boolean
	case lex.OUTPUTFILE:
		return ast.OutputFile
	case lex.INPUTFILE:
		return ast.InputFile
	case lex.IDENT:
		// must be a class
		return p.makeClassType(r.Text)
	default:
		panic(p.Errorf(r, "expected type, got %s, %q", r.Token, r.Text))
	}
}

func (p *Parser) makeArrayType(baseType ast.Type, nDims int, elementType ast.Type) *ast.ArrayType {
	key := elementType.Key() + "[]"
	if existing, ok := p.types[key]; ok {
		return existing.(*ast.ArrayType)
	}
	if p.types == nil {
		p.types = map[ast.TypeKey]ast.Type{}
	}
	typ := &ast.ArrayType{
		Base:        baseType,
		NDims:       nDims,
		ElementType: elementType,
		TypeKey:     key,
	}
	p.types[key] = typ
	return typ
}

func (p *Parser) makeClassType(name string) *ast.ClassType {
	key := ast.TypeKey(name)
	if existing, ok := p.types[key]; ok {
		return existing.(*ast.ClassType)
	}
	if p.types == nil {
		p.types = map[ast.TypeKey]ast.Type{}
	}
	typ := &ast.ClassType{
		TypeKey: key,
	}
	p.types[key] = typ
	return typ
}
