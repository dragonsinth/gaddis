package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

type Begin struct {
	ast.SourceInfo
	Label *Label
}

func (i Begin) Exec(p *Execution) {
}

func (i Begin) String() string {
	return fmt.Sprintf("begin %s", i.Label)
}

type Call struct {
	ast.SourceInfo
	Scope *ast.Scope
	Label *Label
}

func (i Call) Exec(p *Execution) {
	nArg := len(i.Scope.Params)
	args := slices.Clone(p.PopN(nArg))
	locals := slices.Clone(args)
	for _, decl := range i.Scope.Locals {
		locals = append(locals, zeroValue(decl.Type))
	}

	p.Stack = append(p.Stack, Frame{
		Scope:  i.Scope,
		Return: p.PC,
		Args:   args,
		Locals: locals,
		Eval:   make([]any, 0, 16),
	})
	p.Frame = &p.Stack[len(p.Stack)-1]
	p.PC = i.Label.PC - 1 // will advance to next instruction
}

func (i Call) String() string {
	return fmt.Sprintf("call %s", i.Label)
}

type Return struct {
	ast.SourceInfo
	NVal int
}

func (i Return) Exec(p *Execution) {
	p.PC = p.Frame.Return
	p.Stack = p.Stack[:len(p.Stack)-1]
	if len(p.Stack) > 0 {
		if len(p.Frame.Eval) != i.NVal {
			panic(p.Frame.Eval)
		}
		// copy return value(s) from the old frame to the new frame
		rets := p.Frame.Eval
		p.Frame = &p.Stack[len(p.Stack)-1]
		p.Frame.Eval = append(p.Frame.Eval, rets...)
	} else {
		if len(p.Frame.Eval) != 0 {
			panic(p.Frame.Eval)
		}
		p.Frame = nil
	}
}

func (i Return) String() string {
	return fmt.Sprintf("return %d", i.NVal)
}
