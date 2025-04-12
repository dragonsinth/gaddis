package parse

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
)

// TODO: only allow modules and functions in the global block?

const maxErrors = 20

func Parse(input string) (*ast.Program, []ast.Comment, []ast.Error) {
	l := lex.New(input)
	p := New(l)
	ret := p.parseGlobalBlock()
	ret.Block.End = toSourceInfo(l.Lex()).End
	errors := p.errors
	if len(errors) > maxErrors {
		errors = errors[:maxErrors]
	}
	return ret, p.comments, errors
}

func ParseExpr(input string) (ast.Expression, error) {
	l := lex.New(input)
	p := New(l)
	return p.safeParseExpression()
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

	types map[ast.TypeKey]ast.Type
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

func (p *Parser) parseGlobalBlock() *ast.Program {
	bl := p.doParseBlock(true)
	return &ast.Program{Block: bl}
}

func (p *Parser) parseBlock(endTokens ...lex.Token) *ast.Block {
	return p.doParseBlock(false, endTokens...)
}

func (p *Parser) doParseBlock(isGlobal bool, endTokens ...lex.Token) *ast.Block {
	var stmts []ast.Statement
	start := p.SafePeek()
	for {
		peek := p.SafePeek()
		if peek.Token == lex.EOF || slices.Contains(endTokens, peek.Token) {
			return &ast.Block{SourceInfo: spanResult(start, peek), Statements: stmts}
		}

		st := p.safeParseStatement(isGlobal)
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

func (p *Parser) safeParseStatement(isGlobalBlock bool) ast.Statement {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(ast.Error); ok {
				p.errors = append(p.errors, pe)
			} else {
				panic(e)
			}
		}
	}()

	return p.parseStatement(isGlobalBlock)
}

func (p *Parser) safeParseExpression() (_ ast.Expression, err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(ast.Error); ok {
				err = pe
			} else {
				panic(e)
			}
		}
	}()

	return p.parseExpression(), nil
}

type EmptyStatement struct {
	ast.Statement
}

