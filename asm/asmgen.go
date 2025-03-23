package asm

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
	"github.com/dragonsinth/gaddis/lib"
)

func Assemble(prog *ast.Program) *Assembly {
	tv := &TempVisitor{currScope: prog.Scope}
	prog.Visit(tv)

	v := &Visitor{}

	// Map the global scope up front.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt:
			v.newLabel(stmt.Name)
		case *ast.FunctionStmt:
			v.newLabel(stmt.Name)
		default:
			// nothing
		}
	}

	// Emit the global block's begin statement.
	globalLabel := v.newLabel("global!")
	v.code = append(v.code, Begin{
		baseInst: baseInst{prog.Block.SourceInfo},
		Scope:    prog.Scope,
		Label:    globalLabel,
		NParams:  0,
		NLocals:  len(prog.Scope.Locals),
	})

	// Emit all global block non-decls.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt, *ast.FunctionStmt:
			// nothing
		default:
			stmt.Visit(v)
		}
	}

	// If there is a module named main with no arguments, call it at the very end.
	finalReturnSi := prog.Block.SourceInfo.Tail()
	ref := prog.Scope.Lookup("main")
	if ref != nil && ref.ModuleStmt != nil && len(ref.ModuleStmt.Params) == 0 {
		scope := ref.ModuleStmt.Scope
		v.code = append(v.code, Call{
			baseInst: baseInst{ref.ModuleStmt.SourceInfo.Head()},
			Label:    v.refLabel("main"),
			NArgs:    len(scope.Params),
		})
		finalReturnSi = ref.ModuleStmt.SourceInfo.Head()
	}

	// terminate the program cleanly
	v.code = append(v.code, End{
		baseInst: baseInst{finalReturnSi},
		Label:    globalLabel,
	})

	// Now emit all modules and functions.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt, *ast.FunctionStmt:
			stmt.Visit(v)
		default:
			// nothing
		}
	}

	strings := make([]string, len(v.strings))
	for s, i := range v.strings {
		strings[i] = s
	}

	return &Assembly{
		GlobalScope: prog.Scope,
		Code:        v.code,
		Labels:      v.labels,
		Strings:     strings,
	}
}

type Visitor struct {
	base.Visitor
	code    []Inst
	labels  map[string]*Label
	strings map[string]int
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	if vd.IsConst || vd.IsParam {
		return false
	}
	// emit an assignment
	if vd.Expr != nil || len(vd.Dims) > 0 {
		if vd.Expr != nil {
			v.maybeCast(vd.Type, vd.Expr)
		} else {
			v.outputArrayInitializer(vd.SourceInfo, vd.Type.AsArrayType(), vd.Dims, nil)
		}
		v.varRefDecl(vd, vd, true)
		v.store(vd)
	}
	return false
}

func (v *Visitor) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	for _, arg := range d.Exprs {
		if lit, ok := arg.(*ast.Literal); ok && lit.IsTabLiteral {
			v.code = append(v.code, Literal{
				baseInst: baseInst{arg.GetSourceInfo()},
				Typ:      lit.Type,
				Val:      lib.TabDisplay,
			})
		} else {
			arg.Visit(v)
		}
	}
	v.code = append(v.code, LibCall{
		baseInst: baseInst{d.GetSourceInfo()},
		Name:     "Display",
		Type:     ast.UnresolvedType,
		Index:    lib.IndexOf("Display"),
		NArg:     len(d.Exprs),
	})
	return false
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	typ := i.Ref.GetType().AsPrimitive()
	name := "Input" + typ.String()
	v.code = append(v.code, LibCall{
		baseInst: baseInst{i.SourceInfo},
		Name:     name,
		Type:     typ,
		Index:    lib.IndexOf(name),
		NArg:     0,
	})
	v.varRef(i.Ref, true)
	if ar, ok := i.Ref.(*ast.ArrayRef); ok {
		if ar.RefExpr.GetType() == ast.String {
			v.code = append(v.code, StoreChar{baseInst{i.SourceInfo}})
			return false
		}
	}
	v.store(i)
	return false
}

