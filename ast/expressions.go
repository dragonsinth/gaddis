package ast

type Expression interface {
	Node
	GetType() Type
	isExpression()
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

func (il *IntegerLiteral) GetType() Type {
	return Integer
}

func (*IntegerLiteral) isExpression() {}

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

func (rl *RealLiteral) GetType() Type {
	return Real
}

func (*RealLiteral) isExpression() {}

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

func (sl *StringLiteral) GetType() Type {
	return String
}

func (*StringLiteral) isExpression() {}

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

func (cl CharacterLiteral) GetType() Type {
	return Character
}

func (*CharacterLiteral) isExpression() {}

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

func (bl BooleanLiteral) GetType() Type {
	return Boolean
}

func (*BooleanLiteral) isExpression() {}

type UnaryOperation struct {
	SourceInfo
	Op   Operator
	Type Type
	Expr Expression
}

func (uo *UnaryOperation) Visit(v Visitor) {
	if !v.PreVisitUnaryOperation(uo) {
		return
	}
	uo.Expr.Visit(v)
	v.PostVisitUnaryOperation(uo)
}

func (uo *UnaryOperation) GetType() Type {
	return uo.Type
}

func (*UnaryOperation) isExpression() {}

type BinaryOperation struct {
	SourceInfo
	Op   Operator
	Type Type
	Lhs  Expression
	Rhs  Expression
}

func (bo *BinaryOperation) Visit(v Visitor) {
	if !v.PreVisitBinaryOperation(bo) {
		return
	}
	bo.Lhs.Visit(v)
	bo.Rhs.Visit(v)
	v.PostVisitBinaryOperation(bo)
}

func (bo *BinaryOperation) GetType() Type {
	return bo.Type
}

func (*BinaryOperation) isExpression() {}

type VariableExpression struct {
	SourceInfo
	Name string

	Ref  *VarDecl // resolve symbols
	Type Type     // type checking
}

func (ve *VariableExpression) Visit(v Visitor) {
	if !v.PreVisitVariableExpression(ve) {
		return
	}
	v.PostVisitVariableExpression(ve)
}

func (ve *VariableExpression) GetType() Type {
	return ve.Type
}

func (*VariableExpression) isExpression() {}
