package typecheck

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
	"maps"
	"slices"
)

// TODO(scottb):
// - non-private fieldname collisions?
// - public name collisions between super/sub?

func SuperCheck(prog *ast.Program) []ast.Error {
	ch := &superChecker{
		done: map[string]map[string]*method{},
	}
	for _, class := range prog.Scope.Classes {
		ch.check(class)
	}
	return ch.Errors
}

type method struct {
	Name     string
	Id       int
	Callable ast.Callable
}

type superChecker struct {
	base.Visitor
	done map[string]map[string]*method
}

func (ch *superChecker) check(class *ast.ClassStmt) map[string]*method {
	if table := ch.done[class.Name]; table != nil {
		return table
	}

	var superMap map[string]*method
	var superMethods []ast.Callable
	var superFields []*ast.VarDecl
	if super := class.Type.Extends; super != nil {
		superMap = ch.check(super.Class)
		superMethods = super.Scope.Methods
		superFields = super.Scope.Fields
	}

	myMap := map[string]*method{}
	maps.Copy(myMap, superMap)
	myMethods := slices.Clone(superMethods)
	myFields := slices.Clone(superFields)

	for _, stmt := range class.Block.Statements {
		switch stmt := stmt.(type) {
		case *ast.ModuleStmt:
			if stmt.IsConstructor {
				continue // cannot virtual call
			}
			hdr := ast.SourceInfo{
				Start: stmt.Start,
				End:   stmt.Block.Start,
			}
			var id int
			if superMethod := superMap[stmt.Name]; superMethod != nil {
				// type check
				sm, ok := superMethod.Callable.(*ast.ModuleStmt)
				if !ok {
					ch.Errorf(hdr, "Module cannot override super Function %s", stmt.Name)
					continue
				}
				if !ch.checkParams(hdr, stmt.Params, sm.Params) {
					continue
				}

				// override
				id = superMethod.Id
			} else {
				// a new method
				id = len(myMethods)
				myMethods = append(myMethods, nil)
			}
			stmt.Id = id
			myMethods[id] = stmt
			myMap[stmt.Name] = &method{
				Name:     stmt.Name,
				Id:       id,
				Callable: stmt,
			}
		case *ast.FunctionStmt:
			hdr := ast.SourceInfo{
				Start: stmt.Start,
				End:   stmt.Block.Start,
			}
			var id int
			if superMethod := superMap[stmt.Name]; superMethod != nil {
				// type check
				sf, ok := superMethod.Callable.(*ast.FunctionStmt)
				if !ok {
					ch.Errorf(hdr, "Function cannot override super Module %s", stmt.Name)
					continue
				}
				if !ch.checkReturnType(hdr, stmt.Type, sf.Type) || !ch.checkParams(hdr, stmt.Params, sf.Params) {
					continue
				}

				// override
				id = superMethod.Id
			} else {
				// a new method
				id = len(myMethods)
				myMethods = append(myMethods, nil)
			}

			stmt.Id = id
			myMethods[id] = stmt
			myMap[stmt.Name] = &method{
				Name:     stmt.Name,
				Id:       id,
				Callable: stmt,
			}
		case *ast.DeclareStmt:
			for _, decl := range stmt.Decls {
				decl.Id = len(myFields)
				myFields = append(myFields, decl)
			}
		}
	}

	class.Scope.Methods = myMethods
	class.Scope.Fields = myFields
	ch.done[class.Name] = myMap
	return myMap
}

func (ch *superChecker) checkReturnType(si ast.SourceInfo, sub ast.Type, super ast.Type) bool {
	if sub != super {
		ch.Errorf(si, "return: expected to match super type %s, got %s", super, sub)
		return false
	}
	return true
}

func (ch *superChecker) checkParams(si ast.SourceInfo, subs []*ast.VarDecl, supers []*ast.VarDecl) bool {
	if len(subs) != len(supers) {
		ch.Errorf(si, "expected %d params to match super method, got %d", len(supers), len(subs))
		return false
	}
	ok := true
	for i := range subs {
		sub := subs[i].Type
		super := supers[i].Type
		if sub != super {
			ch.Errorf(si, "argument %d: expected to match super type %s, got %s", i, super, sub)
			ok = false
		}
	}
	return ok
}
