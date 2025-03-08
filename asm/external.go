package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"github.com/dragonsinth/gaddis/lib"
	"math/rand"
	"reflect"
	"sort"
)

var (
	libFuncs, libIndex = makeLibFuncs()
	_                  = libFuncs
)

type ExecutionContext struct {
	Rng *rand.Rand
	builtins.IoContext
}

func (ctx *ExecutionContext) random(lo int64, hi int64) int64 {
	return lo + ctx.Rng.Int63n(hi-lo+1)
}

type LibFunc struct {
	Name string
	Func reflect.Value
}

func (ctx *ExecutionContext) CreateLibrary() []LibFunc {
	funcMap := map[string]any{
		"Display":      ctx.Display,
		"InputInteger": ctx.InputInteger,
		"InputReal":    ctx.InputReal,
		"InputString":  ctx.InputString,
		"InputBoolean": ctx.InputBoolean,
	}
	for name, f := range lib.External {
		if _, ok := funcMap[name]; ok {
			panic(name)
		}
		funcMap[name] = f
	}
	funcMap["random"] = ctx.random // override random

	ret := make([]LibFunc, len(funcMap))
	for name, f := range funcMap {
		idx, ok := libIndex[name]
		if !ok {
			panic(name)
		}
		ret[idx] = LibFunc{Name: name, Func: reflect.ValueOf(f)}
	}
	return ret
}

func makeLibFuncs() ([]string, map[string]int) {
	funcs := []string{
		"Display",
		"InputInteger",
		"InputReal",
		"InputString",
		"InputBoolean",
	}
	index := map[string]int{}
	for name := range lib.External {
		funcs = append(funcs, name)
	}
	sort.Strings(funcs)
	for i, name := range funcs {
		index[name] = i
	}
	return funcs, index
}

func libFunc(name string) int {
	i, ok := libIndex[name]
	if !ok {
		panic(name)
	}
	return i
}

type LibCall struct {
	ast.SourceInfo
	Name  string
	Index int
	NArg  int
}

func (i LibCall) Exec(p *Execution) {
	args := p.PopN(i.NArg)
	fn := p.Lib[i.Index].Func
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
	return fmt.Sprintf("libcall %s(%d)", i.Name, i.NArg)
}
