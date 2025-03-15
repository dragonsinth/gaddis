package asm

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
