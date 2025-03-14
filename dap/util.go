package dap

import (
	"bytes"
	"encoding/json"
	api "github.com/google/go-dap"
	"os"
	"strings"
)

type launchArgs struct {
	Name        string `json:"name"`
	Program     string `json:"program"`
	WorkDir     string `json:"workDir"`
	TestMode    bool   `json:"testMode"`
	StopOnEntry bool   `json:"stopOnEntry"`
	NoDebug     bool   `json:"noDebug"`
}

func newEvent(event string) *api.Event {
	return &api.Event{
		ProtocolMessage: api.ProtocolMessage{
			Seq:  0,
			Type: "event",
		},
		Event: event,
	}
}

func newRequest(command string) *api.Request {
	return &api.Request{
		ProtocolMessage: api.ProtocolMessage{
			Seq:  0,
			Type: "request",
		},
		Command: command,
	}
}

func newResponse(requestSeq int, command string) *api.Response {
	return &api.Response{
		ProtocolMessage: api.ProtocolMessage{
			Seq:  0,
			Type: "response",
		},
		Command:    command,
		RequestSeq: requestSeq,
		Success:    true,
	}
}

func newErrorResponse(requestSeq int, command string, message string) *api.ErrorResponse {
	er := &api.ErrorResponse{}
	er.Response = *newResponse(requestSeq, command)
	er.Success = false
	er.Message = message
	er.Body.Error = &api.ErrorMessage{Format: message}
	return er
}

func tryReadInput(program string) []byte {
	buf, _ := os.ReadFile(program + ".in")
	return buf
}

func fromJson(buf json.RawMessage, obj any) error {
	dec := json.NewDecoder(bytes.NewReader(buf))
	return dec.Decode(obj)
}

func toJson(obj any) string {
	var buf strings.Builder
	e := json.NewEncoder(&buf)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	if err := e.Encode(obj); err != nil {
		panic(err)
	}
	return buf.String()
}
