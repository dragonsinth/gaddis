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
	v := &Visitor{
		ind:      "",
		out:      &sb,
		line:     0,
		col:      0,
		comments: comments,
	}
	for _, stmt := range prog.Block.Statements {
		stmt.Visit(v)
	}
	// handle any trailing comments
	if len(v.comments) > 0 {
		v.emitCommentsUpTo(v.comments[len(v.comments)-1].Start.Line + 1)
	}
	return sb.String()
}

type Visitor struct {
	ind string
	out io.StringWriter

	line     int
	col      int
	comments []ast.Comment
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
		v.output(vd.Type.BaseType().String())
		if vd.IsRef {
			v.output(" Ref")
		}
		v.output(" ")
		v.output(vd.Name)
		for typ := vd.Type; typ.IsArrayType(); typ = typ.AsArrayType().ElementType {
			v.output("[]")
		}
	} else {
		v.output(vd.Name)
		for _, dim := range vd.DimExprs {
			v.output("[")
			dim.Visit(v)
			v.output("]")
		}
	}
	if vd.Expr != nil {
		// special format handling for array initializers
		if len(vd.DimExprs) > 0 {
			v.output(" =")
		} else {
			v.output(" = ")
		}
		vd.Expr.Visit(v)
	}
	return false
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
}

