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
	return fmt.Sprintf("%s(%d)", l.Name, l.PC)
}

type Dup struct {
	ast.SourceInfo
}

func (i Dup) Exec(p *Execution) {
	v := p.Pop()
	p.Push(v)
	p.Push(v)
}

func (i Dup) String() string {
	return "dup"
}

type Pop struct {
	ast.SourceInfo
}

func (i Pop) Exec(p *Execution) {
	p.Pop()
}

func (i Pop) String() string {
	return "pop"
}
