package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type Store struct {
	baseInst
}

func (i Store) Exec(p *Execution) {
	ref := p.Pop().(*any)
	val := p.Pop()
	*ref = val
}

func (i Store) String() string {
	return "store"
}

type StoreChar struct {
	baseInst
}

func (i StoreChar) Exec(p *Execution) {
	ref := p.Pop().(*byte)
	val := p.Pop().(byte)
	*ref = val
}

func (i StoreChar) String() string {
	return "store char"
}

type BinOpInt struct {
	baseInst
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
	Op ast.Operator
}

func (i BinOpStr) Exec(p *Execution) {
	b := p.Pop().([]byte)
	a := p.Pop().([]byte)
	p.Push(ast.ByteStringOp(i.Op, a, b))
}

func (i BinOpStr) String() string {
	return fmt.Sprintf("%s str", i.Op.Name())
}

type BinOpChar struct {
	baseInst
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
