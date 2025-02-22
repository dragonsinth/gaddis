package astprint

import (
	"github.com/dragonsinth/gaddis/ast"
	"io"
	"strconv"
	"strings"
)

// TODO: emit extra newlines and comments to turn this into a pretty printer.

func Print(prog *ast.Program, comments []ast.Comment) string {
	// visit the statements in the global block
	var sb strings.Builder
	v := New("", &sb)
	for _, stmt := range prog.Block.Statements {
		stmt.Visit(v)
	}
	return sb.String()
}

func New(indent string, out io.StringWriter) *Visitor {
	return &Visitor{
		ind: indent,
		out: out,
	}
}

type Visitor struct {
	ind string
	out io.StringWriter
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
	if vd.IsParam {
		v.output(vd.Type.String())
		if vd.IsRef {
			v.output(" Ref")
		}
		v.output(" ")
		v.output(vd.Name)
	} else {
		v.output(vd.Name)
		if vd.Expr != nil {
			v.output(" = ")
			vd.Expr.Visit(v)
		}
	}
	return false
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
}

func (v *Visitor) PreVisitDeclareStmt(stmt *ast.DeclareStmt) bool {
	v.indent()
	if stmt.IsConst {
		v.output("Declare ")
	} else {
		v.output("Constant ")
	}
	v.output(stmt.Decls[0].Type.String())
	v.output(" ")
	for i, decl := range stmt.Decls {
		if i > 0 {
			v.output(", ")
		}
		decl.Visit(v)
	}
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitDeclareStmt(stmt *ast.DeclareStmt) {
}

func (v *Visitor) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	v.indent()
	v.output("Display ")
	for i, arg := range d.Exprs {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitDisplayStmt(d *ast.DisplayStmt) {
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	v.indent()
	v.output("Input ")
	v.output(i.Var.Name)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitInputStmt(i *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(s *ast.SetStmt) bool {
	v.indent()
	v.output("Set ")
	v.output(s.Var.Name)
	v.output(" = ")
	s.Expr.Visit(v)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitSetStmt(s *ast.SetStmt) {
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	v.indent()
	is.If.Visit(v)

	for _, cb := range is.ElseIf {
		v.output("Else ")
		cb.Visit(v)
	}
	if is.Else != nil {
		v.output("Else\n")
		is.Else.Visit(v)
	}
	v.output("End If\n")
	return false
}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {
}

func (v *Visitor) PreVisitCondBlock(cb *ast.CondBlock) bool {
	v.output("If ")
	cb.Expr.Visit(v)
	v.output(" Then\n")
	cb.Block.Visit(v)
	return false
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	v.indent()
	v.output("Select ")
	ss.Expr.Visit(v)
	v.output("\n")

	oldInd := v.ind
	v.ind += "\t"

	for _, cb := range ss.Cases {
		cb.Visit(v)
	}
	if ss.Default != nil {
		v.indent()
		v.output("Default:\n")
		ss.Default.Visit(v)
	}

	v.ind = oldInd
	v.indent()
	v.output("End Select\n")
	return false
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	v.indent()
	v.output("Case ")
	cb.Expr.Visit(v)
	v.output(":\n")
	cb.Block.Visit(v)
	return false
}

func (v *Visitor) PostVisitCaseBlock(cb *ast.CaseBlock) {
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	v.output("Do\n")
	ds.Block.Visit(v)
	if ds.Not {
		v.output("Until ")
	} else {
		v.output("While ")
	}
	ds.Expr.Visit(v)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	v.output("While ")
	ws.Expr.Visit(v)
	v.output("\n")
	ws.Block.Visit(v)
	v.indent()
	v.output("End While\n")
	return false
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	v.output("For ")
	v.output(fs.Var.Name)
	v.output(" = ")
	fs.StartExpr.Visit(v)
	v.output(" To ")
	fs.StopExpr.Visit(v)
	if fs.StepExpr != nil {
		v.output(" Step ")
		fs.StepExpr.Visit(v)
	}
	v.output("\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("End For\n")
	return false
}

func (v *Visitor) PostVisitForStmt(ws *ast.ForStmt) {
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.indent()
	v.output("Call ")
	v.output(cs.Name)
	v.output("(")
	for i, arg := range cs.Args {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.indent()
	v.output("Module ")
	v.output(ms.Name)
	v.output("(")
	for i, param := range ms.Params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
	v.output(")\n")
	ms.Block.Visit(v)
	v.indent()
	v.output("End Module\n")
	return false
}

func (v *Visitor) PreVisitReturnStmt(rs *ast.ReturnStmt) bool {
	v.indent()
	v.output("Return ")
	rs.Expr.Visit(v)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.indent()
	v.output("Function ")
	v.output(fs.Type.String())
	v.output(" ")
	v.output(fs.Name)
	v.output("(")
	for i, param := range fs.Params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
	v.output(")\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("End Function\n")
	return false
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {}

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
		v.output("NOT ")
	case ast.NEG:
		v.output("-")
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
	v.output("(")
	bo.Lhs.Visit(v)
	v.output(" ")
	v.output(bo.Op.String())
	v.output(" ")
	bo.Rhs.Visit(v)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	v.output(ve.Name)
	return true
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	v.output(ce.Name)
	v.output("(")
	for i, arg := range ce.Args {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {
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
