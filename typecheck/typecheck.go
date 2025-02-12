package typecheck

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

func TypeCheck(globalBlock *ast.Block) []ast.Error {
	// visit the statements in the global block
	v := New()
	for _, stmt := range globalBlock.Statements {
		stmt.Visit(v)
	}
	slices.SortFunc(v.errors, func(a, b ast.Error) int {
		return b.Start.Pos - a.Start.Pos
	})
	return v.errors
}

func New() *Visitor {
	return &Visitor{}
}

type Visitor struct {
	errors []ast.Error
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitBlock(bl *ast.Block) bool {
	return true
}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {
}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	return true
}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
}

func (v *Visitor) PreVisitConstantStmt(cs *ast.ConstantStmt) bool {
	return true
}

func (v *Visitor) PostVisitConstantStmt(cs *ast.ConstantStmt) {
	for _, vd := range cs.Decls {
		if !ast.CanCoerce(cs.Type, vd.Expr.Type()) {
			v.Errorf(vd.Expr, "%s not assignable to %s", vd.Expr.Type(), cs.Type)
		}
		// TODO: initializer must be a constant expression?
	}
}

func (v *Visitor) PreVisitDeclareStmt(ds *ast.DeclareStmt) bool {
	return true
}

func (v *Visitor) PostVisitDeclareStmt(ds *ast.DeclareStmt) {
	for _, vd := range ds.Decls {
		if vd.Expr != nil {
			if !ast.CanCoerce(ds.Type, vd.Expr.Type()) {
				v.Errorf(vd.Expr, "%s not assignable to %s", vd.Expr.Type(), ds.Type)
			}
		}
	}
}

func (v *Visitor) PreVisitDisplayStmt(ds *ast.DisplayStmt) bool {
	return true
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {}

func (v *Visitor) PreVisitInputStmt(is *ast.InputStmt) bool {
	return true
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {}

func (v *Visitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	return true
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
	exprType := ss.Expr.Type()
	refType := ss.Ref.Type
	if !ast.CanCoerce(refType, exprType) {
		v.Errorf(ss.Expr, "%s not assignable to %s", exprType, refType)
	}
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	return true
}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {}

func (v *Visitor) PreVisitCondBlock(cb *ast.CondBlock) bool {
	return true
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {
	if cb.Expr.Type() != ast.Boolean {
		v.Errorf(cb.Expr, "expected Boolean, got %s", cb.Expr.Type())
	}
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	return true
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
	dstType := ss.Expr.Type()
	for _, cb := range ss.Cases {
		typ := ast.AreComparableTypes(dstType, cb.Expr.Type())
		if typ == ast.UnresolvedType {
			v.Errorf(cb.Expr, "case %s not comparable to select %s", cb.Expr.Type(), ss.Expr.Type())
		} else {
			dstType = typ
		}
		cb.Visit(v)
	}
	ss.Type = dstType
}

func (v *Visitor) PreVisitCaseBlock(cb *ast.CaseBlock) bool {
	return true
}

func (v *Visitor) PostVisitCaseBlock(cb *ast.CaseBlock) {
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	return true
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
	if ds.Expr.Type() != ast.Boolean {
		v.Errorf(ds.Expr, "expected Boolean, got %s", ds.Expr.Type())
	}
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	return true
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
	if ws.Expr.Type() != ast.Boolean {
		v.Errorf(ws.Expr, "expected Boolean, got %s", ws.Expr.Type())
	}
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	return true
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
	refType := fs.Ref.Type
	if !ast.IsNumericType(refType) {
		v.Errorf(fs, "loop variable must be a number, got %s %s", refType, fs.Ref.Name)
	}
	if !ast.CanCoerce(refType, fs.StartExpr.Type()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StartExpr.Type(), refType)
	}
	if !ast.CanCoerce(refType, fs.StopExpr.Type()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StopExpr.Type(), refType)
	}
	if fs.StepExpr != nil {
		if !ast.CanCoerce(refType, fs.StepExpr.Type()) {
			v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StepExpr.Type(), refType)
		}
	}
}

func (v *Visitor) PreVisitIntegerLiteral(il *ast.IntegerLiteral) bool {
	return true
}

func (v *Visitor) PostVisitIntegerLiteral(il *ast.IntegerLiteral) {}

func (v *Visitor) PreVisitRealLiteral(rl *ast.RealLiteral) bool {
	return true
}

func (v *Visitor) PostVisitRealLiteral(rl *ast.RealLiteral) {}

func (v *Visitor) PreVisitStringLiteral(sl *ast.StringLiteral) bool {
	return true
}

func (v *Visitor) PostVisitStringLiteral(sl *ast.StringLiteral) {}

func (v *Visitor) PreVisitCharacterLiteral(cl *ast.CharacterLiteral) bool {
	return true
}

func (v *Visitor) PostVisitCharacterLiteral(cl *ast.CharacterLiteral) {}

func (v *Visitor) PreVisitBooleanLiteral(bl *ast.BooleanLiteral) bool {
	return true
}

func (v *Visitor) PostVisitBooleanLiteral(bl *ast.BooleanLiteral) {}

func (v *Visitor) PreVisitUnaryOperation(uo *ast.UnaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitUnaryOperation(uo *ast.UnaryOperation) {
	typ := uo.Expr.Type()
	switch uo.Op {
	case ast.NOT:
		if typ != ast.Boolean {
			v.Errorf(uo.Expr, "operator %s expects operand of type %s to be Boolean", uo.Op, typ)
		}
	case ast.NEG:
		if !ast.IsNumericType(typ) {
			v.Errorf(uo.Expr, "operator %s expects operand of type %s to be numeric", uo.Op, typ)
		}
	default:
		panic(uo.Op)
	}
}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
	aTyp := bo.Lhs.Type()
	bTyp := bo.Rhs.Type()
	switch bo.Op {
	case ast.ADD, ast.SUB, ast.MUL, ast.DIV, ast.EXP, ast.MOD:
		// TODO: special case ADD as concat?
		if !ast.IsNumericType(aTyp) {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be numeric", bo.Op, aTyp)
		}
		if !ast.IsNumericType(bTyp) {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be numeric", bo.Op, bTyp)
		}
		rTyp := ast.AreComparableTypes(aTyp, bTyp)
		if rTyp == ast.UnresolvedType {
			panic(bo) // should not happen
		}
		bo.Typ = rTyp
	case ast.EQ, ast.NEQ:
		rTyp := ast.AreComparableTypes(aTyp, bTyp)
		if rTyp == ast.UnresolvedType {
			v.Errorf(bo, "operator %s not supported for types %s and %s", aTyp, bTyp)
		}
		bo.Typ = ast.Boolean
	case ast.LT, ast.GT, ast.LTE, ast.GTE:
		if !ast.AreComparableOrderedTypes(aTyp, bTyp) {
			v.Errorf(bo, "operator %s not supported for types %s and %s", aTyp, bTyp)
		}
		bo.Typ = ast.Boolean
	case ast.AND, ast.OR:
		if aTyp != ast.Boolean {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be Boolean", bo.Op, aTyp)
		}
		if bTyp != ast.Boolean {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be Boolean", bo.Op, bTyp)
		}
		bo.Typ = ast.Boolean
	default:
		panic(bo.Op)
	}
}

func (v *Visitor) PreVisitVariableExpression(ve *ast.VariableExpression) bool {
	return true
}

func (v *Visitor) PostVisitVariableExpression(ve *ast.VariableExpression) {}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.errors = append(v.errors, ast.Error{
		SourceInfo: si.GetSourceInfo(),
		Desc:       fmt.Sprintf("type error: "+fmtStr, args...),
	})
}
