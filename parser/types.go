package parser

import "github.com/dragonsinth/gaddis/ast"

func isNumericType(a ast.Type) bool {
	return a == ast.Integer || a == ast.Real
}

func areComparableTypes(a ast.Type, b ast.Type) ast.Type {
	if a == b {
		return a
	}
	if isNumericType(a) && isNumericType(b) {
		return ast.Real // promote
	}
	return ast.InvalidType
}

func areOrderedTypes(a ast.Type, b ast.Type) bool {
	typ := areComparableTypes(a, b)
	if typ == ast.InvalidType {
		return false // must be comparable
	}
	if typ == ast.Boolean {
		return false // cannot order booleans
	}
	return true
}
