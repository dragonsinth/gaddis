package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Label struct {
	Name string
	PC   int
}

func (l *Label) String() string {
	return fmt.Sprintf("%s:%s", PcRef(l.PC), l.Name)
}

type Dup struct {
	baseInst
	ast.SourceInfo
	Skip int
}

func (i Dup) Exec(p *Execution) {
	tip := len(p.Frame.Eval) - 1
	val := p.Frame.Eval[tip-i.Skip]
	p.Push(val)
}

func (i Dup) String() string {
	return "dup"
}

type Pop struct {
	baseInst
	ast.SourceInfo
}

func (i Pop) Exec(p *Execution) {
	p.Pop()
}

func (i Pop) String() string {
	return "pop"
}

type Deref struct {
	baseInst
	ast.SourceInfo
}

func (i Deref) Exec(p *Execution) {
	ref := p.Pop().(*any)
	p.Push(*ref)
}

func (i Deref) String() string {
	return "deref"
}
