package gaddis

import (
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/astprint"
	"github.com/dragonsinth/gaddis/collect"
	"github.com/dragonsinth/gaddis/controlflow"
	"github.com/dragonsinth/gaddis/parse"
	"github.com/dragonsinth/gaddis/typecheck"
)

func Compile(src string) (prog *ast.Program, outSrc string, errs []ast.Error) {
	// parse and report lex/parse errors
	var comments []ast.Comment
	prog, comments, errs = parse.Parse(src)
	if len(errs) > 0 {
		return
	}

	outSrc = astprint.Print(prog, comments)

	// report collection and resolution errors together
	errs = collect.Collect(prog)
	if len(errs) > 0 {
		return
	}

	// errors related to inheritance
	errs = typecheck.SuperCheck(prog)
	if len(errs) > 0 {
		return
	}

	// resolves types, report type checking errors
	errs = typecheck.TypeCheck(prog, prog.Scope)
	if len(errs) > 0 {
		return
	}

	errs = controlflow.ControlFlow(prog)
	if len(errs) > 0 {
		return
	}

	return
}
