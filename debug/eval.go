package debug

import (
	"errors"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/parse"
	"github.com/dragonsinth/gaddis/typecheck"
)

func (ds *Session) evaluateExprInFrame(fr *asm.Frame, exprStr string) (any, ast.Type, error) {
	expr, err := parse.ParseExpr(exprStr)
	if err != nil {
		return nil, nil, errors.New("syntax error")
	}

	errs := typecheck.TypeCheck(expr, fr.Scope)
	if len(errs) > 0 {
		return nil, nil, errors.New(errs[0].Desc)
	}

	// generate new instructions
	evalInst := ds.Source.Assembled.AssembleExpression(expr)
	// copy all the state from the main executor
	p := *ds.Exec
	p.PC = len(p.Code) // start at the end
	p.Code = append(p.Code, evalInst...)
	p.Stack = append(p.Stack, asm.Frame{
		Scope:  ast.NewEvalScope(expr, fr.Scope),
		Start:  p.PC,
		Return: 0,
		Args:   fr.Args,
		Params: fr.Params,
		Locals: fr.Locals,
		Eval:   make([]any, 0, 16),
	})
	evalFrame := &p.Stack[len(p.Stack)-1]
	p.Frame = evalFrame

	if err := p.Run(); err != nil {
		return nil, nil, err
	}
	// should have left exactly 1 value on the stack
	return evalFrame.Eval[0], expr.GetType(), nil
}
