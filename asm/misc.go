package asm

import (
	"fmt"
)

type Label struct {
	Name string
	PC   int
}

func (l *Label) String() string {
	return fmt.Sprintf("%d:%s", l.PC, l.Name)
}

type Dup struct {
	baseInst
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
	baseInst
}

func (i Pop) Exec(p *Execution) {
	p.Pop()
}

func (i Pop) String() string {
	return "pop"
}
