package controlflow

import (
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/base"
)

// ControlFlow analyzes unreachable code, no return statement, read-before-write on locals.
func ControlFlow(prog *ast.Program) []ast.Error {
	cv := &Visitor{}
	prog.Block.Visit(cv)
	cv.pop(prog.Block)
	if len(cv.stack) != 0 {
		panic("here")
	}

	av := &AssignmentVisitor{
		// Explicitly don't analyze global variables.
		// read-before-write on globals too difficult.
		currScope: nil,
	}
	prog.Block.Visit(av)

	return append(cv.Errors, av.Errors...)
}

type Flow int

const (
	INVALID Flow = iota

	CONTINUE // execution continues
	MAYBE    // execution may or may not continue
	HALT     // execution halts (return, error, infinite loop)
)

type stackElement struct {
	Statement ast.Statement
	Flow      Flow
}

type Visitor struct {
	base.Visitor
	stack []stackElement
}

var _ ast.Visitor = &Visitor{}

func (v *Visitor) PostVisitBlock(bl *ast.Block) {
	pl := v.popList(len(bl.Statements))
	flow := CONTINUE
	warned := false
	for _, stmt := range bl.Statements {
		if flow == HALT && !warned {
			warned = true
			v.Errorf(stmt, "unreachable code")
		}

		f := pl.pop(stmt)
		switch f {
		case HALT:
			flow = HALT
		case CONTINUE:
		case MAYBE:
			if flow < MAYBE {
				flow = MAYBE
			}
		default:
			panic(f)
		}
	}
	v.push(bl, flow)
}

func (v *Visitor) PostVisitDeclareStmt(ds *ast.DeclareStmt) {
	v.push(ds, CONTINUE)
}

func (v *Visitor) PostVisitDisplayStmt(ds *ast.DisplayStmt) {
	v.push(ds, CONTINUE)
}

func (v *Visitor) PostVisitInputStmt(is *ast.InputStmt) {
	v.push(is, CONTINUE)
}

func (v *Visitor) PostVisitSetStmt(ss *ast.SetStmt) {
	v.push(ss, CONTINUE)
}

func (v *Visitor) PostVisitIfStmt(is *ast.IfStmt) {
	alt := alternatives{}
	pl := v.popList(len(is.Cases))

	hasDefault := false
	for _, cb := range is.Cases {
		alt.Add(pl.pop(cb.Block))
		if cb.Expr == nil {
			hasDefault = true
		}
	}
	if !hasDefault {
		alt.Add(CONTINUE) // every alternative could be skipped
	}
	v.push(is, alt.Result())
}

func (v *Visitor) PostVisitSelectStmt(ss *ast.SelectStmt) {
	count := len(ss.Cases)
	alt := alternatives{}
	pl := v.popList(count)

	hasDefault := false
	for _, cas := range ss.Cases {
		alt.Add(pl.pop(cas.Block))
		if cas.Expr == nil {
			hasDefault = true
		}
	}
	if !hasDefault {
		alt.Add(CONTINUE) // every alternative could be skipped
	}
	v.push(ss, alt.Result())
}

func (v *Visitor) PostVisitDoStmt(ds *ast.DoStmt) {
	flow := v.pop(ds.Block)
	infinite := false
	if eval := ds.Expr.ConstEval(); eval != nil {
		val := eval.(bool)
		if !ds.Until && val {
			// While True
			infinite = true
		} else if ds.Until && !val {
			// Until False
			infinite = true
		}
	}
	if infinite {
		if flow == CONTINUE {
			v.Errorf(ds, "infinite loop")
		}
		// the inner loop must return (or if it doesn't, at least suppress more errors)
		flow = HALT
	}
	v.push(ds, flow)
}

func (v *Visitor) PostVisitWhileStmt(ws *ast.WhileStmt) {
	flow := v.pop(ws.Block)
	var infinite, skipped bool
	if eval := ws.Expr.ConstEval(); eval != nil {
		val := eval.(bool)
		if val {
			infinite = true
		} else {
			skipped = true
		}
	}

	if skipped {
		flow = CONTINUE
	} else if infinite {
		if flow == CONTINUE {
			v.Errorf(ws, "infinite loop")
		}
		// the inner loop must return (or if it doesn't, at least suppress more errors)
		flow = HALT
	}

	v.push(ws, flow)
}

func (v *Visitor) PostVisitForStmt(fs *ast.ForStmt) {
	// assume a for loop might run 0 times
	// TODO: const eval start/stop/step
	alt := alternatives{}
	alt.Add(CONTINUE)
	alt.Add(v.pop(fs.Block))
	v.push(fs, alt.Result())
}

func (v *Visitor) PostVisitCallStmt(cs *ast.CallStmt) {
	v.push(cs, CONTINUE)
}

func (v *Visitor) PostVisitModuleStmt(ms *ast.ModuleStmt) {
	v.pop(ms.Block)
	v.push(ms, CONTINUE)
}

func (v *Visitor) PostVisitReturnStmt(rs *ast.ReturnStmt) {
	v.push(rs, HALT)
}

func (v *Visitor) PostVisitFunctionStmt(fs *ast.FunctionStmt) {
	flow := v.pop(fs.Block)
	if flow != HALT {
		v.Errorf(fs, "not all control paths return a value")
	}
	v.push(fs, CONTINUE)
}

func (v *Visitor) push(stmt ast.Statement, flow Flow) {
	v.stack = append(v.stack, stackElement{Statement: stmt, Flow: flow})
}

func (v *Visitor) pop(stmt ast.Statement) Flow {
	last := v.stack[len(v.stack)-1]
	if last.Statement != stmt {
		panic(fmt.Sprintf("expected statement %T %v, got %T %v", stmt, stmt, last.Statement, last.Statement))
	}
	v.stack = v.stack[:len(v.stack)-1]
	return last.Flow
}

func (v *Visitor) popList(n int) poppedList {
	ret := v.stack[len(v.stack)-n:]
	v.stack = v.stack[:len(v.stack)-n]
	return poppedList{stack: ret}
}

type poppedList struct {
	stack []stackElement
}

func (pl *poppedList) pop(stmt ast.Statement) Flow {
	next := pl.stack[0]
	if next.Statement != stmt {
		panic(fmt.Sprintf("expected statement %T %v, got %T %v", stmt, stmt, next.Statement, next.Statement))
	}
	pl.stack = pl.stack[1:]
	return next.Flow
}
