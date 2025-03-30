package parse

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
)

var (
	binaryOpMap = map[lex.Token]ast.Operator{
		lex.ADD: ast.ADD,
		lex.SUB: ast.SUB,
		lex.MUL: ast.MUL,
		lex.DIV: ast.DIV,
		lex.EXP: ast.EXP,
		lex.MOD: ast.MOD,
		lex.EQ:  ast.EQ,
		lex.NEQ: ast.NEQ,
		lex.LT:  ast.LT,
		lex.LTE: ast.LTE,
		lex.GT:  ast.GT,
		lex.GTE: ast.GTE,
		lex.AND: ast.AND,
		lex.OR:  ast.OR,
	}
	binaryPrecedence = [][]ast.Operator{
		{ast.EXP},
		{ast.MUL, ast.DIV, ast.MOD},
		{ast.ADD, ast.SUB},
		{ast.EQ, ast.NEQ, ast.LT, ast.LTE, ast.GT, ast.GTE},
		{ast.AND, ast.OR},
	}
	binaryLevels = len(binaryPrecedence) - 1
)
