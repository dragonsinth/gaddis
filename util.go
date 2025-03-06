package gaddis

import (
	"io"
)

type SyncWriter interface {
	io.Writer
	Sync() error
}

func Synced(w io.Writer) SyncWriter {
	if sw, ok := w.(SyncWriter); ok {
		return sw
	}
	return noopSyncWriter{w}
}

type noopSyncWriter struct {
	io.Writer
}

func (noopSyncWriter) Sync() error {
	return nil
}
