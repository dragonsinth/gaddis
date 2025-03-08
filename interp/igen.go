package interp

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
	"github.com/dragonsinth/gaddis/gogen/builtins"
)

func Compile(prog *ast.Program) *Compilation {
	v := &Visitor{
		globalIds: map[*ast.VarDecl]int{},
		localIds:  map[*ast.VarDecl]int{},
		functions: map[*ast.FunctionStmt]*Label{},
		modules:   map[*ast.ModuleStmt]*Label{},
	}

	// Map the global scope up front.
	globalId := 0
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.DeclareStmt:
			// emit declaration only, not assignment
			for _, decl := range stmt.Decls {
				if !decl.IsConst {
					v.globalIds[decl] = globalId
					globalId++
				}
			}
		case *ast.ModuleStmt:
			v.modules[stmt] = &Label{Name: stmt.Name}
		case *ast.FunctionStmt:
			v.functions[stmt] = &Label{Name: stmt.Name}
		default:
			// nothing
		}
	}

	// Emit the global block's begin statement.
	v.code = append(v.code, Begin{
		SourceInfo: prog.Block.SourceInfo,
		Label:      &Label{Name: ":start", PC: 0},
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
	ref := prog.Scope.Lookup("main")
	if ref != nil && ref.ModuleStmt != nil && len(ref.ModuleStmt.Params) == 0 {
		scope := ref.ModuleStmt.Scope
		v.code = append(v.code, Call{
			SourceInfo: prog.Block.SourceInfo.Tail(),
			Scope:      scope,
			Label:      v.modules[ref.ModuleStmt],
		})
	}

	// terminate the program cleanly
	v.code = append(v.code, Return{
		SourceInfo: prog.Block.SourceInfo.Tail(), // should be end
		NVal:       0,
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

	return &Compilation{
		GlobalScope: prog.Scope,
		Code:        v.code,
	}
}

type Visitor struct {
	base.Visitor
	code      []Inst
	globalIds map[*ast.VarDecl]int
	localIds  map[*ast.VarDecl]int
	functions map[*ast.FunctionStmt]*Label
	modules   map[*ast.ModuleStmt]*Label
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	if vd.IsConst || vd.IsParam {
		return false
	}
	// emit an assignment
	if vd.Expr != nil {
		v.maybeCast(vd.Type, vd.Expr)
		v.varRefDecl(vd, vd, true)
		v.store(vd)
	}
	return false
}

func (v *Visitor) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	for _, arg := range d.Exprs {
		if _, ok := arg.(*ast.TabLiteral); ok {
			v.code = append(v.code, Literal{
				SourceInfo: arg.GetSourceInfo(),
				Val:        builtins.TabDisplay,
			})
		} else {
			arg.Visit(v)
		}
	}
	v.code = append(v.code, LibCall{
		SourceInfo: d.GetSourceInfo(),
		Name:       "Display",
		Index:      libFunc("Display"),
		NArg:       len(d.Exprs),
	})
	return false
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	name := "Input" + i.Var.Type.AsPrimitive().String()
	v.code = append(v.code, LibCall{
		SourceInfo: i.SourceInfo,
		Name:       name,
		Index:      libFunc(name),
		NArg:       0,
	})
	v.varRef(i.Var, true)
	v.store(i)
	return false
}

func (v *Visitor) PreVisitSetStmt(i *ast.SetStmt) bool {
	v.maybeCast(i.Var.Type, i.Expr)
	v.varRef(i.Var, true)
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
			v.code = append(v.code, JumpFalse{SourceInfo: cb.SourceInfo, Label: lbl})
		}
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, Jump{si, endLabel})

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
			v.code = append(v.code, Dup{SourceInfo: cb.SourceInfo})
			v.maybeCast(ss.Type, cb.Expr)
			v.code = append(v.code, makeBinaryOp(ss.Type.AsPrimitive(), cb.Expr, ast.EQ))
			lbl = &Label{Name: "case"}
			v.code = append(v.code, JumpFalse{SourceInfo: cb.SourceInfo, Label: lbl})
		} else {
			hasDefault = true
		}
		// we selected this block; remove the original switch expr
		v.code = append(v.code, Pop{SourceInfo: cb.SourceInfo})
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, Jump{si, endLabel})

		if lbl != nil {
			lbl.PC = len(v.code)
		}
	}

	if !hasDefault {
		// remove the original switch expr
		v.code = append(v.code, Pop{SourceInfo: ss.SourceInfo.Tail()})
	}

	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	startLabel := &Label{Name: "do", PC: len(v.code)}
	ds.Block.Visit(v)
	ds.Expr.Visit(v)
	if ds.Not {
		v.code = append(v.code, JumpFalse{SourceInfo: ds.Expr.GetSourceInfo(), Label: startLabel})
	} else {
		v.code = append(v.code, JumpTrue{SourceInfo: ds.Expr.GetSourceInfo(), Label: startLabel})
	}
	return false
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	startLabel := &Label{Name: "while", PC: len(v.code)}
	endLabel := &Label{Name: "wend", PC: 0}
	ws.Expr.Visit(v)
	v.code = append(v.code, JumpFalse{SourceInfo: ws.Expr.GetSourceInfo(), Label: endLabel})
	ws.Block.Visit(v)
	v.code = append(v.code, Jump{SourceInfo: ws.Tail(), Label: startLabel})
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitForStmt(fs *ast.ForStmt) bool {
	endLabel := &Label{Name: "fend", PC: 0}

	refType := fs.Var.Type
	v.varRef(fs.Var, true)
	v.maybeCast(refType, fs.StartExpr)
	v.maybeCast(refType, fs.StopExpr)
	v.stepExpr(fs)
	switch refType {
	case ast.Integer:
		v.code = append(v.code, ForInt{SourceInfo: fs.SourceInfo})
	case ast.Real:
		v.code = append(v.code, ForReal{SourceInfo: fs.SourceInfo})
	default:
		panic(refType)
	}
	v.code = append(v.code, JumpFalse{SourceInfo: fs.SourceInfo, Label: endLabel})

	startLabel := &Label{Name: "for", PC: len(v.code)}
	fs.Block.Visit(v)
	si := ast.SourceInfo{Start: fs.Block.End, End: fs.End}

	// end of loop re-test / increment
	v.varRef(fs.Var, true)
	v.maybeCast(refType, fs.StopExpr)
	v.stepExpr(fs)
	switch refType {
	case ast.Integer:
		v.code = append(v.code, StepInt{SourceInfo: fs.SourceInfo})
	case ast.Real:
		v.code = append(v.code, StepReal{SourceInfo: fs.SourceInfo})
	default:
		panic(refType)
	}
	v.code = append(v.code, JumpTrue{SourceInfo: si, Label: startLabel})
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) stepExpr(fs *ast.ForStmt) {
	refType := fs.Var.Type
	if fs.StepExpr != nil {
		v.maybeCast(refType, fs.StepExpr)
		return
	}
	var val any
	switch refType {
	case ast.Integer:
		val = int64(1)
	case ast.Real:
		val = float64(1)
	default:
		panic(refType)
	}
	v.code = append(v.code, Literal{SourceInfo: fs.StopExpr.GetSourceInfo().Tail(), Val: val})
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	v.outputArguments(cs.Args, cs.Ref.Params)
	v.code = append(v.code, Call{
		SourceInfo: cs.SourceInfo,
		Scope:      cs.Ref.Scope,
		Label:      v.modules[cs.Ref],
	})
	return false
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	v.outputArguments(ce.Args, ce.Ref.Params)
	if ce.Ref.IsExternal {
		v.code = append(v.code, LibCall{
			SourceInfo: ce.SourceInfo,
			Name:       ce.Name,
			Index:      libFunc(ce.Name),
			NArg:       len(ce.Args),
		})
	} else {
		v.code = append(v.code, Call{
			SourceInfo: ce.SourceInfo,
			Scope:      ce.Ref.Scope,
			Label:      v.functions[ce.Ref],
		})
	}
	return false
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	v.mapScope(ms.Scope)
	lbl := v.modules[ms]
	lbl.PC = len(v.code)
	v.code = append(v.code, Begin{SourceInfo: ms.SourceInfo, Label: lbl})
	return true
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.code = append(v.code, Return{SourceInfo: ms.SourceInfo.Tail(), NVal: 0})
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	v.mapScope(fs.Scope)
	lbl := v.functions[fs]
	lbl.PC = len(v.code)
	v.code = append(v.code, Begin{SourceInfo: fs.SourceInfo, Label: lbl})
	return true
}

