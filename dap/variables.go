package dap

import (
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/parse"
	api "github.com/google/go-dap"
)

func (h *Session) onVariablesRequest(request *api.VariablesRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
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
			Value:              asm.DebugStringVal(val),
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
		if frameId != targetFrameId {
			return
		}
		ids := getScopeIds(frameId)
		switch targetScopeId {
		case ids.localId:
			for i := range fr.Locals {
				addVar(fr.Locals[i], fr.Scope.Locals[i], i)
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
		if frameId != targetFrameId {
			return
		}

		decl, ref := func() (*ast.VarDecl, *any) {
			ids := getScopeIds(frameId)
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
			err = fmt.Errorf("failed to parse value %q of type %s", value, decl.Type)
			return
		}

		*ref = val
		typStr = decl.Type.String()
		valStr = asm.DebugStringVal(val)
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
		VariablesReference: request.Arguments.VariablesReference,
		NamedVariables:     0,
		IndexedVariables:   0,
	}
	h.send(response)
}
