package gogen

import (
	_ "embed"
	"github.com/dragonsinth/gaddis/ast"
	"io"
	"strconv"
	"strings"
)

var (
	//go:embed builtins/builtins.go
	builtins string
)

const oldPrefix = "package main_template\n"
const newPrefix = "package main\n"

func Generate(globalBlock *ast.Block) string {
	var sb strings.Builder

	data := strings.TrimPrefix(builtins, oldPrefix)
	sb.WriteString(newPrefix)
	sb.WriteString(data)
	sb.WriteString("\nfunc main() {\n")
	v := New("", &sb)
	// TODO: consider special-casing the global block to declare vars in global vs. other statements in main()
	globalBlock.Visit(v)
	sb.WriteString("}\n")
	return sb.String()
}

func New(indent string, out io.StringWriter) *GoGenerator {
	return &GoGenerator{
		ind: indent,
		out: out,
	}
}

type GoGenerator struct {
	ind        string
	out        io.StringWriter
	selectType ast.Type // type of most recently enclosing select statement
}

func (g *GoGenerator) PreVisitDoStmt(ds *ast.DoStmt) bool {
	g.indent()
	g.output("for {\n")
	ds.Block.Visit(g)
	g.indent()

	g.output("\tif ")
	if ds.Not {
		ds.Expr.Visit(g)
	} else {
		g.output("!(")
		ds.Expr.Visit(g)
		g.output(")")
	}
	g.output(" {\n")
	g.indent()
	g.output("\t\tbreak\n")
	g.indent()
	g.output("\t}\n")

	g.indent()
	g.output("}\n")
	return false
}

func (g *GoGenerator) PostVisitDoStmt(ds *ast.DoStmt) {
}

func (g *GoGenerator) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	g.indent()
	g.output("for ")
	ws.Expr.Visit(g)
	g.output(" {\n")
	ws.Block.Visit(g)
	g.indent()
	g.output("}\n")
	return false
}

func (g *GoGenerator) PostVisitWhileStmt(ws *ast.WhileStmt) {
}

func (g *GoGenerator) PreVisitForStmt(ws *ast.ForStmt) bool {
	g.indent()
	g.output("for ")
	g.ident(ws.Ref)
	g.output(" = ")
	ws.StartExpr.Visit(g)
	g.output("; true; ")
	g.ident(ws.Ref)
	if ws.StepExpr != nil {
		g.output(" += ")
		ws.StepExpr.Visit(g)
		g.output(" {\n")

	} else {
		g.output("++ {\n")
	}

	ws.Block.Visit(g)

	g.indent()
	g.output("\tif ")
	g.ident(ws.Ref)
	g.output(" == ")
	ws.StopExpr.Visit(g)
	g.output(" {\n")
	g.indent()
	g.output("\t\tbreak\n")
	g.indent()
	g.output("\t}\n")

	g.indent()
	g.output("}\n")
	return false
}

func (g *GoGenerator) PostVisitForStmt(ws *ast.ForStmt) {
	//TODO implement me
	panic("implement me")
}

var _ ast.Visitor = &GoGenerator{}

func (g *GoGenerator) PreVisitBlock(bl *ast.Block) bool {
	g.ind = g.ind + "\t"
	return true
}

func (g *GoGenerator) PostVisitBlock(bl *ast.Block) {
	g.ind = g.ind[:len(g.ind)-1]
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
	g.output(goTypes[vd.Type])
	if vd.Expr != nil {
		g.output(" = ")
		g.maybeCast(vd.Type, vd.Expr)
	}
	g.output("\n")
	// Also emit a no-op assignment to avoid Go unreferenced variable errors.
	g.indent()
	g.output("_ = ")
	g.ident(vd)
	g.output("\n")
	return false
}

func (g *GoGenerator) PostVisitVarDecl(vd *ast.VarDecl) {
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
	g.maybeCast(s.Ref.Type, s.Expr)
	g.output("\n")
	return false
}

func (g *GoGenerator) PostVisitSetStmt(s *ast.SetStmt) {
}

