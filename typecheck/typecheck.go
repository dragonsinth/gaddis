package typecheck

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

// TODO(scottb):
// - access control checks
// - super constructor invocation
// - calling constructor outside of call / super

func TypeCheck(node ast.Node, scope *ast.Scope) []ast.Error {
	v := &Visitor{}
	for s := scope; true; s = s.Parent {
		if s.IsGlobal {
			v.globalScope = scope
			break
		}
	}
	v.PushScope(scope)
	node.Visit(v)
	return v.Errors
}

type Visitor struct {
	base.Visitor
	globalScope *ast.Scope
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PostVisitVarDecl(vd *ast.VarDecl) {
	if vd.Expr != nil {
		if !ast.CanCoerce(vd.Type, vd.Expr.GetType()) {
			v.Errorf(vd.Expr, "%s not assignable to %s", vd.Expr.GetType(), vd.Type)
		}
	}
	if vd.IsConst {
		// replace with a brand new literal!
		val := vd.Expr.ConstEval()
		if val == nil {
			v.Visitor.Errorf(vd.Expr, "expression must be constant")
		} else {
			if vd.Type == ast.Real {
				val = ast.EnsureReal(val)
			}
			vd.Expr = &ast.Literal{SourceInfo: vd.Expr.GetSourceInfo(), Type: vd.Type.AsPrimitive(), Val: val}
		}
	}

	for _, d := range vd.DimExprs {
		val := d.ConstEval()
		if val == nil || d.GetType() != ast.Integer {
			v.Visitor.Errorf(d, "dim expression must be a constant integer")
		} else if dim := val.(int64); dim < 0 {
			v.Visitor.Errorf(d, "dim expression must be a positive integer")
		} else {
			vd.Dims = append(vd.Dims, int(dim))
		}
	}
	if vd.Expr != nil && len(vd.Dims) > 0 {
		ai := vd.Expr.(*ast.ArrayInitializer)
		ai.Dims = vd.Dims
	}
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {
	for _, expr := range ds.Exprs {
		if !expr.GetType().IsPrimitive() {
			v.Errorf(expr, "display value must be a primitive type")
		}
	}
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
	ref := is.Ref
	if !ref.CanReference() {
		v.Errorf(ref, "Input argument must be a reference")
	} else if !ref.GetType().IsPrimitive() {
		v.Errorf(ref, "Input argument must be a primitive type")
	}
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
	exprType := ss.Expr.GetType()
	refType := ss.Ref.GetType()
	if !ss.Ref.CanReference() {
		v.Errorf(ss.Ref, "set argument must be a reference")
	} else if !ast.CanCoerce(refType, exprType) {
		v.Errorf(ss.Expr, "%s not assignable to %s", exprType, refType)
	} else if refType.IsArrayType() {
		v.Errorf(ss.Expr, "arrays cannot be assigned to")
	}
}

func (v *Visitor) PostVisitOpenStmt(os *ast.OpenStmt) {
	file := os.File
	if !file.CanReference() {
		v.Errorf(file, "Open file argument must be a reference")
	} else if !file.GetType().IsFileType() {
		v.Errorf(file, "expected file type; got %s", file.GetType())
	}
	if os.Name.GetType() != ast.String {
		v.Errorf(os, "expected String; got %s", os.Name)
	}
}

func (v *Visitor) PostVisitCloseStmt(cs *ast.CloseStmt) {
	file := cs.File
	if !file.GetType().IsFileType() {
		v.Errorf(file, "expected file type; got %s", file.GetType())
	}
}

func (v *Visitor) PostVisitReadStmt(rs *ast.ReadStmt) {
	file := rs.File
	if !file.GetType().IsFileType() {
		v.Errorf(file, "expected file type; got %s", file.GetType())
	}
	for _, expr := range rs.Exprs {
		if !expr.CanReference() {
			v.Errorf(expr, "Read data argument must be a reference")
		} else if !expr.GetType().IsPrimitive() {
			v.Errorf(expr, "Read data argument must be a primitive type")
		}
	}
}

func (v *Visitor) PostVisitWriteStmt(ws *ast.WriteStmt) {
	file := ws.File
	if !file.GetType().IsFileType() {
		v.Errorf(file, "expected file type; got %s", file.GetType())
	}
	for _, expr := range ws.Exprs {
		if !expr.GetType().IsPrimitive() {
			v.Errorf(expr, "Write data argument must be a primitive type")
		}
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
	// early out erroring if the loop var is jacked
	refType := fs.Ref.GetType()
	if _, ok := fs.Ref.(*ast.VariableExpr); !ok {
		v.Errorf(fs.Ref, "loop counter must be a plain variable")
		return
	}
	if !refType.IsNumeric() {
		v.Errorf(fs.Ref, "loop counter must be a number, got %s", refType)
		return
	}
	// check start/stop/step
	if !ast.CanCoerce(refType, fs.StartExpr.GetType()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StartExpr.GetType(), refType)
	}
	if !ast.CanCoerce(refType, fs.StopExpr.GetType()) {
		v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StopExpr.GetType(), refType)
	}
	if fs.StepExpr != nil {
		if !ast.CanCoerce(refType, fs.StepExpr.GetType()) {
			v.Errorf(fs.StartExpr, "%s not assignable to %s", fs.StepExpr.GetType(), refType)
			return
		}

		val := fs.StepExpr.ConstEval()
		var intVal int64
		var floatVal float64
		switch val := val.(type) {
		case int64:
			intVal = val
			floatVal = float64(val)
		case float64:
			floatVal = val
		default:
			v.Errorf(fs.StepExpr, "step expression must be a constant number, got %s", val)
			return
		}

		switch refType {
		case ast.Integer:
			fs.StepExpr = &ast.Literal{
				SourceInfo:   fs.StepExpr.GetSourceInfo(),
				Type:         ast.Integer,
				Val:          intVal,
				IsTabLiteral: false,
			}
		case ast.Real:
			fs.StepExpr = &ast.Literal{
				SourceInfo:   fs.StepExpr.GetSourceInfo(),
				Type:         ast.Real,
				Val:          floatVal,
				IsTabLiteral: false,
			}
		default:
			panic(refType)
		}
	}
}

func (v *Visitor) PostVisitForEachStmt(fs *ast.ForEachStmt) {
	refType := fs.Ref.GetType()
	if _, ok := fs.Ref.(*ast.VariableExpr); !ok {
		v.Errorf(fs.Ref, "loop counter must be a plain variable")
		return
	}

	arrayType := fs.ArrayExpr.GetType()
	if !arrayType.IsArrayType() {
		v.Errorf(fs.ArrayExpr, "For Each In expression must be an array")
		return
	}

	elementType := arrayType.AsArrayType().ElementType
	if !ast.CanCoerce(refType, elementType) {
		v.Errorf(fs.ArrayExpr, "%s element is not assignable to %s", arrayType, refType)
	}
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {
	var decl *ast.Decl
	if cs.Qualifier == nil {
		decl = v.Scope().Lookup(cs.Name)
		if decl == nil {
			v.Errorf(cs, "unresolved symbol: %s", cs.Name)
			return
		}
	} else {
		decl = v.qualifiedLookup(cs.Qualifier, cs.Name)
		if decl == nil {
			v.Errorf(cs, "unresolved symbol: %s in Type %s", cs.Name, cs.Qualifier.GetType())
			return
		}
	}

	if decl.ModuleStmt == nil {
		v.Errorf(cs, "expected Module ref, got: %s", decl)
		return
	}

	// check the number and type of each argument
	cs.Ref = decl.ModuleStmt
	if decl.ModuleStmt.Enclosing != nil && cs.Qualifier == nil {
		// synthesize this ref of the immediate enclosing class type
		cs.Qualifier = &ast.ThisRef{
			SourceInfo: cs.Head(),
			Type:       v.Scope().EnclosingClass().Type,
		}
	}
	v.checkArgumentList(cs, cs.Args, cs.Ref.Params)
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.PushScope(ms.Scope)
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.PopScope()
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	ref := v.Scope().FunctionStmt
	if ref == nil {
		v.Errorf(rs, "return statement without enclosing Function")
	} else {
		rs.Ref = ref
	}

	returnType := rs.Ref.Type
	exprType := rs.Expr.GetType()
	if !ast.CanCoerce(returnType, exprType) {
		v.Errorf(rs.Expr, "return: %s not assignable to %s", exprType, returnType)
	}
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.PushScope(fs.Scope)
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.PopScope()
}

func (v *Visitor) PreVisitClassStmt(cs *ast.ClassStmt) bool {
	v.PushScope(cs.Scope)
	return true
}

func (v *Visitor) PostVisitClassStmt(cs *ast.ClassStmt) {
	v.PopScope()
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
	ve.Type = ast.UnresolvedType

	var decl *ast.Decl
	if ve.Qualifier == nil {
		decl = v.Scope().Lookup(ve.Name)
		if decl == nil {
			v.Errorf(ve, "unresolved symbol: %s", ve.Name)
			return
		}
	} else {
		decl = v.qualifiedLookup(ve.Qualifier, ve.Name)
		if decl == nil {
			v.Errorf(ve, "unresolved symbol: %s in Type %s", ve.Name, ve.Qualifier.GetType())
			return
		}
	}

	if decl.VarDecl == nil {
		v.Errorf(ve, "expected variable ref, got: %s", decl)
		return
	}

	if decl.VarDecl.Enclosing != nil && ve.Qualifier == nil {
		// synthesize this ref of the immediate enclosing class type
		ve.Qualifier = &ast.ThisRef{
			SourceInfo: ve.Head(),
			Type:       v.Scope().EnclosingClass().Type,
		}
	}

	ve.Ref = decl.VarDecl
	ve.Type = ve.Ref.Type
}

func (v *Visitor) PostVisitCallExpr(ce *ast.CallExpr) {
	ce.Type = ast.UnresolvedType

	var decl *ast.Decl
	if ce.Qualifier == nil {
		decl = v.Scope().Lookup(ce.Name)
		if decl == nil {
			v.Errorf(ce, "unresolved symbol: %s", ce.Name)
			return
		}
	} else {
		decl = v.qualifiedLookup(ce.Qualifier, ce.Name)
		if decl == nil {
			v.Errorf(ce, "unresolved symbol: %s in Type %s", ce.Name, ce.Qualifier.GetType())
			return
		}
	}

	if decl.FunctionStmt == nil {
		v.Errorf(ce, "expected Function ref, got: %s", decl)
		return
	}

	if decl.FunctionStmt.Enclosing != nil && ce.Qualifier == nil {
		// synthesize this ref of the immediate enclosing class type
		ce.Qualifier = &ast.ThisRef{
			SourceInfo: ce.Head(),
			Type:       v.Scope().EnclosingClass().Type,
		}
	}

	// Assign the return type, check the number and type of each argument
	ce.Ref = decl.FunctionStmt
	ce.Type = ce.Ref.Type
	v.checkArgumentList(ce, ce.Args, ce.Ref.Params)
}

func (v *Visitor) PostVisitArrayRef(ar *ast.ArrayRef) {
	if indexTyp := ar.IndexExpr.GetType(); indexTyp != ast.UnresolvedType && indexTyp != ast.Integer {
		v.Errorf(ar.IndexExpr, "index expression must be of type Integer")
	}

	if refType := ar.Qualifier.GetType(); refType == ast.UnresolvedType {
		// reduce error spam
		ar.Type = ast.UnresolvedType
	} else if refType == ast.String {
		ar.Type = ast.Character
	} else if refType.IsArrayType() {
		ar.Type = refType.AsArrayType().ElementType
	} else {
		v.Errorf(ar.Qualifier, "array reference expression must be a String or Array type")
	}
}

func (v *Visitor) PostArrayInitializer(ar *ast.ArrayInitializer) {
	typ := ar.Type.BaseType()
	for i, arg := range ar.Args {
		if !ast.CanCoerce(typ, arg.GetType()) {
			v.Errorf(arg, "initializer %d: %s is not assignable to %s", i+1, arg.GetType(), typ)
		}
	}
}

func (v *Visitor) PostVisitNewExpr(ne *ast.NewExpr) {
	ne.Type = ast.UnresolvedType

	classDecl := v.globalScope.Lookup(ne.Name)
	if classDecl == nil {
		v.Errorf(ne, "unresolved symbol: %s", ne.Name)
		return
	}
	cls := classDecl.ClassStmt
	if cls == nil {
		v.Errorf(ne, "expected Class type, got %s", classDecl)
		return
	}
	// find the actual constructor
	modDecl := cls.Scope.Decls[ne.Name]
	if modDecl == nil {
		// There's no constructor; assume a default constructor
		if len(ne.Args) > 0 {
			v.Errorf(ne, "expected 0 args, got %d", len(ne.Args))
		}
		return
	}

	ne.Type = cls.Type
	ctor := modDecl.ModuleStmt
	ne.Ctor = ctor
	v.checkArgumentList(ne, ne.Args, ctor.Params)
}

func (v *Visitor) qualifiedLookup(qual ast.Expression, name string) *ast.Decl {
	qualType := qual.GetType()
	if qualType == ast.UnresolvedType {
		return nil
	}
	if !qualType.IsClassType() {
		v.Errorf(qual, "expected qualifier to be class type, got: %s", qualType)
		return nil
	}
	for ct := qualType.AsClassType(); ct != nil; ct = ct.Extends {
		if decl := ct.Scope.Decls[name]; decl != nil {
			return decl
		}
	}
	return nil
}

func (v *Visitor) checkArgumentList(si ast.HasSourceInfo, args []ast.Expression, params []*ast.VarDecl) {
	for i, c := 0, min(len(args), len(params)); i < c; i++ {
		arg, param := args[i], params[i]
		if !ast.CanCoerce(param.Type, arg.GetType()) {
			v.Errorf(arg, "argument %d: %s not assignable to %s", i+1, arg.GetType(), param.Type)
		}
		if param.IsRef {
			// must be an exact type match for reference
			if ar, ok := arg.(*ast.ArrayRef); ok {
				if ar.Qualifier.GetType() == ast.String {
					// TODO(scottb): could allow this with an auto temp var?
					v.Errorf(arg, "argument %d: expression may not be a string index", i+1)
				}
			} else if !arg.CanReference() {
				v.Errorf(arg, "argument %d: expression must be a reference", i+1)
			} else if arg.GetType() != param.Type {
				v.Errorf(arg, "argument %d: %s not assignable to %s", i+1, arg.GetType(), param.Type)
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
