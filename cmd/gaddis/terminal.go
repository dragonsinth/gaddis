package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func terminalCmd(port int) error {
	if port <= 0 {
		return errors.New("missing/invalid port")
	}

	var conn *net.TCPConn
	{
		c, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			return fmt.Errorf("could not connect to port %d: %v", port, err)
		}
		conn = c.(*net.TCPConn)
	}
	_ = conn.SetKeepAlive(true)
	defer func() {
		_ = conn.Close()
	}()
	log.Println("connected: ", conn.RemoteAddr().String())

	// stdin on background
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			log.Println("stdin err", err)
		} else {
			log.Println("closing stdin")
			_ = conn.CloseWrite()
		}
	}()

	if _, err := io.Copy(os.Stdout, conn); err != nil {
		return fmt.Errorf("stdout err: %v", err)
	}
	return nil
}
