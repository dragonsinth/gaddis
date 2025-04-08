package gogen

import (
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lib"
	"io"
	"slices"
	"strconv"
	"text/template"
)

func GoGenerate(prog *ast.Program, isTest bool) string {
	var sb bytes.Buffer
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
		case *ast.ModuleStmt, *ast.FunctionStmt, *ast.ClassStmt:
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
		case *ast.ModuleStmt, *ast.FunctionStmt, *ast.ClassStmt:
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

type sw interface {
	io.Writer
	io.StringWriter
}

func New(indent string, out sw) *Visitor {
	return &Visitor{
		ind: indent,
		out: out,
	}
}

type Visitor struct {
	ind string
	out sw
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
	if isStringCharAssignment(lhs) {
		ar := lhs.(*ast.ArrayRef)
		// special case string index assignment
		if ar.Qualifier.CanReference() {
			// stringWithCharUpdateRef(str *string, idx int64, c byte)
			v.output("stringWithCharUpdateRef(")
			v.varRef(ar.Qualifier, true)
		} else {
			// eval and ignore: stringWithCharUpdate(str string, idx int64, c byte)
			v.output("stringWithCharUpdate(")
			ar.Qualifier.Visit(v)
		}
		v.output(", ")
		v.maybeCast(ast.Integer, ar.IndexExpr)
		v.output(", ")
		emitRhs()
		v.output(")")
	} else {
		v.varRef(lhs, false) // assignment auto-refs
		v.output(" = ")
		emitRhs()
	}
}

func isStringCharAssignment(expr ast.Expression) bool {
	ar, ok := expr.(*ast.ArrayRef)
	if !ok {
		return false
	}
	return ar.Qualifier.GetType() == ast.String
}

func (v *Visitor) PreVisitOpenStmt(os *ast.OpenStmt) bool {
	v.indent()
	os.File.Visit(v)
	v.output(" = Open" + os.File.GetType().String())
	v.output("(")
	os.File.Visit(v)
	v.output(", ")
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

func (v *Visitor) PreVisitDeleteStmt(ds *ast.DeleteStmt) bool {
	v.indent()
	v.output("DeleteFile(")
	ds.File.Visit(v)
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitDeleteStmt(ds *ast.DeleteStmt) {
}

func (v *Visitor) PreVisitRenameStmt(rs *ast.RenameStmt) bool {
	v.indent()
	v.output("RenameFile(")
	rs.OldFile.Visit(v)
	v.output(", ")
	rs.NewFile.Visit(v)
	v.output(")\n")
	return false
}

func (v *Visitor) PostVisitRenameStmt(rs *ast.RenameStmt) {
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
		if ForInteger(&count_, 1) <= 10 {
			for {
				Display(count_)
				if !StepInteger(&count_, 1) <= 10 {
					break
				}
			}
		}
	*/

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
	v.output("if For")
	v.output(refType.String())
	v.output("(")
	v.varRef(fs.Ref, true)
	v.output(", ")
	v.maybeCast(refType, fs.StartExpr)
	if isNegative {
		v.output(") >= ")
	} else {
		v.output(") <= ")
	}
	v.maybeCast(refType, fs.StopExpr)
	v.output(" {\n")

	v.ind += "\t"
	v.indent()
	v.output("for {\n")
	fs.Block.Visit(v)

	v.ind += "\t"

	v.indent()
	v.output("if Step")
	v.output(refType.String())
	v.output("(")
	v.varRef(fs.Ref, true)
	v.output(", ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	if isNegative {
		v.output(") < ")
	} else {
		v.output(") > ")
	}
	v.maybeCast(refType, fs.StopExpr)
	v.output(" {\n")
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

func (v *Visitor) PreVisitForStmt2(fs *ast.ForStmt) bool {
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
	v.varRef(fs.Ref, false)
	v.output(" = ")
	v.maybeCast(refType, fs.StartExpr)
	v.output("\n")

	v.indent()
	v.output("for Step")
	v.typeName(fs.Ref.GetType())
	v.output("(")
	v.varRef(fs.Ref, true)
	v.output(", ")
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
	} else {
		v.output("1")
	}
	if isNegative {
		v.output(") >= ")
	} else {
		v.output(") <= ")
	}
	v.maybeCast(refType, fs.StopExpr)
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

	if isExternalDeleteInsert(cs) {
		// special case
		v.output(cs.Ref.Name)
		v.output("StringRef(")
		// emit arg 0
		v.varRef(cs.Args[0], true)
		v.output(", ")
		// emit the rest of the argument list
		v.outputArgumentList(cs.Args[1:], cs.Ref.Params[1:])
		v.output(")\n")
		return false
	}

	if cs.Qualifier != nil {
		if cs.Ref.IsConstructor {
			v.maybeCast(cs.Ref.Enclosing, cs.Qualifier)
		} else {
			cs.Qualifier.Visit(v)
		}
		v.output(".")
		if !cs.Ref.IsConstructor {
			v.output("face.")
		}
	}

	if cs.Ref.IsExternal {
		v.output(cs.Name)
	} else {
		v.ident(cs.Ref)
	}
	v.output("(")
	v.outputArgumentList(cs.Args, cs.Ref.Params)
	v.output(")\n")
	return false
}

func isExternalDeleteInsert(cs *ast.CallStmt) bool {
	return cs.Ref.IsExternal && (cs.Name == "delete" || cs.Name == "insert")
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.indent()
	v.output("func ")
	if enc := ms.Enclosing; enc != nil {
		v.output("(this *")
		v.ident(enc)
		v.output(") ")
	}
	v.ident(ms)
	v.output("(")
	v.outputParameterList(ms.Params)
	v.output(")")
	if ms.IsConstructor {
		v.output(" *")
		v.ident(ms.Scope.Parent.ClassStmt)
	}
	v.output(" {\n")
	ms.Block.Visit(v)
	if ms.IsConstructor {
		v.indent()
		v.output("\treturn this\n")
	}
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
	if enc := fs.Enclosing; enc != nil {
		v.output("(this *")
		v.ident(enc)
		v.output(") ")
	}
	v.ident(fs)
	v.output("(")
	v.outputParameterList(fs.Params)
	v.output(") ")
	v.typeName(fs.Type)
	v.output("{\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("}\n")

	return false
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {}

func (v *Visitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	// This gets super, super weird.
	// First, emit the interface def.
	v.output("type I")
	v.output(cs.Name)
	v.output(" interface {\n")
	if cs.Extends != "" {
		v.output("\tI")
		v.output(cs.Extends)
		v.output("\n")
	}
	for _, stmt := range cs.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt:
			if stmt.IsConstructor {
				continue
			}
			v.output("\t")
			v.ident(stmt)
			v.output("(")
			v.outputParameterList(stmt.Params)
			v.output(")\n")
		case *ast.FunctionStmt:
			v.output("\t")
			v.ident(stmt)
			v.output("(")
			v.outputParameterList(stmt.Params)
			v.output(") ")
			v.typeName(stmt.Type)
			v.output("\n")
		}
	}
	v.output("}\n")

	// Next, emit the struct def.
	v.output("type ")
	v.ident(cs)
	v.output(" struct {\n")

	if cs.Extends != "" {
		v.output("\t")
		v.output(cs.Extends)
		v.output("_\n")
	}

	v.output("\tface I")
	v.output(cs.Name)
	v.output("\n")
	for _, stmt := range cs.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.DeclareStmt:
			for _, decl := range stmt.Decls {
				v.output("\t")
				v.ident(decl)
				v.output(" ")
				v.typeName(decl.Type)
				v.output("\n")
			}
		}
	}
	v.output("}\n")

	if cs.Extends == "" {
		if err := rootTmpl.Execute(v.out, struct {
			Shape string
		}{
			Shape: cs.Name,
		}); err != nil {
			panic(err)
		}
	} else {
		if err := subTmpl.Execute(v.out, struct {
			Shape  string
			Circle string
		}{
			Shape:  cs.Extends,
			Circle: cs.Name,
		}); err != nil {
			panic(err)
		}
	}

	// emit array initializers
	for _, field := range cs.Scope.Fields {
		if field.Enclosing != cs.Type {
			continue // skip super fields
		}
		if !field.Type.IsArrayType() {
			continue
		}
		v.output("\tthis.")
		v.ident(field)
		v.output(" = ")
		v.outputArrayInitializer(field.Type.AsArrayType(), field.Dims, nil)
		v.output("\n")
	}

	v.output("\treturn this\n")
	v.output("}\n")

	// Emit toString()
	if err := toStringTmpl.Execute(v.out, struct {
		Shape string
	}{
		Shape: cs.Name,
	}); err != nil {
		panic(err)
	}

	// Emit method bodies
	for _, stmt := range cs.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt, *ast.FunctionStmt:
			stmt.Visit(v)
		}
	}

	return false
}

