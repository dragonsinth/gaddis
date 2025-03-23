package ast

import (
	"github.com/dragonsinth/gaddis/lib"
	"reflect"
)

var ExternalScope = computeExternalScope()

func computeExternalScope() *Scope {
	ret := &Scope{
		Parent:     nil,
		IsExternal: true,
		Decls:      map[string]*Decl{},
	}

	for _, entry := range lib.GetLibrary() {
		ret.Decls[entry.Name] = translateMethod(ret, entry.Name, entry.FuncPtr.Type())
	}

	return ret
}

func translateMethod(scope *Scope, name string, methodType reflect.Type) *Decl {
	var params []*VarDecl
	for i := 0; i < methodType.NumIn(); i++ {
		p := methodType.In(i)
		typ, isRef := translateType(p)
		params = append(params, &VarDecl{
			Type:    typ,
			IsParam: true,
			IsRef:   isRef,
		})
	}

	switch methodType.NumOut() {
	case 0:
		ms := &ModuleStmt{
			SourceInfo: SourceInfo{},
			Name:       name,
			Params:     params,
			IsExternal: true,
			Scope:      scope,
		}
		return &Decl{ModuleStmt: ms}
	case 1:
		returnType, _ := translateType(methodType.Out(0))
		fs := &FunctionStmt{
			SourceInfo: SourceInfo{},
			Name:       name,
			Type:       returnType,
			Params:     params,
			IsExternal: true,
			Scope:      scope,
		}
		return &Decl{FunctionStmt: fs}
	default:
		panic(name)
	}
}

func translateType(inType reflect.Type) (Type, bool) {
	isRef := false
	if inType.Kind() == reflect.Ptr {
		isRef = true
		inType = inType.Elem()
	}

	t, ok := reverseTypeMap[inType.String()]
	if !ok {
		panic(inType.String())
	}
	return t, isRef
}

var reverseTypeMap = map[string]Type{
	"int64":   Integer,
	"float64": Real,
	"[]uint8": String,
	"uint8":   Character,
	"bool":    Boolean,
}
