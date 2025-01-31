package ast

type Type int

const (
	InvalidType = Type(iota)
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
