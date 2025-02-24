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

func GoGenerate(prog *ast.Program) string {
	var sb strings.Builder

	data := strings.TrimPrefix(builtins, oldPrefix)
	sb.WriteString(newPrefix)
	sb.WriteString(data)
	v := New("", &sb)

	// First, iterate the global block and emit any declarations.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.DeclareStmt:
			// emit declaration only, not assignment
			for _, decl := range stmt.Decls {
				v.indent()
				v.output("var ")
				v.ident(decl)
				v.output(" ")
				v.output(goTypes[decl.Type])
				v.output("\n")
			}
		case *ast.ModuleStmt, *ast.FunctionStmt:
			stmt.Visit(v)
		default:
			// nothing
		}
	}

	// Now, emit any non-declarations into the main function.
	sb.WriteString("\nfunc main() {\n")
	v.PreVisitBlock(prog.Block)
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.DeclareStmt:
			// emit either an assignment or a dummy assignment
			for _, decl := range stmt.Decls {
				v.indent()
				if decl.Expr != nil {
					v.ident(decl)
					v.output(" = ")
					v.maybeCast(decl.Type, decl.Expr)
				} else {
					v.output("_ = ")
					v.ident(decl)
				}
				v.output("\n")
			}
		case *ast.ModuleStmt, *ast.FunctionStmt:
			// nothing
		default:
			stmt.Visit(v)
		}
	}
	// If there is a module named main with no arguments, call it at the very end.
	ref := prog.Scope.Lookup("main")
	if ref != nil && ref.ModuleStmt != nil && len(ref.ModuleStmt.Params) == 0 {
		v.indent()
		v.output("main_()\n")
	}
	v.PostVisitBlock(prog.Block)
	v.output("}\n")
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
	if vd.IsParam {
		v.ident(vd)
		v.output(" ")
		if vd.IsRef {
			v.output("*")
		}
		v.output(goTypes[vd.Type])
		return false
	}
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
	v.output("var _ = ")
	v.ident(vd)
	v.output("\n")

	return false
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
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
		if _, ok := arg.(*ast.TabLiteral); ok {
			v.output("TabDisplay")
		} else {
			arg.Visit(v)
		}
	}
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitDisplayStmt(d *ast.DisplayStmt) {
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	v.indent()
	i.Var.Visit(v)
	v.output(" = ")
	v.output("input")
	v.output(i.Var.Type.String())
	v.output("()\n")
	return false
}

