package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type UnaryOpInt struct {
	baseInst
	Op ast.Operator
}

func (i UnaryOpInt) Exec(p *Execution) {
	a := p.Pop().(int64)
	switch i.Op {
	case ast.NEG:
		p.Push(-a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpInt) String() string {
	return fmt.Sprintf("%s int", i.Op.Name())
}

type UnaryOpFloat struct {
	baseInst
	Op ast.Operator
}

func (i UnaryOpFloat) Exec(p *Execution) {
	a := toFloat64(p.Pop())
	switch i.Op {
	case ast.NEG:
		p.Push(-a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpFloat) String() string {
	return fmt.Sprintf("%s real", i.Op.Name())
}

type UnaryOpBool struct {
	baseInst
	Op ast.Operator
}

func (i UnaryOpBool) Exec(p *Execution) {
	a := p.Pop().(bool)
	switch i.Op {
	case ast.NOT:
		p.Push(!a)
	default:
		panic(i.Op)
	}
}

func (i UnaryOpBool) String() string {
	return i.Op.Name()
}
