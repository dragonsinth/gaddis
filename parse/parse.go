package parse

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
	"strconv"
)

// TODO: only allow modules and functions in the global block?

const maxErrors = 20

func Parse(input []byte) (*ast.Block, []ast.Comment, []ast.Error) {
	l := lex.New(input)
	p := New(l)
	ret := p.parseBlock(lex.EOF)
	errs := p.errors
	if len(errs) > maxErrors {
		errs = errs[:maxErrors]
	}
	return ret, p.comments, errs
}

func New(l *lex.Lexer) *Parser {
	return &Parser{
		lex:  l,
		next: nil,
	}
}

type Parser struct {
	lex  *lex.Lexer
	next *lex.Result

	comments []ast.Comment
	errors   []ast.Error
}

func (p *Parser) Peek() lex.Result {
	return p.errCheck(p.SafePeek())
}

func (p *Parser) Next() lex.Result {
	return p.errCheck(p.SafeNext())
}

func (p *Parser) SafePeek() lex.Result {
	if p.next == nil {
		v := p.nextNonComment()
		p.next = &v
	}
	return *p.next
}

func (p *Parser) SafeNext() lex.Result {
	if p.next != nil {
		v := *p.next
		p.next = nil
		return v
	}
	return p.nextNonComment()
}

func (p *Parser) nextNonComment() lex.Result {
	for {
		r := p.lex.Lex()
		if r.Token != lex.COMMENT {
			return r
		}
		p.comments = append(p.comments, ast.Comment{SourceInfo: toSourceInfo(r), Text: r.Text})
	}
}

func (p *Parser) parseBlock(endTokens ...lex.Token) *ast.Block {
	var stmts []ast.Statement
	for {
		peek := p.SafePeek()
		if slices.Contains(endTokens, peek.Token) {
			si := toSourceInfo(peek)
			if len(stmts) > 0 {
				si = mergeSourceInfo(stmts[0], stmts[len(stmts)-1])
			}
			return &ast.Block{SourceInfo: si, Statements: stmts}
		}

		st := p.safeParseStatement()
		if _, ok := st.(EmptyStatement); ok {
			// nothing
		} else if st != nil {
			stmts = append(stmts, st)
		} else {
			if len(p.errors) > maxErrors {
				return &ast.Block{}
			}

			// something went wrong... consume tokens until we hit an EOL then try to keep going
			for tok := p.SafePeek(); tok.Token != lex.EOL && tok.Token != lex.EOF; tok = p.SafePeek() {
				p.SafeNext()
			}
		}
	}
}

func (p *Parser) safeParseStatement() ast.Statement {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(ast.Error); ok {
				p.errors = append(p.errors, pe)
			} else {
				panic(e)
			}
		}
	}()

	return p.parseStatement()
}

type EmptyStatement struct {
	ast.Statement
}

func (p *Parser) parseStatement() ast.Statement {
	r := p.Next()
	switch r.Token {
	case lex.EOL, lex.EOF:
		return EmptyStatement{}
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
		varExpr := p.parseVariableExpression()
		return &ast.InputStmt{SourceInfo: spanAst(r, varExpr), Var: varExpr}
	case lex.SET:
		varExpr := p.parseVariableExpression()
		p.parseTok(lex.ASSIGN)
		expr := p.parseExpression()
		return &ast.SetStmt{SourceInfo: spanAst(r, expr), Var: varExpr, Expr: expr}
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
		p.parseEol()

		var cases []*ast.CaseBlock
		for p.hasTok(lex.CASE) {
			cr := p.parseTok(lex.CASE)
			caseExpr := p.parseExpression()
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
		return &ast.SelectStmt{SourceInfo: spanResult(r, rEnd), Type: ast.UnresolvedType, Expr: expr, Cases: cases, Default: def}
	case lex.DO:
		p.parseEol()
		block := p.parseBlock(lex.WHILE, lex.UNTIL)
		rEnd := p.Next()
		not := false
		switch rEnd.Token {
		case lex.WHILE:
		case lex.UNTIL:
			not = true
		default:
			panic(p.Errorf(r, "expected While or Until, got %s %q", r.Token, r.Text))
		}
		expr := p.parseExpression()
		return &ast.DoStmt{SourceInfo: spanAst(r, expr), Block: block, Not: not, Expr: expr}
	case lex.WHILE:
		expr := p.parseExpression()
		p.parseEol()
		block := p.parseBlock(lex.END)
		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.WHILE)
		return &ast.WhileStmt{SourceInfo: spanResult(r, rEnd), Expr: expr, Block: block}
	case lex.FOR:
		varExpr := p.parseVariableExpression()
		p.parseTok(lex.ASSIGN)
		startExpr := p.parseExpression()
		p.parseTok(lex.TO)
		stopExpr := p.parseExpression()

		var stepExpr ast.Expression
		if p.hasTok(lex.STEP) {
			p.parseTok(lex.STEP)
			stepExpr = p.parseExpression()
		}
		p.parseEol()
		block := p.parseBlock(lex.END)
		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.FOR)
		return &ast.ForStmt{SourceInfo: spanResult(r, rEnd), Var: varExpr, StartExpr: startExpr, StopExpr: stopExpr, StepExpr: stepExpr, Block: block}
	case lex.CALL:
		rNext := p.parseTok(lex.IDENT)
		name := rNext.Text

		var args []ast.Expression
		p.parseTok(lex.LPAREN)
		if !p.hasTok(lex.RPAREN) {
			args = append(args, p.parseExpression())
		}
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			args = append(args, p.parseExpression())
		}
		rEnd := p.parseTok(lex.RPAREN)
		return &ast.CallStmt{SourceInfo: spanResult(r, rEnd), Name: name, Args: args}
	case lex.MODULE:
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
	default:
		panic(p.Errorf(r, "expected statement, got %s %q", r.Token, r.Text))
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
		panic(p.Errorf(r, "expected constant initializer, got %s %q", r.Token, r.Text))
	}
	return &ast.VarDecl{SourceInfo: si, Name: r.Text, Type: typ, Expr: expr, IsConst: isConst}
}

