package controlflow

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

type AssignmentVisitor struct {
	base.Visitor

	written       []bool
	pendingWrites map[*ast.VariableExpr]bool
}

var _ ast.Visitor = &AssignmentVisitor{}

func (v *AssignmentVisitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if v.isLocal(vd) {
		// arrays are always initialized
		if vd.Expr != nil || len(vd.DimExprs) > 0 {
			v.written[vd.Id] = true
		}
	}
}

func (v *AssignmentVisitor) PreVisitInputStmt(is *ast.InputStmt) bool {
	v.pendingWrite(is.Ref)
	return true
}

func (v *AssignmentVisitor) PostVisitInputStmt(is *ast.InputStmt) {
	v.finishWrite(is.Ref)
}

func (v *AssignmentVisitor) PreVisitSetStmt(ss *ast.SetStmt) bool {
	v.pendingWrite(ss.Ref)
	return true
}

func (v *AssignmentVisitor) PostVisitSetStmt(ss *ast.SetStmt) {
	v.finishWrite(ss.Ref)
}

func (v *AssignmentVisitor) PreVisitOpenStmt(os *ast.OpenStmt) bool {
	v.pendingWrite(os.File)
	return true
}

func (v *AssignmentVisitor) PostVisitOpenStmt(os *ast.OpenStmt) {
	v.finishWrite(os.File)
}

func (v *AssignmentVisitor) PreVisitReadStmt(rs *ast.ReadStmt) bool {
	for _, expr := range rs.Exprs {
		v.pendingWrite(expr)
	}
	return true
}

func (v *AssignmentVisitor) PostVisitReadStmt(rs *ast.ReadStmt) {
	for _, expr := range rs.Exprs {
		v.finishWrite(expr)
	}
}

func (v *AssignmentVisitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	// TODO(scottb): delete `for` operation in favor of `step` only?
	// Pros:
	// - stop / step expr can now reference loop var
	// Cons:
	// - loop var evaluated twice on entry...
	// - but maybe that's okay.. disallow array/field refs from being loop vars maybe!?

	// have to do this one manually: none of the expressions are allowed to reference the count variable
	v.pendingWrite(fs.Ref)
	fs.Ref.Visit(v)
	fs.StartExpr.Visit(v)
	fs.StopExpr.Visit(v)
	if fs.StepExpr != nil {
		fs.StepExpr.Visit(v)
	}

	// but it's definitely assigned before the inner block
	v.finishWrite(fs.Ref)
	fs.Block.Visit(v)
	return false
}

func (v *AssignmentVisitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	// have to do this one manually: none of the expressions are allowed to reference the count variable
	v.pendingWrite(fs.Ref)
	fs.Ref.Visit(v)
	fs.ArrayExpr.Visit(v)

	// but it's definitely assigned before the inner block
	v.finishWrite(fs.Ref)
	fs.Block.Visit(v)
	return false
}

func (v *AssignmentVisitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	for i, arg := range cs.Args {
		if cs.Ref.Params[i].IsRef {
			v.pendingWrite(arg)
		}
	}
	return true
}

func (v *AssignmentVisitor) PostVisitCallStmt(cs *ast.CallStmt) {
	for i, arg := range cs.Args {
		if cs.Ref.Params[i].IsRef {
			v.finishWrite(arg)
		}
	}
}

func (v *AssignmentVisitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.written = make([]bool, len(ms.Scope.Locals))
	v.PushScope(ms.Scope)
	v.pendingWrites = map[*ast.VariableExpr]bool{}
	return true
}

func (v *AssignmentVisitor) PostVisitModuleStmt(_ *ast.ModuleStmt) {
	v.PopScope()
	if len(v.pendingWrites) != 0 {
		panic("here")
	}
}

func (v *AssignmentVisitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.written = make([]bool, len(fs.Scope.Locals))
	v.PushScope(fs.Scope)
	v.pendingWrites = map[*ast.VariableExpr]bool{}
	return true
}

func (v *AssignmentVisitor) PostVisitFunctionStmt(_ *ast.FunctionStmt) {
	v.PopScope()
	if len(v.pendingWrites) != 0 {
		panic("here")
	}
}

func (v *AssignmentVisitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
	if lr := v.isLocalRef(ve); lr != nil {
		if !v.pendingWrites[lr] && !v.written[ve.Ref.Id] {
			v.Errorf(ve, "local variable %s: read before write", ve.Ref.Name)
		}
	}
}

func (v *AssignmentVisitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	for i, arg := range ce.Args {
		if ce.Ref.Params[i].IsRef {
			v.pendingWrite(arg)
		}
	}
	return true
}

func (v *AssignmentVisitor) PostVisitCallExpr(ce *ast.CallExpr) {
	for i, arg := range ce.Args {
		if ce.Ref.Params[i].IsRef {
			v.finishWrite(arg)
		}
	}
}

func (v *AssignmentVisitor) pendingWrite(expr ast.Expression) {
	if lr := v.isLocalRef(expr); lr != nil {
		v.pendingWrites[lr] = true
	}
}

func (v *AssignmentVisitor) finishWrite(expr ast.Expression) {
	if lr := v.isLocalRef(expr); lr != nil {
		v.written[lr.Ref.Id] = true
		delete(v.pendingWrites, lr)
	}
}

func (v *AssignmentVisitor) isLocalRef(expr ast.Expression) *ast.VariableExpr {
	if ve, ok := expr.(*ast.VariableExpr); ok {
		if v.isLocal(ve.Ref) {
			return ve
		}
	}
	return nil
}

func (v *AssignmentVisitor) isLocal(ref *ast.VarDecl) bool {
	return ref.Scope == v.Scope() && !ref.IsParam && !ref.IsConst && ref.Enclosing == nil
}
