package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"strconv"
)

type UnaryOpInt struct {
	baseInst
	ast.SourceInfo
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
	ast.SourceInfo
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
	ast.SourceInfo
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

type IncrInt struct {
	baseInst
	ast.SourceInfo
	Val int64
}

func (i IncrInt) Exec(p *Execution) {
	ref := p.Pop().(*any)
	refVal := (*ref).(int64)
	*ref = refVal + i.Val
}

func (i IncrInt) String() string {
	return "incr int " + strconv.FormatInt(i.Val, 10)
}

type IncrReal struct {
	baseInst
	ast.SourceInfo
	Val float64
}

func (i IncrReal) Exec(p *Execution) {
	ref := p.Pop().(*any)
	refVal := (*ref).(float64)
	*ref = refVal + i.Val
}

func (i IncrReal) String() string {
	return "incr real " + strconv.FormatFloat(i.Val, 'g', -1, 64)
}
