package ast

type TypeKey string

type Type interface {
	IsPrimitive() bool
	AsPrimitive() PrimitiveType
	Key() TypeKey
	String() string
	IsNumeric() bool
	isType()
}

type PrimitiveType int

func (t PrimitiveType) IsPrimitive() bool {
	return true
}

func (t PrimitiveType) AsPrimitive() PrimitiveType {
	return t
}

func (t PrimitiveType) IsNumeric() bool {
	return t == Integer || t == Real
}

func (t PrimitiveType) isType() {
}

const (
	UnresolvedType = PrimitiveType(iota)
	Integer
	Real
	String
	Character
	Boolean
)

var typeNames = [...]string{"INVALID", "Integer", "Real", "String", "Character", "Boolean"}

var _ Type = UnresolvedType

func (t PrimitiveType) Key() TypeKey {
	return TypeKey(typeNames[t])
}

func (t PrimitiveType) String() string {
	return typeNames[t]
}

func CanCoerce(dst Type, src Type) bool {
	if dst == src {
		return true
	}
	if dst == Real && src == Integer {
		return true // promote
	}
	return false
}

func AreComparableTypes(a Type, b Type) Type {
	if a == b {
		return a
	}
	if a.IsNumeric() && b.IsNumeric() {
		return Real // promote
	}
	return UnresolvedType
}

func AreComparableOrderedTypes(a Type, b Type) bool {
	typ := AreComparableTypes(a, b)
	if typ == UnresolvedType {
		return false // must be comparable
	}
	if typ == Boolean {
		return false // cannot order booleans
	}
	return true
}
