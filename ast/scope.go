package ast

import "fmt"

// TODO: more than just var decls (module, function, class).

type Scope struct {
	Parent       *Scope
	IsGlobal     bool          // if true, global scope
	ModuleStmt   *ModuleStmt   // if true, enclosing module
	FunctionStmt *FunctionStmt // if true, enclosing function
	Decls        map[string]*Decl
}

func (s *Scope) String() string {
	if s.IsGlobal {
		return "Global Scope"
	} else if s.ModuleStmt != nil {
		return fmt.Sprintf("Module %s Scope", s.ModuleStmt.Name)
	} else if s.FunctionStmt != nil {
		return fmt.Sprintf("Function %s %s Scope", s.FunctionStmt.Type, s.FunctionStmt.Name)
	} else {
		panic("unset")
	}
}

type Decl struct {
	ModuleStmt   *ModuleStmt
	FunctionStmt *FunctionStmt
	VarDecl      *VarDecl
}

func (d *Decl) String() string {
	if d.ModuleStmt != nil {
		return fmt.Sprintf("Module %s", d.ModuleStmt.Name)
	} else if d.FunctionStmt != nil {
		return fmt.Sprintf("Function %s %s", d.FunctionStmt.Type, d.FunctionStmt.Name)
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

func (s *Scope) AddVariable(vd *VarDecl) {
	name := vd.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	s.Decls[name] = &Decl{VarDecl: vd}
}

func NewGlobalScope() *Scope {
	return &Scope{
		IsGlobal: true,
		Decls:    map[string]*Decl{},
	}
}

func NewModuleScope(ms *ModuleStmt, parent *Scope) *Scope {
	return &Scope{
		Parent:     parent,
		ModuleStmt: ms,
		Decls:      map[string]*Decl{},
	}
}

func NewFunctionScope(fs *FunctionStmt, parent *Scope) *Scope {
	return &Scope{
		Parent:       parent,
		FunctionStmt: fs,
		Decls:        map[string]*Decl{},
	}
}