var rootTmpl = template.Must(template.New("").Parse(`
func New{{.Shape}}(face I{{.Shape}}) *{{.Shape}}_ {
	this := &{{.Shape}}_{face: face}
	if face == nil {
		this.face = this
	}
`))

var subTmpl = template.Must(template.New("").Parse(`
func New{{.Circle}}(face I{{.Circle}}) *{{.Circle}}_ {
	this := &{{.Circle}}_{face: face}
	if face == nil {
		this.face = this
	}
	this.{{.Shape}}_ = *New{{.Shape}}(this.face)
`))

var toStringTmpl = template.Must(template.New("").Parse(`
func (this *{{.Shape}}_) String() string {
	if this == nil { return "<nil>" }
	return "<{{.Shape}}>"
}
`))

func (v *Visitor) PostVisitClassStmt(cs *ast.ClassStmt) {}

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
	return false
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	if ce.Qualifier != nil {
		ce.Qualifier.Visit(v)
		v.output(".face.")
	}

	if ce.Ref.IsExternal {
		if ce.Ref.Name == "append" {
			v.output("appendString") // special case append vice builtin append
		} else {
			v.output(ce.Ref.Name)
		}
	} else {
		v.ident(ce.Ref)
	}
	v.output("(")
	v.outputArgumentList(ce.Args, ce.Ref.Params)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {}

