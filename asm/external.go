package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"reflect"
)

type LibCall struct {
	baseInst
	Name  string
	Type  ast.PrimitiveType
	Index int
	NArg  int
}

func (i LibCall) Exec(p *Execution) {
	args := p.PopN(i.NArg)
	fn := p.Lib[i.Index].FuncPtr
	var ret []reflect.Value
	if fn.Type().IsVariadic() {
		ret = fn.CallSlice([]reflect.Value{reflect.ValueOf(args)})
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
