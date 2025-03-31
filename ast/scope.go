package ast

import "fmt"

// TODO: more than just var decls (module, function, class).
// TODO: construct scopes during Parse? Automatically manage scope in base visitor?

type Scope struct {
	SourceInfo   SourceInfo
	Parent       *Scope
	IsExternal   bool // if true, external scope
	IsGlobal     bool // if true, global scope
	IsEval       bool
	ModuleStmt   *ModuleStmt   // if true, enclosing module
	FunctionStmt *FunctionStmt // if true, enclosing function
	ClassStmt    *ClassStmt    // if true, enclosing class
	Decls        map[string]*Decl
	Params       []*VarDecl
	Locals       []*VarDecl
}

func (s *Scope) Desc() string {
	if s.IsExternal {
		return "external"
	} else if s.IsGlobal {
		return "global"
	} else if s.IsEval {
		return "eval"
	} else if s.ModuleStmt != nil {
		return fmt.Sprintf("%s()", s.ModuleStmt.Name)
	} else if s.FunctionStmt != nil {
		return fmt.Sprintf("%s()", s.FunctionStmt.Name)
	} else if s.ClassStmt != nil {
		return s.ClassStmt.Name
	} else {
		panic("unset")
	}
}

func (s *Scope) String() string {
	if s.IsExternal {
		return "external"
	} else if s.IsGlobal {
		return "global"
	} else if s.IsEval {
		return "eval"
	} else if s.ModuleStmt != nil {
		return fmt.Sprintf("Module %s", s.ModuleStmt.Name)
	} else if s.FunctionStmt != nil {
		return fmt.Sprintf("Function %s %s", s.FunctionStmt.Type, s.FunctionStmt.Name)
	} else if s.ClassStmt != nil {
		return fmt.Sprintf("Class %s", s.ClassStmt.Name)
	} else {
		panic("unset")
	}
}

type Decl struct {
	ModuleStmt   *ModuleStmt
	FunctionStmt *FunctionStmt
	ClassStmt    *ClassStmt
	VarDecl      *VarDecl
}

func (d *Decl) String() string {
	if d.ModuleStmt != nil {
		return fmt.Sprintf("Module %s", d.ModuleStmt.Name)
	} else if d.FunctionStmt != nil {
		if d.FunctionStmt.IsExternal {
			return fmt.Sprintf("External Function %s %s", d.FunctionStmt.Type, d.FunctionStmt.Name)
		} else {
			return fmt.Sprintf("Function %s %s", d.FunctionStmt.Type, d.FunctionStmt.Name)
		}
	} else if d.ClassStmt != nil {
		return fmt.Sprintf("Class %s", d.ClassStmt.Name)
	} else if d.VarDecl != nil {
		if d.VarDecl.IsParam {
			if d.VarDecl.IsRef {
				return fmt.Sprintf("parameter %s Ref %s", d.VarDecl.Type, d.VarDecl.Name)
			} else {
				return fmt.Sprintf("parameter %s %s", d.VarDecl.Type, d.VarDecl.Name)
			}
		} else if d.VarDecl.IsConst {
			return fmt.Sprintf("Constant %s %s", d.VarDecl.Type, d.VarDecl.Name)
		} else {
			return fmt.Sprintf("Declare %s %s", d.VarDecl.Type, d.VarDecl.Name)
		}
	} else {
		panic("unset")
	}
}

func (s *Scope) Lookup(name string) *Decl {
	if v := s.Decls[name]; v != nil {
		return v
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

func (s *Scope) AddModule(ms *ModuleStmt) {
	name := ms.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	s.Decls[name] = &Decl{ModuleStmt: ms}
}

func (s *Scope) AddFunction(fs *FunctionStmt) {
	name := fs.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	s.Decls[name] = &Decl{FunctionStmt: fs}
}

func (s *Scope) AddClass(cs *ClassStmt) {
	name := cs.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	s.Decls[name] = &Decl{ClassStmt: cs}
}

func (s *Scope) AddVariable(vd *VarDecl) {
	name := vd.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	if !vd.IsConst {
		if vd.IsParam {
			vd.Id = len(s.Params)
			s.Params = append(s.Params, vd)
		} else {
			vd.Id = len(s.Locals)
			s.Locals = append(s.Locals, vd)
		}
	}
	s.Decls[name] = &Decl{VarDecl: vd}
	vd.Scope = s
}

func NewGlobalScope(bl *Block) *Scope {
	return &Scope{
		SourceInfo: bl.SourceInfo,
		Parent:     ExternalScope,
		IsGlobal:   true,
		Decls:      map[string]*Decl{},
	}
}

func NewModuleScope(ms *ModuleStmt, parent *Scope) *Scope {
	return &Scope{
		SourceInfo: ms.Block.SourceInfo,
		Parent:     parent,
		ModuleStmt: ms,
		Decls:      map[string]*Decl{},
	}
}

func NewFunctionScope(fs *FunctionStmt, parent *Scope) *Scope {
	return &Scope{
		SourceInfo:   fs.Block.SourceInfo,
		Parent:       parent,
		FunctionStmt: fs,
		Decls:        map[string]*Decl{},
	}
}

func NewClassScope(cs *ClassStmt, parent *Scope) *Scope {
	return &Scope{
		SourceInfo: cs.Block.SourceInfo,
		Parent:     parent,
		ClassStmt:  cs,
		Decls:      map[string]*Decl{},
	}
}

func NewEvalScope(expr Expression, parent *Scope) *Scope {
	return &Scope{
		SourceInfo: expr.GetSourceInfo(),
		Parent:     parent,
		IsEval:     true,
		Decls:      map[string]*Decl{},
	}
}