func (v *Visitor) PreVisitArrayRef(ar *ast.ArrayRef) bool {
	// TODO: this won't be sufficient for passing in as a reference param
	ar.Qualifier.Visit(v)
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

func (v *Visitor) PreVisitNewExpr(ne *ast.NewExpr) bool {
	v.output("New")
	v.output(ne.Name)
	v.output("(nil)")
	if ne.Ctor != nil {
		v.output(".")
		v.ident(ne.Ctor)
		v.output("(")
		v.outputArgumentList(ne.Args, ne.Ctor.Params)
		v.output(")")
	}
	return false
}

func (v *Visitor) PostVisitNewExpr(ne *ast.NewExpr) {
}

func (v *Visitor) PreVisitThisRef(ref *ast.ThisRef) bool {
	v.output("this")
	return false
}

func (v *Visitor) PostVisitThisRef(ref *ast.ThisRef) {
}

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
	} else if ast.IsSubclass(dstType, exp.GetType()) {
		v.output("(&")
		exp.Visit(v)
		v.output(".")
		v.ident(dstType.AsClassType())
		v.output(")")
	} else {
		exp.Visit(v)
	}
}

func (v *Visitor) varRef(expr ast.Expression, needRef bool) {
	switch ve := expr.(type) {
	case *ast.VariableExpr:
		v.varRefDecl(ve.Qualifier, ve.Ref, needRef)
	case *ast.ArrayRef:
		if needRef {
			v.output("&")
		}
		expr.Visit(v)
	default:
		panic("implement me")
	}
}

func (v *Visitor) varRefDecl(qual ast.Expression, decl *ast.VarDecl, needRef bool) {
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
		// Local or Field
		needClose := false
		if decl.IsRef == needRef {
			// if we have a ref and need a ref, or we have a val and need a val, we good
		} else if needRef {
			v.output("(")
			v.output("&")
			needClose = true
		} else {
			// Take the value (it's a reference) then dereference it.
			v.output("(")
			v.output("*")
			needClose = true
		}
		if qual != nil {
			qual.Visit(v)
			v.output(".")
		}
		v.ident(decl)
		if needClose {
			v.output(")")
		}
	}
}

func (v *Visitor) outputArgumentList(args []ast.Expression, params []*ast.VarDecl) {
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
}

func (v *Visitor) outputParameterList(params []*ast.VarDecl) {
	for i, param := range params {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
}

func (v *Visitor) typeName(t ast.Type) {
	if t.IsArrayType() {
		v.output("[]")
		v.typeName(t.AsArrayType().ElementType)
		return
	}
	if t.IsClassType() {
		v.output("*")
		v.ident(t.AsClassType())
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
	case ast.OutputFile:
		v.output("OutputFile{}")
	case ast.AppendFile:
		v.output("AppendFile{}")
	case ast.InputFile:
		v.output("InputFile{}")
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
