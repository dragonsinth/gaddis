package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

func (p *Parser) parseModuleStmt(r lex.Result) *ast.ModuleStmt {
	rNext := p.parseTok(lex.IDENT)
	name := rNext.Text

	var params []*ast.VarDecl
	p.parseTok(lex.LPAREN)
	if !p.hasTok(lex.RPAREN) {
		params = append(params, p.parseParamDecl())
	}
	for p.hasTok(lex.COMMA) {
		p.parseTok(lex.COMMA)
		params = append(params, p.parseParamDecl())
	}
	p.parseTok(lex.RPAREN)
	p.parseEol()
	block := p.parseBlock(lex.END)
	p.parseTok(lex.END)
	rEnd := p.parseTok(lex.MODULE)
	return &ast.ModuleStmt{SourceInfo: spanResult(r, rEnd), Name: name, Params: params, Block: block}
}

func (p *Parser) parseFunctionStmt(r lex.Result) *ast.FunctionStmt {
	returnType := p.parseType()
	rNext := p.parseTok(lex.IDENT)
	name := rNext.Text

	var params []*ast.VarDecl
	p.parseTok(lex.LPAREN)
	if !p.hasTok(lex.RPAREN) {
		params = append(params, p.parseParamDecl())
	}
	for p.hasTok(lex.COMMA) {
		p.parseTok(lex.COMMA)
		params = append(params, p.parseParamDecl())
	}
	p.parseTok(lex.RPAREN)
	p.parseEol()
	block := p.parseBlock(lex.END)
	p.parseTok(lex.END)
	rEnd := p.parseTok(lex.FUNCTION)
	return &ast.FunctionStmt{SourceInfo: spanResult(r, rEnd), Name: name, Type: returnType, Params: params, Block: block}
}
