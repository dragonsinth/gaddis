package dap

import (
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/parse"
	api "github.com/google/go-dap"
	"strconv"
)

type variable struct {
	name string
	typ  ast.Type
	ref  *any
	id   int
}

func (h *Session) createVariable(name string, typ ast.Type, ref *any) api.Variable {
	if ref == nil || *ref == nil {
		return api.Variable{
			Name:         name,
			Value:        "<nil>",
			Type:         typ.String(),
			EvaluateName: name,
		}
	}
	if !typ.IsArrayType() {
		return api.Variable{
			Name:         name,
			Value:        asm.DebugStringVal(typ, *ref),
			Type:         typ.String(),
			EvaluateName: name,
		}
	}

	val := (*ref).([]any)

	v, ok := h.variablesByPtr[ref]
	if !ok {
		if h.variablesByPtr == nil {
			h.variablesByPtr = map[*any]variable{}
		}
		if h.variablesById == nil {
			h.variablesById = map[int]variable{}
		}
		v = variable{
			name: name,
			typ:  typ,
			ref:  ref,
			id:   len(h.variablesByPtr) + 1,
		}
		h.variablesByPtr[ref] = v
		h.variablesById[v.id] = v
	}
	return api.Variable{
		Name:               name,
		Value:              asm.DebugStringVal(typ, *ref),
		Type:               typ.String(),
		EvaluateName:       name,
		VariablesReference: v.id,
		IndexedVariables:   len(val),
	}
}

func (h *Session) onVariablesRequest(request *api.VariablesRequest) {
	if h.pausedSessionRequiredError(request) {
		return
	}

	response := &api.VariablesResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	targetVarId := request.Arguments.VariablesReference
	if targetVarId < 1<<14 {
		v, ok := h.variablesById[targetVarId]
		if !ok {
			h.send(response)
			return
		}
		if v.typ.IsArrayType() {
			elementType := v.typ.AsArrayType().ElementType
			val := (*v.ref).([]any)
			for i := range val {
				response.Body.Variables = append(response.Body.Variables, h.createVariable(strconv.Itoa(i), elementType, &val[i]))
			}
		} else {
			panic("dunno what this is")
		}
	}

	targetScopeId := getScopeId(targetVarId)
	targetFrameId := getFrameId(targetScopeId)

	addVar := func(ref *any, vd *ast.VarDecl) {
		response.Body.Variables = append(response.Body.Variables, h.createVariable(vd.Name, vd.Type, ref))
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
				addVar(&fr.Locals[i], fr.Scope.Locals[i])
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
				param := fr.Scope.Params[i]
				if param.IsRef {
					addVar(fr.Params[i].(*any), param)
				} else {
					addVar(&fr.Params[i], param)
				}
			}
		case ids.argsId:
			for i := range fr.Args {
				param := fr.Scope.Params[i]
				if param.IsRef {
					addVar(fr.Args[i].(*any), param)
				} else {
					addVar(&fr.Args[i], param)
				}
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

	var err error
	var typ ast.Type
	var ref *any
	if targetVarId < 1<<14 {
		// resolve list/map id variable
		err = func() error {
			v, ok := h.variablesById[targetVarId]
			if !ok {
				return fmt.Errorf("unknown variable %s", name)
			}
			if v.typ.IsArrayType() {
				elementType := v.typ.AsArrayType().ElementType
				val := (*v.ref).([]any)
				if id, err := strconv.Atoi(name); err != nil || id < 0 || id >= len(val) {
					return fmt.Errorf("invalid array index %s", name)
				} else {
					typ, ref = elementType, &val[id]
				}
			} else {
				panic("dunno what this is")
			}
			return nil
		}()
	} else {
		// resolve a scope
		h.sess.GetStackFrames(func(fr *asm.Frame, frameId int, inst asm.Inst, _ int) {
			err = func() error {
				if fr.Native != nil {
					return nil
				}
				if frameId != targetFrameId {
					return nil
				}
				ids := getScopeIds(frameId)

				// check for trying to set eval stack
				if targetScopeId == ids.localId {
					if n, _ := fmt.Sscanf(name, "[%d]", new(int)); n > 0 {
						return errors.New("eval stack may not be reassigned")
					}
				}

				typ, ref = func() (ast.Type, *any) {
					switch targetScopeId {
					case ids.localId:
						for i, vd := range fr.Scope.Locals {
							if vd.Name == name {
								return vd.Type, &fr.Locals[i]
							}
						}
					case ids.paramId:
						for i, vd := range fr.Scope.Params {
							if vd.Name == name {
								if vd.IsRef {
									return vd.Type, fr.Params[i].(*any)
								} else {
									return vd.Type, &fr.Params[i]
								}
							}
						}
					case ids.argsId:
						for i, vd := range fr.Scope.Params {
							if vd.Name == name {
								if vd.IsRef {
									return vd.Type, fr.Args[i].(*any)
								} else {
									return vd.Type, &fr.Args[i]
								}
							}
						}
					}
					return nil, nil
				}()

				return nil
			}()
		})
	}
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}

	// try to set
	err = func() error {
		if ref == nil {
			return fmt.Errorf("unknown variable %s", name)

		}
		if !typ.IsPrimitive() {
			return fmt.Errorf("variable %s of type %s is not primitive", name, typ)
		}

		var val any
		if value == "<nil>" {
			val = nil
		} else if lit := parse.ParseLiteral(value, ast.SourceInfo{}, typ.AsPrimitive()); lit != nil {
			val = lit.Val
			if str, ok := val.(string); ok {
				val = []byte(str)
			}
		} else {
			return fmt.Errorf("cannot parse <%s> into type %s", value, typ)
		}

		*ref = val

		response := &api.SetVariableResponse{}
		response.Response = *newResponse(request.Seq, request.Command)
		response.Body = api.SetVariableResponseBody{
			Type:  typ.String(),
			Value: asm.DebugStringVal(typ, val),
		}
		h.send(response)

		return nil
	}()
	if err != nil {
		h.send(newErrorResponse(request.Seq, request.Command, err.Error()))
		return
	}
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
