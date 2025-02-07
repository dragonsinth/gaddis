package parser

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
	"strconv"
)

// TODO: defer scope creation / symbol resolution pass
// TODO: defer type checking / type propagation

func Parse(input []byte) (ret *ast.Block, err error) {
	l := lex.NewLexer(input)
	p := NewParser(l)
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	// parse the global lock
	p.maybeParseEol()
	ret = p.parseBlock(lex.EOF)
	p.parseTok(lex.EOF)
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

func (p *Parser) parseBlock(endTokens ...lex.Token) *ast.Block {
	parentScope := p.currScope
	defer func() {
		p.currScope = parentScope
	}()
	p.currScope = ast.NewScope(parentScope)

	var stmts []ast.Statement
	for {
		if slices.Contains(endTokens, p.Peek().Token) {
			si := toSourceInfo(p.Peek())
			if len(stmts) > 0 {
				si = mergeSourceInfo(stmts[0], stmts[len(stmts)-1])
			}
			return &ast.Block{SourceInfo: si, Scope: p.currScope, Statements: stmts}
		}
		stmts = append(stmts, p.parseStatement())
		p.parseEol()
	}
}

func (p *Parser) parseStatement() ast.Statement {
	r := p.Next()
	switch r.Token {
	case lex.CONSTANT:
		typ := p.parseType()
		var decls []*ast.VarDecl
		lastDecl := p.parseVarDecl(typ, true)
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, true)
			decls = append(decls, lastDecl)
		}
		return &ast.DeclareStmt{SourceInfo: spanAst(r, lastDecl), Type: typ, Decls: decls}
	case lex.DECLARE:
		typ := p.parseType()
		var decls []*ast.VarDecl
		lastDecl := p.parseVarDecl(typ, false)
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, false)
			decls = append(decls, lastDecl)
		}
		return &ast.DeclareStmt{SourceInfo: spanAst(r, lastDecl), Type: typ, Decls: decls}
	case lex.DISPLAY:
		var exprs []ast.Expression
		lastExpr := p.parseExpression()
		exprs = append(exprs, lastExpr)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastExpr = p.parseExpression()
			exprs = append(exprs, lastExpr)
		}
		return &ast.DisplayStmt{SourceInfo: spanAst(r, lastExpr), Exprs: exprs}
	case lex.INPUT:
		id := p.parseTok(lex.IDENT)
		name := id.Text
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.InputStmt{SourceInfo: spanResult(r, id), Name: name, Ref: ref}
	case lex.SET:
		id := p.parseTok(lex.IDENT)
		name := id.Text
		ref := p.currScope.Lookup(name)
		if ref == nil {
			panic(fmt.Errorf("%d:%d: unresolved reference: %s", r.Pos.Line, r.Pos.Column, name))
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

		return &ast.SetStmt{SourceInfo: spanAst(r, expr), Name: name, Ref: ref, Expr: expr}
	case lex.IF:
		ifCond := p.parseIfCondBlock(r.Pos)

		// loop for else-if
		var elseIfs []*ast.CondBlock
		var elseBlock *ast.Block
		for p.hasTok(lex.ELSE) {
			p.parseTok(lex.ELSE)
			if p.hasTok(lex.IF) {
				// an else if block
				r := p.parseTok(lex.IF)
				elseIfCond := p.parseIfCondBlock(r.Pos)
				elseIfs = append(elseIfs, elseIfCond)
			} else {
				// this is the final else block
				p.parseEol()
				elseBlock = p.parseBlock(lex.END)
			}
		}

		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.IF)
		return &ast.IfStmt{SourceInfo: spanResult(r, rEnd), If: ifCond, ElseIf: elseIfs, Else: elseBlock}
	case lex.SELECT:
		expr := p.parseExpression()
		exprType := expr.Type()
		p.parseEol()

		var cases []*ast.CaseBlock
		for p.hasTok(lex.CASE) {
			cr := p.parseTok(lex.CASE)
			caseExpr := p.parseExpression()
			if areComparableTypes(exprType, caseExpr.Type()) == ast.InvalidType {
				panic(fmt.Errorf("%d:%d type error: case type %s not comparable to select type %s", cr.Pos.Line, cr.Pos.Column, caseExpr.Type(), exprType))
			}
			p.parseTok(lex.COLON)
			p.parseEol()
			block := p.parseBlock(lex.CASE, lex.DEFAULT, lex.END)
			cases = append(cases, &ast.CaseBlock{SourceInfo: spanAst(cr, block), Expr: caseExpr, Block: block})
		}

		var def *ast.Block
		if p.hasTok(lex.DEFAULT) {
			p.parseTok(lex.DEFAULT)
			p.parseTok(lex.COLON)
			p.parseEol()
			def = p.parseBlock(lex.END)
		}

		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.SELECT)
		return &ast.SelectStmt{SourceInfo: spanResult(r, rEnd), Expr: expr, Cases: cases, Default: def} // TODO
	default:
		panic(fmt.Errorf("%d:%d: expected statement, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

func (p *Parser) parseVarDecl(typ ast.Type, isConst bool) *ast.VarDecl {
	r := p.parseTok(lex.IDENT)
	si := toSourceInfo(r)
	var expr ast.Expression
	if p.hasTok(lex.ASSIGN) {
		p.parseTok(lex.ASSIGN)
		expr = p.parseExpression()
		si = spanAst(r, expr)
	} else if !isConst {
		expr = nil
	} else {
		r := p.Peek()
		panic(fmt.Errorf("%d:%d: expected constant initializer, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
	decl := &ast.VarDecl{SourceInfo: si, Name: r.Text, Type: typ, Expr: expr, IsConst: isConst}
	if existing, ok := p.currScope.Decls[decl.Name]; ok {
		panic(fmt.Errorf("%d:%d: identifier %s %s already defined in this scope", r.Pos.Line, r.Pos.Column, existing.Type, existing.Name))
	}
	p.currScope.Decls[decl.Name] = decl
	return decl
}

func (p *Parser) parseIfCondBlock(pos lex.Position) *ast.CondBlock {
	expr := p.parseExpression()
	if expr.Type() != ast.Boolean {
		panic(fmt.Errorf("%d:%d type error: expected Boolean, got %s", pos.Line, pos.Column, expr.Type()))
	}
	p.parseTok(lex.THEN)
	p.parseEol()
	block := p.parseBlock(lex.END, lex.ELSE)
	return &ast.CondBlock{SourceInfo: mergeSourceInfo(expr, block), Expr: expr, Block: block}
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
	case lex.BOOLEAN:
		return ast.Boolean
	default:
		panic(fmt.Errorf("%d:%d expected type, got %s, %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseBinaryOperations(binaryLevels)
}

func (p *Parser) parseBinaryOperations(level int) ast.Expression {
	if level < 0 {
		return p.parseTerminal()
	}
	ops := binaryPrecedence[level]
	ret := p.parseBinaryOperations(level - 1)
	for {
		r := p.Peek()
		op, ok := binaryOpMap[r.Token]
		if ok && slices.Contains(ops, op) {
			// generate a new operation and continue
			p.Next()
			rhs := p.parseBinaryOperations(level - 1)
			ret = p.tryCreateBinaryOperation(r, op, ret, rhs)
		} else {
			return ret
		}
	}
}

func (p *Parser) tryCreateBinaryOperation(r lex.Result, op ast.Operator, lhs ast.Expression, rhs ast.Expression) *ast.BinaryOperation {
	// TODO: more semantic / type checking pass
	aTyp := lhs.Type()
	bTyp := rhs.Type()
	si := mergeSourceInfo(lhs, rhs)
	switch op {
	case ast.ADD, ast.SUB, ast.MUL, ast.DIV, ast.EXP, ast.MOD:
		if !isNumericType(aTyp) {
			panic(fmt.Errorf("%d:%d operator %s expects left hand operand of type %s to be numeric", r.Pos.Line, r.Pos.Column, r.Text, aTyp))
		}
		if !isNumericType(bTyp) {
			panic(fmt.Errorf("%d:%d operator %s expects right hand operand of type %s to be numeric", r.Pos.Line, r.Pos.Column, r.Text, bTyp))
		}
		rTyp := areComparableTypes(aTyp, bTyp)
		if rTyp == ast.InvalidType {
			panic(fmt.Errorf("%d:%d binary operation %s not supported for types %s and %s", r.Pos.Line, r.Pos.Column, r.Text, aTyp, bTyp))
		}
		return &ast.BinaryOperation{SourceInfo: si, Op: op, Typ: rTyp, Lhs: lhs, Rhs: rhs}
	case ast.EQ, ast.NEQ:
		rTyp := areComparableTypes(aTyp, bTyp)
		if rTyp == ast.InvalidType {
			panic(fmt.Errorf("%d:%d binary operation %s not supported for types %s and %s", r.Pos.Line, r.Pos.Column, r.Text, aTyp, bTyp))
		}
		return &ast.BinaryOperation{SourceInfo: si, Op: op, Typ: ast.Boolean, Lhs: lhs, Rhs: rhs}
	case ast.LT, ast.GT, ast.LTE, ast.GTE:
		if !areOrderedTypes(aTyp, bTyp) {
			panic(fmt.Errorf("%d:%d binary operation %s not supported for types %s and %s", r.Pos.Line, r.Pos.Column, r.Text, aTyp, bTyp))
		}
		return &ast.BinaryOperation{SourceInfo: si, Op: op, Typ: ast.Boolean, Lhs: lhs, Rhs: rhs}
	case ast.AND, ast.OR:
		if aTyp != ast.Boolean {
			panic(fmt.Errorf("%d:%d operator %s expects left hand operand of type %s to be boolean", r.Pos.Line, r.Pos.Column, r.Text, aTyp))
		}
		if bTyp != ast.Boolean {
			panic(fmt.Errorf("%d:%d operator %s expects right hand operand of type %s to be boolean", r.Pos.Line, r.Pos.Column, r.Text, bTyp))
		}
		return &ast.BinaryOperation{SourceInfo: si, Op: op, Typ: ast.Boolean, Lhs: lhs, Rhs: rhs}
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
		return &ast.VariableExpression{SourceInfo: toSourceInfo(r), Name: name, Ref: ref}
	case lex.INT_LIT:
		v, err := strconv.ParseInt(r.Text, 0, 64)
		if err != nil {
			panic(err)
		}
		return &ast.IntegerLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.REAL_LIT:
		v, err := strconv.ParseFloat(r.Text, 64)
		if err != nil {
			panic(err)
		}
		return &ast.RealLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.STR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.StringLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.CHR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil || len(v) > 1 {
			panic(fmt.Errorf("%d:%d: invalid string literal %s", r.Pos.Line, r.Pos.Column, r.Text))
		}
		return &ast.CharacterLiteral{SourceInfo: toSourceInfo(r), Val: v[0]}
	case lex.NOT:
		expr := p.parseExpression()
		if expr.Type() != ast.Boolean {
			panic(fmt.Errorf("%d:%d operator %s expects operand of type %s to be boolean", r.Pos.Line, r.Pos.Column, r.Text, expr.Type()))
		}
		return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NOT, Typ: ast.Boolean, Expr: expr}
	case lex.TRUE:
		return &ast.BooleanLiteral{SourceInfo: toSourceInfo(r), Val: true}
	case lex.FALSE:
		return &ast.BooleanLiteral{SourceInfo: toSourceInfo(r), Val: false}
	case lex.LPAREN:
		expr := p.parseExpression()
		p.parseTok(lex.RPAREN)
		return expr
	default:
		panic(fmt.Errorf("%d:%d expected expression, got %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

func (p *Parser) parseEol() {
	p.parseTok(lex.EOL)
	p.maybeParseEol()
}

func (p *Parser) maybeParseEol() {
	for p.hasTok(lex.EOL) {
		p.parseTok(lex.EOL)
	}
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

func errCheck(r lex.Result) lex.Result {
	if r.Token == lex.ILLEGAL {
		panic(fmt.Errorf("%d:%d: %s; %w", r.Pos.Line, r.Pos.Column, r.Token, r.Error))
	}
	return r
}
