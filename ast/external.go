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

	for name, f := range lib.External {
		fs := translateMethod(name, reflect.TypeOf(f))
		fs.Scope = ret
		ret.Decls[name] = &Decl{FunctionStmt: fs}
	}

	return ret
}

func translateMethod(name string, methodType reflect.Type) *FunctionStmt {
	if methodType.NumOut() != 1 {
		panic(name)
	}
	returnType := translateType(methodType.Out(0))

	var params []*VarDecl
	for i := 0; i < methodType.NumIn(); i++ {
		p := methodType.In(i)
		params = append(params, &VarDecl{
			Type:    translateType(p),
			IsParam: true,
		})
	}

	return &FunctionStmt{
		SourceInfo: SourceInfo{},
		Name:       name,
		Type:       returnType,
		Params:     params,
		IsExternal: true,
		Scope:      nil,
	}
}

func translateType(inType reflect.Type) Type {
	t, ok := reverseTypeMap[inType.Name()]
	if !ok {
		panic(inType)
	}
	return t
}

var reverseTypeMap = map[string]Type{
	"int64":   Integer,
	"float64": Real,
	"string":  String,
	"byte":    Character,
	"bool":    Boolean,
}
