package dap

import (
	"bytes"
	"encoding/json"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"os"
	"strings"
)

type launchArgs struct {
	Name        string `json:"name"`
	Program     string `json:"program"`
	WorkDir     string `json:"workDir"`
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

func tryReadInput(program string) *bytes.Reader {
	buf, _ := os.ReadFile(program + ".in")
	return bytes.NewReader(buf)
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

func dapSource(source debug.Source) *api.Source {
	return &api.Source{
		Name:             source.Name,
		Path:             source.Path,
		SourceReference:  0,
		PresentationHint: "",
		Origin:           "",
		Sources:          nil,
		AdapterData:      nil,
		Checksums:        []api.Checksum{{Algorithm: "SHA256", Checksum: source.Sum}},
	}
}

func getChecksum(source api.Source) string {
	for _, cs := range source.Checksums {
		if cs.Algorithm == "SHA256" {
			return cs.Checksum
		}
	}
	return ""
}
