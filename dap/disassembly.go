package dap

import (
	"github.com/dragonsinth/gaddis/asm"
	api "github.com/google/go-dap"
)

func (h *Session) onDisassembleRequest(request *api.DisassembleRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	args := request.Arguments
	origPC := asm.RefPc(args.MemoryReference)
	if origPC < 0 {
		// TODO: return all invalid?
		h.send(newErrorResponse(request.Seq, request.Command, "unable to disassemble"))
		return
	}

	origPC += args.Offset / 4
	start := origPC + args.InstructionOffset

	response := &api.DisassembleResponse{}
	response.Response = *newResponse(request.Seq, request.Command)

	lastLine := -1
	source := h.sess.Source
	srcRef := h.source
	for i := 0; i < args.InstructionCount; i++ {
		pc := i + start
		if pc < 0 || pc >= source.Breakpoints.NInst {
			response.Body.Instructions = append(response.Body.Instructions, api.DisassembledInstruction{
				Address:     asm.PcRef(pc),
				Instruction: "noop",
			})
		} else {
			inst := source.Assembled.Code[pc]
			si := inst.GetSourceInfo()
			pos := si.Start
			di := api.DisassembledInstruction{
				Address:     asm.PcRef(pc),
				Instruction: inst.String(),
				Symbol:      inst.Sym(),
			}
			if pos.Line != lastLine {
				if srcRef != nil {
					di.Location = srcRef
					srcRef = nil
				}
				di.Line = pos.Line + h.lineOff
				di.Column = pos.Column + h.colOff
				lastLine = pos.Line
			}
			response.Body.Instructions = append(response.Body.Instructions, di)
		}
	}

	h.send(response)
}
