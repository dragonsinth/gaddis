package ast

type Scope struct {
	Parent *Scope
	Decls  map[string]*VarDecl
}

func (s *Scope) Lookup(name string) *VarDecl {
	if v := s.Decls[name]; v != nil {
		return v
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent: parent,
		Decls:  map[string]*VarDecl{},
	}
}
