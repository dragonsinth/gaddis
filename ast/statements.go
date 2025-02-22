package ast

type Statement interface {
	Node
	isStatement()
}

type Block struct {
	SourceInfo
	Statements []Statement

	Scope *Scope // collect symbols
}

func (bl *Block) Visit(v Visitor) {
	if !v.PreVisitBlock(bl) {
		return
	}
	for _, stmt := range bl.Statements {
		stmt.Visit(v)
	}
	v.PostVisitBlock(bl)
}

type DeclareStmt struct {
	SourceInfo
	IsConst bool
	Type    Type
	Decls   []*VarDecl
}

func (ds *DeclareStmt) Visit(v Visitor) {
	if !v.PreVisitDeclareStmt(ds) {
		return
	}
	for _, d := range ds.Decls {
		d.Visit(v)
	}
	v.PostVisitDeclareStmt(ds)
}

func (*DeclareStmt) isStatement() {
}

type DisplayStmt struct {
	SourceInfo
	Exprs []Expression
}

func (ds *DisplayStmt) Visit(v Visitor) {
	if !v.PreVisitDisplayStmt(ds) {
		return
	}
	for _, e := range ds.Exprs {
		e.Visit(v)
	}
	v.PostVisitDisplayStmt(ds)
}

func (*DisplayStmt) isStatement() {
}

type InputStmt struct {
	SourceInfo
	Var *VariableExpression
}

func (is *InputStmt) Visit(v Visitor) {
	if !v.PreVisitInputStmt(is) {
		return
	}
	is.Var.Visit(v)
	v.PostVisitInputStmt(is)
}

func (*InputStmt) isStatement() {
}

type SetStmt struct {
	SourceInfo
	Var  *VariableExpression
	Expr Expression
}

func (ss *SetStmt) Visit(v Visitor) {
	if !v.PreVisitSetStmt(ss) {
		return
	}
	ss.Var.Visit(v)
	ss.Expr.Visit(v)
	v.PostVisitSetStmt(ss)
}

func (*SetStmt) isStatement() {
}

type IfStmt struct {
	SourceInfo
	If     *CondBlock
	ElseIf []*CondBlock
	Else   *Block
}

func (is *IfStmt) Visit(v Visitor) {
	if !v.PreVisitIfStmt(is) {
		return
	}
	is.If.Visit(v)
	for _, cb := range is.ElseIf {
		cb.Visit(v)
	}
	if is.Else != nil {
		is.Else.Visit(v)
	}
	v.PostVisitIfStmt(is)
}

func (*IfStmt) isStatement() {
}

type CondBlock struct {
	SourceInfo
	Expr  Expression
	Block *Block
}

func (cb *CondBlock) Visit(v Visitor) {
	if !v.PreVisitCondBlock(cb) {
		return
	}
	cb.Expr.Visit(v)
	cb.Block.Visit(v)
	v.PostVisitCondBlock(cb)
}

type SelectStmt struct {
	SourceInfo
	Type    Type
	Expr    Expression
	Cases   []*CaseBlock
	Default *Block
}

func (ss *SelectStmt) Visit(v Visitor) {
	if !v.PreVisitSelectStmt(ss) {
		return
	}
	ss.Expr.Visit(v)
	for _, cb := range ss.Cases {
		cb.Visit(v)
	}
	if ss.Default != nil {
		ss.Default.Visit(v)
	}
	v.PostVisitSelectStmt(ss)
}

func (*SelectStmt) isStatement() {
}

type CaseBlock struct {
	SourceInfo
	Expr  Expression
	Block *Block
}

func (cb *CaseBlock) Visit(v Visitor) {
	if !v.PreVisitCaseBlock(cb) {
		return
	}
	cb.Expr.Visit(v)
	cb.Block.Visit(v)
	v.PostVisitCaseBlock(cb)
}

type DoStmt struct {
	SourceInfo
	Block *Block
	Not   bool
	Expr  Expression
}

func (ds *DoStmt) Visit(v Visitor) {
	if !v.PreVisitDoStmt(ds) {
		return
	}
	ds.Block.Visit(v)
	ds.Expr.Visit(v)
	v.PostVisitDoStmt(ds)
}

func (*DoStmt) isStatement() {
}

type WhileStmt struct {
	SourceInfo
	Expr  Expression
	Block *Block
}

func (ws *WhileStmt) Visit(v Visitor) {
	if !v.PreVisitWhileStmt(ws) {
		return
	}
	ws.Expr.Visit(v)
	ws.Block.Visit(v)
	v.PostVisitWhileStmt(ws)
}

func (*WhileStmt) isStatement() {
}

type ForStmt struct {
	SourceInfo
	Var       *VariableExpression
	StartExpr Expression
	StopExpr  Expression
	StepExpr  Expression
	Block     *Block
}

func (ws *ForStmt) Visit(v Visitor) {
	if !v.PreVisitForStmt(ws) {
		return
	}
	ws.Var.Visit(v)
	ws.StartExpr.Visit(v)
	ws.StopExpr.Visit(v)
	if ws.StepExpr != nil {
		ws.StepExpr.Visit(v)
	}
	ws.Block.Visit(v)
	v.PostVisitForStmt(ws)
}

func (*ForStmt) isStatement() {
}

type CallStmt struct {
	SourceInfo
	Name string
	Args []Expression

	Ref *ModuleStmt // resolve
}

func (cs *CallStmt) Visit(v Visitor) {
	if !v.PreVisitCallStmt(cs) {
		return
	}
	for _, arg := range cs.Args {
		arg.Visit(v)
	}
	v.PostVisitCallStmt(cs)
}
func (*CallStmt) isStatement() {
}

type ModuleStmt struct {
	SourceInfo
	Name   string
	Params []*VarDecl
	Block  *Block
}

func (ms *ModuleStmt) Visit(v Visitor) {
	if !v.PreVisitModuleStmt(ms) {
		return
	}
	for _, param := range ms.Params {
		param.Visit(v)
	}
	ms.Block.Visit(v)
	v.PostVisitModuleStmt(ms)
}

func (ms *ModuleStmt) GetName() string {
	return ms.Name
}

func (*ModuleStmt) isStatement() {
}
