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
	EQ
	NEQ
	LT
	LTE
	GT
	GTE
	AND
	OR
	NOT
)

var operators = [...]string{
	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	EXP: "^",
	MOD: "MOD",
	EQ:  "==",
	NEQ: "!=",
	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">",
	AND: "AND",
	OR:  "OR",
	NOT: "NOT",
}

func (op Operator) String() string {
	return operators[op]
}
