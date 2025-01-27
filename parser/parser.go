package parser

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
	"strconv"
)

func Parse(input []byte) (ret *ast.Block, err error) {
	l := lex.NewLexer(input)
	p := NewParser(l)
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	// parse the global lock
	ret = p.parseBlock(lex.EOF)
	return
}

func NewParser(l *lex.Lexer) *Parser {
	return &Parser{
		lex:  l,
		next: nil,
	}
}

type Parser struct {
	lex       *lex.Lexer
	next      *lex.Result
	currScope *ast.Scope
}

func (p *Parser) Peek() lex.Result {
	if p.next == nil {
		v := p.lex.Lex()
		p.next = &v
	}
	return errCheck(*p.next)
}

func (p *Parser) Next() lex.Result {
	if p.next != nil {
		v := *p.next
		p.next = nil
		return v
	}
	return errCheck(p.lex.Lex())
}

func errCheck(r lex.Result) lex.Result {
	if r.Token == lex.ILLEGAL {
		panic(fmt.Errorf("%d:%d: %s; %w", r.Pos.Line, r.Pos.Column, r.Token, r.Error))
	}
	return r
}

func (p *Parser) parseBlock(expectToken lex.Token) *ast.Block {
	parentScope := p.currScope
	defer func() {
		p.currScope = parentScope
	}()
	p.currScope = ast.NewScope(parentScope)

	var stmts []ast.Statement
	for !p.hasTok(expectToken) {
		for p.hasTok(lex.EOL) {
			p.parseTok(lex.EOL)
		}
		stmts = append(stmts, p.parseStatement())
		p.parseTok(lex.EOL)
	}
	return &ast.Block{Scope: p.currScope, Statements: stmts}
}

func (p *Parser) parseStatement() ast.Statement {
	r := p.Next()
	switch r.Token {
	case lex.CONSTANT:
		typ := p.parseType()
		var decls []*ast.VarDecl
		decls = append(decls, p.parseVarDecl(typ, true))
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			decls = append(decls, p.parseVarDecl(typ, true))
		}
		return &ast.DeclareStmt{typ, decls}
	case lex.DECLARE:
		typ := p.parseType()
		var decls []*ast.VarDecl
		decls = append(decls, p.parseVarDecl(typ, false))
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			decls = append(decls, p.parseVarDecl(typ, false))
		}
		return &ast.DeclareStmt{typ, decls}
	case lex.DISPLAY:
		var exprs []ast.Expression
		exprs = append(exprs, p.parseExpression())
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			exprs = append(exprs, p.parseExpression())
		}
		return &ast.DisplayStmt{exprs}
	case lex.INPUT:
		name := p.parseIdentifer()
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.InputStmt{name, ref}
	case lex.SET:
		name := p.parseIdentifer()
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		p.parseTok(lex.ASSIGN)
		expr := p.parseExpression()
		return &ast.SetStmt{name, ref, expr}
	default:
		panic(fmt.Errorf("%d:%d: expected statement, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

func (p *Parser) parseVarDecl(typ ast.Type, isConst bool) *ast.VarDecl {
	r := p.parseTok(lex.IDENT)
	var expr ast.Expression
	if p.hasTok(lex.ASSIGN) {
		p.parseTok(lex.ASSIGN)
		expr = p.parseExpression()
	} else if !isConst {
		expr = nil
	} else {
		r := p.Peek()
		panic(fmt.Errorf("%d:%d: expected constant initializer, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
	decl := &ast.VarDecl{r.Text, typ, expr, isConst}
	if existing, ok := p.currScope.Decls[decl.Name]; ok {
		panic(fmt.Errorf("%d:%d: identifier %s %s already defined in this scope", r.Pos.Line, r.Pos.Column, existing.Type, existing.Name))
	}
	p.currScope.Decls[decl.Name] = decl
	return decl
}

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
	default:
		panic(fmt.Errorf("%d:%d expected type, got %s, %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

func (p *Parser) parseIdentifer() string {
	r := p.parseTok(lex.IDENT)
	return r.Text
}

func (p *Parser) hasTok(expect lex.Token) bool {
	r := p.Peek()
	return r.Token == expect
}

func (p *Parser) parseTok(expect lex.Token) lex.Result {
	r := p.Next()
	if r.Token != expect {
		panic(fmt.Errorf("%d:%d expected token %q, got %s %q", r.Pos.Line, r.Pos.Column, expect, r.Token, r.Text))
	}
	return r
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseOperators(2)
}

var precedence = [][]lex.Token{
	{lex.EXP},
	{lex.MUL, lex.DIV, lex.MOD},
	{lex.ADD, lex.SUB},
}

func (p *Parser) parseOperators(level int) ast.Expression {
	if level < 0 {
		return p.parseTerminal()
	}
	ops := precedence[level]
	ret := p.parseOperators(level - 1)
	for {
		r := p.Peek()
		if slices.Contains(ops, r.Token) {
			// generate a new binary operation and continue
			p.Next()
			rhs := p.parseOperators(level - 1)
			ret = ast.NewBinaryOperation(r, ret, rhs)
		} else {
			return ret
		}
	}
}

func (p *Parser) parseTerminal() ast.Expression {
	r := p.Next()
	switch r.Token {
	case lex.IDENT:
		name := r.Text
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.VariableExpression{name, ref}
	case lex.INT_LIT:
		v, err := strconv.ParseInt(r.Text, 0, 64)
		if err != nil {
			panic(err)
		}
		return &ast.IntegerLiteral{v}
	case lex.REAL_LIT:
		v, err := strconv.ParseFloat(r.Text, 64)
		if err != nil {
			panic(err)
		}
		return &ast.RealLiteral{v}
	case lex.STR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.StringLiteral{v}
	case lex.CHR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil || len(v) > 1 {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.CharacterLiteral{v[0]}
	case lex.LPAREN:
		expr := p.parseExpression()
		p.parseTok(lex.RPAREN)
		return expr
	default:
		panic(fmt.Errorf("%d:%d expected expression, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}
