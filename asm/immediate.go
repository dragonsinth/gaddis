package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/lib"
	"strconv"
)

type Literal struct {
	baseInst
	ast.SourceInfo
	Typ ast.Type
	Val any
	Id  int // only for strings
}

func (i Literal) Exec(p *Execution) {
	p.Push(i.Val)
}

func (i Literal) String() string {
	var typ, str string
	switch i.Typ {
	case ast.UnresolvedType:
		typ = "unk"
	case ast.Integer:
		typ = "int"
		str = strconv.FormatInt(i.Val.(int64), 10)
	case ast.Real:
		typ = "real"
		str = strconv.FormatFloat(i.Val.(float64), 'g', -1, 64)
	case ast.String:
		typ = "str"
		if i.Val == lib.TabDisplay {
			str = "tab"
		} else {
			str = fmt.Sprintf("[%d]", i.Id)
		}
	case ast.Character:
		typ = "chr"
		str = strconv.QuoteRune(rune(i.Val.(byte)))
	case ast.Boolean:
		typ = "bool"
		if i.Val.(bool) {
			str = "true"
		} else {
			str = "false"
		}
	case ast.OutputFile:
		typ = "out_file"
		str = "{}"
	case ast.AppendFile:
		typ = "app_file"
		str = "{}"
	case ast.InputFile:
		typ = "in_file"
		str = "{}"
	}
	return fmt.Sprintf("literal %s %s", typ, str)
}

var litTypes = []string{
	ast.UnresolvedType: "unk",
	ast.Integer:        "int",
	ast.Real:           "real",
	ast.String:         "str",
	ast.Character:      "char",
	ast.Boolean:        "bool",
	ast.OutputFile:     "out_file",
	ast.AppendFile:     "app_file",
	ast.InputFile:      "in_file",
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
	return fmt.Sprintf("&global[%d] #%s", i.Index, i.Name)
}

func (i GlobalRef) Sym() string {
	return i.Name
}

type GlobalVal struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i GlobalVal) Exec(p *Execution) {
	val := p.Stack[0].Locals[i.Index]
	if val == nil {
		panic(fmt.Sprintf("global variable %s read before assignment", i.Name))
	}
	p.Push(val)
}

func (i GlobalVal) String() string {
	return fmt.Sprintf("global[%d] #%s", i.Index, i.Name)
}

func (i GlobalVal) Sym() string {
	return i.Name
}

type ParamRef struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i ParamRef) Exec(p *Execution) {
	p.Push(&p.Frame.Params[i.Index])
}

func (i ParamRef) String() string {
	return fmt.Sprintf("&param[%d] #%s", i.Index, i.Name)
}

func (i ParamRef) Sym() string {
	return i.Name
}

type ParamVal struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i ParamVal) Exec(p *Execution) {
	val := p.Frame.Params[i.Index]
	if val == nil {
		// ths shouldn't really ever happen; this instruction is only for reading ref params, which should always
		// point to a valid memory location.
		panic(fmt.Sprintf("compiler bug!?: param %s read before assignment", i.Name))
	}
	p.Push(val)
}

func (i ParamVal) String() string {
	return fmt.Sprintf("param[%d] #%s", i.Index, i.Name)
}

func (i ParamVal) Sym() string {
	return i.Name
}

type ParamPtr struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i ParamPtr) Exec(p *Execution) {
	val := p.Frame.Params[i.Index]
	if val == nil {
		// ths shouldn't really ever happen; this instruction is only for reading ref params, which should always
		// point to a valid memory location.
		panic(fmt.Sprintf("compiler bug!?: param %s read before assignment", i.Name))
	}
	val = *val.(*any)
	if val == nil {
		// The _value_ however could be nil.
		panic(fmt.Sprintf("param variable %s read before assignment", i.Name))
	}
	p.Push(val)
}

func (i ParamPtr) String() string {
	return fmt.Sprintf("*param[%d] #%s", i.Index, i.Name)
}

func (i ParamPtr) Sym() string {
	return i.Name
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
	return fmt.Sprintf("&local[%d] #%s", i.Index, i.Name)
}

func (i LocalRef) Sym() string {
	return i.Name
}

type LocalVal struct {
	ast.SourceInfo
	Name  string
	Index int
}

func (i LocalVal) Exec(p *Execution) {
	val := p.Frame.Locals[i.Index]
	if val == nil {
		panic(fmt.Sprintf("local %s read before assignment", i.Name))
	}
	p.Push(val)
}

func (i LocalVal) String() string {
	return fmt.Sprintf("local[%d] #%s", i.Index, i.Name)
}

func (i LocalVal) Sym() string {
	return i.Name
}