func (v *Visitor) PostVisitInputStmt(i *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(s *ast.SetStmt) bool {
	v.indent()
	s.Var.Visit(v)
	v.output(" = ")
	v.maybeCast(s.Var.Type, s.Expr)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitSetStmt(s *ast.SetStmt) {
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	v.indent()

	for i, cb := range is.Cases {
		if i > 0 {
			v.output(" else ")
		}
		if cb.Expr != nil {
			v.output("if ")
			cb.Expr.Visit(v)
			v.output(" {\n")
		} else {
			v.output("{\n")
		}
		cb.Block.Visit(v)
		v.indent()
		v.output("}")
	}
	v.output("\n")
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

	v.ind = oldInd
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	v.indent()
	if cb.Expr != nil {
		v.output("case ")
		v.maybeCast(v.selectType, cb.Expr)
		v.output(":\n")
	} else {
		v.output("default:\n")
	}
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
	fs.Var.Visit(v)
	v.output(" = ")
	fs.StartExpr.Visit(v)
	v.output("\n")

	v.indent()
	v.output("for step")
	refType := fs.Var.Type
	v.output(refType.String())
	v.output("(")
	fs.Var.Visit(v)
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
	fs.Var.Visit(v)
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

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.indent()
	v.ident(cs.Ref)
	v.outputArguments(cs.Args, cs.Ref.Params)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.indent()
	v.output("func ")
	v.ident(ms)
	v.output("(")
	for i, param := range ms.Params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
	v.output(") {\n")
	ms.Block.Visit(v)
	v.indent()
	v.output("}\n")

	// Also emit a no-op assignment to avoid Go unreferenced variable errors.
	v.indent()
	v.output("var _ = ")
	v.ident(ms)
	v.output("\n")

	return false
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {}

func (v *Visitor) PreVisitReturnStmt(rs *ast.ReturnStmt) bool {
	v.indent()
	v.output("return ")
	v.maybeCast(rs.Ref.Type, rs.Expr)
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.indent()
	v.output("func ")
	v.ident(fs)
	v.output("(")
	for i, param := range fs.Params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
	v.output(") ")
	v.output(goTypes[fs.Type])
	v.output("{\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("}\n")

	// Also emit a no-op assignment to avoid Go unreferenced variable errors.
	v.indent()
	v.output("var _ = ")
	v.ident(fs)
	v.output("\n")

	return false
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

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

func (v *Visitor) PreVisitTabLiteral(tl *ast.TabLiteral) bool {
	v.output("Tab")
	return false
}

func (v *Visitor) PostVisitTabLiteral(tl *ast.TabLiteral) {}

func (v *Visitor) PreVisitBooleanLiteral(cl *ast.BooleanLiteral) bool {
	v.output(strconv.FormatBool(cl.Val))
	return true
}

func (v *Visitor) PostVisitBooleanLiteral(cl *ast.BooleanLiteral) {
}

func (v *Visitor) PreVisitParenExpr(pe *ast.ParenExpr) bool {
	// unary/binary operation emit parens to force order of operations
	return true
}

func (v *Visitor) PostVisitParenExpr(pe *ast.ParenExpr) {}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	switch uo.Op {
	case ast.NOT:
		v.output("!")
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
	dstType := ast.AreComparableTypes(bo.Lhs.GetType(), bo.Rhs.GetType())

	// must special case exp and mod
	if bo.Op == ast.MOD || bo.Op == ast.EXP {
		if bo.Op == ast.MOD {
			v.output("mod")
		} else {
			v.output("exp")
		}
		v.output(bo.Type.String())
		v.output("(")
		v.maybeCast(dstType, bo.Lhs)
		v.output(", ")
		v.maybeCast(dstType, bo.Rhs)
		v.output(")")
	} else {
		v.output("(")
		v.maybeCast(dstType, bo.Lhs)
		v.output(" ")
		v.output(goBinaryOperators[bo.Op])
		v.output(" ")
		v.maybeCast(dstType, bo.Rhs)
		v.output(")")
	}
	return false
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	if ve.Ref.IsRef {
		// if we get here, we need to dereference
		v.output("*")
	}
	v.ident(ve.Ref)
	return true
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	v.indent()
	v.ident(ce.Ref)
	v.outputArguments(ce.Args, ce.Ref.Params)
	return false
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {}

func (v *Visitor) indent() {
	v.output(v.ind)
}

func (v *Visitor) output(s string) {
	_, err := v.out.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (v *Visitor) ident(named ast.HasName) {
	v.output(named.GetName())
	v.output("_")
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	if dstType == ast.Real && exp.GetType() == ast.Integer {
		v.output("float64(")
		exp.Visit(v)
		v.output(")")
	} else if dstType == ast.Integer && exp.GetType() == ast.Real {
		v.output("int64(")
		exp.Visit(v)
		v.output(")")
	} else {
		exp.Visit(v)
	}
}

func (v *Visitor) outputArguments(args []ast.Expression, params []*ast.VarDecl) {
	v.output("(")
	for i, arg := range args {
		if i > 0 {
			v.output(", ")
		}
		param := params[i]
		if param.IsRef {
			// special case
			// TODO: other types of references
			ref := arg.(*ast.VariableExpr).Ref
			if !ref.IsRef {
				// take the address
				v.output("&")
			}
			v.ident(ref)
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
	v.output(")")
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
