package ast

type Type int

const (
	UnresolvedType = Type(iota)
	Integer
	Real
	String
	Character
	Boolean
)

var typeNames = [...]string{"INVALID", "Integer", "Real", "String", "Character", "Boolean"}

func (t Type) String() string {
	return typeNames[t]
}

func IsNumericType(t Type) bool {
	return t == Integer || t == Real
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
	if IsNumericType(a) && IsNumericType(b) {
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
