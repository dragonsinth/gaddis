package gaddis

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/parse"
	"github.com/dragonsinth/gaddis/typecheck"
)

func Compile(src []byte) (*ast.Block, []ast.Comment, []ast.Error) {
	block, comments, errs := parse.Parse(src)
	if len(errs) > 0 {
		return block, comments, errs
	}

	errs = typecheck.TypeCheck(block)
	if len(errs) > 0 {
		return block, comments, errs
	}

	return block, comments, nil
}
