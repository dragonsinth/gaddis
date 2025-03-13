package dap

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

type Terminal struct {
	Port     int
	Listener net.Listener
	Ready    chan struct{}
	Conn     *net.TCPConn // cannot read until ready is closed
}

func (t *Terminal) Read(p []byte) (n int, err error) {
	<-t.Ready
	if t.Conn != nil {
		return t.Conn.Read(p)
	}
	return 0, io.EOF
}

func (t *Terminal) Close() {
	_ = t.Listener.Close()
	<-t.Ready
	if t.Conn != nil {
		// try to friendly drain and close the conn for the client
		_ = t.Conn.CloseWrite()
		_ = t.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, _ = io.Copy(io.Discard, t.Conn)
		_ = t.Conn.Close()
	}
}

func (t *Terminal) Output(line string) {
	<-t.Ready
	if t.Conn != nil {
		_, _ = t.Conn.Write([]byte(line))
	}
}

func (t *Terminal) Interrupt() {
	<-t.Ready
	if t.Conn != nil {
		_ = t.Conn.SetReadDeadline(time.Now())
	}
}

func (t *Terminal) Continue() {
	<-t.Ready
	if t.Conn != nil {
		_ = t.Conn.SetReadDeadline(time.Time{})
	}
}

func StartTerminal() (*Terminal, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("creating terminal listener: %v", err)
	}
	log.Println("Started terminal connection at ", listener.Addr())

	t := &Terminal{
		Listener: listener,
		Port:     listener.Addr().(*net.TCPAddr).Port,
		Ready:    make(chan struct{}),
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic:", r)
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, false)
				log.Println(string(buf))
			}
		}()
		defer func() {
			close(t.Ready)
		}()
		defer func() {
			_ = listener.Close()
		}()

		c, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Errorf("accepting terminal listener: %v", err))
			return
		}
		conn := c.(*net.TCPConn)

		log.Println("Accepted terminal connection from", conn.RemoteAddr())
		t.Conn = conn
		_ = t.Conn.SetKeepAlive(true)
	}()

	return t, nil
}
