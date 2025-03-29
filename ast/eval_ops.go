package ast

import (
	"math"
)

func AnyOp(op Operator, argType PrimitiveType, a, b any) any {
	defer func() {
		// evaluation can panic; just return nil
		recover()
	}()

	switch argType {
	case Integer:
		return IntegerOp(op, a.(int64), b.(int64))
	case Real:
		a = EnsureReal(a)
		b = EnsureReal(b)
		return RealOp(op, a.(float64), b.(float64))
	case String:
		return StringOp(op, a.(string), b.(string))
	case Character:
		return CharacterOp(op, a.(byte), b.(byte))
	case Boolean:
		return BooleanOp(op, a.(bool), b.(bool))
	default:
		panic(a)
	}
}

func IntegerOp(op Operator, a, b int64) any {
	switch op {
	case ADD:
		return a + b
	case SUB:
		return a - b
	case MUL:
		return a * b
	case DIV:
		return a / b
	case EXP:
		return expInteger(a, b)
	case MOD:
		return a % b
	case EQ:
		return a == b
	case NEQ:
		return a != b
	case LT:
		return a < b
	case LTE:
		return a <= b
	case GT:
		return a > b
	case GTE:
		return a >= b
	default:
		panic(op)
	}
}

func RealOp(op Operator, a, b float64) any {
	switch op {
	case ADD:
		return a + b
	case SUB:
		return a - b
	case MUL:
		return a * b
	case DIV:
		return a / b
	case EXP:
		return math.Pow(a, b)
	case MOD:
		return math.Mod(a, b)
	case EQ:
		return a == b
	case NEQ:
		return a != b
	case LT:
		return a < b
	case LTE:
		return a <= b
	case GT:
		return a > b
	case GTE:
		return a >= b
	default:
		panic(op)
	}
}

func StringOp(op Operator, a, b string) any {
	switch op {
	case EQ:
		return a == b
	case NEQ:
		return a != b
	case LT:
		return a < b
	case LTE:
		return a <= b
	case GT:
		return a > b
	case GTE:
		return a >= b
	default:
		panic(op)
	}
}

func CharacterOp(op Operator, a, b byte) any {
	switch op {
	case EQ:
		return a == b
	case NEQ:
		return a != b
	case LT:
		return a < b
	case LTE:
		return a <= b
	case GT:
		return a > b
	case GTE:
		return a >= b
	default:
		panic(op)
	}
}

func BooleanOp(op Operator, a, b bool) any {
	switch op {
	case EQ:
		return a == b
	case NEQ:
		return a != b
	case AND:
		return a && b
	case OR:
		return a || b
	default:
		panic(op)
	}
}

func expInteger(base, exp int64) int64 {
	if exp < 0 {
		return 0 // Or handle negative exponents as needed (e.g., return 1 / intPow(base, -exp))
	}
	if exp == 0 {
		return 1
	}

	result := int64(1)
	for {
		if exp&1 == 1 { // Check if the least significant bit of exp is 1
			result *= base
		}
		exp >>= 1 // Right shift exp (equivalent to dividing by 2)
		if exp == 0 {
			break
		}
		base *= base // Square the base
	}
	return result
}

func EnsureReal(x any) float64 {
	switch v := x.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	default:
		panic(x)
	}
}
