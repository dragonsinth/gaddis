package ast

type TypeKey string

type Type interface {
	Key() TypeKey
	String() string
	IsNumeric() bool
	IsPrimitive() bool
	AsPrimitive() PrimitiveType
	IsArrayType() bool
	AsArrayType() *ArrayType
	BaseType() Type

	isType()
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

type PrimitiveType int

func (t PrimitiveType) Key() TypeKey { return TypeKey(typeNames[t]) }

func (t PrimitiveType) String() string { return typeNames[t] }

func (t PrimitiveType) IsPrimitive() bool { return true }

func (t PrimitiveType) AsPrimitive() PrimitiveType { return t }

func (t PrimitiveType) IsNumeric() bool { return t == Integer || t == Real }

func (t PrimitiveType) IsArrayType() bool { return false }

func (t PrimitiveType) AsArrayType() *ArrayType { return nil }

func (t PrimitiveType) BaseType() Type { return t }

func (t PrimitiveType) isType() {
}

type ArrayType struct {
	ElementType Type
}

func (t *ArrayType) Key() TypeKey {
	return TypeKey(t.ElementType.String() + "[]")
}

func (t *ArrayType) String() string {
	return t.ElementType.String() + "[]"
}

func (t *ArrayType) IsPrimitive() bool { return false }

func (t *ArrayType) AsPrimitive() PrimitiveType { return UnresolvedType }

func (t *ArrayType) IsNumeric() bool { return false }

func (t *ArrayType) IsArrayType() bool { return true }

func (t *ArrayType) AsArrayType() *ArrayType { return t }

func (t *ArrayType) BaseType() Type { return t.ElementType.BaseType() }

func (t *ArrayType) isType() {
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

func IsOrderedType(typ Type) bool {
	if typ == UnresolvedType || typ == Boolean {
		return false // cannot order booleans
	}
	return typ.IsPrimitive() // the other primitive types are ordered
}
