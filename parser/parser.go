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
		return &ast.DeclareStmt{Type: typ, Decls: decls}
	case lex.DECLARE:
		typ := p.parseType()
		var decls []*ast.VarDecl
		decls = append(decls, p.parseVarDecl(typ, false))
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			decls = append(decls, p.parseVarDecl(typ, false))
		}
		return &ast.DeclareStmt{Type: typ, Decls: decls}
	case lex.DISPLAY:
		var exprs []ast.Expression
		exprs = append(exprs, p.parseExpression())
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			exprs = append(exprs, p.parseExpression())
		}
		return &ast.DisplayStmt{Exprs: exprs}
	case lex.INPUT:
		name := p.parseIdentifer()
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.InputStmt{Name: name, Ref: ref}
	case lex.SET:
		name := p.parseIdentifer()
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		p.parseTok(lex.ASSIGN)
		expr := p.parseExpression()

		// type check
		exprType := expr.Type()
		if ref.Type != exprType {
			if ref.Type != ast.Real || exprType != ast.Integer {
				panic(fmt.Errorf("%d:%d type error: %s not assignable to %s %s", r.Pos.Line, r.Pos.Column, exprType, ref.Type, name))
			}
		}

		return &ast.SetStmt{Name: name, Ref: ref, Expr: expr}
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
	decl := &ast.VarDecl{Name: r.Text, Type: typ, Expr: expr, IsConst: isConst}
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

var tokenMap = map[lex.Token]ast.Operator{
	lex.ADD: ast.ADD,
	lex.SUB: ast.SUB,
	lex.MUL: ast.MUL,
	lex.DIV: ast.DIV,
	lex.EXP: ast.EXP,
	lex.MOD: ast.MOD,
}

var precedence = [][]ast.Operator{
	{ast.EXP},
	{ast.MUL, ast.DIV, ast.MOD},
	{ast.ADD, ast.SUB},
}

func (p *Parser) parseOperators(level int) ast.Expression {
	if level < 0 {
		return p.parseTerminal()
	}
	ops := precedence[level]
	ret := p.parseOperators(level - 1)
	for {
		r := p.Peek()
		op, ok := tokenMap[r.Token]
		if ok && slices.Contains(ops, op) {
			// generate a new binary operation and continue
			p.Next()
			rhs := p.parseOperators(level - 1)
			ret = p.tryCreateBinaryOperation(r, op, ret, rhs)
		} else {
			return ret
		}
	}
}

func (p *Parser) tryCreateBinaryOperation(r lex.Result, op ast.Operator, lhs ast.Expression, rhs ast.Expression) *ast.BinaryOperation {
	// TODO: more semantic / type checking pass
	switch op {
	case ast.ADD, ast.SUB, ast.MUL, ast.DIV, ast.EXP, ast.MOD:
		aTyp := lhs.Type()
		bTyp := rhs.Type()
		var typ ast.Type
		if aTyp == ast.Integer && bTyp == ast.Integer {
			typ = ast.Integer
		} else if aTyp == ast.Integer && bTyp == ast.Real {
			typ = ast.Real
		} else if aTyp == ast.Real && bTyp == ast.Integer {
			typ = ast.Real
		} else if aTyp == ast.Real && bTyp == ast.Real {
			typ = ast.Real
		} else {
			panic(fmt.Errorf("%d:%d unsupported binary operation %s not supported for types %s and %s", r.Pos.Line, r.Pos.Column, r.Text, aTyp, bTyp))
		}
		return &ast.BinaryOperation{Op: op, Typ: typ, Lhs: lhs, Rhs: rhs}
	default:
		panic(fmt.Errorf("%d:%d unsupported binary operation: %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
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
		return &ast.VariableExpression{Name: name, Ref: ref}
	case lex.INT_LIT:
		v, err := strconv.ParseInt(r.Text, 0, 64)
		if err != nil {
			panic(err)
		}
		return &ast.IntegerLiteral{Val: v}
	case lex.REAL_LIT:
		v, err := strconv.ParseFloat(r.Text, 64)
		if err != nil {
			panic(err)
		}
		return &ast.RealLiteral{Val: v}
	case lex.STR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.StringLiteral{Val: v}
	case lex.CHR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil || len(v) > 1 {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.CharacterLiteral{Val: v[0]}
	case lex.LPAREN:
		expr := p.parseExpression()
		p.parseTok(lex.RPAREN)
		return expr
	default:
		panic(fmt.Errorf("%d:%d expected expression, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}
