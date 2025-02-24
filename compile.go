package gaddis

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/collect"
	"github.com/dragonsinth/gaddis/controlflow"
	"github.com/dragonsinth/gaddis/parse"
	"github.com/dragonsinth/gaddis/resolve"
	"github.com/dragonsinth/gaddis/typecheck"
)

func Compile(src string) (*ast.Program, []ast.Comment, []ast.Error) {
	// parse and report lex/parse errors
	prog, comments, errs := parse.Parse(src)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	// report collection and resolution errors together
	errs = collect.Collect(prog)
	errs = append(errs, resolve.Resolve(prog)...)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	// resolves types, report type checking errors
	errs = typecheck.TypeCheck(prog)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	errs = controlflow.ControlFlow(prog)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	return prog, comments, nil
}
