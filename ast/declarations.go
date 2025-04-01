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
	Enclosing *ClassType

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

type Callable interface {
	Statement
	HasName
	GetEnclosing() *ClassType
	GetParams() []*VarDecl
	GetScope() *Scope
}

type ModuleStmt struct {
	SourceInfo
	Name   string
	Params []*VarDecl
	Block  *Block

	IsExternal    bool
	IsPrivate     bool
	IsConstructor bool
	Enclosing     *ClassType

	Scope *Scope // collect
	Id    int    // only if method
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

func (ms *ModuleStmt) GetEnclosing() *ClassType {
	return ms.Enclosing
}

func (ms *ModuleStmt) GetParams() []*VarDecl {
	return ms.Params
}

func (ms *ModuleStmt) GetScope() *Scope {
	return ms.Scope
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
	Enclosing  *ClassType

	Scope *Scope // collect
	Id    int    // only if method
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

func (fs *FunctionStmt) GetEnclosing() *ClassType {
	return fs.Enclosing
}

func (fs *FunctionStmt) GetParams() []*VarDecl {
	return fs.Params
}

func (fs *FunctionStmt) GetScope() *Scope {
	return fs.Scope
}

func (*FunctionStmt) isStatement() {
}
