package gaddis

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/collect"
	"github.com/dragonsinth/gaddis/controlflow"
	"github.com/dragonsinth/gaddis/parse"
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

	// resolves types, report type checking errors
	errs = typecheck.TypeCheck(prog, prog.Scope)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	errs = controlflow.ControlFlow(prog)
	if len(errs) > 0 {
		return prog, comments, errs
	}

	return prog, comments, nil
}
