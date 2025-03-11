package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

type Begin struct {
	baseInst
	NArgs int
	Label *Label
}

func (i Begin) Exec(p *Execution) {
}

func (i Begin) String() string {
	return fmt.Sprintf("begin(%d) :%s", i.NArgs, i.Label.Name)
}

func (i Begin) Sym() string {
	return i.Label.Name
}

type End struct {
	baseInst
	Label *Label
}

func (i End) Exec(p *Execution) {
	panic("unreachable")
}

func (i End) String() string {
	return fmt.Sprintf("end :%s", i.Label.Name)
}

func (i End) Sym() string {
	return i.Label.Name
}

type Call struct {
	baseInst
	Scope *ast.Scope
	Label *Label
}

func (i Call) Exec(p *Execution) {
	if len(p.Stack) >= MAX_STACK {
		panic("stack overflow")
	}

	nArg := len(i.Scope.Params)
	args := slices.Clone(p.PopN(nArg))
	locals := make([]any, len(i.Scope.Locals))

	p.Stack = append(p.Stack, Frame{
		Scope:  i.Scope,
		Start:  i.Label.PC,
		Return: p.PC,
		Args:   args,
		Locals: append(args, locals...),
		Eval:   make([]any, 0, 16),
	})
	p.Frame = &p.Stack[len(p.Stack)-1]
	p.PC = i.Label.PC - 1 // will advance to next instruction
}

func (i Call) String() string {
	return fmt.Sprintf("call(%d) :%s", len(i.Scope.Params), i.Label.Name)
}

func (i Call) Sym() string {
	return i.Label.Name
}

type Return struct {
	baseInst
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
	if i.NVal == 0 {
		return "return"
	} else {
		return fmt.Sprintf("return(%d)", i.NVal)
	}
}
