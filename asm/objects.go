package asm

import (
	"github.com/dragonsinth/gaddis/ast"
	"strconv"
	"strings"
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

type NewArray struct {
	baseInst
	Typ  *ast.ArrayType
	Size int
}

func (n NewArray) Exec(p *Execution) {
	arr := p.PopN(n.Size)
	// make a copy
	val := append([]any{}, arr...)
	p.Push(val)
}

func (n NewArray) String() string {
	return "new array " + litTypes[n.Typ.Base.AsPrimitive()] + arrayTypeTail(n.Typ.NDims, n.Size)
}

func arrayTypeTail(dims int, sz int) string {
	var sb strings.Builder
	sb.WriteRune('[')
	sb.WriteString(strconv.Itoa(sz))
	sb.WriteRune(']')
	for i := 1; i < dims; i++ {
		sb.WriteString("[]")
	}
	return sb.String()
}
