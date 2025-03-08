package asm

import (
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
)

type Store struct {
	ast.SourceInfo
}

func (i Store) Exec(p *Execution) {
	ref := p.Pop().(*any)
	val := p.Pop()
	*ref = val
}

func (i Store) String() string {
	return "store"
}

type BinOpInt struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpInt) Exec(p *Execution) {
	b := p.Pop().(int64)
	a := p.Pop().(int64)
	switch i.Op {
	case ast.ADD:
		p.Push(a + b)
	case ast.SUB:
		p.Push(a - b)
	case ast.MUL:
		p.Push(a * b)
	case ast.DIV:
		p.Push(a / b)
	case ast.EXP:
		p.Push(builtins.ExpInteger(a, b))
	case ast.MOD:
		p.Push(builtins.ModInteger(a, b))
	case ast.EQ:
		p.Push(a == b)
	case ast.NEQ:
		p.Push(a != b)
	case ast.LT:
		p.Push(a < b)
	case ast.LTE:
		p.Push(a <= b)
	case ast.GT:
		p.Push(a > b)
	case ast.GTE:
		p.Push(a >= b)
	default:
		panic(i.Op)
	}
}

func (i BinOpInt) String() string {
	return fmt.Sprintf("%s_int", i.Op.Name())
}

type BinOpReal struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpReal) Exec(p *Execution) {
	b := toFloat64(p.Pop())
	a := toFloat64(p.Pop())
	switch i.Op {
	case ast.ADD:
		p.Push(a + b)
	case ast.SUB:
		p.Push(a - b)
	case ast.MUL:
		p.Push(a * b)
	case ast.DIV:
		p.Push(a / b)
	case ast.EXP:
		p.Push(builtins.ExpReal(a, b))
	case ast.MOD:
		p.Push(builtins.ModReal(a, b))
	case ast.EQ:
		p.Push(a == b)
	case ast.NEQ:
		p.Push(a != b)
	case ast.LT:
		p.Push(a < b)
	case ast.LTE:
		p.Push(a <= b)
	case ast.GT:
		p.Push(a > b)
	case ast.GTE:
		p.Push(a >= b)
	default:
		panic(i.Op)
	}
}

func (i BinOpReal) String() string {
	return fmt.Sprintf("%s_real", i.Op.Name())
}

type BinOpStr struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpStr) Exec(p *Execution) {
	b := p.Pop().([]byte)
	a := p.Pop().([]byte)
	cmp := bytes.Compare(a, b)
	switch i.Op {
	case ast.EQ:
		p.Push(cmp == 0)
	case ast.NEQ:
		p.Push(cmp != 0)
	case ast.LT:
		p.Push(cmp < 0)
	case ast.LTE:
		p.Push(cmp <= 0)
	case ast.GT:
		p.Push(cmp > 0)
	case ast.GTE:
		p.Push(cmp >= 0)
	default:
		panic(i.Op)
	}
}

func (i BinOpStr) String() string {
	return fmt.Sprintf("%s_str", i.Op.Name())
}

type BinOpChar struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpChar) Exec(p *Execution) {
	b := p.Pop().(byte)
	a := p.Pop().(byte)
	switch i.Op {
	case ast.EQ:
		p.Push(a == b)
	case ast.NEQ:
		p.Push(a != b)
	case ast.LT:
		p.Push(a < b)
	case ast.LTE:
		p.Push(a <= b)
	case ast.GT:
		p.Push(a > b)
	case ast.GTE:
		p.Push(a >= b)
	default:
		panic(i.Op)
	}
}

func (i BinOpChar) String() string {
	return fmt.Sprintf("%s_char", i.Op.Name())
}

type BinOpBool struct {
	ast.SourceInfo
	Op ast.Operator
}

func (i BinOpBool) Exec(p *Execution) {
	b := p.Pop().(bool)
	a := p.Pop().(bool)
	switch i.Op {
	case ast.EQ:
		p.Push(a == b)
	case ast.NEQ:
		p.Push(a != b)
	case ast.AND:
		p.Push(a && b)
	case ast.OR:
		p.Push(a || b)
	default:
		panic(i.Op)
	}
}

func (i BinOpBool) String() string {
	return fmt.Sprintf("%s_bool", i.Op.Name())
}
