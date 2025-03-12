package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"slices"
)

type Begin struct {
	baseInst
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
	baseInst
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

type Call struct {
	baseInst
	Label *Label
	NArgs int
}

func (i Call) Exec(p *Execution) {
	if len(p.Stack) >= MaxStack {
		panic("stack overflow")
	}

	be := p.Code[i.Label.PC].(Begin)

	args := slices.Clone(p.PopN(i.NArgs))

	p.Stack = append(p.Stack, Frame{
		Scope:  be.Scope,
		Start:  i.Label.PC,
		Return: p.PC,
		Args:   args,
	})
	p.Frame = &p.Stack[len(p.Stack)-1]
	p.PC = i.Label.PC - 1 // will advance to next instruction
}

func (i Call) String() string {
	return fmt.Sprintf("call(%d) %s", i.NArgs, i.Label)
}

func (i Call) Sym() string {
	return i.Label.Name
}

type Return struct {
	baseInst
	NVal int
}

func (i Return) Exec(p *Execution) {
	if len(p.Frame.Eval) != i.NVal {
		panic(p.Frame.Eval)
	}
	p.PC = p.Frame.Return
	p.Stack = p.Stack[:len(p.Stack)-1]

	// copy return value(s) from the old frame to the new frame
	rets := p.Frame.Eval
	p.Frame = &p.Stack[len(p.Stack)-1]
	p.Frame.Eval = append(p.Frame.Eval, rets...)
}

func (i Return) String() string {
	if i.NVal == 0 {
		return "return"
	} else {
		return fmt.Sprintf("return(%d)", i.NVal)
	}
}
