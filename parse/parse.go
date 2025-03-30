package parse

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
	"strconv"
	"strings"
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
	for {
		peek := p.SafePeek()
		if peek.Token == lex.EOF || slices.Contains(endTokens, peek.Token) {
			si := toSourceInfo(peek)
			if len(stmts) > 0 {
				si = mergeSourceInfo(stmts[0], stmts[len(stmts)-1])
			}
			return &ast.Block{SourceInfo: si, Statements: stmts}
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
		lastDecl := p.parseVarDecl(typ, true)
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, true)
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
		lastDecl := p.parseVarDecl(typ, false)
		decls = append(decls, lastDecl)
		for p.hasTok(lex.COMMA) {
			p.parseTok(lex.COMMA)
			lastDecl = p.parseVarDecl(typ, false)
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
		rNext := p.parseTok(lex.IDENT)
		name := rNext.Text
		p.parseTok(lex.LPAREN)
		args := p.parseCommaExpressions(lex.RPAREN)
		rEnd := p.parseTok(lex.RPAREN)
		return &ast.CallStmt{SourceInfo: spanResult(r, rEnd), Name: name, Args: args}
	case lex.MODULE:
		if !isGlobalBlock {
			panic(p.Errorf(r, "Module may only be declared in the global scope"))
		}
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
	case lex.RETURN:
		expr := p.parseExpression()
		return &ast.ReturnStmt{SourceInfo: spanAst(r, expr), Expr: expr}

	case lex.FUNCTION:
		if !isGlobalBlock {
			panic(p.Errorf(r, "Function may only be declared in the global scope"))
		}

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
	case lex.OPEN:
		return &ast.OpenStmt{SourceInfo: toSourceInfo(r), File: p.parseExpression(), Name: p.parseExpression()}
	case lex.CLOSE:
		return &ast.CloseStmt{SourceInfo: toSourceInfo(r), File: p.parseExpression()}
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
	default:
		panic(p.Errorf(r, "expected statement, got %s %q", r.Token, r.Text))
	}
}

func (p *Parser) parseVarDecl(typ ast.Type, isConst bool) *ast.VarDecl {
	isFileType := typ.IsFileType()

	r := p.parseTok(lex.IDENT)
	rEnd := r

	if isConst && isFileType {
		panic(p.Errorf(r, "file types cannot be constant"))
	}

	var dims []ast.Expression
	if !isConst {
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
	} else if isConst {
		r := p.Peek()
		panic(p.Errorf(r, "expected constant initializer, got %s %q", r.Token, r.Text))
	}
	return &ast.VarDecl{SourceInfo: si, Name: r.Text, Type: typ, DimExprs: dims, Expr: expr, IsConst: isConst}
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
	default:
		panic(p.Errorf(r, "expected type, got %s, %q", r.Token, r.Text))
	}
}

func (p *Parser) parseCommaExpressions(endTokens ...lex.Token) []ast.Expression {
	var exprs []ast.Expression
	if peek := p.Peek(); peek.Token != lex.EOF && !slices.Contains(endTokens, peek.Token) {
		for {
			expr := p.parseExpression()
			exprs = append(exprs, expr)
			if p.hasTok(lex.COMMA) {
				p.parseTok(lex.COMMA)
			} else {
				break
			}
		}
	}
	return exprs
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseBinaryOperations(binaryLevels)
}