func (p *Parser) parseParamDecl() *ast.VarDecl {
	rStart := p.Peek()
	typ := p.parseType()
	isRef := false
	if p.hasTok(lex.REF) {
		p.parseTok(lex.REF)
		isRef = true
	}
	r := p.parseTok(lex.IDENT)
	return &ast.VarDecl{SourceInfo: spanResult(rStart, r), Name: r.Text, Type: typ, IsParam: true, IsRef: isRef}
}

func (p *Parser) parseIfCondBlock(start lex.Position) *ast.CondBlock {
	expr := p.parseExpression()
	p.parseTok(lex.THEN)
	p.parseEol()
	block := p.parseBlock(lex.END, lex.ELSE)
	return &ast.CondBlock{SourceInfo: ast.SourceInfo{Start: toPos(start), End: block.End}, Expr: expr, Block: block}
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
		panic(p.Errorf(r, "expected type, got %s, %q", r.Token, r.Text))
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
			ret = &ast.BinaryOperation{SourceInfo: mergeSourceInfo(ret, rhs), Op: op, Type: ast.UnresolvedType, Lhs: ret, Rhs: rhs}
		} else {
			return ret
		}
	}
}

func (p *Parser) parseVariableExpression() *ast.VariableExpression {
	r := p.parseTok(lex.IDENT)
	return &ast.VariableExpression{SourceInfo: toSourceInfo(r), Name: r.Text}
}

func (p *Parser) parseTerminal() ast.Expression {
	r := p.Next()
	switch r.Token {
	case lex.IDENT:
		return &ast.VariableExpression{SourceInfo: toSourceInfo(r), Name: r.Text}
	case lex.INT_LIT:
		v, err := strconv.ParseInt(r.Text, 0, 64)
		if err != nil {
			// should happen in lexer
			panic(p.Errorf(r, "invalid Integer literal %s", r.Text))
		}
		return &ast.IntegerLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.REAL_LIT:
		v, err := strconv.ParseFloat(r.Text, 64)
		if err != nil {
			// should happen in lexer
			panic(p.Errorf(r, "invalid Real literal %s", r.Text))
		}
		return &ast.RealLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.STR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil {
			// should happen in lexer
			panic(p.Errorf(r, "invalid String literal %s", r.Text))
		}
		return &ast.StringLiteral{SourceInfo: toSourceInfo(r), Val: v}
	case lex.CHR_LIT:
		v, err := strconv.Unquote(r.Text)
		if err != nil || len(v) > 1 {
			// should happen in lexer
			panic(p.Errorf(r, "invalid Character literal %s", r.Text))
		}
		return &ast.CharacterLiteral{SourceInfo: toSourceInfo(r), Val: v[0]}
	case lex.NOT:
		expr := p.parseExpression()
		return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NOT, Type: ast.Boolean, Expr: expr}
	case lex.SUB:
		expr := p.parseExpression()
		return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NEG, Type: expr.GetType(), Expr: expr}
	case lex.TRUE:
		return &ast.BooleanLiteral{SourceInfo: toSourceInfo(r), Val: true}
	case lex.FALSE:
		return &ast.BooleanLiteral{SourceInfo: toSourceInfo(r), Val: false}
	case lex.LPAREN:
		expr := p.parseExpression()
		p.parseTok(lex.RPAREN)
		return expr
	default:
		panic(p.Errorf(r, "expected expression, got %s %q", r.Token, r.Text))
	}
}

func (p *Parser) parseEol() {
	if !p.hasTok(lex.EOF) {
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
		panic(p.Errorf(r, "expected token %q, got %s %q", expect, r.Token, r.Text))
	}
	return r
}

func (p *Parser) errCheck(r lex.Result) lex.Result {
	if r.Token == lex.ILLEGAL {
		panic(p.Errorf(r, "illegal token %q; %s", r.Text, r.Error.Error()))
	}
	return r
}

func (p *Parser) Errorf(r lex.Result, fmtStr string, args ...any) ast.Error {
	return ast.Error{
		SourceInfo: toSourceInfo(r),
		Desc:       fmt.Sprintf("syntax error: "+fmtStr, args...),
	}
}