func (v *Visitor) mapScope(scope *ast.Scope) {
	for i, decl := range scope.Params {
		v.localIds[decl] = i
	}
	for i, decl := range scope.Locals {
		v.localIds[decl] = i + len(scope.Params)
	}
}

func (v *Visitor) PostVisitFunctionStmt(ms *ast.FunctionStmt) {
	v.code = append(v.code, Return{SourceInfo: ms.SourceInfo.Tail(), NVal: 1})
}

func (v *Visitor) PostVisitIntegerLiteral(l *ast.IntegerLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        l.Val,
	})
}

func (v *Visitor) PostVisitRealLiteral(l *ast.RealLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        l.Val,
	})
}

func (v *Visitor) PostVisitStringLiteral(l *ast.StringLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        l.Val,
	})
}

func (v *Visitor) PostVisitCharacterLiteral(l *ast.CharacterLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        l.Val,
	})
}

func (v *Visitor) PostVisitTabLiteral(l *ast.TabLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        "\t",
	})
}

func (v *Visitor) PostVisitBooleanLiteral(l *ast.BooleanLiteral) {
	v.code = append(v.code, Literal{
		SourceInfo: l.SourceInfo,
		Val:        l.Val,
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
		v.code = append(v.code, UnaryOpInt{SourceInfo: l.SourceInfo, Op: l.Op})
	case ast.Real:
		v.code = append(v.code, UnaryOpFloat{SourceInfo: l.SourceInfo, Op: l.Op})
	case ast.Boolean:
		v.code = append(v.code, UnaryOpBool{SourceInfo: l.SourceInfo, Op: l.Op})
	default:
		panic(l.Type)
	}
}

func (v *Visitor) PostVisitVariableExpr(ve *ast.VariableExpr) {
	v.varRef(ve, false) // if we get here, we need a value
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	exp.Visit(v)
	if dstType == ast.Real && exp.GetType() == ast.Integer {
		v.code = append(v.code, IntToReal{exp.GetSourceInfo()})
	} else if dstType == ast.Integer && exp.GetType() == ast.Real {
		v.code = append(v.code, RealToInt{exp.GetSourceInfo()})
	}
}

func (v *Visitor) varRef(expr *ast.VariableExpr, needRef bool) {
	v.varRefDecl(expr, expr.Ref, needRef)
}

func (v *Visitor) varRefDecl(hs ast.HasSourceInfo, decl *ast.VarDecl, needRef bool) {
	isRef := decl.IsRef
	if decl.IsConst {
		if isRef || needRef {
			panic("here")
		}
		v.code = append(v.code, Literal{
			SourceInfo: hs.GetSourceInfo(),
			Val:        decl.Expr.ConstEval(),
		})
	} else if decl.Scope.IsGlobal {
		if isRef {
			panic("here")
		}
		if needRef {
			v.code = append(v.code, GlobalRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      v.globalIds[decl],
			})
		} else {
			v.code = append(v.code, GlobalVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      v.globalIds[decl],
			})
		}
		return
	} else {
		// Local
		if decl.IsRef == needRef {
			// if we have a ref and need a ref, or we have a val and need a val, we good
			v.code = append(v.code, LocalVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      v.localIds[decl],
			})
		} else if needRef {
			v.code = append(v.code, LocalRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      v.localIds[decl],
			})
		} else {
			// Take the value (it's a reference) then derefence it.
			v.code = append(v.code, LocalPtr{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      v.localIds[decl],
			})
		}
	}
}

func (v *Visitor) store(hs ast.HasSourceInfo) {
	v.code = append(v.code, Store{hs.GetSourceInfo()})
}

func (v *Visitor) outputArguments(args []ast.Expression, params []*ast.VarDecl) {
	for i, arg := range args {
		param := params[i]
		if param.IsRef {
			// special case
			// TODO: other types of references
			ve := arg.(*ast.VariableExpr)
			v.varRef(ve, true)
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
}

func makeBinaryOp(t ast.PrimitiveType, hs ast.HasSourceInfo, op ast.Operator) Inst {
	si := hs.GetSourceInfo()
	switch t {
	case ast.Integer:
		return BinOpInt{SourceInfo: si, Op: op}
	case ast.Real:
		return BinOpReal{SourceInfo: si, Op: op}
	case ast.String:
		return BinOpStr{SourceInfo: si, Op: op}
	case ast.Character:
		return BinOpChar{SourceInfo: si, Op: op}
	case ast.Boolean:
		return BinOpBool{SourceInfo: si, Op: op}
	default:
		panic(t)
	}
}
