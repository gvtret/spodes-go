package main

import (
	"log"
	"net"
	"os"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

func main() {
	// Configure logging to a file
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	listenAddr := "127.0.0.1:4059"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}
	defer listener.Close()
	log.Printf("HDLC server listening on %s", listenAddr)

	time.Sleep(5 * time.Second)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)

			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted connection from %s", conn.RemoteAddr())

	hdlcConn := hdlc.NewHDLCConnection(nil) // Use default config
	// Server address is 0x01, Client is 0x02
	hdlcConn.SetAddress([]byte{0x01}, []byte{0x02})

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		log.Printf("Server received %d bytes: %x", n, buf[:n])

		responses, err := hdlcConn.Handle(buf[:n])
		if err != nil {
			log.Printf("Error handling HDLC frame: %v", err)
			// Decide if we should close the connection based on the error type
			if hdlcErr, ok := err.(*hdlc.HDLCError); ok && hdlcErr.ShouldExit {
				return
			}
			continue
		}

		for _, resp := range responses {
			log.Printf("Server sending response: %x", resp)
			_, err := conn.Write(resp)
			if err != nil {
				log.Printf("Error writing to connection: %v", err)
				return
			}
		}

		// After handling raw frames, check for reassembled PDUs
		for len(hdlcConn.ReassembledData) > 0 {
			pdu, err := hdlcConn.Read()
			if err != nil {
				log.Printf("Error reading PDU: %v", err)
				break // Or continue, depending on desired error handling
			}
			log.Printf("Server received reassembled PDU: %s", string(pdu))
			// Here, the application layer would process the PDU
		}
	}
}
