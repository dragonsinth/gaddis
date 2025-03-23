package asm

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
)

type OffsetType int

const (
	OffsetTypeString OffsetType = iota
	OffsetTypeArray
	OffsetTypeObject
)

type OffsetRef struct {
	baseInst
	OffsetType OffsetType
}

func (i OffsetRef) Exec(p *Execution) {
	idx := p.Pop().(int64)

	switch i.OffsetType {
	case OffsetTypeString:
		ref := p.Pop().([]byte)
		p.Push(&ref[idx])
	case OffsetTypeArray, OffsetTypeObject:
		arr := p.Pop().([]any)
		p.Push(&arr[idx])
	default:
		panic(i.OffsetType)
	}
}

func (i OffsetRef) String() string {
	switch i.OffsetType {
	case OffsetTypeString:
		return "&offset str"
	case OffsetTypeArray:
		return "&offset arr"
	case OffsetTypeObject:
		return "&offset obj"
	default:
		panic(i.OffsetType)
	}
}

type OffsetVal struct {
	baseInst
	OffsetType OffsetType
}

func (i OffsetVal) Exec(p *Execution) {
	idx := p.Pop().(int64)

	switch i.OffsetType {
	case OffsetTypeString:
		ref := p.Pop().([]byte)
		p.Push(ref[idx])
	case OffsetTypeArray, OffsetTypeObject:
		arr := p.Pop().([]any)
		p.Push(arr[idx])
	default:
		panic(i.OffsetType)
	}
}

func (i OffsetVal) String() string {
	switch i.OffsetType {
	case OffsetTypeString:
		return "offset str"
	case OffsetTypeArray:
		return "offset arr"
	case OffsetTypeObject:
		return "offset obj"
	default:
		panic(i.OffsetType)
	}
}

type ArrayNew struct {
	baseInst
	Typ  *ast.ArrayType
	Size int
}

func (n ArrayNew) Exec(p *Execution) {
	arr := p.PopN(n.Size)
	// make a copy
	val := append([]any{}, arr...)
	p.Push(val)
}

func (n ArrayNew) String() string {
	return "array new " + litTypes[n.Typ.Base.AsPrimitive()] + arrayTypeSized(n.Typ.NDims, n.Size)
}

type ArrayClone struct {
	baseInst
	Typ   *ast.ArrayType
	NDims int
}

func (d ArrayClone) Exec(p *Execution) {
	p.Push(arrayClone(d.NDims, p.Pop().([]any)))
}

func arrayClone(dims int, v []any) []any {
	ret := make([]any, len(v))
	if dims == 1 {
		copy(ret, v)
	} else {
		next := dims - 1
		for i := range v {
			ret[i] = arrayClone(next, v[i].([]any))
		}
	}
	return ret
}

func (d ArrayClone) String() string {
	return "array clone " + litTypes[d.Typ.Base.AsPrimitive()] + arrayTypeTail(d.NDims)
}

func arrayTypeSized(dims int, sz int) string {
	return fmt.Sprintf("[%d]", sz) + arrayTypeTail(dims-1)
}

func arrayTypeTail(dims int) string {
	if dims == 0 {
		return ""
	}
	return "[]" + arrayTypeTail(dims-1)
}

type ArrayLen struct {
	baseInst
}

func (a ArrayLen) Exec(p *Execution) {
	arr := p.Pop().([]any)
	p.Push(int64(len(arr)))
}

func (a ArrayLen) String() string {
	return "array len"
}
