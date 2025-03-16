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
	NEG
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
	GTE: ">=",
	AND: "AND",
	OR:  "OR",
	NOT: "NOT",
	NEG: "-",
}

func (op Operator) String() string {
	return operators[op]
}

var names = [...]string{
	ADD: "add",
	SUB: "sub",
	MUL: "mul",
	DIV: "div",
	EXP: "exp",
	MOD: "mod",
	EQ:  "eq",
	NEQ: "neq",
	LT:  "lt",
	LTE: "lte",
	GT:  "gt",
	GTE: "gte",
	AND: "and",
	OR:  "or",
	NOT: "not",
	NEG: "neg",
}

func (op Operator) Name() string {
	return names[op]
}
