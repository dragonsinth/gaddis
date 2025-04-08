package asmgen

import (
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
	"github.com/dragonsinth/gaddis/lib"
)

func AssembleExpression(as *asm.Assembly, expr ast.Expression) []asm.Inst {
	v := &Visitor{
		labels: as.Labels,
	}
	expr.Visit(v)
	v.code = append(v.code, asm.Halt{SourceInfo: expr.GetSourceInfo().Tail(), NVal: 1})
	return v.code
}

func Assemble(prog *ast.Program) *asm.Assembly {
	tv := &TempVisitor{}
	prog.Visit(tv)

	v := &Visitor{}
	nClasses := len(prog.Scope.Classes)
	v.vtables = make([]asm.Vtable, nClasses)

	// Map the global scope up front.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case ast.Callable:
			v.newLabel(stmt)
		case *ast.ClassStmt:
			// synthesize a "New" function
			v.newLabelName(stmt.Name + "$new$")
			for _, cs := range stmt.Block.Statements {
				switch cs := cs.(type) {
				case ast.Callable:
					v.newLabel(cs)
				}
			}
		default:
			// nothing
		}
	}

	// Emit the global block's begin statement.
	globalLabel := &asm.Label{Name: "global$"}
	v.code = append(v.code, asm.Begin{
		SourceInfo: prog.Block.SourceInfo,
		Scope:      prog.Scope,
		Label:      globalLabel,
		NParams:    0,
		NLocals:    len(prog.Scope.Locals),
	})

	// Emit all global block non-decls.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt, *ast.FunctionStmt, *ast.ClassStmt:
		default:
			stmt.Visit(v)
		}
	}

	// If there is a module named main with no arguments, call it at the very end.
	finalReturnSi := prog.Block.SourceInfo.Tail()
	ref := prog.Scope.Lookup("main")
	if ref != nil && ref.ModuleStmt != nil && len(ref.ModuleStmt.Params) == 0 {
		scope := ref.ModuleStmt.Scope
		v.code = append(v.code, asm.Call{
			SourceInfo: ref.ModuleStmt.SourceInfo.Head(),
			Label:      v.refLabel(ref.ModuleStmt),
			NArgs:      len(scope.Params),
		})
		finalReturnSi = ref.ModuleStmt.SourceInfo.Head()
	}

	// terminate the program cleanly
	v.code = append(v.code, asm.End{
		SourceInfo: finalReturnSi,
		Label:      globalLabel,
	})

	// Now emit all modules and functions.
	for _, stmt := range prog.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt, *ast.FunctionStmt:
			stmt.Visit(v)
		case *ast.ClassStmt:
			v.emitNewFunction(stmt)
			for _, cs := range stmt.Block.Statements {
				switch cs := cs.(type) {
				case ast.Callable:
					cs.Visit(v)
				}
			}
		default:
			// nothing
		}
	}

	strings := make([]string, len(v.strings))
	for s, i := range v.strings {
		strings[i] = s
	}

	// compute vtables now that all the labels are known
	lblTables := make([][]*asm.Label, nClasses)
	classes := make([]string, nClasses)
	for i, c := range prog.Scope.Classes {
		classes[i] = c.Name
		for _, call := range c.Scope.Methods {
			lbl := v.refLabel(call)
			lblTables[i] = append(lblTables[i], lbl)
			v.vtables[i] = append(v.vtables[i], lbl.PC)
		}
	}

	return &asm.Assembly{
		GlobalScope: prog.Scope,
		Code:        v.code,
		Labels:      v.labels,
		Strings:     strings,
		Classes:     classes,
		Vtables:     lblTables,
	}
}

