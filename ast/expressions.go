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

type TabLiteral struct {
	SourceInfo
}

func (tl *TabLiteral) Visit(v Visitor) {
	if !v.PreVisitTabLiteral(tl) {
		return
	}
	v.PostVisitTabLiteral(tl)
}

func (tl *TabLiteral) GetType() Type {
	return String
}

func (*TabLiteral) isExpression() {}

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

type ParenExpr struct {
	SourceInfo
	Expr Expression
}

func (pe *ParenExpr) Visit(v Visitor) {
	if !v.PreVisitParenExpr(pe) {
		return
	}
	pe.Expr.Visit(v)
	v.PostVisitParenExpr(pe)
}

func (pe *ParenExpr) GetType() Type {
	return pe.Expr.GetType()
}

func (*ParenExpr) isExpression() {}

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

type VariableExpr struct {
	SourceInfo
	Name string

	Ref  *VarDecl // resolve symbols
	Type Type     // type checking
}

func (ve *VariableExpr) Visit(v Visitor) {
	if !v.PreVisitVariableExpr(ve) {
		return
	}
	v.PostVisitVariableExpr(ve)
}

func (ve *VariableExpr) GetType() Type {
	return ve.Type
}

func (*VariableExpr) isExpression() {}

type CallExpr struct {
	SourceInfo
	Name string
	Args []Expression

	Ref  *FunctionStmt // resolve
	Type Type          // type checking
}

func (ce *CallExpr) Visit(v Visitor) {
	if !v.PreVisitCallExpr(ce) {
		return
	}
	for _, arg := range ce.Args {
		arg.Visit(v)
	}
	v.PostVisitCallExpr(ce)
}

func (ce *CallExpr) GetType() Type {
	return ce.Type
}

func (*CallExpr) isExpression() {
}
