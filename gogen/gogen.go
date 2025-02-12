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

func GoGenerate(globalBlock *ast.Block) string {
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

func New(indent string, out io.StringWriter) *Visitor {
	return &Visitor{
		ind: indent,
		out: out,
	}
}

type Visitor struct {
	ind        string
	out        io.StringWriter
	selectType ast.Type // type of most recently enclosing select statement
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitBlock(bl *ast.Block) bool {
	v.ind = v.ind + "\t"
	return true
}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {
	v.ind = v.ind[:len(v.ind)-1]
}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	v.indent()
	if vd.IsConst {
		v.output("const ")
	} else {
		v.output("var ")
	}
	v.ident(vd)
	v.output(" ")
	v.output(goTypes[vd.Type])
	if vd.Expr != nil {
		v.output(" = ")
		v.maybeCast(vd.Type, vd.Expr)
	}
	v.output("\n")
	// Also emit a no-op assignment to avoid Go unreferenced variable errors.
	v.indent()
	v.output("_ = ")
	v.ident(vd)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
}

func (v *Visitor) PreVisitConstantStmt(stmt *ast.ConstantStmt) bool {
	return true
}

func (v *Visitor) PostVisitConstantStmt(stmt *ast.ConstantStmt) {
}

func (v *Visitor) PreVisitDeclareStmt(stmt *ast.DeclareStmt) bool {
	return true
}

func (v *Visitor) PostVisitDeclareStmt(stmt *ast.DeclareStmt) {
}

func (v *Visitor) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	v.indent()
	v.output("display(")
	for i, arg := range d.Exprs {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitDisplayStmt(d *ast.DisplayStmt) {
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	v.indent()
	v.ident(i.Ref)
	v.output(" = ")
	v.output("input")
	v.output(i.Ref.Type.String())
	v.output("()\n")
	return false
}

func (v *Visitor) PostVisitInputStmt(i *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(s *ast.SetStmt) bool {
	v.indent()
	v.ident(s.Ref)
	v.output(" = ")
	v.maybeCast(s.Ref.Type, s.Expr)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitSetStmt(s *ast.SetStmt) {
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	v.indent()
	is.If.Visit(v)

	for _, cb := range is.ElseIf {
		v.output(" else ")
		cb.Visit(v)
	}
	if is.Else != nil {
		v.output(" else {\n")
		is.Else.Visit(v)
		v.indent()
		v.output("}")
	}
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {
}

func (v *Visitor) PreVisitCondBlock(cb *ast.CondBlock) bool {
	v.output("if ")
	cb.Expr.Visit(v)
	v.output(" {\n")
	cb.Block.Visit(v)
	v.indent()
	v.output("}")
	return false
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	v.indent()

	v.output("switch (")
	v.maybeCast(ss.Type, ss.Expr)
	v.output(") {\n")

	oldInd, oldType := v.ind, v.selectType
	v.ind += "\t"
	v.selectType = ss.Type
	defer func() {
		v.ind, v.selectType = oldInd, oldType
	}()

	for _, cb := range ss.Cases {
		cb.Visit(v)
	}
	if ss.Default != nil {
		v.indent()
		v.output("default:\n")
		ss.Default.Visit(v)
	}

	v.ind = oldInd
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	v.indent()
	v.output("case ")
	v.maybeCast(v.selectType, cb.Expr)
	v.output(":\n")
	cb.Block.Visit(v)
	return false
}

func (v *Visitor) PostVisitCaseBlock(cb *ast.CaseBlock) {
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	v.indent()
	v.output("for {\n")
	ds.Block.Visit(v)
	v.indent()

	v.output("\tif ")
	if ds.Not {
		ds.Expr.Visit(v)
	} else {
		v.output("!(")
		ds.Expr.Visit(v)
		v.output(")")
	}
	v.output(" {\n")
	v.indent()
	v.output("\t\tbreak\n")
	v.indent()
	v.output("\t}\n")

	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	v.indent()
	v.output("for ")
	ws.Expr.Visit(v)
	v.output(" {\n")
	ws.Block.Visit(v)
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	v.indent()
	v.ident(fs.Ref)
	v.output(" = ")
	fs.StartExpr.Visit(v)
	v.output("\n")

	v.indent()
	v.output("for step")
	refType := fs.Ref.Type
	v.output(refType.String())
	v.output("(")
	v.ident(fs.Ref)
	v.output(", ")
	v.maybeCast(refType, fs.StopExpr)
	v.output(", ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	v.output(") {\n")
	fs.Block.Visit(v)

	v.indent()
	v.output("\t")
	v.ident(fs.Ref)
	v.output(" += ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	v.output("\n")
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
}

func (v *Visitor) PreVisitIntegerLiteral(il *ast.IntegerLiteral) bool {
	v.output(strconv.FormatInt(il.Val, 10))
	return true
}

func (v *Visitor) PostVisitIntegerLiteral(l *ast.IntegerLiteral) {
}

func (v *Visitor) PreVisitRealLiteral(rl *ast.RealLiteral) bool {
	v.output(strconv.FormatFloat(rl.Val, 'f', -1, 64))
	return true
}

func (v *Visitor) PostVisitRealLiteral(l *ast.RealLiteral) {
}

func (v *Visitor) PreVisitStringLiteral(sl *ast.StringLiteral) bool {
	v.output(strconv.Quote(sl.Val))
	return true
}

func (v *Visitor) PostVisitStringLiteral(sl *ast.StringLiteral) {
}

func (v *Visitor) PreVisitCharacterLiteral(cl *ast.CharacterLiteral) bool {
	v.output(strconv.QuoteRune(rune(cl.Val)))
	return true
}

func (v *Visitor) PostVisitCharacterLiteral(cl *ast.CharacterLiteral) {
}

func (v *Visitor) PreVisitBooleanLiteral(cl *ast.BooleanLiteral) bool {
	v.output(strconv.FormatBool(cl.Val))
	return true
}

func (v *Visitor) PostVisitBooleanLiteral(cl *ast.BooleanLiteral) {
}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	switch uo.Op {
	case ast.NOT:
		v.output(" !")
	case ast.NEG:
		v.output(" -")
	default:
		panic(uo.Op)
	}
	v.output("(")
	uo.Expr.Visit(v)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitUnaryOperation(uo *ast.UnaryOperation) {
}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	dstType := ast.AreComparableTypes(bo.Lhs.Type(), bo.Rhs.Type())

	// must special case exp and mod
	if bo.Op == ast.MOD || bo.Op == ast.EXP {
		if bo.Op == ast.MOD {
			v.output("mod")
		} else {
			v.output("exp")
		}
		v.output(bo.Typ.String())
		v.output("(")
		v.maybeCast(dstType, bo.Lhs)
		v.output(", ")
		v.maybeCast(dstType, bo.Rhs)
		v.output(")")
	} else {
		v.output("(")
		v.maybeCast(dstType, bo.Lhs)
		v.output(goBinaryOperators[bo.Op])
		v.maybeCast(dstType, bo.Rhs)
		v.output(")")
	}
	return false
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (v *Visitor) PreVisitVariableExpression(ve *ast.VariableExpression) bool {
	v.ident(ve.Ref)
	return true
}

func (v *Visitor) PostVisitVariableExpression(ve *ast.VariableExpression) {
}

func (v *Visitor) indent() {
	v.output(v.ind)
}

func (v *Visitor) output(s string) {
	_, err := v.out.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (v *Visitor) ident(ref *ast.VarDecl) {
	v.output(ref.Name)
	v.output("_")
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	if dstType == ast.Real && exp.Type() == ast.Integer {
		v.output("float64(")
		exp.Visit(v)
		v.output(")")
	} else if dstType == ast.Integer && exp.Type() == ast.Real {
		v.output("int64(")
		exp.Visit(v)
		v.output(")")
	} else {
		exp.Visit(v)
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
