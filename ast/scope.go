package ast

import "fmt"

// TODO: more than just var decls (module, function, class).

type Scope struct {
	Parent *Scope
	Decls  map[string]*Decl
}

type Decl struct {
	ModuleStmt *ModuleStmt
	VarDecl    *VarDecl
}

func (d *Decl) String() string {
	if d.ModuleStmt != nil {
		return fmt.Sprintf("Module %s", d.ModuleStmt.Name)
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

func (s *Scope) AddVariable(vd *VarDecl) {
	name := vd.Name
	if s.Decls[name] != nil {
		panic(name)
	}
	s.Decls[name] = &Decl{VarDecl: vd}
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent: parent,
		Decls:  map[string]*Decl{},
	}
}
