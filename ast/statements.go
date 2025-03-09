package ast

type Statement interface {
	Node
	isStatement()
}

type DeclareStmt struct {
	SourceInfo
	Decls []*VarDecl
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
	Ref Expression
}

func (is *InputStmt) Visit(v Visitor) {
	if !v.PreVisitInputStmt(is) {
		return
	}
	is.Ref.Visit(v)
	v.PostVisitInputStmt(is)
}

func (*InputStmt) isStatement() {
}

type SetStmt struct {
	SourceInfo
	Ref  Expression
	Expr Expression
}

func (ss *SetStmt) Visit(v Visitor) {
	if !v.PreVisitSetStmt(ss) {
		return
	}
	ss.Ref.Visit(v)
	ss.Expr.Visit(v)
	v.PostVisitSetStmt(ss)
}

func (*SetStmt) isStatement() {
}

type IfStmt struct {
	SourceInfo
	Cases []*CondBlock // may end with an else block
}

func (is *IfStmt) Visit(v Visitor) {
	if !v.PreVisitIfStmt(is) {
		return
	}
	for _, cb := range is.Cases {
		cb.Visit(v)
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
	if cb.Expr != nil {
		cb.Expr.Visit(v)
	}
	cb.Block.Visit(v)
	v.PostVisitCondBlock(cb)
}

type SelectStmt struct {
	SourceInfo
	Type  Type
	Expr  Expression
	Cases []*CaseBlock
}

func (ss *SelectStmt) Visit(v Visitor) {
	if !v.PreVisitSelectStmt(ss) {
		return
	}
	ss.Expr.Visit(v)
	for _, cb := range ss.Cases {
		cb.Visit(v)
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
	if cb.Expr != nil {
		cb.Expr.Visit(v)
	}
	cb.Block.Visit(v)
	v.PostVisitCaseBlock(cb)
}

type DoStmt struct {
	SourceInfo
	Block *Block
	Until bool // if true, UNTIL, if false WHILE
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
	Ref       Expression
	StartExpr Expression
	StopExpr  Expression
	StepExpr  Expression
	Block     *Block
}

func (ws *ForStmt) Visit(v Visitor) {
	if !v.PreVisitForStmt(ws) {
		return
	}
	ws.Ref.Visit(v)
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

	Scope *Scope // collect
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

type ReturnStmt struct {
	SourceInfo
	Expr Expression

	Ref *FunctionStmt // resolve
}

func (rs *ReturnStmt) Visit(v Visitor) {
	if !v.PreVisitReturnStmt(rs) {
		return
	}
	rs.Expr.Visit(v)
	v.PostVisitReturnStmt(rs)
}

func (*ReturnStmt) isStatement() {
}

type FunctionStmt struct {
	SourceInfo
	Name   string
	Type   Type
	Params []*VarDecl
	Block  *Block

	IsExternal bool

	Scope *Scope // collect
}

func (fs *FunctionStmt) Visit(v Visitor) {
	if !v.PreVisitFunctionStmt(fs) {
		return
	}
	for _, param := range fs.Params {
		param.Visit(v)
	}
	fs.Block.Visit(v)
	v.PostVisitFunctionStmt(fs)
}

func (fs *FunctionStmt) GetName() string {
	return fs.Name
}

func (*FunctionStmt) isStatement() {
}
