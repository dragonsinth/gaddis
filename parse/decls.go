package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

type varDeclOpts struct {
	isConst   bool
	isPrivate bool
	isField   *ast.ClassType
}

func (p *Parser) parseVarDecl(typ ast.Type, opts varDeclOpts) *ast.VarDecl {
	isFileType := typ.IsFileType()

	r := p.parseTok(lex.IDENT)
	rEnd := r

	if opts.isConst && isFileType {
		panic(p.Errorf(r, "file types cannot be constant"))
	}

	var dims []ast.Expression
	if !opts.isConst {
		baseType := typ
		for p.hasTok(lex.LBRACKET) {
			p.parseTok(lex.LBRACKET)
			dims = append(dims, p.parseExpression())
			rEnd = p.parseTok(lex.RBRACKET)
			typ = p.makeArrayType(baseType, len(dims), typ)
		}
	}

	si := spanResult(r, rEnd)
	var expr ast.Expression

	if p.hasTok(lex.ASSIGN) {
		if isFileType {
			panic(p.Errorf(r, "file types cannot have initializers"))
		}
		rAssign := p.parseTok(lex.ASSIGN)
		if len(dims) > 0 {
			expr = p.parseArrayInitializer(rAssign, typ.(*ast.ArrayType))
		} else {
			expr = p.parseExpression()
		}
		si = spanAst(r, expr)
	} else if opts.isConst {
		r := p.Peek()
		panic(p.Errorf(r, "expected constant initializer, got %s %q", r.Token, r.Text))
	}
	return &ast.VarDecl{
		SourceInfo: si,
		Name:       r.Text,
		Type:       typ,
		DimExprs:   dims,
		Expr:       expr,
		IsConst:    opts.isConst,
		IsField:    opts.isField,
		IsPrivate:  opts.isPrivate,
	}
}
