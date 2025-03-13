package dap

import (
	"bufio"
	"errors"
	"github.com/dragonsinth/gaddis/debug"
	api "github.com/google/go-dap"
	"io"
	"log"
	"sync"
)

// NewSession return a new DAP Session.
func NewSession(rw *bufio.ReadWriter, dbgLog *log.Logger) *Session {
	return &Session{
		rw:           rw,
		dbgLog:       dbgLog,
		sendQueue:    make(chan api.Message, 1024),
		bpsBySum:     map[string][]int{},
		sourceByPath: map[string]*debug.Source{},
		sourceBySum:  map[string]*debug.Source{},
	}
}

// Session manages the debug session and implements DAP.
type Session struct {
	// rw is used to read requests and write events/responses
	rw     *bufio.ReadWriter
	dbgLog *log.Logger

	// sendQueue is used to capture messages from multiple request
	// processing goroutines while writing them to the client connection
	// from a single goroutine via sendFromQueue. We must keep track of
	// the multiple channel senders with a wait group to make sure we do
	// not close this channel prematurely. Closing this channel will signal
	// the sendFromQueue goroutine that it can exit.
	sendQueue chan api.Message

	instBps      []int
	bpsBySum     map[string][]int
	sourceByPath map[string]*debug.Source
	sourceBySum  map[string]*debug.Source

	lineOff, colOff int

	sess       *debug.Session
	source     *api.Source
	launchArgs launchArgs
	runId      int // new per sess

	canTerminal bool
	terminal    *Terminal
	terminalPid int
}

// Run runs the session for as long as it last.
func (h *Session) Run() error {
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(h.sendQueue)

	defer func() {

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		h.sendFromQueue()
	}()

	defer func() {
		if h.sess != nil {
			h.sess.Host.SuppressAllEvents()
			h.sess.Halt()
		}
	}()

	defer func() {
		if h.terminal != nil {
			h.terminal.Close()
			if h.terminalPid > 0 {
				// TODO: kill?
			}
		}
	}()

	for {
		if err := h.handleRequest(); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (h *Session) handleRequest() error {
	msg, err := api.ReadProtocolMessage(h.rw.Reader)
	if err != nil {
		return err
	}
	if req, ok := msg.(api.RequestMessage); ok {
		h.dbgLog.Printf("Received request: %s", toJson(msg))
		h.dispatchRequest(req)
	} else if rsp, ok := msg.(api.ResponseMessage); ok {
		h.dbgLog.Printf("Received response: %s", toJson(msg))
		h.dispatchResponse(rsp)
	} else {
		log.Println("error: not a request message!\n", toJson(msg))
	}
	return nil
}

// send lets the sender goroutine know via a channel that there is
// a message to be sent to client.
func (h *Session) send(message api.Message) {
	select {
	case h.sendQueue <- message:
	default:
		// just drop messages if the queue is that backed up
	}
}

// sendFromQueue run in a separate goroutine to listen on a
// channel for messages to send back to the client. It will
// return once the channel is closed or the outbound conn
// becomes unwritable.
func (h *Session) sendFromQueue() {
	seq := 1
	for message := range h.sendQueue {
		switch m := message.(type) {
		case api.RequestMessage:
			m.GetRequest().Seq = seq
		case api.ResponseMessage:
			m.GetResponse().Seq = seq
		case api.EventMessage:
			m.GetEvent().Seq = seq
		default:
			panic(m)
		}
		seq++
		if err := api.WriteProtocolMessage(h.rw.Writer, message); err != nil {
			log.Println("Error writing message:", err)
			return
		}
		h.dbgLog.Printf("Message sent\n%s", toJson(message))
		if err := h.rw.Flush(); err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}

func (h *Session) stdout(line string) {
	h.send(&api.OutputEvent{
		Event: *newEvent("output"),
		Body: api.OutputEventBody{
			Category: "stdout",
			Output:   line,
			Source:   h.source,
		},
	})
}
