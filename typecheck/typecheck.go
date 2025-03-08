package typecheck

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

func TypeCheck(program *ast.Program) []ast.Error {
	// visit the statements in the global block
	v := &Visitor{}
	program.Block.Visit(v)
	return v.Errors
}

type Visitor struct {
	base.Visitor
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if vd.Expr != nil {
		if !ast.CanCoerce(vd.Type, vd.Expr.GetType()) {
			v.Errorf(vd.Expr, "%s not assignable to %s", vd.Expr.GetType(), vd.Type)
		}
	}
	if vd.IsConst {
		val := vd.Expr.ConstEval()
		if val == nil {
			v.Visitor.Errorf(vd.Expr, "expression must be constant")
		} else {
			// replace with a brand new literal!
			si := vd.Expr.GetSourceInfo()
			switch vd.Type.AsPrimitive() {
			case ast.Integer:
				vd.Expr = &ast.IntegerLiteral{SourceInfo: si, Val: val.(int64)}
			case ast.Real:
				vd.Expr = &ast.RealLiteral{SourceInfo: si, Val: ast.EnsureReal(val)}
			case ast.String:
				vd.Expr = &ast.StringLiteral{SourceInfo: si, Val: val.(string)}
			case ast.Character:
				vd.Expr = &ast.CharacterLiteral{SourceInfo: si, Val: val.(byte)}
			case ast.Boolean:
				vd.Expr = &ast.BooleanLiteral{SourceInfo: si, Val: val.(bool)}
			default:
				panic(vd.Type)
			}
		}
	}
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
	// TODO: variable is non-primitive?
	if is.Var.Ref.IsConst {
		v.Errorf(is, "input variable must be a reference")
	}
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
	exprType := ss.Expr.GetType()
	refType := ss.Var.Type
	if !ast.CanCoerce(refType, exprType) {
		v.Errorf(ss.Expr, "%s not assignable to %s", exprType, refType)
	}
	if ss.Var.Ref.IsConst {
		v.Errorf(ss.Var, "set variable must be a reference")
	}
}

func (v *Visitor) PostVisitCondBlock(cb *ast.CondBlock) {
	if cb.Expr != nil && cb.Expr.GetType() != ast.Boolean {
		v.Errorf(cb.Expr, "expected Boolean, got %s", cb.Expr.GetType())
	}
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
	dstType := ss.Expr.GetType()
	for _, cb := range ss.Cases {
		if cb.Expr != nil {
			typ := ast.AreComparableTypes(dstType, cb.Expr.GetType())
			if typ == ast.UnresolvedType {
				v.Errorf(cb.Expr, "case %s not comparable to select %s", cb.Expr.GetType(), ss.Expr.GetType())
			} else {
				dstType = typ
			}
		}
	}
	ss.Type = dstType
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
	if ds.Expr.GetType() != ast.Boolean {
		v.Errorf(ds.Expr, "expected Boolean, got %s", ds.Expr.GetType())
	}
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
	if ws.Expr.GetType() != ast.Boolean {
		v.Errorf(ws.Expr, "expected Boolean, got %s", ws.Expr.GetType())
	}
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
	refType := fs.Var.Type
	if !refType.IsNumeric() {
		v.Errorf(fs.Var, "loop variable must be a number, got %s %s", refType, fs.Var.Name)
	}
	if fs.Var.Ref.IsConst {
		v.Errorf(fs.Var, "loop variable must be a reference")
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

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {
	// check the number and type of each argument
	v.checkArgumentList(cs, cs.Args, cs.Ref.Params)
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	returnType := rs.Ref.Type
	exprType := rs.Expr.GetType()
	if !ast.CanCoerce(returnType, exprType) {
		v.Errorf(rs.Expr, "return: %s not assignable to %s", exprType, returnType)
	}
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
		if ok && !typ.IsNumeric() {
			v.Errorf(uo.Expr, "operator %s expects operand of type %s to be numeric", op, typ)
		}
	default:
		panic(op)
	}
	uo.Type = typ
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
		if aOk && !aTyp.IsNumeric() {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be numeric", op, aTyp)
		}
		if bOk && !bTyp.IsNumeric() {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be numeric", op, bTyp)
		}
		if ok {
			rTyp := ast.AreComparableTypes(aTyp, bTyp)
			if rTyp == ast.UnresolvedType {
				v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
			}
			bo.Type = rTyp
			bo.ArgType = rTyp
		}
	case ast.EQ, ast.NEQ:
		rTyp := ast.AreComparableTypes(aTyp, bTyp)
		if ok && rTyp == ast.UnresolvedType {
			v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
		}
		bo.Type = ast.Boolean
		bo.ArgType = rTyp
	case ast.LT, ast.GT, ast.LTE, ast.GTE:
		rTyp := ast.AreComparableTypes(aTyp, bTyp)
		if ok && !ast.IsOrderedType(rTyp) {
			v.Errorf(bo, "operator %s not supported for types %s and %s", op, aTyp, bTyp)
		}
		bo.Type = ast.Boolean
		bo.ArgType = rTyp
	case ast.AND, ast.OR:
		if aOk && aTyp != ast.Boolean {
			v.Errorf(bo.Lhs, "operator %s expects left hand operand of type %s to be Boolean", op, aTyp)
		}
		if bOk && bTyp != ast.Boolean {
			v.Errorf(bo.Rhs, "operator %s expects right hand operand of type %s to be Boolean", op, bTyp)
		}
		bo.Type = ast.Boolean
		bo.ArgType = ast.Boolean
	default:
		panic(op)
	}
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
	ve.Type = ve.Ref.Type
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {
	// Assign the return type.
	ce.Type = ce.Ref.Type

	// check the number and type of each argument
	v.checkArgumentList(ce, ce.Args, ce.Ref.Params)
}

func (v *Visitor) checkArgumentList(si ast.HasSourceInfo, args []ast.Expression, params []*ast.VarDecl) {
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
			case *ast.VariableExpr:
				ref := arg.Ref
				if ref.IsConst {
					v.Errorf(arg, "argument %d: expression must be a reference", i+1)
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
		v.Errorf(si, "expected %d args, got %d", len(params), len(args))
	}
}

func (v *Visitor) Errorf(si ast.HasSourceInfo, fmtStr string, args ...any) {
	v.Visitor.Errorf(si, "type error: "+fmtStr, args...)
}
