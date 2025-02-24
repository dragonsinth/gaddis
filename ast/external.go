package ast

import (
	main_template "github.com/dragonsinth/gaddis/gogen/builtins"
	"reflect"
	"strings"
)

var ExternalScope = computeExternalScope()

func computeExternalScope() *Scope {
	ret := &Scope{
		Parent:     nil,
		IsExternal: true,
		Decls:      map[string]*Decl{},
	}

	libType := reflect.TypeOf(main_template.Lib)
	for i := 0; i < libType.NumMethod(); i++ {
		fs := translateMethod(libType.Method(i))
		fs.Scope = ret
		lower := strings.ToLower(fs.Name)
		camel := lower[:1] + fs.Name[1:]
		ret.Decls[camel] = &Decl{FunctionStmt: fs}
	}

	return ret
}

func translateMethod(m reflect.Method) *FunctionStmt {
	if m.Type.NumOut() != 1 {
		panic(m)
	}
	returnType := translateType(m.Type.Out(0))

	var params []*VarDecl
	for i := 1; i < m.Type.NumIn(); i++ {
		p := m.Type.In(i)
		params = append(params, &VarDecl{
			Type:    translateType(p),
			IsParam: true,
		})
	}

	return &FunctionStmt{
		SourceInfo: SourceInfo{},
		Name:       m.Name,
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
