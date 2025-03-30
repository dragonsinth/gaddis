package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"slices"
	"strconv"
	"strings"
)

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
		return p.parseUnaryOperations()
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

func (p *Parser) parseUnaryOperations() ast.Expression {
	for {
		r := p.Peek()
		switch r.Token {
		case lex.NOT:
			p.Next()
			expr := p.parseUnaryOperations()
			return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NOT, Type: ast.Boolean, Expr: expr}
		case lex.SUB:
			p.Next()
			expr := p.parseUnaryOperations()
			return &ast.UnaryOperation{SourceInfo: spanAst(r, expr), Op: ast.NEG, Type: ast.UnresolvedType, Expr: expr}
		default:
			return p.parsePostfixOps()
		}
	}
}

func (p *Parser) parsePostfixOps() ast.Expression {
	expr := p.parsePrimary()
	for {
		r := p.Peek()
		switch r.Token {
		case lex.LBRACKET:
			p.Next()
			indexExpr := p.parseExpression()
			rEnd := p.parseTok(lex.RBRACKET)
			expr = &ast.ArrayRef{SourceInfo: spanResult(r, rEnd), Type: ast.UnresolvedType, Qualifier: expr, IndexExpr: indexExpr}
		case lex.DOT:
			p.Next()
			rEnd := p.parseTok(lex.IDENT)
			expr = &ast.VariableExpr{SourceInfo: spanResult(r, rEnd), Type: ast.UnresolvedType, Name: rEnd.Text, Qualifier: expr}
		case lex.LPAREN:
			p.Next()
			// This is actually a call expression, the LHS better be a variable reference...
			varRef, ok := expr.(*ast.VariableExpr)
			if !ok {
				panic(p.Errorf(r, "call qualifier must be an identifier, got %T", expr))
			}
			// Parse the args...
			args := p.parseCommaExpressions(lex.RPAREN)
			rEnd := p.parseTok(lex.RPAREN)
			// Promote the variable reference to a call expression.
			return &ast.CallExpr{SourceInfo: spanResult(r, rEnd), Name: varRef.Name, Qualifier: varRef.Qualifier, Args: args}
		default:
			return expr
		}
	}
}

func (p *Parser) parsePrimary() ast.Expression {
	r := p.Next()
	switch r.Token {
	case lex.IDENT:
		return &ast.VariableExpr{SourceInfo: toSourceInfo(r), Name: r.Text}
	case lex.INT_LIT:
		return p.parseLiteral(r, ast.Integer)
	case lex.REAL_LIT:
		return p.parseLiteral(r, ast.Real)
	case lex.STR_LIT, lex.TAB_LIT:
		return p.parseLiteral(r, ast.String)
	case lex.CHR_LIT:
		return p.parseLiteral(r, ast.Character)
	case lex.NEW:
		// Must be a call expression
		rNext := p.parseTok(lex.IDENT)
		name := rNext.Text
		p.parseTok(lex.LPAREN)
		args := p.parseCommaExpressions(lex.RPAREN)
		rEnd := p.parseTok(lex.RPAREN)
		return &ast.NewExpr{SourceInfo: spanResult(r, rEnd), Name: name, Args: args}
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
