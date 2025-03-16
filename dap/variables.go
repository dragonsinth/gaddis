package dap

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/parse"
	api "github.com/google/go-dap"
)

func (h *Session) onVariablesRequest(request *api.VariablesRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	targetVarId := request.Arguments.VariablesReference
	targetScopeId := getScopeId(targetVarId)
	targetFrameId := getFrameId(targetScopeId)
	_ = getVarIndex(targetVarId) // TODO: individual variables

	response := &api.VariablesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	addVar := func(val any, vd *ast.VarDecl, id int) {
		response.Body.Variables = append(response.Body.Variables, api.Variable{
			Name:               vd.Name,
			Value:              asm.DebugStringVal(vd.Type, val),
			Type:               vd.Type.String(),
			PresentationHint:   nil,
			EvaluateName:       vd.Name,
			VariablesReference: 0, // TODO: map/list
			NamedVariables:     0, // map
			IndexedVariables:   0, // list
			MemoryReference:    "",
		})
	}

	h.sess.GetStackFrames(func(fr *asm.Frame, frameId int, inst asm.Inst, _ int) {
		if fr.Native != nil {
			return
		}
		if frameId != targetFrameId {
			return
		}
		ids := getScopeIds(frameId)
		switch targetScopeId {
		case ids.localId:
			for i := range fr.Locals {
				addVar(fr.Locals[i], fr.Scope.Locals[i], i)
			}
			// TODO: consider how to get type information for eval stack.
			for i := range fr.Eval {
				response.Body.Variables = append(response.Body.Variables, api.Variable{
					Name:  fmt.Sprintf("[%d]", i),
					Value: asm.DebugStringVal(ast.UnresolvedType, fr.Eval[i]),
				})
			}
		case ids.paramId:
			for i := range fr.Params {
				addVar(fr.Params[i], fr.Scope.Params[i], i)
			}
		case ids.argsId:
			for i := range fr.Args {
				addVar(fr.Args[i], fr.Scope.Params[i], i)
			}
		}
	})
	h.send(response)
}

func (h *Session) onSetVariableRequest(request *api.SetVariableRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}
	name := request.Arguments.Name
	value := request.Arguments.Value

	targetVarId := request.Arguments.VariablesReference
	targetScopeId := getScopeId(targetVarId)
	targetFrameId := getFrameId(targetScopeId)
	_ = getVarIndex(targetVarId) // TODO: individual variables

	var err error
	var typStr string
	var valStr string
	h.sess.GetStackFrames(func(fr *asm.Frame, frameId int, inst asm.Inst, _ int) {
		if fr.Native != nil {
			return
		}
		if frameId != targetFrameId {
			return
		}
		ids := getScopeIds(frameId)

		// check for trying to set eval stack
		if targetScopeId == ids.localId {
			if n, _ := fmt.Sscanf(name, "[%d]", new(int)); n > 0 {
				err = errors.New("eval stack may not be reassigned")
				return
			}
		}

		decl, ref := func() (*ast.VarDecl, *any) {
			switch targetScopeId {
			case ids.localId:
				for i, vd := range fr.Scope.Locals {
					if vd.Name == name {
						return vd, &fr.Locals[i]
					}
				}
			case ids.paramId:
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Params[i]
					}
				}
			case ids.argsId:
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Args[i]
					}
				}
			}
			return nil, nil
		}()

		if decl == nil {
			err = fmt.Errorf("unknown variable %s", name)
			return
		}
		if !decl.Type.IsPrimitive() {
			err = fmt.Errorf("variable %s of type %s is not primitive", name, decl.Type)
			return
		}

		var val any
		if value == "<nil>" {
			val = nil
		} else if lit := parse.ParseLiteral(value, ast.SourceInfo{}, decl.Type.AsPrimitive()); lit != nil {
			val = lit.Val
			if str, ok := val.(string); ok {
				val = []byte(str)
			}
		} else {
			err = fmt.Errorf("cannot parse <%s> into type %s", value, decl.Type)
			return
		}

		*ref = val
		typStr = decl.Type.String()
		valStr = asm.DebugStringVal(decl.Type, val)
	})
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}

	response := &api.SetVariableResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.SetVariableResponseBody{
		Value:              valStr,
		Type:               typStr,
		VariablesReference: 0,
		NamedVariables:     0,
		IndexedVariables:   0,
	}
	h.send(response)
}

func (h *Session) onEvaluateRequest(request *api.EvaluateRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}

	val, typ, err := h.sess.EvaluateExpressionInFrame(request.Arguments.FrameId, request.Arguments.Expression)
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}

	response := &api.EvaluateResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body.Result = asm.DebugStringVal(typ, val)
	response.Body.Type = typ.String()
	h.send(response)
}
