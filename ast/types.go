package ast

type Type int

const (
	InvalidType = Type(iota)
	Integer
	Real
	String
	Character
)

func (t Type) String() string {
	return [...]string{"Integer", "Real", "String", "Character"}[t]
}
