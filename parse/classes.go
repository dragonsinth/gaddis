package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

func (p *Parser) parseClassBody(r lex.Result) *ast.ClassStmt {
	rNext := p.parseTok(lex.IDENT)
	name := rNext.Text
	var extends string
	if p.hasTok(lex.EXTENDS) {
		p.parseTok(lex.EXTENDS)
		rNext = p.parseTok(lex.IDENT)
		extends = rNext.Text
	}

	classType := p.makeClassType(name)
	if extends != "" {
		classType.Extends = p.makeClassType(extends)
	}

	var stmts []ast.Statement
	for {
		peek := p.SafePeek()
		if peek.Token == lex.END {
			p.parseTok(lex.END)
			rEnd := p.parseTok(lex.CLASS)

			si := toSourceInfo(peek)
			if len(stmts) > 0 {
				si = mergeSourceInfo(stmts[0], stmts[len(stmts)-1])
			}
			return &ast.ClassStmt{
				SourceInfo: spanResult(r, rEnd),
				Name:       name,
				Extends:    extends,
				Type:       classType,
				Block:      &ast.Block{SourceInfo: si, Statements: stmts},
				Scope:      nil,
			}
		}

		st := p.safeParseClassElement()
		if _, ok := st.(EmptyStatement); ok {
			// nothing
		} else if st != nil {
			stmts = append(stmts, st)
		} else {
			if len(p.errors) > maxErrors {
				return &ast.ClassStmt{}
			}

			// something went wrong... consume tokens until we hit an EOL then try to keep going
			for tok := p.SafePeek(); tok.Token != lex.EOL && tok.Token != lex.EOF; tok = p.SafePeek() {
				p.SafeNext()
			}
		}
	}
}

func (p *Parser) safeParseClassElement() ast.Statement {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(ast.Error); ok {
				p.errors = append(p.errors, pe)
			} else {
				panic(e)
			}
		}
	}()

	return p.parseClassElement()
}

func (p *Parser) parseClassElement() ast.Statement {
	r := p.Next()
	isPrivate := false
	switch r.Token {
	case lex.EOL, lex.EOF:
		return EmptyStatement{}
	case lex.PUBLIC:
	case lex.PRIVATE:
		isPrivate = true
	default:
		panic(p.Errorf(r, "expected class element, got %s %q", r.Token, r.Text))
	}

	rNext := p.Peek()
	switch rNext.Token {
	case lex.MODULE:
		p.Next()
		ms := p.parseModuleStmt(r)
		ms.IsMethod = true
		ms.IsPrivate = isPrivate
		return ms
	case lex.FUNCTION:
		p.Next()
		fs := p.parseFunctionStmt(r)
		fs.IsMethod = true
		fs.IsPrivate = isPrivate
		return fs
	default:
		typ := p.parseType()
		if p.hasTok(lex.APPENDMODE) {
			p.parseTok(lex.APPENDMODE)
			typ = ast.AppendFile
		}

		var decls []*ast.VarDecl
		lastDecl := p.parseVarDecl(typ, varDeclOpts{isField: true, isPrivate: isPrivate})
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, varDeclOpts{isField: true, isPrivate: isPrivate})
			decls = append(decls, lastDecl)
		}
		return &ast.DeclareStmt{
			SourceInfo: spanAst(r, lastDecl),
			Type:       typ,
			IsConst:    false,
			IsField:    true,
			IsPrivate:  isPrivate,
			Decls:      decls,
		}
	}
}
