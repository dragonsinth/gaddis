package dap

import (
	"fmt"
	api "github.com/google/go-dap"
	"strconv"
)

func (h *Session) onDisassembleRequest(request *api.DisassembleRequest) {
	if h.sess == nil {
		h.send(newErrorResponse(request.Seq, request.Command, "no session found"))
		return
	}

	args := request.Arguments
	origPC := refPc(args.MemoryReference)
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
				Address:     "",
				Instruction: "invalid",
			})
		} else {
			inst := source.Assembled.Code[pc]
			si := inst.GetSourceInfo()
			pos := si.Start
			di := api.DisassembledInstruction{
				Address:     pcRef(pc),
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

func pcRef(pc int) string {
	pc = pc * 4
	pc += 0x1000 // start program here
	ret := fmt.Sprintf("0x%04X", pc)
	return ret
}

func refPc(s string) int {
	pc, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return -1
	}
	if pc < 0x1000 || pc > 0x7fff {
		return -1
	}
	pc -= 0x1000
	pc = pc / 4
	return int(pc)
}