func (v *Visitor) PreVisitDeclareStmt(ds *ast.DeclareStmt) bool {
	v.bol(ds.Start)
	defer v.eol(ds.End)

	if ds.IsConst {
		v.output("Constant ")
	} else if ds.IsPrivate {
		v.output("Private ")
	} else if ds.IsField {
		v.output("Public ")
	} else {
		v.output("Declare ")
	}
	// special hack case
	if ds.Type == ast.AppendFile {
		v.output("OutputFile AppendMode")
	} else {
		v.output(ds.Type.String())
	}
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

	if ds.IsPrint {
		v.output("Print")
	} else {
		v.output("Display")
	}
	v.outputArguments(" ", ds.Exprs)
	return false
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {
}

func (v *Visitor) PreVisitInputStmt(is *ast.InputStmt) bool {
	v.bol(is.Start)
	defer v.eol(is.End)

	v.output("Input ")
	is.Ref.Visit(v)
	return false
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
}

func (v *Visitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	v.bol(ss.Start)
	defer v.eol(ss.End)

	v.output("Set ")
	ss.Ref.Visit(v)
	v.output(" = ")
	ss.Expr.Visit(v)
	return false
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
}

func (v *Visitor) PreVisitOpenStmt(os *ast.OpenStmt) bool {
	v.bol(os.Start)
	defer v.eol(os.End)

	v.output("Open ")
	os.File.Visit(v)
	v.output(" ")
	os.Name.Visit(v)
	return false
}

func (v *Visitor) PostVisitOpenStmt(os *ast.OpenStmt) {}

func (v *Visitor) PreVisitCloseStmt(cs *ast.CloseStmt) bool {
	v.bol(cs.Start)
	defer v.eol(cs.End)

	v.output("Close ")
	cs.File.Visit(v)
	return false
}

func (v *Visitor) PostVisitCloseStmt(cs *ast.CloseStmt) {}

func (v *Visitor) PreVisitReadStmt(rs *ast.ReadStmt) bool {
	v.bol(rs.Start)
	defer v.eol(rs.End)

	v.output("Read ")
	rs.File.Visit(v)
	v.outputArguments(" ", rs.Exprs)
	return false
}

func (v *Visitor) PostVisitReadStmt(rs *ast.ReadStmt) {}

func (v *Visitor) PreVisitWriteStmt(ws *ast.WriteStmt) bool {
	v.bol(ws.Start)
	defer v.eol(ws.End)

	v.output("Write ")
	ws.File.Visit(v)
	v.outputArguments(" ", ws.Exprs)
	return false
}

func (v *Visitor) PostVisitWriteStmt(ws *ast.WriteStmt) {}

func (v *Visitor) PreVisitDeleteStmt(ds *ast.DeleteStmt) bool {
	v.bol(ds.Start)
	defer v.eol(ds.End)

	v.output("Delete ")
	ds.File.Visit(v)
	return false
}

func (v *Visitor) PostVisitDeleteStmt(ds *ast.DeleteStmt) {}

func (v *Visitor) PreVisitRenameStmt(rs *ast.RenameStmt) bool {
	v.bol(rs.Start)
	defer v.eol(rs.End)

	v.output("Rename ")
	rs.OldFile.Visit(v)
	v.output(", ")
	rs.NewFile.Visit(v)
	return false
}

func (v *Visitor) PostVisitRenameStmt(rs *ast.RenameStmt) {}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	v.bol(is.Start)
	defer v.eol(is.End)

	for i, cb := range is.Cases {
		if cb.Expr != nil {
			if i > 0 {
				v.bol(cb.Start)
				v.output("Else ")
			}
			v.output("If ")
			cb.Expr.Visit(v)
			v.output(" Then")
		} else {
			v.bol(cb.Start)
			v.output("Else")
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
	if ds.Until {
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
	fs.Ref.Visit(v)
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

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
}

func (v *Visitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	v.bol(fs.Start)
	defer v.eol(fs.End)

	v.output("For Each ")
	fs.Ref.Visit(v)
	v.output(" In ")
	fs.ArrayExpr.Visit(v)
	v.eol(fs.Start)
	fs.Block.Visit(v)

	v.bol(fs.End)
	v.output("End For")
	return false
}

func (v *Visitor) PostVisitForEachStmt(fs *ast.ForEachStmt) {
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.bol(cs.Start)
	defer v.eol(cs.End)

	v.output("Call ")
	if cs.Qualifier != nil {
		cs.Qualifier.Visit(v)
		v.output(".")
	}
	v.output(cs.Name)
	v.output("(")
	v.outputArguments("", cs.Args)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.bol(ms.Start)
	defer v.eol(ms.End)

	if ms.IsPrivate {
		v.output("Private ")
	} else if ms.Enclosing != nil {
		v.output("Public ")
	}

	v.output("Module ")
	v.output(ms.Name)
	v.output("(")
	v.outputParams(ms.Params)
	v.output(")")
	v.eol(ms.Start)

	ms.Block.Visit(v)

	v.bol(ms.End)
	v.output("End Module")
	return false
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {}

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

	if fs.IsPrivate {
		v.output("Private ")
	} else if fs.Enclosing != nil {
		v.output("Public ")
	}

	v.output("Function ")
	v.output(fs.Type.String())
	v.output(" ")
	v.output(fs.Name)
	v.output("(")
	v.outputParams(fs.Params)
	v.output(")")
	v.eol(fs.Start)

	fs.Block.Visit(v)

	v.bol(fs.End)
	v.output("End Function")
	return false
}

func (v *Visitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	v.bol(cs.Start)
	defer v.eol(cs.End)

	v.output("Class ")
	v.output(cs.Name)
	if cs.Extends != "" {
		v.output(" Extends ")
		v.output(cs.Extends)
	}
	v.eol(cs.Start)

	cs.Block.Visit(v)

	v.bol(cs.End)
	v.output("End Class")
	return false
}

func (v *Visitor) PostVisitClassStmt(cs *ast.ClassStmt) {}

func (v *Visitor) PreVisitLiteral(l *ast.Literal) bool {
	if l.IsTabLiteral {
		v.output("Tab")
		return false
	}
	switch l.Type {
	case ast.Integer:
		v.output(strconv.FormatInt(l.Val.(int64), 10))
	case ast.Real:
		v.output(strconv.FormatFloat(l.Val.(float64), 'f', -1, 64))
	case ast.String:
		v.output(strconv.Quote(l.Val.(string)))
	case ast.Character:
		v.output(strconv.QuoteRune(rune(l.Val.(byte))))
	case ast.Boolean:
		if l.Val.(bool) {
			v.output("True")
		} else {
			v.output("False")
		}
	default:
		panic(l.Type)
	}
	return false
}

func (v *Visitor) PostVisitLiteral(l *ast.Literal) {}

func (v *Visitor) PreVisitParenExpr(pe *ast.ParenExpr) bool {
	v.output("(")
	pe.Expr.Visit(v)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitParenExpr(pe *ast.ParenExpr) {}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	switch uo.Op {
	case ast.NOT:
		v.output("NOT ")
	case ast.NEG:
		v.output("-")
	default:
		panic(uo.Op)
	}
	uo.Expr.Visit(v)
	return false
}

func (v *Visitor) PostVisitUnaryOperation(uo *ast.UnaryOperation) {
}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	bo.Lhs.Visit(v)
	v.output(" ")
	v.output(bo.Op.String())
	v.output(" ")
	bo.Rhs.Visit(v)
	return false
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	if ve.Qualifier != nil {
		ve.Qualifier.Visit(v)
		v.output(".")
	}
	v.output(ve.Name)
	return false
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	if ce.Qualifier != nil {
		ce.Qualifier.Visit(v)
		v.output(".")
	}
	v.output(ce.Name)
	v.output("(")
	v.outputArguments("", ce.Args)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {
}

func (v *Visitor) PreVisitArrayRef(ar *ast.ArrayRef) bool {
	ar.Qualifier.Visit(v)
	v.output("[")
	ar.IndexExpr.Visit(v)
	v.output("]")
	return false
}

func (v *Visitor) PostVisitArrayRef(ar *ast.ArrayRef) {}

func (v *Visitor) PreVisitArrayInitializer(ai *ast.ArrayInitializer) bool {
	lastPos := ai.SourceInfo.Start
	nextPos := ai.Args[0].GetSourceInfo().Start

	indent := false
	if lastPos.Line < nextPos.Line {
		v.eol(lastPos)
		v.ind = v.ind + "\t"
		indent = true
		v.bol(nextPos)
	} else {
		v.output(" ")
	}

	for i, arg := range ai.Args {
		nextPos = arg.GetSourceInfo().Start
		if i > 0 {
			if lastPos.Line < nextPos.Line {
				v.output(",")
				v.eol(lastPos)
				if !indent {
					v.ind = v.ind + "\t"
					indent = true
				}
				v.bol(nextPos)
			} else {
				v.output(", ")
			}
		}
		arg.Visit(v)
		lastPos = nextPos
	}

	if indent {
		v.ind = v.ind[:len(v.ind)-1]
	}
	return false
}

func (v *Visitor) PostArrayInitializer(ai *ast.ArrayInitializer) {}

func (v *Visitor) PreVisitNewExpr(ne *ast.NewExpr) bool {
	v.output("New ")
	v.output(ne.Name)
	v.output("(")
	v.outputArguments("", ne.Args)
	v.output(")")
	return false
}

func (v *Visitor) PostVisitNewExpr(ne *ast.NewExpr) {
}

func (v *Visitor) PreVisitThisRef(ref *ast.ThisRef) bool {
	panic("should not print after type check")
}

func (v *Visitor) PostVisitThisRef(ref *ast.ThisRef) {
}

func (v *Visitor) outputArguments(hdr string, exprs []ast.Expression) {
	if len(exprs) == 0 {
		return
	}
	v.output(hdr)
	for i, expr := range exprs {
		if i > 0 {
			v.output(", ")
		}
		expr.Visit(v)
	}
}

func (v *Visitor) outputParams(parms []*ast.VarDecl) {
	for i, param := range parms {
		if i > 0 {
			v.output(", ")
		}
		param.Visit(v)
	}
}

func (v *Visitor) output(s string) {
	if strings.ContainsFunc(s, func(r rune) bool {
		return r == '\n'
	}) {
		panic("newline in output")
	}
	v.col += len(s)
	_, _ = v.out.WriteString(s)
}

func (v *Visitor) bol(pos ast.Position) {
	v.emitCommentsUpTo(pos.Line)
	_, _ = v.out.WriteString(v.ind)
	v.col += len(v.ind)
}

func (v *Visitor) eol(pos ast.Position) {
	if comment := v.peekComment(); comment != nil {
		if comment.Start.Line < pos.Line {
			panic("should not get here")
		}
		// Trailing line comment
		if comment.Start.Line == pos.Line {
			_, _ = v.out.WriteString(" // ")
			_, _ = v.out.WriteString(commentText(*comment))
			v.popComment()
		}
	}
	if len(v.comments) > 0 && v.comments[0].Start.Line <= pos.Line {
		v.comments = v.comments[1:]
	}
	_, _ = v.out.WriteString("\n")
	v.line++
	v.col = 0
}

func (v *Visitor) emitCommentsUpTo(line int) {
	if v.col != 0 {
		panic(v.col)
	}
	blankLines := 1
	for v.line < line {
		if comment := v.peekComment(); comment != nil {
			// full line comment
			if comment.Start.Line <= v.line {
				_, _ = v.out.WriteString(v.ind)
				_, _ = v.out.WriteString("// ")
				_, _ = v.out.WriteString(commentText(*comment))
				v.popComment()
				blankLines = 0
			}
		}
		if blankLines < 2 {
			_, _ = v.out.WriteString("\n")
		}
		blankLines++
		v.line++
		v.col = 0
	}
}

func (v *Visitor) peekComment() *ast.Comment {
	if len(v.comments) > 0 {
		return &v.comments[0]
	}
	return nil
}

func (v *Visitor) popComment() {
	v.comments = v.comments[1:]
}

func commentText(comment ast.Comment) string {
	text := comment.Text
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimSpace(text)
	return text
}
