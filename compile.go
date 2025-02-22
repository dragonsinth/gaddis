package gaddis

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/collect"
	"github.com/dragonsinth/gaddis/parse"
	"github.com/dragonsinth/gaddis/resolve"
	"github.com/dragonsinth/gaddis/typecheck"
)

func Compile(src []byte) (*ast.Block, []ast.Comment, []ast.Error) {
	// parse and report lex/parse errors
	block, comments, errs := parse.Parse(src)
	if len(errs) > 0 {
		return block, comments, errs
	}

	// report collection and resolution errors together
	errs = collect.Collect(block)
	errs = append(errs, resolve.Resolve(block)...)
	if len(errs) > 0 {
		return block, comments, errs
	}

	// resolves types, report type checking errors
	errs = typecheck.TypeCheck(block)
	if len(errs) > 0 {
		return block, comments, errs
	}

	return block, comments, nil
}