func (v *Visitor) PreVisitSetStmt(i *ast.SetStmt) bool {
	v.maybeCast(i.Ref.GetType(), i.Expr)
	v.varRef(i.Ref, true)
	if ar, ok := i.Ref.(*ast.ArrayRef); ok {
		if ar.RefExpr.GetType() == ast.String {
			v.code = append(v.code, StoreChar{baseInst{i.SourceInfo}})
			return false
		}
	}
	v.store(i)
	return false
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	endLabel := &Label{Name: "endif"}
	for _, cb := range is.Cases {
		var lbl *Label
		if cb.Expr != nil {
			cb.Expr.Visit(v)
			lbl = &Label{Name: "else"}
			v.code = append(v.code, JumpFalse{baseInst: baseInst{cb.SourceInfo}, Label: lbl})
		}
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, Jump{baseInst{si}, endLabel})

		if lbl != nil {
			lbl.PC = len(v.code)
		}
	}
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	endLabel := &Label{Name: "endif"}

	// Evaluate the switch expr first.
	v.maybeCast(ss.Type, ss.Expr)
	hasDefault := false
	for _, cb := range ss.Cases {
		var lbl *Label
		if cb.Expr != nil {
			// Duplicate the switch expression in case we fail.
			v.code = append(v.code, Dup{baseInst: baseInst{cb.SourceInfo}})
			v.maybeCast(ss.Type, cb.Expr)
			v.code = append(v.code, makeBinaryOp(ss.Type.AsPrimitive(), cb.Expr, ast.EQ))
			lbl = &Label{Name: "case"}
			v.code = append(v.code, JumpFalse{baseInst: baseInst{cb.SourceInfo}, Label: lbl})
		} else {
			hasDefault = true
		}
		// we selected this block; remove the original switch expr
		v.code = append(v.code, Pop{baseInst: baseInst{cb.SourceInfo}})
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, Jump{baseInst{si}, endLabel})

		if lbl != nil {
			lbl.PC = len(v.code)
		}
	}

	if !hasDefault {
		// remove the original switch expr
		v.code = append(v.code, Pop{baseInst: baseInst{ss.SourceInfo.Tail()}})
	}

	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	startLabel := &Label{Name: "do", PC: len(v.code)}
	ds.Block.Visit(v)
	ds.Expr.Visit(v)
	if ds.Until {
		v.code = append(v.code, JumpFalse{baseInst: baseInst{ds.Expr.GetSourceInfo()}, Label: startLabel})
	} else {
		v.code = append(v.code, JumpTrue{baseInst: baseInst{ds.Expr.GetSourceInfo()}, Label: startLabel})
	}
	return false
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	startLabel := &Label{Name: "while", PC: len(v.code)}
	endLabel := &Label{Name: "wend", PC: 0}
	ws.Expr.Visit(v)
	v.code = append(v.code, JumpFalse{baseInst: baseInst{ws.Expr.GetSourceInfo()}, Label: endLabel})
	ws.Block.Visit(v)
	v.code = append(v.code, Jump{baseInst: baseInst{ws.Tail()}, Label: startLabel})
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	op := ast.LTE
	intVal := int64(1)
	floatVal := float64(1)
	if fs.StepExpr != nil {
		stepLit := fs.StepExpr.ConstEval()
		switch val := stepLit.(type) {
		case int64:
			if val < 0 {
				op = ast.GTE
			}
			intVal = val
			floatVal = float64(val)
		case float64:
			if val < 0 {
				op = ast.GTE
			}
			floatVal = val
		default:
			panic(stepLit)
		}
	}

	refType := fs.Ref.GetType()
	endLabel := &Label{Name: "fend", PC: 0}

	// store
	v.maybeCast(refType, fs.StartExpr)
	v.varRef(fs.Ref, true)
	v.code = append(v.code, Store{baseInst{fs.StartExpr.GetSourceInfo()}})

	startLabel := &Label{Name: "for", PC: len(v.code)}

	// test
	v.varRef(fs.Ref, false)
	v.maybeCast(refType, fs.StopExpr)
	switch refType {
	case ast.Integer:
		v.code = append(v.code, BinOpInt{baseInst: baseInst{fs.StopExpr.GetSourceInfo()}, Op: op})
	case ast.Real:
		v.code = append(v.code, BinOpReal{baseInst: baseInst{fs.StopExpr.GetSourceInfo()}, Op: op})
	default:
		panic(refType)
	}
	v.code = append(v.code, JumpFalse{baseInst: baseInst{fs.StopExpr.GetSourceInfo()}, Label: endLabel})

	fs.Block.Visit(v)

	// post loop increment+jump
	si := fs.SourceInfo.Tail()

	v.varRefDecl(si, fs.Ref.(*ast.VariableExpr).Ref, true)
	switch refType {
	case ast.Integer:
		v.code = append(v.code, IncrInt{baseInst: baseInst{si}, Val: intVal})
	case ast.Real:
		v.code = append(v.code, IncrReal{baseInst: baseInst{si}, Val: floatVal})
	default:
		panic(refType)
	}

	v.code = append(v.code, Jump{baseInst: baseInst{si}, Label: startLabel})
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	// attribute all the for/step/jumps to top line of the for loop
	si := fs.SourceInfo.Head()
	endLabel := &Label{Name: "fend", PC: 0}

	// initialize the index expression
	fs.Index.Visit(v)

	// test
	startLabel := &Label{Name: "for", PC: len(v.code)}
	v.varRefDecl(si, fs.Index, false)
	fs.ArrayExpr.Visit(v)
	v.code = append(v.code, ArrayLen{baseInst: baseInst{si}})
	v.code = append(v.code, BinOpInt{baseInst: baseInst{si}, Op: ast.LT})
	v.code = append(v.code, JumpFalse{baseInst: baseInst{si}, Label: endLabel})

	// assign current element value
	// ref = arr[idx]
	// arr, idx, array ref, ref, store
	fs.ArrayExpr.Visit(v)
	v.varRefDecl(si, fs.Index, false)
	v.code = append(v.code, OffsetVal{baseInst: baseInst{si}, OffsetType: OffsetTypeArray})
	v.varRef(fs.Ref, true)
	v.code = append(v.code, Store{baseInst{si}})

	fs.Block.Visit(v)

	// post loop increment+jump
	si = fs.SourceInfo.Tail()
	v.varRefDecl(si, fs.Index, true)
	v.code = append(v.code, IncrInt{baseInst: baseInst{si}, Val: 1})
	v.code = append(v.code, Jump{baseInst: baseInst{si}, Label: startLabel})
	endLabel.PC = len(v.code)

	return false
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	v.code = append(v.code, Return{
		baseInst: baseInst{rs.SourceInfo},
		NVal:     1,
	})
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.outputArguments(cs.Args, cs.Ref.Params)
	if cs.Ref.IsExternal {
		v.code = append(v.code, LibCall{
			baseInst: baseInst{cs.SourceInfo},
			Name:     cs.Name,
			Type:     ast.UnresolvedType,
			Index:    lib.IndexOf(cs.Name),
			NArg:     len(cs.Args),
		})
	} else {
		v.code = append(v.code, Call{
			baseInst: baseInst{cs.SourceInfo},
			Label:    v.refLabel(cs.Ref.Name),
			NArgs:    len(cs.Args),
		})
	}
	return false
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	lbl := v.refLabel(ms.Name)
	lbl.PC = len(v.code)
	v.code = append(v.code, Begin{
		baseInst: baseInst{ms.SourceInfo},
		Scope:    ms.Scope,
		Label:    lbl,
		NParams:  len(ms.Scope.Params),
		NLocals:  len(ms.Scope.Locals),
	})
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.code = append(v.code, End{
		baseInst: baseInst{ms.SourceInfo.Tail()},
		Label:    v.refLabel(ms.Name),
	})
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	lbl := v.refLabel(fs.Name)
	lbl.PC = len(v.code)
	v.code = append(v.code, Begin{
		baseInst: baseInst{fs.SourceInfo},
		Scope:    fs.Scope,
		Label:    lbl,
		NParams:  len(fs.Scope.Params),
		NLocals:  len(fs.Scope.Locals),
	})
	return true
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.code = append(v.code, End{
		baseInst: baseInst{fs.SourceInfo.Tail()},
		Label:    v.refLabel(fs.Name),
	})
}

