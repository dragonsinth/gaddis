package ast

type Expression interface {
	Node
	GetType() Type
	ConstEval() any
	CanReference() bool
	isExpression()
}

type baseExpression struct {
}

func (baseExpression) ConstEval() any { return nil }

func (baseExpression) CanReference() bool { return false }

func (baseExpression) isExpression() {}

type Literal struct {
	SourceInfo
	baseExpression
	Type         PrimitiveType
	Val          any
	IsTabLiteral bool
}

func (l *Literal) Visit(v Visitor) {
	if !v.PreVisitLiteral(l) {
		return
	}
	v.PostVisitLiteral(l)
}

func (l *Literal) GetType() Type {
	return l.Type
}

func (l *Literal) ConstEval() any {
	return l.Val
}

type ParenExpr struct {
	SourceInfo
	baseExpression
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

func (pe *ParenExpr) ConstEval() any {
	return pe.Expr.ConstEval()
}

type UnaryOperation struct {
	SourceInfo
	baseExpression
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

func (uo *UnaryOperation) ConstEval() any {
	expr := uo.Expr.ConstEval()
	if expr == nil {
		return nil
	}

	switch uo.Op {
	case NEG:
		switch v := expr.(type) {
		case int64:
			return -v
		case float64:
			return -v
		default:
			panic(v)
		}
	case NOT:
		switch v := expr.(type) {
		case bool:
			return !v
		default:
			panic(v)
		}
	default:
		panic(uo.Op)
	}
}

type BinaryOperation struct {
	SourceInfo
	baseExpression
	Op      Operator
	Type    Type
	Lhs     Expression
	Rhs     Expression
	ArgType Type
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

func (bo *BinaryOperation) ConstEval() any {
	lhs := bo.Lhs.ConstEval()
	rhs := bo.Rhs.ConstEval()
	if lhs == nil || rhs == nil {
		return nil
	}
	ret := AnyOp(bo.Op, bo.ArgType.AsPrimitive(), lhs, rhs)
	if bo.ArgType == Real {
		ret = EnsureReal(ret)
	}
	return ret
}

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

func (ve *VariableExpr) ConstEval() any {
	if ve.Ref == nil {
		return nil
	}
	if ve.Ref.IsConst {
		ret := ve.Ref.Expr.ConstEval()
		if ve.Ref.Type == Real && ve.Ref.Expr.GetType() == Integer {
			return float64(ret.(int64))
		}
		return ret
	}
	return nil
}

func (ve *VariableExpr) CanReference() bool {
	return !ve.Ref.IsConst
}

func (*VariableExpr) isExpression() {}

type CallExpr struct {
	SourceInfo
	baseExpression
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

type ArrayRef struct {
	SourceInfo
	baseExpression
	RefExpr   Expression
	IndexExpr Expression
	Type      Type
}

func (ar *ArrayRef) Visit(v Visitor) {
	if !v.PreVisitArrayRef(ar) {
		return
	}
	ar.RefExpr.Visit(v)
	ar.IndexExpr.Visit(v)
	v.PostVisitArrayRef(ar)
}

func (ar *ArrayRef) CanReference() bool {
	return true
}

func (ar *ArrayRef) GetType() Type {
	return ar.Type
}

type ArrayInitializer struct {
	SourceInfo
	baseExpression
	Args []Expression
	Type *ArrayType
	Dims []int
}

func (ai *ArrayInitializer) Visit(v Visitor) {
	if !v.PreVisitArrayInitializer(ai) {
		return
	}
	for _, expr := range ai.Args {
		expr.Visit(v)
	}
	v.PostArrayInitializer(ai)
}

func (ai *ArrayInitializer) GetType() Type {
	return ai.Type
}
