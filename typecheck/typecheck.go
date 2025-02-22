package typecheck

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

func TypeCheck(globalBlock *ast.Block) []ast.Error {
	// visit the statements in the global block
	v := New()
	globalBlock.Visit(v)
	slices.SortFunc(v.errors, func(a, b ast.Error) int {
		return a.Start.Pos - b.Start.Pos
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

func (v *Visitor) PreVisitDeclareStmt(ds *ast.DeclareStmt) bool {
	return true
}

func (v *Visitor) PostVisitDeclareStmt(ds *ast.DeclareStmt) {
	for _, vd := range ds.Decls {
		if vd.Expr != nil {
			if !ast.CanCoerce(ds.Type, vd.Expr.GetType()) {
				v.Errorf(vd.Expr, "%s not assignable to %s", vd.Expr.GetType(), ds.Type)
			}
			// TODO: initializer must be a constant expression?
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

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
	// TODO: variable is non-primitive?
	if is.Var.Ref.IsConst {
		v.Errorf(is, "input variable may not be a constant")
	}
}

func (v *Visitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	return true
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
	exprType := ss.Expr.GetType()
	refType := ss.Var.Type
	if !ast.CanCoerce(refType, exprType) {
		v.Errorf(ss.Expr, "%s not assignable to %s", exprType, refType)
	}
	if ss.Var.Ref.IsConst {
		v.Errorf(ss.Var, "loop variable may not be a constant")
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
	if cb.Expr.GetType() != ast.Boolean {
		v.Errorf(cb.Expr, "expected Boolean, got %s", cb.Expr.GetType())
	}
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	return true
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
	dstType := ss.Expr.GetType()
	for _, cb := range ss.Cases {
		typ := ast.AreComparableTypes(dstType, cb.Expr.GetType())
		if typ == ast.UnresolvedType {
			v.Errorf(cb.Expr, "case %s not comparable to select %s", cb.Expr.GetType(), ss.Expr.GetType())
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
	if ds.Expr.GetType() != ast.Boolean {
		v.Errorf(ds.Expr, "expected Boolean, got %s", ds.Expr.GetType())
	}
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	return true
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
	if ws.Expr.GetType() != ast.Boolean {
		v.Errorf(ws.Expr, "expected Boolean, got %s", ws.Expr.GetType())
	}
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	return true
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
	refType := fs.Var.Type
	if !ast.IsNumericType(refType) {
		v.Errorf(fs.Var, "loop variable must be a number, got %s %s", refType, fs.Var.Name)
	}
	if fs.Var.Ref.IsConst {
		v.Errorf(fs.Var, "loop variable may not be a constant")
	}
	if !ast.CanCoerce(refType, fs.StartExpr.GetType()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StartExpr.GetType(), refType)
	}
	if !ast.CanCoerce(refType, fs.StopExpr.GetType()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StopExpr.GetType(), refType)
	}
	if fs.StepExpr != nil {
		if !ast.CanCoerce(refType, fs.StepExpr.GetType()) {
			v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StepExpr.GetType(), refType)
		}
	}
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	return true
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {
	// check the number and type of each argument
	args := cs.Args
	params := cs.Ref.Params
	for i, c := 0, min(len(args), len(params)); i < c; i++ {
		arg, param := args[i], params[i]
		if !ast.CanCoerce(param.Type, arg.GetType()) {
			v.Errorf(arg, "argument %d: %s not assignable to %s", i+1, arg.GetType(), param.Type)
		}
		if param.IsRef {
			// must be an exact type match for reference
			if arg.GetType() != param.Type {
				v.Errorf(arg, "argument %d: %s not assignable to %s", i+1, arg.GetType(), param.Type)
			}
			// the argument must be referencable thing
			switch arg := arg.(type) {
			case *ast.VariableExpression:
				ref := arg.Ref
				if ref.IsConst {
					v.Errorf(arg, "argument %d: expression may not be a constant", i+1)
				}
			default:
				// TODO: array references, class field references?
				v.Errorf(arg, "argument %d: expression must be a reference", i+1)
			}
		} else if !ast.CanCoerce(param.Type, arg.GetType()) {
			v.Errorf(arg, "argument %d: %s not assignable to %s", i+1, arg.GetType(), param.Type)
		}
	}
	if len(args) != len(params) {
		v.Errorf(cs, "expected %d args, got %d", len(args), len(params))
	}
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {}

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
	typ := uo.Expr.GetType()
	ok := typ != ast.UnresolvedType // reduce error reporting spam
	op := uo.Op
	switch op {
	case ast.NOT:
		if ok && typ != ast.Boolean {
			v.Errorf(uo.Expr, "operator %s expects operand of type %s to be Boolean", op, typ)
		}
	case ast.NEG:
		if ok && !ast.IsNumericType(typ) {
			v.Errorf(uo.Expr, "operator %s expects operand of type %s to be numeric", op, typ)
		}
	default:
		panic(op)
	}
}

func (v *Visitor) PreVisitBinaryOperation(bo *ast.BinaryOperation) bool {
	return true
}

func (v *Visitor) PostVisitBinaryOperation(bo *ast.BinaryOperation) {
	aTyp := bo.Lhs.GetType()
	bTyp := bo.Rhs.GetType()
	aOk := aTyp != ast.UnresolvedType
	bOk := bTyp != ast.UnresolvedType
	ok := aOk && bOk
	op := bo.Op
	switch op {
	case ast.ADD, ast.SUB, ast.MUL, ast.DIV, ast.EXP, ast.MOD:
		// TODO: special case ADD as concat?
		if aOk && !ast.IsNumericType(aTyp) {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be numeric", op, aTyp)
		}
		if bOk && !ast.IsNumericType(bTyp) {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be numeric", op, bTyp)
		}
		if ok {
			rTyp := ast.AreComparableTypes(aTyp, bTyp)
			if rTyp == ast.UnresolvedType {
				v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
			}
			bo.Type = rTyp
		}
	case ast.EQ, ast.NEQ:
		rTyp := ast.AreComparableTypes(aTyp, bTyp)
		if ok && rTyp == ast.UnresolvedType {
			v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
		}
		bo.Type = ast.Boolean
	case ast.LT, ast.GT, ast.LTE, ast.GTE:
		if ok && !ast.AreComparableOrderedTypes(aTyp, bTyp) {
			v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
		}
		bo.Type = ast.Boolean
	case ast.AND, ast.OR:
		if aOk && aTyp != ast.Boolean {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be Boolean", op, aTyp)
		}
		if bOk && bTyp != ast.Boolean {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be Boolean", op, bTyp)
		}
		bo.Type = ast.Boolean
	default:
		panic(op)
	}
}

func (v *Visitor) PreVisitVariableExpression(ve *ast.VariableExpression) bool {
	return true
}

func (v *Visitor) PostVisitVariableExpression(ve *ast.VariableExpression) {
	ve.Type = ve.Ref.Type
}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.errors = append(v.errors, ast.Error{
		SourceInfo: si.GetSourceInfo(),
		Desc:       fmt.Sprintf("type error: "+fmtStr, args...),
	})
}
