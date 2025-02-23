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

func (v *Visitor) PreVisitDeclareStmt(ds *ast.DeclareStmt) bool {
	v.bol(ds.Start)
	defer v.eol(ds.End)

	if ds.IsConst {
		v.output("Declare ")
	} else {
		v.output("Constant ")
	}
	v.output(ds.Decls[0].Type.String())
	v.output(" ")
	for i, decl := range ds.Decls {
		if i > 0 {
			v.output(", ")
		}
		decl.Visit(v)
	}
	return false
}

func (v *Visitor) PostVisitDeclareStmt(ds *ast.DeclareStmt) {
}

func (v *Visitor) PreVisitDisplayStmt(ds *ast.DisplayStmt) bool {
	v.bol(ds.Start)
	defer v.eol(ds.End)

	v.output("Display ")
	for i, arg := range ds.Exprs {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	return false
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {
}

func (v *Visitor) PreVisitInputStmt(is *ast.InputStmt) bool {
	v.bol(is.Start)
	defer v.eol(is.End)

	v.output("Input ")
	v.output(is.Var.Name)
	return false
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	v.bol(ss.Start)
	defer v.eol(ss.End)

	v.output("Set ")
	v.output(ss.Var.Name)
	v.output(" = ")
	ss.Expr.Visit(v)
	return false
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	v.bol(is.Start)
	defer v.eol(is.End)

	for i, cb := range is.Cases {
		if i > 0 {
			v.output("Else")
		}
		if cb.Expr != nil {
			if i > 0 {
				v.output(" ")
			}
			v.output("If ")
			cb.Expr.Visit(v)
			v.output(" Then")
		}
		v.eol(cb.Start)
		cb.Block.Visit(v)
	}
	v.bol(is.End)
	v.output("End If")
	return false
}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {
}

func (v *Visitor) PreVisitCondBlock(cb *ast.CondBlock) bool {
	return false
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	v.bol(ss.Start)
	defer v.eol(ss.End)

	v.output("Select ")
	ss.Expr.Visit(v)
	v.eol(ss.Start)

	oldInd := v.ind
	v.ind += "\t"

	for _, cb := range ss.Cases {
		cb.Visit(v)
	}

	v.ind = oldInd
	v.bol(ss.End)
	v.output("End Select")
	return false
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	v.bol(cb.Start)
	if cb.Expr != nil {
		v.output("Case ")
		cb.Expr.Visit(v)
	} else {
		v.output("Default")
	}
	v.output(":")
	v.eol(cb.Start)
	cb.Block.Visit(v)
	return false
}

func (v *Visitor) PostVisitCaseBlock(cb *ast.CaseBlock) {
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	v.bol(ds.Start)
	defer v.eol(ds.End)

	v.output("Do")
	v.eol(ds.Start)

	ds.Block.Visit(v)
	v.bol(ds.End)
	if ds.Not {
		v.output("Until ")
	} else {
		v.output("While ")
	}
	ds.Expr.Visit(v)
	return false
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	v.bol(ws.Start)
	defer v.eol(ws.End)

	v.output("While ")
	ws.Expr.Visit(v)
	v.eol(ws.Start)

	ws.Block.Visit(v)

	v.bol(ws.End)
	v.output("End While")
	return false
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	v.bol(fs.Start)
	defer v.eol(fs.End)

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
	v.eol(fs.Start)
	fs.Block.Visit(v)

	v.bol(fs.End)
	v.output("End For")
	return false
}

func (v *Visitor) PostVisitForStmt(ws *ast.ForStmt) {
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.bol(cs.Start)
	defer v.eol(cs.End)

	v.output("Call ")
	v.output(cs.Name)
	v.output("(")
	for i, arg := range cs.Args {
		if i > 0 {
			v.output(", ")
		}
		arg.Visit(v)
	}
	v.output(")")
	return false
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.bol(ms.Start)
	defer v.eol(ms.End)

	v.output("Module ")
	v.output(ms.Name)
	v.output("(")
	for i, param := range ms.Params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
	v.output(")")
	v.eol(ms.Start)

	ms.Block.Visit(v)

	v.bol(ms.End)
	v.output("End Module")
	return false
}

func (v *Visitor) PreVisitReturnStmt(rs *ast.ReturnStmt) bool {
	v.bol(rs.Start)
	defer v.eol(rs.End)

	v.output("Return ")
	rs.Expr.Visit(v)
	return false
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.bol(fs.Start)
	defer v.eol(fs.End)

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
	v.output(")")
	v.eol(fs.Start)

	fs.Block.Visit(v)

	v.bol(fs.End)
	v.output("End Function")
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
	v.output(")")
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

func (v *Visitor) bol(node ast.Position) {
	// TODO: extra newlines to catch up.
	v.indent()
}

func (v *Visitor) eol(node ast.Position) {
	// TODO: trailing comment?
	v.output("\n")
}
