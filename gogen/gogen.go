package gogen

import (
	"fmt"
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
	parseGoCode(lib.IoSource, &imports, &code)
	parseGoCode(lib.LibSource, &imports, &code)
	parseGoCode(builtins, &imports, &code)

	slices.Sort(imports)
	imports = slices.Compact(imports)
	sb.WriteString("import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&sb, "\t%q\n", imp)
	}
	sb.WriteString(")\n")
	for _, c := range code {
		sb.WriteString(c)
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')

	v := New("", &sb)

	if isTest {
		// Lock the rng in test mode.
		v.output("func init() { randCtx.Rng.Seed(0) }\n\n")
	}

	// First, iterate the global block and emit any declarations.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.DeclareStmt:
			// emit declaration only, not assignment
			for _, decl := range stmt.Decls {
				if decl.IsConst {
					continue
				}
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
				if decl.IsConst {
					continue
				}
				if decl.Expr != nil {
					v.indent()
					v.ident(decl)
					v.output(" = ")
					v.maybeCast(decl.Type, decl.Expr)
					v.output("\n")
				} else if len(decl.Dims) > 0 {
					v.indent()
					v.ident(decl)
					v.output(" = ")
					v.outputArrayInitializer(decl.Type.AsArrayType(), decl.Dims, nil)
					v.output("\n")
				}
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
	if vd.IsConst {
		return false
	}
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
	v.output("var ")
	v.ident(vd)
	v.output(" ")
	v.typeName(vd.Type)
	if vd.Expr != nil {
		v.output(" = ")
		v.maybeCast(vd.Type, vd.Expr)
	} else if len(vd.Dims) > 0 {
		v.output(" = ")
		v.outputArrayInitializer(vd.Type.AsArrayType(), vd.Dims, []ast.Expression{})
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
	v.emitAssignment(i.Ref, func() {
		v.output(" Input")
		v.output(i.Ref.GetType().AsPrimitive().String())
		v.output("()")
	})
	v.output("\n")
	return false
}

func (v *Visitor) PostVisitInputStmt(i *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(s *ast.SetStmt) bool {
	v.indent()
	v.emitAssignment(s.Ref, func() {
		v.maybeCast(s.Ref.GetType(), s.Expr)
	})
	v.output("\n")
	return false
}

func (v *Visitor) emitAssignment(lhs ast.Expression, emitRhs func()) {
	if ar, ok := lhs.(*ast.ArrayRef); ok && lhs.GetType() == ast.Character {
		// special case string index assignment
		if ar.RefExpr.CanReference() {
			v.varRef(ar.RefExpr, false)
			v.output(" = ")
		}
		// stringWithCharUpdate(c byte, idx int64, str string) string
		v.output("stringWithCharUpdate(")
		emitRhs()
		v.output(", ")
		v.maybeCast(ast.Integer, ar.IndexExpr)
		v.output(", ")
		ar.RefExpr.Visit(v)
		v.output(")")
	} else {
		v.varRef(lhs, false) // assignment auto-refs
		v.output(" = ")
		emitRhs()
	}
}

func (v *Visitor) PreVisitOpenStmt(os *ast.OpenStmt) bool {
	v.indent()
	os.File.Visit(v)
	v.output(" = Open" + os.File.GetType().String())
	v.output("(")
	os.Name.Visit(v)
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitOpenStmt(os *ast.OpenStmt) {}

func (v *Visitor) PreVisitCloseStmt(cs *ast.CloseStmt) bool {
	v.indent()
	if cs.File.GetType() == ast.InputFile {
		v.output("CloseInputFile(")
	} else {
		v.output("CloseOutputFile(")
	}
	cs.File.Visit(v)
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitCloseStmt(cs *ast.CloseStmt) {}

func (v *Visitor) PreVisitReadStmt(rs *ast.ReadStmt) bool {
	for _, expr := range rs.Exprs {
		v.indent()
		v.varRef(expr, false)
		v.output(" = Read")
		v.typeName(expr.GetType().AsPrimitive())
		v.output("(")
		rs.File.Visit(v)
		v.output(")\n")
	}
	return false
}

func (v *Visitor) PostVisitReadStmt(rs *ast.ReadStmt) {}

func (v *Visitor) PreVisitWriteStmt(ws *ast.WriteStmt) bool {
	v.indent()
	v.output("WriteFile(")
	ws.File.Visit(v)
	for _, expr := range ws.Exprs {
		v.output(", ")
		expr.Visit(v)
	}
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitWriteStmt(ws *ast.WriteStmt) {}

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
	isNegative := false
	if fs.StepExpr != nil {
		stepLit := fs.StepExpr.ConstEval()
		switch val := stepLit.(type) {
		case int64:
			isNegative = val < 0
		case float64:
			isNegative = val < 0
		default:
			panic(stepLit)
		}
	}

	refType := fs.Ref.GetType()
	v.indent()
	v.output("for ")
	v.varRef(fs.Ref, false)
	v.output(" = ")
	v.maybeCast(refType, fs.StartExpr)
	v.output("; ")
	v.varRef(fs.Ref, false)
	if isNegative {
		v.output(" >= ")
	} else {
		v.output(" <= ")
	}
	v.maybeCast(refType, fs.StopExpr)
	v.output("; ")
	v.varRef(fs.Ref, false)
	if fs.StepExpr != nil {
		v.output(" += ")
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("++")
	}
	v.output(" {\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
}

func (v *Visitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	v.indent()
	v.output("for _, ")
	v.varRef(fs.Ref, false)
	v.output(" = range ")
	fs.ArrayExpr.Visit(v)
	v.output(" {\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("}\n")
	return false
}

func (v *Visitor) PostVisitForEachStmt(fs *ast.ForEachStmt) {
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.indent()
	if cs.Ref.IsExternal {
		name := cs.Ref.Name
		if name == "delete" || name == "insert" {
			// special case!
			name = name + "String"
			v.varRef(cs.Args[0], false)
			v.output(" = ")
		}
		v.output(name)
	} else {
		v.ident(cs.Ref)
	}
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

	return false
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

func (v *Visitor) PreVisitLiteral(l *ast.Literal) bool {
	switch l.Type {
	case ast.Integer:
		v.output("Integer(")
		v.output(strconv.FormatInt(l.Val.(int64), 10))
		v.output(")")
	case ast.Real:
		v.output("Real(")
		v.output(strconv.FormatFloat(l.Val.(float64), 'f', -1, 64))
		v.output(")")
	case ast.String:
		v.output(strconv.Quote(l.Val.(string)))
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

func (v *Visitor) PreVisitArrayRef(ar *ast.ArrayRef) bool {
	// TODO: this won't be sufficient for passing in as a reference param
	ar.RefExpr.Visit(v)
	v.output("[")
	ar.IndexExpr.Visit(v)
	v.output("]")
	return false
}

func (v *Visitor) PostVisitArrayRef(ar *ast.ArrayRef) {}

func (v *Visitor) PreVisitArrayInitializer(ai *ast.ArrayInitializer) bool {
	v.outputArrayInitializer(ai.Type, ai.Dims, ai.Args)
	return false
}

func (v *Visitor) PostArrayInitializer(ai *ast.ArrayInitializer) {}

func (v *Visitor) outputArrayInitializer(t *ast.ArrayType, dims []int, exprs []ast.Expression) []ast.Expression {
	v.typeName(t)
	v.output("{")
	if len(dims) == 1 {
		// []Integer{1, 2, 3, 4, 5}
		typ := t.BaseType()
		for i := 0; i < dims[0]; i++ {
			if i > 0 {
				v.output(", ")
			}
			if len(exprs) > 0 {
				v.maybeCast(typ, exprs[0])
				exprs = exprs[1:]
			} else {
				v.zero(typ)
			}
		}
	} else {
		// [][]Integer{[]Integer{}}
		for i := 0; i < dims[0]; i++ {
			if i > 0 {
				v.output(", ")
			}
			exprs = v.outputArrayInitializer(t.ElementType.AsArrayType(), dims[1:], exprs)
		}
	}
	v.output("}")
	return exprs
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

func (v *Visitor) ident(named ast.HasName) {
	v.output(named.GetName())
	v.output("_")
}

func (v *Visitor) outputLiteral(typ ast.PrimitiveType, val any) {
	switch typ {
	case ast.Integer:
		v.output(strconv.FormatInt(val.(int64), 10))
	case ast.Real:
		v.output(strconv.FormatFloat(val.(float64), 'f', -1, 64))
	case ast.String:
		v.output(strconv.Quote(val.(string)))
	case ast.Character:
		v.output(strconv.QuoteRune(rune(val.(byte))))
	case ast.Boolean:
		v.output(strconv.FormatBool(val.(bool)))
	default:
		panic(typ)
	}
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	if dstType == ast.Real && exp.GetType() == ast.Integer {
		v.output("Real(")
		exp.Visit(v)
		v.output(")")
	} else if dstType == ast.Integer && exp.GetType() == ast.Real {
		v.output("Integer(")
		exp.Visit(v)
		v.output(")")
	} else {
		exp.Visit(v)
	}
}

func (v *Visitor) varRef(expr ast.Expression, needRef bool) {
	switch ve := expr.(type) {
	case *ast.VariableExpr:
		v.varRefDecl(ve.Ref, needRef)
	case *ast.ArrayRef:
		if needRef {
			v.output("&")
		}
		expr.Visit(v)
	default:
		panic("implement me")
	}
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
		v.outputLiteral(decl.Type.AsPrimitive(), val)
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
			v.output("(")
			v.output("&")
			v.ident(decl)
			v.output(")")
		} else {
			// Take the value (it's a reference) then dereference it.
			v.output("(")
			v.output("*")
			v.ident(decl)
			v.output(")")
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
			v.varRef(arg, true)
		} else if param.Type.IsArrayType() {
			// deep clone the array
			v.output("Clone(")
			arg.Visit(v)
			v.output(")")
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
	v.output(")")
}

func (v *Visitor) typeName(t ast.Type) {
	if t.IsArrayType() {
		v.output("[]")
		v.typeName(t.AsArrayType().ElementType)
		return
	}
	if !t.IsPrimitive() && !t.IsFileType() {
		panic("here")
	}
	v.output(t.String())
}

func (v *Visitor) zero(typ ast.Type) {
	switch typ {
	case ast.Integer, ast.Real, ast.Character:
		v.output("0")
	case ast.Boolean:
		v.output("false")
	case ast.String:
		v.output("\"\"")
	default:
		v.output("nil")
	}
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