func (p *Parser) parseStatement(isGlobalBlock bool) ast.Statement {
	r := p.Next()
	switch r.Token {
	case lex.EOL, lex.EOF:
		return EmptyStatement{}
	case lex.CONSTANT:
		typ := p.parseType()
		var decls []*ast.VarDecl
		lastDecl := p.parseVarDecl(typ, varDeclOpts{isConst: true})
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, varDeclOpts{isConst: true})
			decls = append(decls, lastDecl)
		}
		return &ast.DeclareStmt{SourceInfo: spanAst(r, lastDecl), Type: typ, IsConst: true, Decls: decls}
	case lex.DECLARE:
		typ := p.parseType()

		if p.hasTok(lex.APPENDMODE) {
			p.parseTok(lex.APPENDMODE)
			typ = ast.AppendFile
		}

		var decls []*ast.VarDecl
		lastDecl := p.parseVarDecl(typ, varDeclOpts{})
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, varDeclOpts{})
			decls = append(decls, lastDecl)
		}
		return &ast.DeclareStmt{SourceInfo: spanAst(r, lastDecl), Type: typ, IsConst: false, Decls: decls}
	case lex.DISPLAY, lex.PRINT:
		si := toSourceInfo(r)
		exprs := p.parseCommaExpressions(lex.EOL)
		if len(exprs) > 0 {
			si = mergeSourceInfo(si, exprs[len(exprs)-1])
		}
		return &ast.DisplayStmt{SourceInfo: si, Exprs: exprs, IsPrint: r.Token == lex.PRINT}
	case lex.INPUT:
		refExpr := p.parseExpression()
		return &ast.InputStmt{SourceInfo: spanAst(r, refExpr), Ref: refExpr}
	case lex.SET:
		refExpr := p.parseExpression()
		p.parseTok(lex.ASSIGN)
		expr := p.parseExpression()
		return &ast.SetStmt{SourceInfo: spanAst(r, expr), Ref: refExpr, Expr: expr}
	case lex.IF:
		cases := []*ast.CondBlock{p.parseIfCondBlock(r.Pos)}

		// loop for else-if
		for p.hasTok(lex.ELSE) {
			r := p.parseTok(lex.ELSE)
			cases[len(cases)-1].SourceInfo.End = toSourceInfo(r).End
			if p.hasTok(lex.IF) {
				// an else if block
				p.parseTok(lex.IF)
				elseIfCond := p.parseIfCondBlock(r.Pos)
				cases = append(cases, elseIfCond)
			} else {
				// this is the final else block
				p.parseEol()
				elseBlock := p.parseBlock(lex.END)
				elseCond := &ast.CondBlock{SourceInfo: spanAst(r, elseBlock), Block: elseBlock}
				cases = append(cases, elseCond)
			}
		}

		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.IF)
		cases[len(cases)-1].SourceInfo.End = toSourceInfo(rEnd).End
		return &ast.IfStmt{SourceInfo: spanResult(r, rEnd), Cases: cases}
	case lex.SELECT:
		expr := p.parseExpression()
		p.parseEol()

		// Special case: select is the only block-like construct that does not actually parse a block.
		// So we need to manually consume any lines before the first case statement. After that, any
		// empty lines will be consumed by the last select/case.
		for p.hasTok(lex.EOL) {
			p.parseTok(lex.EOL)
		}

		var cases []*ast.CaseBlock
		for p.hasTok(lex.CASE) {
			cr := p.parseTok(lex.CASE)
			caseExpr := p.parseExpression()
			p.parseTok(lex.COLON)
			p.parseEol()
			block := p.parseBlock(lex.CASE, lex.DEFAULT, lex.END)
			cases = append(cases, &ast.CaseBlock{SourceInfo: spanAst(cr, block), Expr: caseExpr, Block: block})
		}

		if p.hasTok(lex.DEFAULT) {
			cr := p.parseTok(lex.DEFAULT)
			p.parseTok(lex.COLON)
			p.parseEol()
			block := p.parseBlock(lex.END)
			cases = append(cases, &ast.CaseBlock{SourceInfo: spanAst(cr, block), Expr: nil, Block: block})
		}

		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.SELECT)
		return &ast.SelectStmt{SourceInfo: spanResult(r, rEnd), Type: ast.UnresolvedType, Expr: expr, Cases: cases}
	case lex.DO:
		p.parseEol()
		block := p.parseBlock(lex.WHILE, lex.UNTIL)
		rEnd := p.Next()
		until := false
		switch rEnd.Token {
		case lex.WHILE:
		case lex.UNTIL:
			until = true
		default:
			panic(p.Errorf(r, "expected While or Until, got %s %q", r.Token, r.Text))
		}
		expr := p.parseExpression()
		return &ast.DoStmt{SourceInfo: spanAst(r, expr), Block: block, Until: until, Expr: expr}
	case lex.WHILE:
		expr := p.parseExpression()
		p.parseEol()
		block := p.parseBlock(lex.END)
		p.parseTok(lex.END)
		rEnd := p.parseTok(lex.WHILE)
		return &ast.WhileStmt{SourceInfo: spanResult(r, rEnd), Expr: expr, Block: block}
	case lex.FOR:
		if p.hasTok(lex.EACH) {
			p.parseTok(lex.EACH)
			refExpr := p.parseExpression()
			p.parseTok(lex.IN)
			arrayExpr := p.parseExpression()
			p.parseEol()
			block := p.parseBlock(lex.END)
			p.parseTok(lex.END)
			rEnd := p.parseTok(lex.FOR)
			return &ast.ForEachStmt{SourceInfo: spanResult(r, rEnd), Ref: refExpr, ArrayExpr: arrayExpr, Block: block}
		} else {
			refExpr := p.parseExpression()
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
			return &ast.ForStmt{SourceInfo: spanResult(r, rEnd), Ref: refExpr, StartExpr: startExpr, StopExpr: stopExpr, StepExpr: stepExpr, Block: block}
		}
	case lex.CALL:
		expr := p.parseExpression()
		// Better be a call expressions...
		callExpr, ok := expr.(*ast.CallExpr)
		if !ok {
			panic(p.Errorf(r, "Expected call expression, got %T", expr))
		}
		// Promote to a call statement
		return &ast.CallStmt{SourceInfo: spanAst(r, callExpr), Name: callExpr.Name, Qualifier: callExpr.Qualifier, Args: callExpr.Args}
	case lex.MODULE:
		if !isGlobalBlock {
			panic(p.Errorf(r, "Module may only be declared in the global scope"))
		}
		return p.parseModuleStmt(r)
	case lex.RETURN:
		expr := p.parseExpression()
		return &ast.ReturnStmt{SourceInfo: spanAst(r, expr), Expr: expr}

	case lex.FUNCTION:
		if !isGlobalBlock {
			panic(p.Errorf(r, "Function may only be declared in the global scope"))
		}
		return p.parseFunctionStmt(r)
	case lex.OPEN:
		file := p.parseExpression()
		nameExpr := p.parseExpression()
		return &ast.OpenStmt{SourceInfo: spanAst(r, nameExpr), File: file, Name: nameExpr}
	case lex.CLOSE:
		file := p.parseExpression()
		return &ast.CloseStmt{SourceInfo: spanAst(r, file), File: file}
	case lex.READ:
		file := p.parseExpression()

		si := toSourceInfo(r)
		exprs := p.parseCommaExpressions(lex.EOL)
		if len(exprs) > 0 {
			si = mergeSourceInfo(si, exprs[len(exprs)-1])
		}
		return &ast.ReadStmt{SourceInfo: si, File: file, Exprs: exprs}
	case lex.WRITE:
		file := p.parseExpression()

		si := toSourceInfo(r)
		exprs := p.parseCommaExpressions(lex.EOL)
		if len(exprs) > 0 {
			si = mergeSourceInfo(si, exprs[len(exprs)-1])
		}
		return &ast.WriteStmt{SourceInfo: si, File: file, Exprs: exprs}
	case lex.DELETE:
		file := p.parseExpression()
		return &ast.DeleteStmt{SourceInfo: spanAst(r, file), File: file}
	case lex.RENAME:
		oldFile := p.parseExpression()
		p.parseTok(lex.COMMA)
		newFile := p.parseExpression()
		return &ast.RenameStmt{SourceInfo: spanAst(r, oldFile), OldFile: oldFile, NewFile: newFile}
	case lex.CLASS:
		if !isGlobalBlock {
			panic(p.Errorf(r, "Class may only be declared in the global scope"))
		}
		return p.parseClassBody(r)
	default:
		panic(p.Errorf(r, "expected statement, got %s %q", r.Token, r.Text))
	}
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
	name := r.Text

	baseType := typ
	nDims := 0
	for p.hasTok(lex.LBRACKET) {
		p.parseTok(lex.LBRACKET)
		r = p.parseTok(lex.RBRACKET)
		nDims++
		typ = p.makeArrayType(baseType, nDims, typ)
	}

	return &ast.VarDecl{SourceInfo: spanResult(rStart, r), Name: name, Type: typ, IsParam: true, IsRef: isRef}
}

func (p *Parser) parseIfCondBlock(start lex.Position) *ast.CondBlock {
	expr := p.parseExpression()
	p.parseTok(lex.THEN)
	p.parseEol()
	block := p.parseBlock(lex.END, lex.ELSE)
	return &ast.CondBlock{SourceInfo: ast.SourceInfo{Start: toPos(start), End: block.End}, Expr: expr, Block: block}
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
