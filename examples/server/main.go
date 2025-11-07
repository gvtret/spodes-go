package main

import (
	"fmt"
	"net"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:4059")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("HDLC Server listening on localhost:4059")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	hdlcConn := hdlc.NewHDLCConnection()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		responses, err := hdlcConn.Handle(buf[:n])
		if err != nil {
			return
		}

		for _, resp := range responses {
			conn.Write(resp)
		}
	}
}
