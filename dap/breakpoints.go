package dap

import (
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
)

func (h *Session) onSetBreakpointsRequest(request *api.SetBreakpointsRequest) {
	var source *debug.Source
	if !request.Arguments.SourceModified {
		source = h.sourceBySum[getChecksum(request.Arguments.Source)]
		if source == nil {
			source = h.sourceByPath[request.Arguments.Source.Path]
		}
	}
	if source == nil {
		var err error
		source, err = debug.LoadSource(request.Arguments.Source.Path)
		if err != nil {
			h.send(newErrorResponse(request.Seq, request.Command, "error loading source"))
			return
		} else if source.Breakpoints == nil {
			h.send(newErrorResponse(request.Seq, request.Command, "parsing source"))
			return
		}
		h.sourceByPath[source.Path] = source
		h.sourceBySum[source.Sum] = source
	}

	response := &api.SetBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	if h.noDebug {
		h.send(response)
		return
	}

	breakpoints := source.Breakpoints

	var bps []int
	srcPtr := dapSource(*source)
	for _, bp := range request.Arguments.Breakpoints {
		var msg string
		srcLine := bp.Line - h.lineOff
		instLine := breakpoints.InstFromSource(srcLine)
		if instLine > 0 {
			bps = append(bps, srcLine)
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified:             true,
				Message:              msg,
				Source:               srcPtr,
				Line:                 bp.Line,
				Column:               h.colOff,
				InstructionReference: asm.PcRef(instLine),
			})
			srcPtr = nil
		} else {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified: false,
				Message:  "failed",
			})
		}

	}

	h.bpsBySum[source.Sum] = bps
	if h.sess != nil && source.Sum == h.sess.Source.Sum {
		h.sess.SetLineBreakpoints(bps)
	}
	h.send(response)
}

func (h *Session) onSetExceptionBreakpointsRequest(request *api.SetExceptionBreakpointsRequest) {
	response := &api.SetExceptionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	h.send(response)
}

func (h *Session) onSetInstructionBreakpointsRequest(request *api.SetInstructionBreakpointsRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	response := &api.SetInstructionBreakpointsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	breakpoints := h.sess.Source.Breakpoints
	var pcs []int
	for _, bp := range request.Arguments.Breakpoints {
		origPc := asm.RefPc(bp.InstructionReference)
		pc := origPc + (bp.Offset / 4)
		if origPc < 0 || pc < 0 || pc >= breakpoints.NInst {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified: false,
				Message:  "failed",
			})
		} else {
			pcs = append(pcs, pc)
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.Breakpoint{
				Verified:             true,
				Line:                 breakpoints.SourceFromInst(pc) + h.lineOff,
				Column:               h.colOff,
				InstructionReference: asm.PcRef(pc),
			})
		}

	}
	h.sess.SetInstBreakpoints(pcs)
	h.send(response)
}

func (h *Session) onBreakpointLocationsRequest(request *api.BreakpointLocationsRequest) {
	source := h.sourceBySum[getChecksum(request.Arguments.Source)]
	if source == nil {
		source = h.sourceByPath[request.Arguments.Source.Path]
	}
	if source == nil {
		var err error
		source, err = debug.LoadSource(request.Arguments.Source.Path)
		if err != nil {
			h.send(newErrorResponse(request.Seq, request.Command, "error loading source"))
			return
		} else if source.Breakpoints == nil {
			h.send(newErrorResponse(request.Seq, request.Command, "parsing source"))
			return
		}
		h.sourceByPath[source.Path] = source
		h.sourceBySum[source.Sum] = source
	}

	response := &api.BreakpointLocationsResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	startLine := request.Arguments.Line
	endLine := max(startLine, request.Arguments.EndLine)

	for line := startLine; line <= endLine; line++ {
		srcLine := line - h.lineOff
		if source.Breakpoints.ValidSrcLine(srcLine) {
			response.Body.Breakpoints = append(response.Body.Breakpoints, api.BreakpointLocation{
				Line:   line,
				Column: h.colOff,
			})
		}
	}

	h.send(response)
}

func (h *Session) onExceptionInfoRequest(request *api.ExceptionInfoRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}
	if request.Arguments.ThreadId != 1 {
		h.send(newErrorResponse(request.Seq, request.Command, "unknown threadId"))
		return
	}

	response := &api.ExceptionInfoResponse{}
	response.Response = *newResponse(request.Seq, request.Command)
	response.Body = api.ExceptionInfoResponseBody{
		ExceptionId: "",
		Description: "",
		BreakMode:   "always",
	}

	trace, exception := h.sess.GetCurrentException()
	if exception != nil {
		msg := exception.Error()
		response.Body.Description = msg
		response.Body.Details = &api.ExceptionDetails{
			Message:    msg,
			StackTrace: trace,
		}
	}
	h.send(response)
}