type Visitor struct {
	base.Visitor
	code    []asm.Inst
	labels  map[string]*asm.Label
	strings map[string]int
	vtables []asm.Vtable
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PreVisitVarDecl(vd *ast.VarDecl) bool {
	if vd.IsConst || vd.IsParam {
		return false
	}
	if needInitializer(vd) {
		v.varRefDecl(vd, nil, vd, true)
		v.emitInitializer(vd)
		v.store(vd)
	}
	return false
}

func needInitializer(vd *ast.VarDecl) bool {
	return vd.Expr != nil || len(vd.Dims) > 0 || vd.Type.IsFileType()
}

func (v *Visitor) emitInitializer(vd *ast.VarDecl) {
	if vd.Expr != nil {
		v.maybeCast(vd.Type, vd.Expr)
	} else if len(vd.Dims) > 0 {
		v.outputArrayInitializer(vd.SourceInfo, vd.Type.AsArrayType(), vd.Dims, nil)
	} else if vd.Type.IsFileType() {
		switch vd.Type.AsFileType() {
		case ast.OutputFile, ast.AppendFile:
			v.code = append(v.code, asm.Literal{
				SourceInfo: vd.SourceInfo,
				Typ:        vd.Type,
				Val:        lib.OutputFile{},
				Id:         0,
			})
		case ast.InputFile:
			v.code = append(v.code, asm.Literal{
				SourceInfo: vd.SourceInfo,
				Typ:        vd.Type,
				Val:        lib.InputFile{},
				Id:         0,
			})
		default:
			panic(vd.Type)
		}
	} else {
		panic(vd)
	}
}

func (v *Visitor) PreVisitDisplayStmt(d *ast.DisplayStmt) bool {
	for _, arg := range d.Exprs {
		if lit, ok := arg.(*ast.Literal); ok && lit.IsTabLiteral {
			v.code = append(v.code, asm.Literal{
				SourceInfo: arg.GetSourceInfo(),
				Typ:        lit.Type,
				Val:        lib.TabDisplay,
			})
		} else {
			arg.Visit(v)
		}
	}
	v.code = append(v.code, asm.LibCall{
		SourceInfo: d.GetSourceInfo(),
		Name:       "Display",
		Type:       ast.UnresolvedType,
		Index:      lib.IndexOf("Display"),
		NArg:       len(d.Exprs),
	})
	return false
}

func (v *Visitor) PreVisitInputStmt(i *ast.InputStmt) bool {
	v.emitAssignment(i.SourceInfo, i.Ref, func() {
		typ := i.Ref.GetType().AsPrimitive()
		name := "Input" + typ.String()
		v.code = append(v.code, asm.LibCall{
			SourceInfo: i.SourceInfo,
			Name:       name,
			Type:       typ,
			Index:      lib.IndexOf(name),
			NArg:       0,
		})
	})
	return false
}

func (v *Visitor) PreVisitSetStmt(i *ast.SetStmt) bool {
	v.emitAssignment(i.SourceInfo, i.Ref, func() {
		v.maybeCast(i.Ref.GetType(), i.Expr)
	})
	return false
}

func (v *Visitor) emitAssignment(si ast.SourceInfo, lhs ast.Expression, emitRhs func()) {
	if isStringCharAssignment(lhs) {
		ar := lhs.(*ast.ArrayRef)
		if ar.Qualifier.CanReference() {
			// str = stringWithCharUpdate(str string, idx int64, c byte)
			v.varRef(ar.Qualifier, true)
			// duplicate and deref as argument 0
			v.code = append(v.code, asm.Dup{SourceInfo: si, Skip: 0})
			v.code = append(v.code, asm.Deref{SourceInfo: si})
		} else {
			// eval and ignore: stringWithCharUpdate(str string, idx int64, c byte)
			ar.Qualifier.Visit(v)
		}
		ar.IndexExpr.Visit(v)
		emitRhs()
		v.code = append(v.code, asm.LibCall{
			SourceInfo: si,
			Name:       "$stringWithCharUpdate",
			Type:       ast.String,
			Index:      lib.IndexOf("$stringWithCharUpdate"),
			NArg:       3,
		})
		if ar.Qualifier.CanReference() {
			// store the final result
			v.store(si)
		} else {
			v.code = append(v.code, asm.Pop{SourceInfo: si})
		}
	} else {
		v.varRef(lhs, true)
		emitRhs()
		v.store(si)
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
	v.varRef(os.File, true)
	os.File.Visit(v)
	os.Name.Visit(v)
	name := "Open" + os.File.GetType().String()
	v.code = append(v.code, asm.LibCall{
		SourceInfo: os.SourceInfo,
		Name:       name,
		Type:       os.File.GetType(),
		Index:      lib.IndexOf(name),
		NArg:       2,
	})
	v.store(os)
	return false
}

func (v *Visitor) PreVisitCloseStmt(cs *ast.CloseStmt) bool {
	var name string
	if cs.File.GetType() == ast.InputFile {
		name = "CloseInputFile"
	} else {
		name = "CloseOutputFile"
	}
	cs.File.Visit(v)
	v.code = append(v.code, asm.LibCall{
		SourceInfo: cs.SourceInfo,
		Name:       name,
		Type:       ast.UnresolvedType,
		Index:      lib.IndexOf(name),
		NArg:       1,
	})
	return false
}

func (v *Visitor) PreVisitReadStmt(rs *ast.ReadStmt) bool {
	rs.File.Visit(v) // evaluate the file once
	for _, arg := range rs.Exprs {
		v.varRef(arg, true)

		// dup the file
		v.code = append(v.code, asm.Dup{
			SourceInfo: rs.GetSourceInfo(),
			Skip:       1,
		})

		name := "Read" + arg.GetType().String()
		v.code = append(v.code, asm.LibCall{
			SourceInfo: rs.GetSourceInfo(),
			Name:       name,
			Type:       ast.UnresolvedType,
			Index:      lib.IndexOf(name),
			NArg:       1,
		})
		v.store(rs)
	}
	// remove the original file var
	v.code = append(v.code, asm.Pop{SourceInfo: rs.GetSourceInfo()})
	return false
}

func (v *Visitor) PreVisitWriteStmt(ws *ast.WriteStmt) bool {
	ws.File.Visit(v)
	for _, arg := range ws.Exprs {
		arg.Visit(v)
	}
	v.code = append(v.code, asm.LibCall{
		SourceInfo: ws.GetSourceInfo(),
		Name:       "WriteFile",
		Type:       ast.UnresolvedType,
		Index:      lib.IndexOf("WriteFile"),
		NArg:       len(ws.Exprs) + 1,
	})
	return false
}

func (v *Visitor) PreVisitDeleteStmt(ds *ast.DeleteStmt) bool {
	ds.File.Visit(v)
	v.code = append(v.code, asm.LibCall{
		SourceInfo: ds.GetSourceInfo(),
		Name:       "DeleteFile",
		Type:       ast.UnresolvedType,
		Index:      lib.IndexOf("DeleteFile"),
		NArg:       1,
	})
	return false
}

func (v *Visitor) PreVisitRenameStmt(rs *ast.RenameStmt) bool {
	rs.OldFile.Visit(v)
	rs.NewFile.Visit(v)
	v.code = append(v.code, asm.LibCall{
		SourceInfo: rs.GetSourceInfo(),
		Name:       "RenameFile",
		Type:       ast.UnresolvedType,
		Index:      lib.IndexOf("RenameFile"),
		NArg:       2,
	})
	return false
}

func (v *Visitor) PreVisitIfStmt(is *ast.IfStmt) bool {
	endLabel := &asm.Label{Name: "endif"}
	for _, cb := range is.Cases {
		var lbl *asm.Label
		if cb.Expr != nil {
			cb.Expr.Visit(v)
			lbl = &asm.Label{Name: "else"}
			v.code = append(v.code, asm.JumpFalse{SourceInfo: cb.SourceInfo, Label: lbl})
		}
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, asm.Jump{SourceInfo: si, Label: endLabel})

		if lbl != nil {
			lbl.PC = len(v.code)
		}
	}
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitSelectStmt(ss *ast.SelectStmt) bool {
	endLabel := &asm.Label{Name: "endif"}

	// Evaluate the switch expr first.
	v.maybeCast(ss.Type, ss.Expr)
	hasDefault := false
	for _, cb := range ss.Cases {
		var lbl *asm.Label
		if cb.Expr != nil {
			// Duplicate the switch expression in case we fail.
			v.code = append(v.code, asm.Dup{SourceInfo: cb.SourceInfo})
			v.maybeCast(ss.Type, cb.Expr)
			v.code = append(v.code, makeBinaryOp(ss.Type.AsPrimitive(), cb.Expr, ast.EQ))
			lbl = &asm.Label{Name: "case"}
			v.code = append(v.code, asm.JumpFalse{SourceInfo: cb.SourceInfo, Label: lbl})
		} else {
			hasDefault = true
		}
		// we selected this block; remove the original switch expr
		v.code = append(v.code, asm.Pop{SourceInfo: cb.SourceInfo})
		cb.Block.Visit(v)

		// setup a jump to the end of this block
		si := ast.SourceInfo{Start: cb.Block.End, End: cb.End}
		v.code = append(v.code, asm.Jump{SourceInfo: si, Label: endLabel})

		if lbl != nil {
			lbl.PC = len(v.code)
		}
	}

	if !hasDefault {
		// remove the original switch expr
		v.code = append(v.code, asm.Pop{SourceInfo: ss.SourceInfo.Tail()})
	}

	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitDoStmt(ds *ast.DoStmt) bool {
	startLabel := &asm.Label{Name: "do", PC: len(v.code)}
	ds.Block.Visit(v)
	ds.Expr.Visit(v)
	if ds.Until {
		v.code = append(v.code, asm.JumpFalse{SourceInfo: ds.Expr.GetSourceInfo(), Label: startLabel})
	} else {
		v.code = append(v.code, asm.JumpTrue{SourceInfo: ds.Expr.GetSourceInfo(), Label: startLabel})
	}
	return false
}

func (v *Visitor) PreVisitWhileStmt(ws *ast.WhileStmt) bool {
	startLabel := &asm.Label{Name: "while", PC: len(v.code)}
	endLabel := &asm.Label{Name: "wend", PC: 0}
	ws.Expr.Visit(v)
	v.code = append(v.code, asm.JumpFalse{SourceInfo: ws.Expr.GetSourceInfo(), Label: endLabel})
	ws.Block.Visit(v)
	v.code = append(v.code, asm.Jump{SourceInfo: ws.Tail(), Label: startLabel})
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
	endLabel := &asm.Label{Name: "fend", PC: 0}

	// store
	v.varRef(fs.Ref, true)
	// leave an extra ref so we can test after
	v.code = append(v.code, asm.Dup{SourceInfo: fs.Ref.GetSourceInfo(), Skip: 0})
	v.maybeCast(refType, fs.StartExpr)
	v.store(fs.StartExpr)

	startLabel := &asm.Label{Name: "for", PC: len(v.code)}

	// test
	v.code = append(v.code, asm.Deref{SourceInfo: fs.Ref.GetSourceInfo()})
	v.maybeCast(refType, fs.StopExpr)
	switch refType {
	case ast.Integer:
		v.code = append(v.code, asm.BinOpInt{SourceInfo: fs.StopExpr.GetSourceInfo(), Op: op})
	case ast.Real:
		v.code = append(v.code, asm.BinOpReal{SourceInfo: fs.StopExpr.GetSourceInfo(), Op: op})
	default:
		panic(refType)
	}
	v.code = append(v.code, asm.JumpFalse{SourceInfo: fs.StopExpr.GetSourceInfo(), Label: endLabel})

	fs.Block.Visit(v)

	// post loop increment+jump
	si := fs.Ref.GetSourceInfo()

	// leave an extra ref so we can test after
	v.varRef(fs.Ref, true)
	v.code = append(v.code, asm.Dup{SourceInfo: fs.Ref.GetSourceInfo(), Skip: 0})
	switch refType {
	case ast.Integer:
		v.code = append(v.code, asm.IncrInt{SourceInfo: si, Val: intVal})
	case ast.Real:
		v.code = append(v.code, asm.IncrReal{SourceInfo: si, Val: floatVal})
	default:
		panic(refType)
	}

	v.code = append(v.code, asm.Jump{SourceInfo: si, Label: startLabel})
	endLabel.PC = len(v.code)
	return false
}

func (v *Visitor) PreVisitForEachStmt(fs *ast.ForEachStmt) bool {
	// attribute all the for/step/jumps to top line of the for loop
	si := fs.SourceInfo.Head()
	endLabel := &asm.Label{Name: "fend", PC: 0}

	// initialize the index and array expressions
	fs.IndexTemp.Visit(v)
	fs.ArrayTemp.Visit(v)

	startLabel := &asm.Label{Name: "for", PC: len(v.code)}

	indexTempExpr := &ast.VariableExpr{
		SourceInfo: si,
		Name:       fs.IndexTemp.Name,
		Ref:        fs.IndexTemp,
		Type:       fs.IndexTemp.Type,
	}

	arrayTempExpr := &ast.VariableExpr{
		SourceInfo: si,
		Name:       fs.ArrayTemp.Name,
		Ref:        fs.ArrayTemp,
		Type:       fs.ArrayTemp.Type,
	}

	// if !(idx < len(arr)) goto end
	indexTempExpr.Visit(v)
	arrayTempExpr.Visit(v)
	v.code = append(v.code, asm.ArrayLen{SourceInfo: si})
	v.code = append(v.code, asm.BinOpInt{SourceInfo: si, Op: ast.LT})
	v.code = append(v.code, asm.JumpFalse{SourceInfo: si, Label: endLabel})

	// ref = arr[idx]
	v.varRef(fs.Ref, true)
	arrayTempExpr.Visit(v)
	indexTempExpr.Visit(v)
	v.code = append(v.code, asm.ArrayVal{SourceInfo: si, OffsetType: asm.OffsetTypeArray})
	v.store(si)

	fs.Block.Visit(v)

	// post loop increment+jump
	si = fs.SourceInfo.Tail()
	v.varRef(indexTempExpr, true)
	v.code = append(v.code, asm.IncrInt{SourceInfo: si, Val: 1})
	v.code = append(v.code, asm.Jump{SourceInfo: si, Label: startLabel})
	endLabel.PC = len(v.code)

	return false
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	v.code = append(v.code, asm.Return{
		SourceInfo: rs.SourceInfo,
		NVal:       1,
	})
}

func (v *Visitor) PreVisitCallStmt(cs *ast.CallStmt) bool {
	si := cs.SourceInfo
	if isExternalDeleteInsert(cs) {
		// special case this, e.g.: s := insertString(s, pos, add)
		// str = stringWithCharUpdate(str string, idx int64, c byte)
		v.varRef(cs.Args[0], true)
		// duplicate and deref as argument 0
		v.code = append(v.code, asm.Dup{SourceInfo: si, Skip: 0})
		v.code = append(v.code, asm.Deref{SourceInfo: si})
		// emit the rest of the args
		v.outputArguments(cs.Args[1:], cs.Ref.Params[1:])
		v.code = append(v.code, asm.LibCall{
			SourceInfo: si,
			Name:       cs.Name,
			Type:       ast.String,
			Index:      lib.IndexOf(cs.Name),
			NArg:       len(cs.Args),
		})
		// store the result
		v.store(si)
		return false
	}

	nArg := len(cs.Args)
	if cs.Qualifier != nil {
		cs.Qualifier.Visit(v)
		nArg++
	}
	v.outputArguments(cs.Args, cs.Ref.Params)
	if cs.Ref.IsExternal {
		v.code = append(v.code, asm.LibCall{
			SourceInfo: si,
			Name:       cs.Name,
			Type:       ast.UnresolvedType,
			Index:      lib.IndexOf(cs.Name),
			NArg:       nArg,
		})
	} else if cs.Qualifier != nil && !cs.Ref.IsConstructor {
		v.code = append(v.code, asm.VCall{
			SourceInfo: si,
			Class:      cs.Ref.Enclosing.GetName(),
			Name:       cs.Name,
			Index:      cs.Ref.Id,
			NArgs:      nArg,
		})
	} else {
		v.code = append(v.code, asm.Call{
			SourceInfo: si,
			Label:      v.refLabel(cs.Ref),
			NArgs:      nArg,
		})
	}
	return false
}

func isExternalDeleteInsert(cs *ast.CallStmt) bool {
	return cs.Ref.IsExternal && (cs.Name == "delete" || cs.Name == "insert")
}

func (v *Visitor) PreVisitModuleStmt(ms *ast.ModuleStmt) bool {
	return v.preVisitCallable(ms)
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.postVisitCallable(ms)
}

func (v *Visitor) PreVisitFunctionStmt(fs *ast.FunctionStmt) bool {
	return v.preVisitCallable(fs)
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	v.postVisitCallable(fs)
}

func (v *Visitor) preVisitCallable(callable ast.Callable) bool {
	lbl := v.refLabel(callable)
	lbl.PC = len(v.code)
	scope := callable.GetScope()
	v.code = append(v.code, asm.Begin{
		SourceInfo: callable.GetSourceInfo(),
		Scope:      scope,
		Label:      lbl,
		NParams:    len(scope.Params),
		NLocals:    len(scope.Locals),
	})
	return true
}

func (v *Visitor) postVisitCallable(callable ast.Callable) bool {
	v.code = append(v.code, asm.End{
		SourceInfo: callable.GetSourceInfo().Tail(),
		Label:      v.refLabel(callable),
	})
	return true
}

func (v *Visitor) emitNewFunction(cs *ast.ClassStmt) {
	si := cs.Head()
	fs := &ast.FunctionStmt{
		SourceInfo: si,
		Name:       "new$",
		Type:       cs.Type,
		Block:      &ast.Block{SourceInfo: si},
		Enclosing:  cs.Type,
	}
	fs.Scope = ast.NewFunctionScope(fs, cs.Scope)
	vd := &ast.VarDecl{
		SourceInfo: si,
		Name:       "this",
		Type:       cs.Type,
	}
	fs.Scope.AddVariable(vd)

	v.preVisitCallable(fs)

	// this = New Thing()
	v.code = append(v.code, asm.LocalRef{
		SourceInfo: si,
		Name:       "this",
		Index:      0,
	})
	v.code = append(v.code, asm.ObjNew{
		SourceInfo: si,
		Type:       cs.Type,
		Vtable:     &v.vtables[cs.Type.Class.Id],
		NFields:    len(cs.Type.Scope.Fields),
	})
	v.store(si)

	// emit initializers
	for _, field := range cs.Scope.Fields {
		v.code = append(v.code, asm.LocalVal{
			SourceInfo: si,
			Name:       "this",
			Index:      0,
		})
		v.code = append(v.code, asm.FieldRef{
			SourceInfo: si,
			Name:       field.Name,
			Index:      field.Id,
		})
		if needInitializer(field) {
			v.emitInitializer(field)
		} else {
			// always zero-init fields regardless
			v.zero(si, field.Type)
		}
		v.store(si)
	}

	v.code = append(v.code, asm.LocalVal{
		SourceInfo: si,
		Name:       "this",
		Index:      0,
	})
	v.code = append(v.code, asm.Return{
		SourceInfo: si,
		NVal:       1,
	})
	v.postVisitCallable(fs)
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

	v.code = append(v.code, asm.Literal{
		SourceInfo: l.SourceInfo,
		Typ:        l.Type,
		Val:        l.Val,
		Id:         id,
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
		v.code = append(v.code, asm.UnaryOpInt{SourceInfo: l.SourceInfo, Op: l.Op})
	case ast.Real:
		v.code = append(v.code, asm.UnaryOpFloat{SourceInfo: l.SourceInfo, Op: l.Op})
	case ast.Boolean:
		v.code = append(v.code, asm.UnaryOpBool{SourceInfo: l.SourceInfo, Op: l.Op})
	default:
		panic(l.Type)
	}
}

func (v *Visitor) PreVisitVariableExpr(ve *ast.VariableExpr) bool {
	v.varRef(ve, false) // if we get here, we need a value
	return false
}

func (v *Visitor) PreVisitCallExpr(ce *ast.CallExpr) bool {
	nArg := len(ce.Args)
	if ce.Qualifier != nil {
		ce.Qualifier.Visit(v)
		nArg++
	}
	v.outputArguments(ce.Args, ce.Ref.Params)
	if ce.Ref.IsExternal {
		v.code = append(v.code, asm.LibCall{
			SourceInfo: ce.SourceInfo,
			Name:       ce.Name,
			Type:       ce.Type.AsPrimitive(),
			Index:      lib.IndexOf(ce.Name),
			NArg:       nArg,
		})
	} else if ce.Qualifier != nil {
		v.code = append(v.code, asm.VCall{
			SourceInfo: ce.SourceInfo,
			Class:      ce.Ref.Enclosing.GetName(),
			Name:       ce.Name,
			Index:      ce.Ref.Id,
			NArgs:      nArg,
		})
	} else {
		v.code = append(v.code, asm.Call{
			SourceInfo: ce.SourceInfo,
			Label:      v.refLabel(ce.Ref),
			NArgs:      nArg,
		})
	}
	return false
}

func (v *Visitor) PreVisitArrayRef(arr *ast.ArrayRef) bool {
	arr.Qualifier.Visit(v)
	arr.IndexExpr.Visit(v)
	typ := asm.OffsetTypeArray
	if arr.Qualifier.GetType() == ast.String {
		typ = asm.OffsetTypeString
	}

	// if we get here we need a value
	v.code = append(v.code, asm.ArrayVal{
		SourceInfo: arr.SourceInfo,
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

func (v *Visitor) PreVisitNewExpr(ne *ast.NewExpr) bool {
	// Call the new function
	v.code = append(v.code, asm.Call{
		SourceInfo: ne.SourceInfo,
		Label:      v.refLabelName(ne.Name + "$new$"),
		NArgs:      0,
	})

	// Maybe call constructor.
	if ctor := ne.Ctor; ctor != nil {
		// duplicate the `this` pointer so it remains on the stack when the ctor returns
		v.code = append(v.code, asm.Dup{SourceInfo: ne.SourceInfo, Skip: 0})
		v.outputArguments(ne.Args, ctor.Params)
		v.code = append(v.code, asm.Call{
			SourceInfo: ne.SourceInfo,
			Label:      v.refLabel(ctor),
			NArgs:      len(ctor.Params) + 1,
		})
	}
	return false
}

func (v *Visitor) PreVisitThisRef(ref *ast.ThisRef) bool {
	v.code = append(v.code, asm.ParamVal{
		SourceInfo: ref.SourceInfo,
		Name:       "this",
		Index:      0,
	})
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
				v.zero(newLitSi, typ)
			}
		}
	} else {
		for i := 0; i < dims[0]; i++ {
			exprs = v.outputArrayInitializer(newLitSi, t.ElementType.AsArrayType(), dims[1:], exprs)
		}
	}
	v.code = append(v.code, &asm.ArrayNew{
		SourceInfo: newLitSi,
		Typ:        t,
		Size:       dims[0],
	})
	return exprs
}

func (v *Visitor) maybeCast(dstType ast.Type, exp ast.Expression) {
	exp.Visit(v)
	if dstType == ast.Real && exp.GetType() == ast.Integer {
		v.code = append(v.code, asm.IntToReal{SourceInfo: exp.GetSourceInfo()})
	} else if dstType == ast.Integer && exp.GetType() == ast.Real {
		v.code = append(v.code, asm.RealToInt{SourceInfo: exp.GetSourceInfo()})
	}
}

func (v *Visitor) varRef(expr ast.Expression, needRef bool) {
	switch exp := expr.(type) {
	case *ast.VariableExpr:
		v.varRefDecl(expr, exp.Qualifier, exp.Ref, needRef)
	case *ast.ArrayRef:
		typ := asm.OffsetTypeArray
		if exp.Qualifier.GetType() == ast.String {
			typ = asm.OffsetTypeString
		}
		exp.Qualifier.Visit(v)
		exp.IndexExpr.Visit(v)
		if needRef {
			v.code = append(v.code, asm.ArrayRef{
				SourceInfo: exp.GetSourceInfo(),
				OffsetType: typ,
			})
		} else {
			v.code = append(v.code, asm.ArrayVal{
				SourceInfo: exp.GetSourceInfo(),
				OffsetType: typ,
			})
		}
	default:
		panic("implement me")
	}
}

func (v *Visitor) varRefDecl(hs ast.HasSourceInfo, qual ast.Expression, decl *ast.VarDecl, needRef bool) {
	isRef := decl.IsRef
	if decl.IsConst {
		if isRef || needRef {
			panic("here")
		}
		val := decl.Expr.ConstEval()
		if val == nil {
			panic("here")
		}
		v.code = append(v.code, asm.Literal{
			SourceInfo: hs.GetSourceInfo(),
			Typ:        decl.Type.AsPrimitive(),
			Val:        val,
		})
	} else if decl.Scope.IsGlobal {
		if isRef {
			panic("here")
		}
		if needRef {
			v.code = append(v.code, asm.GlobalRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		} else {
			v.code = append(v.code, asm.GlobalVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		}
		return
	} else if decl.IsParam {
		// Param
		if isRef == needRef {
			// if we have a ref and need a ref, or we have a val and need a val, we good
			v.code = append(v.code, asm.ParamVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		} else if needRef {
			v.code = append(v.code, asm.ParamRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		} else {
			// Take the value (it's a reference) then derefence it.
			v.code = append(v.code, asm.ParamPtr{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		}
	} else if qual != nil {
		// Field
		if isRef {
			panic("here")
		}
		qual.Visit(v)
		if needRef {
			v.code = append(v.code, asm.FieldRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		} else {
			v.code = append(v.code, asm.FieldVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		}
	} else {
		// Local
		if isRef {
			panic("here")
		}
		if needRef {
			v.code = append(v.code, asm.LocalRef{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		} else {
			v.code = append(v.code, asm.LocalVal{
				SourceInfo: hs.GetSourceInfo(),
				Name:       decl.Name,
				Index:      decl.Id,
			})
		}
	}
}

func (v *Visitor) store(hs ast.HasSourceInfo) {
	v.code = append(v.code, asm.Store{SourceInfo: hs.GetSourceInfo()})
}

func (v *Visitor) outputArguments(args []ast.Expression, params []*ast.VarDecl) {
	if len(args) != len(params) {
		panic("argument count mismatch")
	}
	for i, arg := range args {
		param := params[i]
		if param.IsRef {
			v.varRef(arg, true)
		} else if param.Type.IsArrayType() {
			arg.Visit(v)
			at := param.Type.AsArrayType()
			v.code = append(v.code, asm.ArrayClone{
				SourceInfo: arg.GetSourceInfo(),
				Typ:        at,
				NDims:      at.NDims,
			})
		} else {
			v.maybeCast(param.Type, arg)
		}
	}
}

func (v *Visitor) newLabel(callable ast.Callable) *asm.Label {
	name := callable.GetName()
	if enc := callable.GetEnclosing(); enc != nil {
		name = enc.GetName() + "$" + name
	}
	return v.newLabelName(name)
}

func (v *Visitor) newLabelName(name string) *asm.Label {
	if v.labels == nil {
		v.labels = map[string]*asm.Label{}
	}
	if _, ok := v.labels[name]; ok {
		panic(name)
	}
	l := &asm.Label{Name: name}
	v.labels[name] = l
	return l
}

func (v *Visitor) refLabel(callable ast.Callable) *asm.Label {
	name := callable.GetName()
	if enc := callable.GetEnclosing(); enc != nil {
		name = enc.GetName() + "$" + name
	}
	return v.refLabelName(name)
}

func (v *Visitor) refLabelName(name string) *asm.Label {
	if _, ok := v.labels[name]; !ok {
		panic(name)
	}
	return v.labels[name]
}

func (v *Visitor) zero(si ast.SourceInfo, typ ast.Type) {
	v.code = append(v.code, asm.Literal{
		SourceInfo: si,
		Typ:        typ,
		Val:        asm.ZeroValue(typ),
		Id:         0,
	})
}

func makeBinaryOp(t ast.PrimitiveType, hs ast.HasSourceInfo, op ast.Operator) asm.Inst {
	si := hs.GetSourceInfo()
	switch t {
	case ast.Integer:
		return asm.BinOpInt{SourceInfo: si, Op: op}
	case ast.Real:
		return asm.BinOpReal{SourceInfo: si, Op: op}
	case ast.String:
		return asm.BinOpStr{SourceInfo: si, Op: op}
	case ast.Character:
		return asm.BinOpChar{SourceInfo: si, Op: op}
	case ast.Boolean:
		return asm.BinOpBool{SourceInfo: si, Op: op}
	default:
		panic(t)
	}
}
