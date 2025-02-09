package ast

import (
	"io"
	"strconv"
	"strings"
)

func DebugString(globalBlock *Block) string {
	// visit the statements in the global block
	var sb strings.Builder
	v := NewStringVisitor("", &sb)
	for _, stmt := range globalBlock.Statements {
		stmt.Visit(v)
	}
	return sb.String()
}

func NewStringVisitor(indent string, out io.StringWriter) *StringVisitor {
	return &StringVisitor{
		ind: indent,
		out: out,
	}
}

type StringVisitor struct {
	ind string
	out io.StringWriter
}

var _ Visitor = &StringVisitor{}

func (v *StringVisitor) PreVisitBlock(bl *Block) bool {
	v.ind = v.ind + "\t"
	return true
}

func (v *StringVisitor) PostVisitBlock(bl *Block) {
	v.ind = v.ind[:len(v.ind)-1]
}

func (v *StringVisitor) PreVisitVarDecl(vd *VarDecl) bool {
	v.ident(vd)
	if vd.Expr != nil {
		v.output(" = ")
		vd.Expr.Visit(v)
	}
	return false
}

func (v *StringVisitor) PostVisitVarDecl(vd *VarDecl) {
}

func (v *StringVisitor) PreVisitConstantStmt(stmt *ConstantStmt) bool {
	v.indent()
	v.output("Constant ")
	v.output(typeNames[stmt.Decls[0].Type])
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

func (v *StringVisitor) PostVisitConstantStmt(stmt *ConstantStmt) {
}

func (v *StringVisitor) PreVisitDeclareStmt(stmt *DeclareStmt) bool {
	v.indent()
	v.output("Declare ")
	v.output(typeNames[stmt.Decls[0].Type])
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

func (v *StringVisitor) PostVisitDeclareStmt(stmt *DeclareStmt) {
}

func (v *StringVisitor) PreVisitDisplayStmt(d *DisplayStmt) bool {
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

func (v *StringVisitor) PostVisitDisplayStmt(d *DisplayStmt) {
}

func (v *StringVisitor) PreVisitInputStmt(i *InputStmt) bool {
	v.indent()
	v.output("Input ")
	v.ident(i.Ref)
	v.output("\n")
	return false
}

func (v *StringVisitor) PostVisitInputStmt(i *InputStmt) {
}

func (v *StringVisitor) PreVisitSetStmt(s *SetStmt) bool {
	v.indent()
	v.output("Set ")
	v.ident(s.Ref)
	v.output(" = ")
	s.Expr.Visit(v)
	v.output("\n")
	return false
}

func (v *StringVisitor) PostVisitSetStmt(s *SetStmt) {
}

func (v *StringVisitor) PreVisitIfStmt(is *IfStmt) bool {
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

func (v *StringVisitor) PostVisitIfStmt(is *IfStmt) {
}

func (v *StringVisitor) PreVisitCondBlock(cb *CondBlock) bool {
	v.output("If ")
	cb.Expr.Visit(v)
	v.output(" Then\n")
	cb.Block.Visit(v)
	return false
}

func (v *StringVisitor) PostVisitCondBlock(cb *CondBlock) {
}

func (v *StringVisitor) PreVisitSelectStmt(ss *SelectStmt) bool {
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

func (v *StringVisitor) PostVisitSelectStmt(ss *SelectStmt) {
}

func (v *StringVisitor) PreVisitCaseBlock(cb *CaseBlock) bool {
	v.indent()
	v.output("Case ")
	cb.Expr.Visit(v)
	v.output(":\n")
	cb.Block.Visit(v)
	return false
}

func (v *StringVisitor) PostVisitCaseBlock(cb *CaseBlock) {
}

func (v *StringVisitor) PreVisitDoStmt(ds *DoStmt) bool {
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

func (v *StringVisitor) PostVisitDoStmt(ds *DoStmt) {
}

func (v *StringVisitor) PreVisitWhileStmt(ws *WhileStmt) bool {
	v.output("While ")
	ws.Expr.Visit(v)
	v.output("\n")
	ws.Block.Visit(v)
	v.indent()
	v.output("End While\n")
	return false
}

func (v *StringVisitor) PostVisitWhileStmt(ws *WhileStmt) {
}

func (v *StringVisitor) PreVisitForStmt(fs *ForStmt) bool {
	v.output("For ")
	v.output(fs.Name)
	v.output(" = ")
	fs.StartExpr.Visit(v)
	v.output(" To ")
	fs.StopExpr.Visit(v)
	if fs.Step != nil {
		v.output(" Step ")
		fs.Step.Visit(v)
	}
	v.output("\n")
	fs.Block.Visit(v)
	v.indent()
	v.output("End For\n")
	return false
}

func (v *StringVisitor) PostVisitForStmt(ws *ForStmt) {
}

func (v *StringVisitor) PreVisitIntegerLiteral(il *IntegerLiteral) bool {
	v.output(strconv.FormatInt(il.Val, 10))
	return true
}

func (v *StringVisitor) PostVisitIntegerLiteral(l *IntegerLiteral) {
}

func (v *StringVisitor) PreVisitRealLiteral(rl *RealLiteral) bool {
	v.output(strconv.FormatFloat(rl.Val, 'f', -1, 64))
	return true
}

func (v *StringVisitor) PostVisitRealLiteral(l *RealLiteral) {
}

func (v *StringVisitor) PreVisitStringLiteral(sl *StringLiteral) bool {
	v.output(strconv.Quote(sl.Val))
	return true
}

func (v *StringVisitor) PostVisitStringLiteral(sl *StringLiteral) {
}

func (v *StringVisitor) PreVisitCharacterLiteral(cl *CharacterLiteral) bool {
	v.output(strconv.QuoteRune(rune(cl.Val)))
	return true
}

func (v *StringVisitor) PostVisitCharacterLiteral(cl *CharacterLiteral) {
}

func (v *StringVisitor) PreVisitBooleanLiteral(cl *BooleanLiteral) bool {
	v.output(strconv.FormatBool(cl.Val))
	return true
}

func (v *StringVisitor) PostVisitBooleanLiteral(cl *BooleanLiteral) {
}

func (v *StringVisitor) PreVisitUnaryOperation(uo *UnaryOperation) bool {
	switch uo.Op {
	case NOT:
		v.output("NOT ")
	case NEG:
		v.output("-")
	default:
		panic(uo.Op)
	}
	v.output("(")
	uo.Expr.Visit(v)
	v.output(")")
	return false
}

func (v *StringVisitor) PostVisitUnaryOperation(uo *UnaryOperation) {
}

func (v *StringVisitor) PreVisitBinaryOperation(bo *BinaryOperation) bool {
	v.output("(")
	bo.Lhs.Visit(v)
	v.output(" ")
	v.output(operators[bo.Op])
	v.output(" ")
	bo.Rhs.Visit(v)
	v.output(")")
	return false
}

func (v *StringVisitor) PostVisitBinaryOperation(bo *BinaryOperation) {
}

func (v *StringVisitor) PreVisitVariableExpression(ve *VariableExpression) bool {
	v.ident(ve.Ref)
	return true
}

func (v *StringVisitor) PostVisitVariableExpression(ve *VariableExpression) {
}

func (v *StringVisitor) indent() {
	v.output(v.ind)
}

func (v *StringVisitor) output(s string) {
	_, err := v.out.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (v *StringVisitor) ident(ref *VarDecl) {
	v.output(ref.Name)
}
