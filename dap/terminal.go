package dap

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

type Terminal struct {
	Listener net.Listener
	Port     int

	Input  <-chan string
	Output chan<- string
	Done   chan struct{}
}

func (t *Terminal) Close() {
	_ = t.Listener.Close()
	close(t.Done)
}

func (t *Terminal) Write(line string) {
	select {
	case t.Output <- line:
	case <-t.Done:
	}
}

func StartTerminal() (*Terminal, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("creating terminal listener: %v", err)
	}
	log.Println("Started terminal connection at ", listener.Addr())

	input := make(chan string)
	output := make(chan string)
	t := &Terminal{
		Listener: listener,
		Port:     listener.Addr().(*net.TCPAddr).Port,

		Input:  input,
		Output: output,
		Done:   make(chan struct{}),
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
			_ = listener.Close()
		}()

		c, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Errorf("accepting terminal listener: %v", err))
			return
		}
		_ = listener.Close() // only accept a single connection

		conn := c.(*net.TCPConn)
		_ = conn.SetKeepAlive(true)

		log.Println("Accepted terminal connection from", conn.RemoteAddr())

		// input loop
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(input)
			scan := bufio.NewScanner(conn)
			for scan.Scan() {
				select {
				case input <- scan.Text():
				case <-t.Done:
					// try to friendly drain and close the conn for the client
					_ = conn.CloseWrite()
					_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
					_, _ = io.Copy(io.Discard, conn)
					_ = conn.Close()
					return
				}
			}
		}()

		// output loop
		for {
			select {
			case <-t.Done:
				_ = conn.CloseWrite()
			case line, ok := <-output:
				if !ok {
					return
				}
				_, _ = conn.Write([]byte(line))
			}
		}
	}()

	return t, nil
}