func (p *Parser) parseBinaryOperations(level int) ast.Expression {
	if level < 0 {
		return p.parseTrailingArrayRefs()
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

func (p *Parser) parseTrailingArrayRefs() ast.Expression {
	expr := p.parseTerminal()
	for p.hasTok(lex.LBRACKET) {
		r := p.parseTok(lex.LBRACKET)
		indexExpr := p.parseExpression()
		rEnd := p.parseTok(lex.RBRACKET)
		expr = &ast.ArrayRef{SourceInfo: spanResult(r, rEnd), Type: ast.UnresolvedType, RefExpr: expr, IndexExpr: indexExpr}
	}
	return expr
}

func (p *Parser) parseTerminal() ast.Expression {
	r := p.Next()
	switch r.Token {
	case lex.IDENT:
		// If the next token is a '(', this is actually a CallExpr.
		if p.hasTok(lex.LPAREN) {
			// This is actually a call expression.
			p.parseTok(lex.LPAREN)
			args := p.parseCommaExpressions(lex.RPAREN)
			rEnd := p.parseTok(lex.RPAREN)
			return &ast.CallExpr{SourceInfo: spanResult(r, rEnd), Name: r.Text, Args: args}
		} else {
			return &ast.VariableExpr{SourceInfo: toSourceInfo(r), Name: r.Text}
		}
	case lex.INT_LIT:
		return p.parseLiteral(r, ast.Integer)
	case lex.REAL_LIT:
		return p.parseLiteral(r, ast.Real)
	case lex.STR_LIT, lex.TAB_LIT:
		return p.parseLiteral(r, ast.String)
	case lex.CHR_LIT:
		return p.parseLiteral(r, ast.Character)
	case lex.NOT:
		expr := p.parseExpression()
		return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NOT, Type: ast.Boolean, Expr: expr}
	case lex.SUB:
		expr := p.parseExpression()
		return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NEG, Type: ast.UnresolvedType, Expr: expr}
	case lex.TRUE, lex.FALSE:
		return p.parseLiteral(r, ast.Boolean)
	case lex.LPAREN:
		expr := p.parseExpression()
		rEnd := p.parseTok(lex.RPAREN)
		return &ast.ParenExpr{SourceInfo: spanResult(r, rEnd), Expr: expr}
	default:
		panic(p.Errorf(r, "expected expression, got %s %q", r.Token, r.Text))
	}
}

func (p *Parser) parseLiteral(r lex.Result, typ ast.PrimitiveType) *ast.Literal {
	lit := ParseLiteral(r.Text, toSourceInfo(r), typ)
	if lit == nil {
		// should not happen; lexer should catch such things
		panic(p.Errorf(r, "invalid %s literal %s", typ.String(), r.Text))
	}
	return lit
}

func (p *Parser) parseArrayInitializer(r lex.Result, typ *ast.ArrayType) ast.Expression {
	for p.hasTok(lex.EOL) {
		p.parseTok(lex.EOL)
	}
	args := []ast.Expression{p.parseExpression()}
	for p.hasTok(lex.COMMA) {
		p.parseTok(lex.COMMA)
		for p.hasTok(lex.EOL) {
			p.parseTok(lex.EOL)
		}
		args = append(args, p.parseExpression())
	}
	si := spanAst(r, args[len(args)-1])
	return &ast.ArrayInitializer{SourceInfo: si, Args: args, Type: typ}
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

func (p *Parser) makeArrayType(baseType ast.Type, nDims int, elementType ast.Type) ast.Type {
	key := elementType.Key() + "[]"
	typ := &ast.ArrayType{
		Base:        baseType,
		NDims:       nDims,
		ElementType: elementType,
		TypeKey:     key,
	}
	if existing, ok := p.types[key]; ok {
		return existing
	}
	if p.types == nil {
		p.types = map[ast.TypeKey]ast.Type{}
	}
	p.types[key] = typ
	return typ
}

func (p *Parser) Errorf(r lex.Result, fmtStr string, args ...any) ast.Error {
	return ast.Error{
		SourceInfo: toSourceInfo(r),
		Desc:       fmt.Sprintf("syntax error: "+fmtStr, args...),
	}
}

func ParseLiteral(input string, si ast.SourceInfo, typ ast.PrimitiveType) *ast.Literal {
	switch typ {
	case ast.Integer:
		v, err := strconv.ParseInt(input, 0, 64)
		if err != nil {
			return nil
		}
		return &ast.Literal{SourceInfo: si, Type: ast.Integer, Val: v}
	case ast.Real:
		v, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return nil
		}
		return &ast.Literal{SourceInfo: si, Type: ast.Real, Val: v}
	case ast.String:
		if input == "Tab" {
			return &ast.Literal{SourceInfo: si, Type: ast.String, Val: "\t", IsTabLiteral: true}
		}
		v, err := strconv.Unquote(input)
		if err != nil {
			return nil
		}
		return &ast.Literal{SourceInfo: si, Type: ast.String, Val: v}
	case ast.Character:
		v, err := strconv.Unquote(input)
		if err != nil || len(v) > 1 {
			return nil
		}
		return &ast.Literal{SourceInfo: si, Type: ast.Character, Val: v[0]}
	case ast.Boolean:
		// ToLower only to support dynamic eval; lexer would not suppport.
		switch strings.ToLower(input) {
		case "true":
			return &ast.Literal{SourceInfo: si, Type: ast.Boolean, Val: true}
		case "false":
			return &ast.Literal{SourceInfo: si, Type: ast.Boolean, Val: false}
		default:
			return nil
		}
	default:
		panic(typ)
	}
}
