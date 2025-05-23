package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Store struct {
	baseInst
	ast.SourceInfo
}

func (i Store) Exec(p *Execution) {
	val := p.Pop()
	ref := p.Pop().(*any)
	*ref = val
}

func (i Store) String() string {
	return "store"
}

type BinOpInt struct {
	baseInst
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpInt) Exec(p *Execution) {
	b := p.Pop().(int64)
	a := p.Pop().(int64)
	p.Push(ast.IntegerOp(i.Op, a, b))
}

func (i BinOpInt) String() string {
	return fmt.Sprintf("%s int", i.Op.Name())
}

type BinOpReal struct {
	baseInst
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpReal) Exec(p *Execution) {
	b := toFloat64(p.Pop())
	a := toFloat64(p.Pop())
	p.Push(ast.RealOp(i.Op, a, b))
}

func (i BinOpReal) String() string {
	return fmt.Sprintf("%s real", i.Op.Name())
}

type BinOpStr struct {
	baseInst
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpStr) Exec(p *Execution) {
	b := p.Pop().(string)
	a := p.Pop().(string)
	p.Push(ast.StringOp(i.Op, a, b))
}

func (i BinOpStr) String() string {
	return fmt.Sprintf("%s str", i.Op.Name())
}

type BinOpChar struct {
	baseInst
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpChar) Exec(p *Execution) {
	b := p.Pop().(byte)
	a := p.Pop().(byte)
	p.Push(ast.CharacterOp(i.Op, a, b))
}

func (i BinOpChar) String() string {
	return fmt.Sprintf("%s char", i.Op.Name())
}

type BinOpBool struct {
	baseInst
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpBool) Exec(p *Execution) {
	b := p.Pop().(bool)
	a := p.Pop().(bool)
	p.Push(ast.BooleanOp(i.Op, a, b))
}

func (i BinOpBool) String() string {
	return fmt.Sprintf("%s bool", i.Op.Name())
}
