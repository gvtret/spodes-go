package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/gvtret/spodes-go/pkg/hdlc"
	"github.com/gvtret/spodes-go/pkg/wrapper"
)

func main() {
	transport := flag.String("transport", "hdlc", "Transport layer to use: 'hdlc' or 'wrapper'")
	flag.Parse()

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
	log.Printf("%s server listening on %s", *transport, listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		switch *transport {
		case "hdlc":
			go handleHDLCConnection(conn)
		case "wrapper":
			go handleWrapperConnection(conn)
		default:
			log.Fatalf("Invalid transport layer specified: %s", *transport)
		}
	}
}

func handleHDLCConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted HDLC connection from %s", conn.RemoteAddr())

	config := hdlc.DefaultConfig()
	config.SrcAddr = []byte{0x01} // Server address
	config.DestAddr = []byte{0x02} // Client address
	hdlcConn := hdlc.NewHDLCConnection(config)

	go func() {
		for frame := range hdlcConn.RetransmitFrames {
			log.Printf("Server retransmitting frame: %x", frame)
			conn.Write(frame)
		}
	}()

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

		for len(hdlcConn.ReassembledData) > 0 {
			pdu, err := hdlcConn.Read()
			if err != nil {
				log.Printf("Error reading PDU: %v", err)
				break
			}
			log.Printf("Server received reassembled PDU: %s", string(pdu))
		}
	}
}

func handleWrapperConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted WRAPPER connection from %s", conn.RemoteAddr())

	wrapperConn := wrapper.NewConn(conn)

	for {
		frame, err := wrapperConn.Receive()
		if err != nil {
			log.Printf("Error receiving wrapper frame: %v", err)
			return
		}

		log.Printf("Server received wrapper frame: %+v", frame)

		// Echo the payload back
		respFrame := &wrapper.Frame{
			Version: wrapper.Version,
			SrcAddr: frame.DstAddr,
			DstAddr: frame.SrcAddr,
			Length:  frame.Length,
			Payload: frame.Payload,
		}

		log.Printf("Server sending response frame: %+v", respFrame)
		err = wrapperConn.Send(respFrame)
		if err != nil {
			log.Printf("Error sending wrapper frame: %v", err)
			return
		}
	}
}
