package ast

type Operator int

const (
	INVALID_OPERATOR Operator = iota
	ADD
	SUB
	MUL
	DIV
	EXP
	MOD
)

var operators = [...]string{
	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	EXP: "^",
	MOD: "MOD",
}

func (op Operator) String() string {
	return operators[op]
}