func (v *Visitor) PostVisitLiteral(l *ast.Literal) {
	var id int
	if l.Type == ast.String && !l.IsTabLiteral {
		val := l.Val.(string)
		if existing, ok := v.strings[val]; ok {
			id = existing
		} else {
			if v.strings == nil {
				v.strings = map[string]int{}
			}
			id = len(v.strings)
			v.strings[val] = id
		}
	}

	v.code = append(v.code, Literal{
		baseInst: baseInst{l.SourceInfo},
		Typ:      l.Type,
		Val:      l.Val,
		Id:       id,
	})
}

func (v *Visitor) PreVisitBinaryOperation(l *ast.BinaryOperation) bool {
	v.maybeCast(l.ArgType, l.Lhs)
	v.maybeCast(l.ArgType, l.Rhs)
	v.code = append(v.code, makeBinaryOp(l.ArgType.AsPrimitive(), l, l.Op))
	return false
}

func (v *Visitor) PostVisitUnaryOperation(l *ast.UnaryOperation) {
	switch l.Type {
	case ast.Integer:
		v.code = append(v.code, UnaryOpInt{baseInst: baseInst{l.SourceInfo}, Op: l.Op})
	case ast.Real:
		v.code = append(v.code, UnaryOpFloat{baseInst: baseInst{l.SourceInfo}, Op: l.Op})
	case ast.Boolean:
		v.code = append(v.code, UnaryOpBool{baseInst: baseInst{l.SourceInfo}, Op: l.Op})
	default:
		panic(l.Type)
	}
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
	v.varRef(ve, false) // if we get here, we need a value
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	v.outputArguments(ce.Args, ce.Ref.Params)
	if ce.Ref.IsExternal {
		v.code = append(v.code, LibCall{
			baseInst: baseInst{ce.SourceInfo},
			Name:     ce.Name,
			Type:     ce.Type.AsPrimitive(),
			Index:    lib.IndexOf(ce.Name),
			NArg:     len(ce.Args),
		})
	} else {
		v.code = append(v.code, Call{
			baseInst: baseInst{ce.SourceInfo},
			Label:    v.refLabel(ce.Ref.Name),
			NArgs:    len(ce.Args),
		})
	}
	return false
}

