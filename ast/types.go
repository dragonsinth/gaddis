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
	IsFileType() bool
	AsFileType() FileType
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

func (t PrimitiveType) IsFileType() bool { return false }

func (t PrimitiveType) AsFileType() FileType { return InvalidFileType }

func (t PrimitiveType) BaseType() Type { return t }

func (t PrimitiveType) isType() {
}

type ArrayType struct {
	Base        Type
	NDims       int
	ElementType Type
	TypeKey     TypeKey
}

func (t *ArrayType) Key() TypeKey {
	return t.TypeKey
}

func (t *ArrayType) String() string {
	return string(t.TypeKey)
}

func (t *ArrayType) IsPrimitive() bool { return false }

func (t *ArrayType) AsPrimitive() PrimitiveType { return UnresolvedType }

func (t *ArrayType) IsNumeric() bool { return false }

func (t *ArrayType) IsArrayType() bool { return true }

func (t *ArrayType) AsArrayType() *ArrayType { return t }

func (t *ArrayType) IsFileType() bool { return false }

func (t *ArrayType) AsFileType() FileType { return InvalidFileType }

func (t *ArrayType) BaseType() Type { return t.Base }

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

const (
	InvalidFileType = FileType(0)
	OutputFile      = FileType(11)
	AppendFile      = FileType(12)
	InputFile       = FileType(13)
)

var fileTypeNames = [...]string{
	InvalidFileType: "INVALID_FILE",
	OutputFile:      "OutputFile",
	AppendFile:      "AppendFile",
	InputFile:       "InputFile",
}

var _ Type = InvalidFileType

type FileType int

func (t FileType) Key() TypeKey { return TypeKey(fileTypeNames[t]) }

func (t FileType) String() string { return fileTypeNames[t] }

func (t FileType) IsPrimitive() bool { return false }

func (t FileType) AsPrimitive() PrimitiveType { return UnresolvedType }

func (t FileType) IsNumeric() bool { return false }

func (t FileType) IsArrayType() bool { return false }

func (t FileType) AsArrayType() *ArrayType { return nil }

func (t FileType) IsFileType() bool { return true }

func (t FileType) AsFileType() FileType { return t }

func (t FileType) BaseType() Type { return t }

func (t FileType) isType() {
}