func (g *GoGenerator) PreVisitIfStmt(is *ast.IfStmt) bool {
	g.indent()
	is.If.Visit(g)

	for _, cb := range is.ElseIf {
		g.output(" else ")
		cb.Visit(g)
	}
	if is.Else != nil {
		g.output(" else {\n")
		is.Else.Visit(g)
		g.indent()
		g.output("}")
	}
	g.output("\n")
	return false
}

func (g *GoGenerator) PostVisitIfStmt(is *ast.IfStmt) {
}

func (g *GoGenerator) PreVisitCondBlock(cb *ast.CondBlock) bool {
	g.output("if ")
	cb.Expr.Visit(g)
	g.output(" {\n")
	cb.Block.Visit(g)
	g.indent()
	g.output("}")
	return false
}

func (g *GoGenerator) PostVisitCondBlock(cb *ast.CondBlock) {
}

func (g *GoGenerator) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	g.indent()

	g.output("switch (")
	g.maybeCast(ss.Type, ss.Expr)
	g.output(") {\n")

	oldInd, oldType := g.ind, g.selectType
	g.ind += "\t"
	g.selectType = ss.Type
	defer func() {
		g.ind, g.selectType = oldInd, oldType
	}()

	for _, cb := range ss.Cases {
		cb.Visit(g)
	}
	if ss.Default != nil {
		g.indent()
		g.output("default:\n")
		ss.Default.Visit(g)
	}

	g.ind = oldInd
	g.indent()
	g.output("}\n")
	return false
}

func (g *GoGenerator) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (g *GoGenerator) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	g.indent()
	g.output("case ")
	g.maybeCast(g.selectType, cb.Expr)
	g.output(":\n")
	cb.Block.Visit(g)
	return false
}

func (g *GoGenerator) PostVisitCaseBlock(cb *ast.CaseBlock) {
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

func (g *GoGenerator) PreVisitBooleanLiteral(cl *ast.BooleanLiteral) bool {
	g.output(strconv.FormatBool(cl.Val))
	return true
}

func (g *GoGenerator) PostVisitBooleanLiteral(cl *ast.BooleanLiteral) {
}

func (g *GoGenerator) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	if uo.Op == ast.NOT {
		g.output("!")
	}
	g.output("(")
	uo.Expr.Visit(g)
	g.output(")")
	return false
}

func (g *GoGenerator) PostVisitUnaryOperation(uo *ast.UnaryOperation) {
}

func (g *GoGenerator) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	dstType := ast.AreComparableTypes(bo.Lhs.Type(), bo.Rhs.Type())

	// must special case exp and mod
	if bo.Op == ast.MOD || bo.Op == ast.EXP {
		if bo.Op == ast.MOD {
			g.output("mod")
		} else {
			g.output("exp")
		}
		g.output(bo.Typ.String())
		g.output("(")
		g.maybeCast(dstType, bo.Lhs)
		g.output(", ")
		g.maybeCast(dstType, bo.Rhs)
		g.output(")")
	} else {
		g.output("(")
		g.maybeCast(dstType, bo.Lhs)
		g.output(goBinaryOperators[bo.Op])
		g.maybeCast(dstType, bo.Rhs)
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

func (g *GoGenerator) maybeCast(dstType ast.Type, exp ast.Expression) {
	if dstType == ast.Real && exp.Type() == ast.Integer {
		g.output("float64(")
		exp.Visit(g)
		g.output(")")
	} else if dstType == ast.Integer && exp.Type() == ast.Real {
		g.output("int64(")
		exp.Visit(g)
		g.output(")")
	} else {
		exp.Visit(g)
	}
}

var goTypes = [...]string{
	ast.Integer:   "int64",
	ast.Real:      "float64",
	ast.String:    "string",
	ast.Character: "character",
	ast.Boolean:   "bool",
}

var goBinaryOperators = [...]string{
	ast.ADD: "+",
	ast.SUB: "-",
	ast.MUL: "*",
	ast.DIV: "/",
	ast.EQ:  "==",
	ast.NEQ: "!=",
	ast.LT:  "<",
	ast.LTE: "<=",
	ast.GT:  ">",
	ast.GTE: ">=",
	ast.AND: "&&",
	ast.OR:  "||",
}