func (v *Visitor) PreVisitArrayRef(arr *ast.ArrayRef) bool {
	arr.RefExpr.Visit(v)
	arr.IndexExpr.Visit(v)
	typ := OffsetTypeArray
	if arr.RefExpr.GetType() == ast.String {
		typ = OffsetTypeString
	}

	// if we get here we need a value
	v.code = append(v.code, OffsetVal{
		baseInst:   baseInst{arr.GetSourceInfo()},
		OffsetType: typ,
	})
	return false
}

func (v *Visitor) PreVisitArrayInitializer(ai *ast.ArrayInitializer) bool {
	si := ai.SourceInfo
	if len(ai.Args) > 0 {
		si = ai.Args[len(ai.Args)-1].GetSourceInfo()
	}
	v.outputArrayInitializer(si, ai.Type, ai.Dims, ai.Args)
	return false
}

func (v *Visitor) outputArrayInitializer(newLitSi ast.SourceInfo, t *ast.ArrayType, dims []int, exprs []ast.Expression) []ast.Expression {
	if len(dims) == 1 {
		typ := t.BaseType()
		for i := 0; i < dims[0]; i++ {
			if len(exprs) > 0 {
				expr := exprs[0]
				v.maybeCast(typ, expr)
				exprs = exprs[1:]
			} else {
				v.zero(newLitSi, typ.AsPrimitive())
			}
		}
	} else {
		for i := 0; i < dims[0]; i++ {
			exprs = v.outputArrayInitializer(newLitSi, t.ElementType.AsArrayType(), dims[1:], exprs)
		}
	}
	v.code = append(v.code, &ArrayNew{
		baseInst: baseInst{newLitSi},
		Typ:      t,
		Size:     dims[0],
	})
	return exprs
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	exp.Visit(v)
	if dstType == ast.Real && exp.GetType() == ast.Integer {
		v.code = append(v.code, IntToReal{baseInst{exp.GetSourceInfo()}})
	} else if dstType == ast.Integer && exp.GetType() == ast.Real {
		v.code = append(v.code, RealToInt{baseInst{exp.GetSourceInfo()}})
	}
}

