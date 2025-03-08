package gogen

import "github.com/dragonsinth/gaddis/ast"

type goStringStruct struct {
	ast.PrimitiveType
}

var goStringType = goStringStruct{ast.String}
