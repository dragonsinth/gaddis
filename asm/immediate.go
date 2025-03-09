package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Literal struct {
	ast.SourceInfo
	Val any
}

func (i Literal) Exec(p *Execution) {
	val := i.Val
	if v, ok := i.Val.(string); ok {
		val = []byte(v)
	}
	p.Push(val)
}

func (i Literal) String() string {
	return fmt.Sprintf("literal %#v", i.Val)
}

type GlobalRef struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i GlobalRef) Exec(p *Execution) {
	p.Push(&p.Stack[0].Locals[i.Index])
}

func (i GlobalRef) String() string {
	return fmt.Sprintf("&global %s(%d)", i.Name, i.Index)
}

type GlobalVal struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i GlobalVal) Exec(p *Execution) {
	val := p.Stack[0].Locals[i.Index]
	if val == nil {
		panic(fmt.Sprintf("global variable %s read before assignment", p.Stack[0].Scope.Locals[i.Index].Name))
	}
	p.Push(val)
}

func (i GlobalVal) String() string {
	return fmt.Sprintf("global %s(%d)", i.Name, i.Index)
}

type LocalRef struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i LocalRef) Exec(p *Execution) {
	p.Push(&p.Frame.Locals[i.Index])
}

func (i LocalRef) String() string {
	return fmt.Sprintf("&local %s(%d)", i.Name, i.Index)
}

type LocalVal struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i LocalVal) Exec(p *Execution) {
	val := p.Frame.Locals[i.Index]
	if val == nil {
		decl := getDeclInScope(p.Frame, i.Index)
		if decl.IsParam {
			// ths shouldn't really ever happen; this instruction is only for reading ref params, which should always
			// point to a valid memory location.
			panic(fmt.Sprintf("compiler bug!?: param %s read before assignment", decl.Name))
		} else {
			panic(fmt.Sprintf("local %s read before assignment", decl.Name))
		}
	}
	p.Push(val)
}

func (i LocalVal) String() string {
	return fmt.Sprintf("local %s(%d)", i.Name, i.Index)
}

type LocalPtr struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i LocalPtr) Exec(p *Execution) {
	val := p.Frame.Locals[i.Index]
	if val == nil {
		// ths shouldn't really ever happen; this instruction is only for reading ref params, which should always
		// point to a valid memory location.
		decl := getDeclInScope(p.Frame, i.Index)
		panic(fmt.Sprintf("compiler bug!?: local variable %s read before assignment", decl.Name))
	}
	val = *val.(*any)
	if val == nil {
		// The _value_ however could be nil.
		decl := getDeclInScope(p.Frame, i.Index)
		panic(fmt.Sprintf("local variable %s read before assignment", decl.Name))
	}
	p.Push(val)
}

func (i LocalPtr) String() string {
	return fmt.Sprintf("*local %s(%d)", i.Name, i.Index)
}

func getDeclInScope(fr *Frame, idx int) *ast.VarDecl {
	nParams := len(fr.Scope.Params)
	if idx < nParams {
		return fr.Scope.Params[idx]
	} else {
		idx -= nParams
		return fr.Scope.Locals[idx]
	}
}
