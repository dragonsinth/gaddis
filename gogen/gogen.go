package gogen

import (
	_ "embed"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lex"
	"io"
	"strconv"
	"strings"
)

var (
	//go:embed tmpl/builtins.go
	builtins string
)

const oldPrefix = "package main_template\n"
const newPrefix = "package main\n"
const suffix = "}\n"

func Generate(globalBlock *ast.Block) string {
	var sb strings.Builder

	sb.WriteString(newPrefix)
	data := strings.TrimPrefix(builtins, oldPrefix)
	data = strings.TrimSuffix(data, suffix)
	sb.WriteString(data)
	v := New("\t", &sb)
	globalBlock.Visit(v)
	sb.WriteString(suffix)
	return sb.String()
}

type Output interface {
	WriteString(string)
}

func New(indent string, out io.StringWriter) *GoGenerator {
	return &GoGenerator{
		ind: indent,
		out: out,
	}
}

type GoGenerator struct {
	ind string
	out io.StringWriter
}

var _ ast.Visitor = &GoGenerator{}

func (g *GoGenerator) PreVisitBlock(bl *ast.Block) bool {
	return true
}

func (g *GoGenerator) PostVisitBlock(bl *ast.Block) {
}

func (g *GoGenerator) PreVisitVarDecl(vd *ast.VarDecl) bool {
	g.indent()
	if vd.IsConst {
		g.output("const ")
	} else {
		g.output("var ")
	}
	g.ident(vd)
	g.output(" ")
	g.output(goType(vd.Type))
	if vd.Expr != nil {
		g.output(" = ")
	}
	return true
}

func (g *GoGenerator) PostVisitVarDecl(vd *ast.VarDecl) {
	g.output("\n")
	// Also emit a no-op assignment to avoid Go unreferenced variable errors.
	g.indent()
	g.output("_ = ")
	g.ident(vd)
	g.output("\n")
}

func (g *GoGenerator) PreVisitConstantStmt(stmt *ast.ConstantStmt) bool {
	return true
}

func (g *GoGenerator) PostVisitConstantStmt(stmt *ast.ConstantStmt) {
}

func (g *GoGenerator) PreVisitDeclareStmt(stmt *ast.DeclareStmt) bool {
	return true
}

func (g *GoGenerator) PostVisitDeclareStmt(stmt *ast.DeclareStmt) {
}

func (g *GoGenerator) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	g.indent()
	g.output("display(")
	for i, arg := range d.Exprs {
		if i > 0 {
			g.output(", ")
		}
		arg.Visit(g)
	}
	g.output(")\n")
	return false
}

func (g *GoGenerator) PostVisitDisplayStmt(d *ast.DisplayStmt) {
}

func (g *GoGenerator) PreVisitInputStmt(i *ast.InputStmt) bool {
	g.indent()
	g.ident(i.Ref)
	g.output(" = ")
	g.output("input")
	g.output(i.Ref.Type.String())
	g.output("()\n")
	return false
}

func (g *GoGenerator) PostVisitInputStmt(i *ast.InputStmt) {
}

func (g *GoGenerator) PreVisitSetStmt(s *ast.SetStmt) bool {
	g.indent()
	g.ident(s.Ref)
	g.output(" = ")
	return true
}

func (g *GoGenerator) PostVisitSetStmt(s *ast.SetStmt) {
	g.output("\n")
}

func (g *GoGenerator) PreVisitIntegerLiteral(il *ast.IntegerLiteral) bool {
	g.output(strconv.FormatInt(il.Val, 10))
	return true
}

func (g *GoGenerator) PostVisitIntegerLiteral(l *ast.IntegerLiteral) {
}

func (g *GoGenerator) PreVisitRealLiteral(rl *ast.RealLiteral) bool {
	g.output(strconv.FormatFloat(rl.Val, 'f', -1, 64))
	return true
}

func (g *GoGenerator) PostVisitRealLiteral(l *ast.RealLiteral) {
}

func (g *GoGenerator) PreVisitStringLiteral(sl *ast.StringLiteral) bool {
	g.output(strconv.Quote(sl.Val))
	return true
}

func (g *GoGenerator) PostVisitStringLiteral(sl *ast.StringLiteral) {
}

func (g *GoGenerator) PreVisitCharacterLiteral(cl *ast.CharacterLiteral) bool {
	g.output(strconv.QuoteRune(rune(cl.Val)))
	return true
}

func (g *GoGenerator) PostVisitCharacterLiteral(cl *ast.CharacterLiteral) {
}

func (g *GoGenerator) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	// must special case exp and mod
	if bo.Op == lex.MOD || bo.Op == lex.EXP {
		if bo.Op == lex.MOD {
			g.output("mod")
		} else {
			g.output("exp")
		}
		g.output(bo.Typ.String())
		g.output("(")
		g.maybeCast(bo.Typ, bo.Lhs)
		g.output(", ")
		g.maybeCast(bo.Typ, bo.Rhs)
		g.output(")")
	} else {
		g.output("(")
		g.maybeCast(bo.Typ, bo.Lhs)
		g.output(operator(bo.Op))
		g.maybeCast(bo.Typ, bo.Rhs)
		g.output(")")
	}
	return false
}

func (g *GoGenerator) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (g *GoGenerator) PreVisitVariableExpression(ve *ast.VariableExpression) bool {
	g.ident(ve.Ref)
	return true
}

func (g *GoGenerator) PostVisitVariableExpression(ve *ast.VariableExpression) {
}

func (g *GoGenerator) indent() {
	g.output(g.ind)
}

func (g *GoGenerator) output(s string) {
	_, err := g.out.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (g *GoGenerator) ident(ref *ast.VarDecl) {
	g.output(ref.Name)
	g.output("_")
}

func (g *GoGenerator) maybeCast(typ ast.Type, exp ast.Expression) {
	if exp.Type() != typ {
		g.output(goType(typ))
		g.output("(")
		defer g.output(")")
	}
	exp.Visit(g)
}

func goType(t ast.Type) string {
	return [...]string{"INVALID", "int64", "float64", "string", "byte"}[t]
}

func operator(op lex.Token) string {
	return [...]string{
		lex.ADD: "+",
		lex.SUB: "-",
		lex.MUL: "*",
		lex.DIV: "/",
		lex.EXP: "^", // TODO: fixme
		lex.MOD: "%",
	}[op]
}
