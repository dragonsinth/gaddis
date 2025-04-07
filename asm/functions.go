package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"reflect"
	"slices"
)

type Call struct {
	ast.SourceInfo
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
	ast.SourceInfo
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

type LibCall struct {
	ast.SourceInfo
	Name  string
	Type  ast.Type
	Index int
	NArg  int
}

func (i LibCall) Exec(p *Execution) {
	args := p.PopN(i.NArg)
	fn := p.Lib[i.Index].FuncPtr
	var ret []reflect.Value
	if fn.Type().IsVariadic() {
		var rArgs []reflect.Value
		for i, c := 0, fn.Type().NumIn()-1; i < c; i++ {
			rArgs = append(rArgs, reflect.ValueOf(args[0]))
			args = args[1:]
		}
		rArgs = append(rArgs, reflect.ValueOf(args))
		ret = fn.CallSlice(rArgs)
	} else {
		rArgs := make([]reflect.Value, i.NArg)
		for i, arg := range args {
			rArgs[i] = reflect.ValueOf(arg)
		}
		ret = fn.Call(rArgs)
	}

	switch len(ret) {
	case 0:
	case 1:
		p.Push(ret[0].Interface())
	default:
		panic(ret)
	}
}

func (i LibCall) String() string {
	return fmt.Sprintf("libcall(%d) %d:%s", i.NArg, i.Index, i.Name)
}

func (i LibCall) Sym() string {
	return i.Name
}
