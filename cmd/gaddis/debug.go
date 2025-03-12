package main

import (
	"bufio"
	"fmt"
	"github.com/dragonsinth/gaddis/dap"
	"io"
	"log"
	"net"
	"os"
	"runtime"
)

// server starts a server that listens on a specified port
// and blocks indefinitely. This server can accept multiple
// client connections at the same time.
func debugServer(port int, dbgLog *log.Logger) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()
	log.Println("Started server at ", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection failed:", err)
			continue
		}
		log.Println("Accepted connection from", conn.RemoteAddr())

		// Handle multiple client connections concurrently
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
				log.Println("Closing connection from", conn.RemoteAddr())
				_ = conn.Close()
			}()

			rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			session := dap.NewSession(rw, dbgLog)
			if err := session.Run(); err != nil {
				log.Println("Error:", err)
			}
		}()
	}
}

func debugCmd(port int, verbose bool) error {
	dbgLog := log.New(io.Discard, "", log.LstdFlags)

	if port < 0 {
		// we have to run on stdout, so create a log file.
		_ = os.Chdir(os.Getenv("PWD"))
		logFile, err := os.OpenFile("gaddis-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			log.SetOutput(logFile)
			defer func() {
				_ = logFile.Close()
			}()
		}
		if verbose {
			dbgLog.SetOutput(logFile)
		}
		log.SetOutput(logFile)
	} else {
		// setup normal stdout/stderr logging
		if verbose {
			dbgLog.SetOutput(os.Stdout)
		}
		log.SetOutput(os.Stderr)

		// if running as a server within vscode, don't emit timestamps (IDE will do this).
		if os.Getenv("VSCODE_CLI") != "" {
			dbgLog.SetFlags(0)
			log.SetFlags(0)
		}
	}

	if verbose {
		dbgLog.Println(os.Getwd())
		dbgLog.Println(os.Args)
		for _, ev := range os.Environ() {
			dbgLog.Println(ev)
		}
	}

	if port >= 0 {
		return debugServer(port, dbgLog)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
		session := dap.NewSession(rw, dbgLog)
		if err := session.Run(); err != nil {
			log.Println("Error:", err)
		}
		return session.Run()
	}
}
