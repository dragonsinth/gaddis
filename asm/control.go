package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

type Begin struct {
	ast.SourceInfo
	Scope   *ast.Scope
	Label   *Label
	NParams int
	NLocals int
}

func (i Begin) Exec(p *Execution) {
	p.Frame.Params = slices.Clone(p.Frame.Args)
	p.Frame.Locals = make([]any, i.NLocals)
	p.Frame.Eval = make([]any, 0, 16)
}

func (i Begin) String() string {
	return fmt.Sprintf("begin(%d,%d) :%s", i.NParams, i.NLocals, i.Label.Name)
}

func (i Begin) Sym() string {
	return i.Label.Name
}

type End struct {
	ast.SourceInfo
	Label *Label
}

func (i End) Exec(p *Execution) {
	if len(p.Frame.Eval) != 0 {
		panic(p.Frame.Eval)
	}
	p.PC = p.Frame.Return
	p.Stack = p.Stack[:len(p.Stack)-1]
	if len(p.Stack) > 0 {
		p.Frame = &p.Stack[len(p.Stack)-1]
	} else {
		p.Frame = nil
	}
}

func (i End) String() string {
	return fmt.Sprintf("end :%s", i.Label.Name)
}

func (i End) Sym() string {
	return i.Label.Name
}

// Halt terminates the program with an expected number of values left on the eval stack.
// Used for expression evaluation.
type Halt struct {
	baseInst
	ast.SourceInfo
	NVal int
}

func (i Halt) Exec(p *Execution) {
	if len(p.Frame.Eval) != i.NVal {
		panic(p.Frame.Eval)
	}
	p.Frame = nil
}

func (i Halt) String() string {
	if i.NVal == 0 {
		return "halt"
	} else {
		return fmt.Sprintf("halt(%d)", i.NVal)
	}
}

type Jump struct {
	ast.SourceInfo
	Label *Label
}

func (i Jump) Exec(p *Execution) {
	p.PC = i.Label.PC - 1 // will advance to next instruction
}

func (i Jump) String() string {
	return fmt.Sprintf("jump %s", PcRef(i.Label.PC))
}

func (i Jump) Sym() string {
	return i.Label.Name
}

type JumpFalse struct {
	ast.SourceInfo
	Label *Label
}

func (i JumpFalse) Exec(p *Execution) {
	v := p.Pop().(bool)
	if !v {
		p.PC = i.Label.PC - 1 // will advance to next instruction
	}
}

func (i JumpFalse) String() string {
	return fmt.Sprintf("jump false %s", PcRef(i.Label.PC))
}

func (i JumpFalse) Sym() string {
	return i.Label.Name
}

type JumpTrue struct {
	ast.SourceInfo
	Label *Label
}

func (i JumpTrue) Exec(p *Execution) {
	v := p.Pop().(bool)
	if v {
		p.PC = i.Label.PC - 1 // will advance to next instruction
	}
}

func (i JumpTrue) String() string {
	return fmt.Sprintf("jump true %s", PcRef(i.Label.PC))
}

func (i JumpTrue) Sym() string {
	return i.Label.Name
}
