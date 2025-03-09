package gogen

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lib"
	"io"
	"slices"
	"strconv"
	"strings"
)

func GoGenerate(prog *ast.Program, isTest bool) string {
	var sb strings.Builder
	sb.WriteString("package main\n")

	var imports, code []string
	parseGoCode(lib.Source, &imports, &code)
	parseGoCode(builtins, &imports, &code)

	slices.Sort(imports)
	imports = slices.Compact(imports)
	for _, imp := range imports {
		sb.WriteString(imp)
		sb.WriteByte('\n')
	}
	for _, c := range code {
		sb.WriteString(c)
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')

	v := New("", &sb)

	if isTest {
		// Lock the rng in test mode.
		v.output("func init() { rng.Seed(0) }\n\n")
	}

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
				v.typeName(decl.Type)
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
		v.ident(vd)
		v.output(" ")
		if vd.IsRef {
			v.output("*")
		}
		v.typeName(vd.Type)
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
	v.typeName(vd.Type)
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
	v.output("Display(")
	for i, arg := range d.Exprs {
		if i > 0 {
			v.output(", ")
		}
		if lit, ok := arg.(*ast.Literal); ok && lit.IsTabLiteral {
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
	v.varRef(i.Var, false)
	v.output(" = ")
	v.output(" Input")
	v.output(i.Var.Type.String())
	v.output("()\n")
	return false
}

func (v *Visitor) PostVisitInputStmt(i *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(s *ast.SetStmt) bool {
	v.indent()
	v.varRef(s.Var, false)
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

	// string equality
	typ := ss.Type
	if typ == ast.String {
		typ = goStringType
	}

	v.output("switch (")
	v.maybeCast(typ, ss.Expr)
	v.output(") {\n")

	v.ind += "\t"
	for _, cb := range ss.Cases {
		v.indent()
		if cb.Expr != nil {
			v.output("case ")
			v.maybeCast(typ, cb.Expr)
			v.output(":\n")
		} else {
			v.output("default:\n")
		}
		cb.Block.Visit(v)
	}
	v.ind = v.ind[:len(v.ind)-1]
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
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
	if ds.Until {
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
	if val := ws.Expr.ConstEval(); val == nil || !val.(bool) {
		ws.Expr.Visit(v)
	}
	v.output(" {\n")
	ws.Block.Visit(v)
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	/*
		if ForInteger(&count_, 1, 10, 1) {
			for {
				Display(count_)
				if !StepInteger(&count_, 10, 1) {
					break
				}
			}
		}
	*/

	refType := fs.Var.Type
	v.indent()
	v.output("if For")
	v.output(refType.String())
	v.output("(")
	v.varRef(fs.Var, true)
	v.output(", ")
	fs.StartExpr.Visit(v)
	v.output(", ")
	v.maybeCast(refType, fs.StopExpr)
	v.output(", ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	v.output(") {\n")

	v.ind += "\t"
	v.indent()
	v.output("for {\n")
	fs.Block.Visit(v)

	v.ind += "\t"

	v.indent()
	v.output("if !Step")
	v.output(refType.String())
	v.output("(")
	v.varRef(fs.Var, true)
	v.output(", ")
	v.maybeCast(refType, fs.StopExpr)
	v.output(", ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	v.output(") {\n")
	v.indent()
	v.output("\tbreak\n")
	v.indent()
	v.output("}\n")
	v.ind = v.ind[:len(v.ind)-1]
	v.indent()
	v.output("}\n")

	v.ind = v.ind[:len(v.ind)-1]
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
	v.typeName(fs.Type)
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

func (v *Visitor) PreVisitLiteral(l *ast.Literal) bool {
	switch l.Type {
	case ast.Integer:
		v.output(strconv.FormatInt(l.Val.(int64), 10))
	case ast.Real:
		v.output(strconv.FormatFloat(l.Val.(float64), 'f', -1, 64))
	case ast.String:
		v.output("String(")
		v.output(strconv.Quote(l.Val.(string)))
		v.output(")")
	case ast.Character:
		v.output(strconv.QuoteRune(rune(l.Val.(byte))))
	case ast.Boolean:
		v.output(strconv.FormatBool(l.Val.(bool)))
	default:
		panic(l.Type)
	}
	return false
}

func (v *Visitor) PostVisitLiteral(l *ast.Literal) {}

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
	argType := bo.ArgType
	if argType == ast.String {
		// force byte[] to string for binary operations
		argType = goStringType
	}

	// must special case exp and mod
	if bo.Op == ast.MOD || bo.Op == ast.EXP {
		if bo.Op == ast.MOD {
			v.output("Mod")
		} else {
			v.output("Exp")
		}
		v.output(bo.Type.String())
		v.output("(")
		v.maybeCast(argType, bo.Lhs)
		v.output(", ")
		v.maybeCast(argType, bo.Rhs)
		v.output(")")
	} else {
		v.output("(")
		v.maybeCast(argType, bo.Lhs)
		v.output(" ")
		v.output(goBinaryOperators[bo.Op])
		v.output(" ")
		v.maybeCast(argType, bo.Rhs)
		v.output(")")
	}
	return false
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	v.varRef(ve, false) // if we get here, we need a value
	return true
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	if ce.Ref.IsExternal {
		if ce.Ref.Name == "append" {
			v.output("appendString") // special case append vice builtin append
		} else {
			v.output(ce.Ref.Name)
		}
	} else {
		v.ident(ce.Ref)
	}
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
	} else if dstType == goStringType && exp.GetType() == ast.String {
		v.output("string(")
		exp.Visit(v)
		v.output(")")
	} else {
		exp.Visit(v)
	}
}

func (v *Visitor) varRef(expr *ast.VariableExpr, needRef bool) {
	v.varRefDecl(expr.Ref, needRef)
}

func (v *Visitor) varRefDecl(decl *ast.VarDecl, needRef bool) {
	isRef := decl.IsRef
	if decl.IsConst {
		if isRef || needRef {
			panic("here")
		}
		val := decl.Expr.ConstEval()
		if val == nil {
			panic("here")
		}
		v.ident(decl)
	} else if decl.Scope.IsGlobal {
		if isRef {
			panic("here")
		}
		if needRef {
			v.output("&")
			v.ident(decl)
		} else {
			v.ident(decl)
		}
		return
	} else {
		// Local
		if decl.IsRef == needRef {
			// if we have a ref and need a ref, or we have a val and need a val, we good
			v.ident(decl)
		} else if needRef {
			v.output("&")
			v.ident(decl)
		} else {
			// Take the value (it's a reference) then derefence it.
			v.output("*")
			v.ident(decl)
		}
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
			v.varRef(arg.(*ast.VariableExpr), true)
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
	v.output(")")
}

func (v *Visitor) typeName(t ast.Type) {
	if !t.IsPrimitive() {
		panic("TODO: implement non-primitive types")
	}
	v.output(goTypes[t.AsPrimitive()])
}

var goTypes = [...]string{
	ast.Integer:   "int64",
	ast.Real:      "float64",
	ast.String:    "String",
	ast.Character: "byte",
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
