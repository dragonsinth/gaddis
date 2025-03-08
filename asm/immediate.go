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
	p.Push(&p.Globals[i.Index])
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
	p.Push(p.Globals[i.Index])
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
	p.Push(p.Frame.Locals[i.Index])
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
	val := p.Frame.Locals[i.Index].(*any)
	p.Push(*val)
}

func (i LocalPtr) String() string {
	return fmt.Sprintf("*local %s(%d)", i.Name, i.Index)
}
