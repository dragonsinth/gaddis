package ast

import (
	"fmt"
	"github.com/dragonsinth/gaddis/lex"
)

type Expression interface {
	fmt.Stringer
	Type() Type
	Visit(v Visitor)
}

type IntegerLiteral struct {
	Val int64
}

func (il *IntegerLiteral) Visit(v Visitor) {
	if !v.PreVisitIntegerLiteral(il) {
		return
	}
	v.PostVisitIntegerLiteral(il)
}

func (il *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", il.Val)
}

func (il *IntegerLiteral) Type() Type {
	return Integer
}

type RealLiteral struct {
	Val float64
}

func (rl *RealLiteral) Visit(v Visitor) {
	if !v.PreVisitRealLiteral(rl) {
		return
	}
	v.PostVisitRealLiteral(rl)
}

func (rl *RealLiteral) String() string {
	return fmt.Sprintf("%f", rl.Val)
}

func (rl *RealLiteral) Type() Type {
	return Real
}

type StringLiteral struct {
	Val string
}

func (sl *StringLiteral) Visit(v Visitor) {
	if !v.PreVisitStringLiteral(sl) {
		return
	}
	v.PostVisitStringLiteral(sl)
}

func (sl *StringLiteral) String() string {
	return fmt.Sprintf("%q", sl.Val)
}

func (sl *StringLiteral) Type() Type {
	return String
}

type CharacterLiteral struct {
	Val byte
}

func (cl *CharacterLiteral) Visit(v Visitor) {
	if !v.PreVisitCharacterLiteral(cl) {
		return
	}
	v.PostVisitCharacterLiteral(cl)
}

func (cl CharacterLiteral) String() string {
	return fmt.Sprintf("%c", cl.Val)
}

func (cl CharacterLiteral) Type() Type {
	return Character
}

type BinaryOperation struct {
	Op  lex.Token
	Typ Type
	Lhs Expression
	Rhs Expression
}

func (bo *BinaryOperation) Visit(v Visitor) {
	if !v.PreVisitBinaryOperation(bo) {
		return
	}
	bo.Lhs.Visit(v)
	bo.Rhs.Visit(v)
	v.PostVisitBinaryOperation(bo)
}

func (bo *BinaryOperation) Type() Type {
	return bo.Typ
}

func (bo *BinaryOperation) String() string {
	return fmt.Sprintf("(%s %s %s)", bo.Lhs, bo.Op, bo.Rhs)
}

func NewBinaryOperation(r lex.Result, lhs Expression, rhs Expression) *BinaryOperation {
	// TODO: more semantic / type checking pass
	switch r.Token {
	case lex.ADD, lex.SUB, lex.MUL, lex.DIV, lex.EXP, lex.MOD:
		aTyp := lhs.Type()
		bTyp := rhs.Type()
		var typ Type
		if aTyp == Integer && bTyp == Integer {
			typ = Integer
		} else if aTyp == Integer && bTyp == Real {
			typ = Real
		} else if aTyp == Real && bTyp == Integer {
			typ = Real
		} else if aTyp == Real && bTyp == Real {
			typ = Real
		} else {
			panic(fmt.Sprintf("%d:%d unsupported binary operation %s not supported for types %s and %s", r.Pos.Line, r.Pos.Column, r.Text, aTyp, bTyp))
		}
		return &BinaryOperation{r.Token, typ, lhs, rhs}
	default:
		panic(fmt.Sprintf("%d:%d unsupported binary operation: %s %q", r.Pos.Line, r.Pos.Column, r.Token, r.Text))
	}
}

type VariableExpression struct {
	Name string
	Ref  *VarDecl
}

func (ve *VariableExpression) Visit(v Visitor) {
	if !v.PreVisitVariableExpression(ve) {
		return
	}
	v.PostVisitVariableExpression(ve)
}

func (ve *VariableExpression) Type() Type {
	return ve.Ref.Type
}

func (ve *VariableExpression) String() string {
	return ve.Name
}
