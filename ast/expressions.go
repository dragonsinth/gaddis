package ast

type Expression interface {
	Node
	Type() Type
}

type IntegerLiteral struct {
	SourceInfo
	Val int64
}

func (il *IntegerLiteral) Visit(v Visitor) {
	if !v.PreVisitIntegerLiteral(il) {
		return
	}
	v.PostVisitIntegerLiteral(il)
}

func (il *IntegerLiteral) Type() Type {
	return Integer
}

type RealLiteral struct {
	SourceInfo
	Val float64
}

func (rl *RealLiteral) Visit(v Visitor) {
	if !v.PreVisitRealLiteral(rl) {
		return
	}
	v.PostVisitRealLiteral(rl)
}

func (rl *RealLiteral) Type() Type {
	return Real
}

type StringLiteral struct {
	SourceInfo
	Val string
}

func (sl *StringLiteral) Visit(v Visitor) {
	if !v.PreVisitStringLiteral(sl) {
		return
	}
	v.PostVisitStringLiteral(sl)
}

func (sl *StringLiteral) Type() Type {
	return String
}

type CharacterLiteral struct {
	SourceInfo
	Val byte
}

func (cl *CharacterLiteral) Visit(v Visitor) {
	if !v.PreVisitCharacterLiteral(cl) {
		return
	}
	v.PostVisitCharacterLiteral(cl)
}

func (cl CharacterLiteral) Type() Type {
	return Character
}

type BooleanLiteral struct {
	SourceInfo
	Val bool
}

func (bl *BooleanLiteral) Visit(v Visitor) {
	if !v.PreVisitBooleanLiteral(bl) {
		return
	}
	v.PostVisitBooleanLiteral(bl)
}

func (bl BooleanLiteral) Type() Type {
	return Boolean
}

type UnaryOperation struct {
	SourceInfo
	Op   Operator
	Typ  Type
	Expr Expression
}

func (uo *UnaryOperation) Visit(v Visitor) {
	if !v.PreVisitUnaryOperation(uo) {
		return
	}
	uo.Expr.Visit(v)
	v.PostVisitUnaryOperation(uo)
}

func (uo *UnaryOperation) Type() Type {
	return uo.Typ
}

type BinaryOperation struct {
	SourceInfo
	Op  Operator
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

type VariableExpression struct {
	SourceInfo
	Name string

	Ref *VarDecl // resolve symbols
	Typ Type     // type checking
}

func (ve *VariableExpression) Visit(v Visitor) {
	if !v.PreVisitVariableExpression(ve) {
		return
	}
	v.PostVisitVariableExpression(ve)
}

func (ve *VariableExpression) Type() Type {
	return ve.Typ
}
