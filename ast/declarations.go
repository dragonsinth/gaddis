package ast

type HasName interface {
	GetName() string
}

type VarDecl struct {
	SourceInfo
	Name      string
	Type      Type
	DimExprs  []Expression
	Expr      Expression
	IsConst   bool
	IsPrivate bool
	IsParam   bool
	IsRef     bool
	IsField   *ClassType

	Scope *Scope // collect
	Id    int
	Dims  []int // typecheck
}

func (vd *VarDecl) Visit(v Visitor) {
	if !v.PreVisitVarDecl(vd) {
		return
	}
	for _, d := range vd.DimExprs {
		d.Visit(v)
	}
	if vd.Expr != nil {
		vd.Expr.Visit(v)
	}
	v.PostVisitVarDecl(vd)
}

func (vd *VarDecl) GetName() string {
	return vd.Name
}

type ModuleStmt struct {
	SourceInfo
	Name   string
	Params []*VarDecl
	Block  *Block

	IsExternal    bool
	IsPrivate     bool
	IsConstructor bool
	IsMethod      *ClassType

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

type FunctionStmt struct {
	SourceInfo
	Name   string
	Type   Type
	Params []*VarDecl
	Block  *Block

	IsExternal bool
	IsPrivate  bool
	IsMethod   *ClassType

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
