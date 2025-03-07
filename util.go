package gaddis

import (
	"io"
)

type Syncer interface {
	Sync() error
}

type SyncWriter interface {
	io.Writer
	Syncer
}

func NoopSyncWriter(w io.Writer) SyncWriter {
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

func WriteSync(w io.Writer, sync Syncer) SyncWriter {
	return writeSync{
		Writer: w,
		Syncer: sync,
	}
}

type writeSync struct {
	io.Writer
	Syncer
}