func (v *Visitor) varRef(expr ast.Expression, needRef bool) {
	switch exp := expr.(type) {
	case *ast.VariableExpr:
		v.varRefDecl(expr, exp.Ref, needRef)
	case *ast.ArrayRef:
		typ := OffsetTypeArray
		if exp.RefExpr.GetType() == ast.String {
			typ = OffsetTypeString
		}
		exp.RefExpr.Visit(v)
		exp.IndexExpr.Visit(v)
		if needRef {
			v.code = append(v.code, OffsetRef{
				baseInst:   baseInst{exp.GetSourceInfo()},
				OffsetType: typ,
			})
		} else {
			v.code = append(v.code, OffsetVal{
				baseInst:   baseInst{exp.GetSourceInfo()},
				OffsetType: typ,
			})
		}
	default:
		panic("implement me")
	}
}

func (v *Visitor) varRefDecl(hs ast.HasSourceInfo, decl *ast.VarDecl, needRef bool) {
	isRef := decl.IsRef
	if decl.IsConst {
		if isRef || needRef {
			panic("here")
		}
		val := decl.Expr.ConstEval()
		if val == nil {
			panic("here")
		}
		v.code = append(v.code, Literal{
			baseInst: baseInst{hs.GetSourceInfo()},
			Typ:      decl.Type.AsPrimitive(),
			Val:      val,
		})
	} else if decl.Scope.IsGlobal {
		if isRef {
			panic("here")
		}
		if needRef {
			v.code = append(v.code, GlobalRef{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		} else {
			v.code = append(v.code, GlobalVal{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		}
		return
	} else if decl.IsParam {
		// Param
		if isRef == needRef {
			// if we have a ref and need a ref, or we have a val and need a val, we good
			v.code = append(v.code, ParamVal{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		} else if needRef {
			v.code = append(v.code, ParamRef{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		} else {
			// Take the value (it's a reference) then derefence it.
			v.code = append(v.code, ParamPtr{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		}
	} else {
		// Local
		if isRef {
			panic("here")
		}
		if needRef {
			v.code = append(v.code, LocalRef{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		} else {
			v.code = append(v.code, LocalVal{
				baseInst: baseInst{hs.GetSourceInfo()},
				Name:     decl.Name,
				Index:    decl.Id,
			})
		}
	}
}

func (v *Visitor) store(hs ast.HasSourceInfo) {
	v.code = append(v.code, Store{baseInst{hs.GetSourceInfo()}})
}

func (v *Visitor) outputArguments(args []ast.Expression, params []*ast.VarDecl) {
	for i, arg := range args {
		param := params[i]
		if param.IsRef {
			v.varRef(arg, true)
		} else if param.Type.IsArrayType() {
			arg.Visit(v)
			at := param.Type.AsArrayType()
			v.code = append(v.code, ArrayClone{
				baseInst: baseInst{arg.GetSourceInfo()},
				Typ:      at,
				NDims:    at.NDims,
			})
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
}

func (v *Visitor) newLabel(name string) *Label {
	if v.labels == nil {
		v.labels = map[string]*Label{}
	}
	if _, ok := v.labels[name]; ok {
		panic(name)
	}
	l := &Label{Name: name}
	v.labels[name] = l
	return l
}

func (v *Visitor) refLabel(name string) *Label {
	if _, ok := v.labels[name]; !ok {
		panic(name)
	}
	return v.labels[name]
}

func (v *Visitor) zero(si ast.SourceInfo, typ ast.PrimitiveType) {
	v.code = append(v.code, Literal{
		baseInst: baseInst{si},
		Typ:      typ,
		Val:      zeroValue(typ),
		Id:       0,
	})
}

func makeBinaryOp(t ast.PrimitiveType, hs ast.HasSourceInfo, op ast.Operator) Inst {
	si := hs.GetSourceInfo()
	switch t {
	case ast.Integer:
		return BinOpInt{baseInst: baseInst{si}, Op: op}
	case ast.Real:
		return BinOpReal{baseInst: baseInst{si}, Op: op}
	case ast.String:
		return BinOpStr{baseInst: baseInst{si}, Op: op}
	case ast.Character:
		return BinOpChar{baseInst: baseInst{si}, Op: op}
	case ast.Boolean:
		return BinOpBool{baseInst: baseInst{si}, Op: op}
	default:
		panic(t)
	}
}
