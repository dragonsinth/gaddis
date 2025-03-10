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
	varId := request.Arguments.VariablesReference
	scopeId := varId / 1024
	isParamScope := varId&512 != 0
	_ = varId % 512 // TODO
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

	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst, _ int) {
		if id != scopeId {
			return
		}
		if isParamScope {
			for i, vd := range fr.Scope.Params {
				addVar(fr.Args[i], vd, i)
			}
		} else {
			nArgs := len(fr.Args)
			for i, vd := range fr.Scope.Params {
				addVar(fr.Locals[i], vd, i)
			}
			for i, vd := range fr.Scope.Locals {
				addVar(fr.Locals[i+nArgs], vd, i)
			}
		}
	})
	h.send(response)
}

func (h *Session) onSetVariableRequest(request *api.SetVariableRequest) {
	name := request.Arguments.Name
	value := request.Arguments.Value
	varId := request.Arguments.VariablesReference
	scopeId := varId / 1024
	isParamScope := varId&512 != 0
	_ = varId % 512 // TODO

	var err error
	var typStr string
	var valStr string
	h.sess.GetStackFrames(func(fr *asm.Frame, id int, inst asm.Inst, _ int) {
		if id != scopeId {
			return
		}

		decl, ref := func() (*ast.VarDecl, *any) {
			if isParamScope {
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Args[i]
					}
				}
			} else {
				nArgs := len(fr.Args)
				for i, vd := range fr.Scope.Params {
					if vd.Name == name {
						return vd, &fr.Locals[i]
					}
				}
				for i, vd := range fr.Scope.Locals {
					if vd.Name == name {
						return vd, &fr.Locals[i+nArgs]
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
