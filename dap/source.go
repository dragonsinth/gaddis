package dap

import (
	"github.com/dragonsinth/gaddis/debug"
	"github.com/dragonsinth/gaddis/lib"
	api "github.com/google/go-dap"
	"path/filepath"
)

func (h *Session) onSourceRequest(request *api.SourceRequest) {
	libSrc := lib.SrcById(request.Arguments.SourceReference)
	if libSrc == nil {
		h.send(newErrorResponse(request.GetSeq(), request.Command, "uknown source"))
		return
	}

	h.send(&api.SourceResponse{
		Response: *newResponse(request.Seq, request.Command),
		Body: api.SourceResponseBody{
			Content:  libSrc.Src,
			MimeType: "text/plain; charset=utf-8",
		},
	})
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

func libSource(filename string) *api.Source {
	if src := lib.SrcByName(filepath.Base(filename)); src != nil {
		return &api.Source{
			Path:            filename,
			SourceReference: src.Id,
		}
	}
	return nil
}

func getChecksum(source api.Source) string {
	for _, cs := range source.Checksums {
		if cs.Algorithm == "SHA256" {
			return cs.Checksum
		}
	}
	return ""
}
