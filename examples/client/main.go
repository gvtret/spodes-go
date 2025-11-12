package main

import (
	"bytes"
	"flag"
	"log"
	"net"
	"time"

	"github.com/gvtret/spodes-go/pkg/hdlc"
	"github.com/gvtret/spodes-go/pkg/wrapper"
)

func main() {
	transport := flag.String("transport", "hdlc", "Transport layer to use: 'hdlc' or 'wrapper'")
	flag.Parse()

	serverAddr := "127.0.0.1:4059"
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()
	log.Printf("Connected to %s server at %s", *transport, serverAddr)

	switch *transport {
	case "hdlc":
		runHDLCClient(conn)
	case "wrapper":
		runWrapperClient(conn)
	default:
		log.Fatalf("Invalid transport layer specified: %s", *transport)
	}
}

func runHDLCClient(conn net.Conn) {
	config := hdlc.DefaultConfig()
	config.SrcAddr = []byte{0x02}  // Client address
	config.DestAddr = []byte{0x01} // Server address
	hdlcConn := hdlc.NewHDLCConnection(config)

	go func() {
		for {
			_, _, err := hdlcConn.Read() // In client, we don't process unsolicited PDUs
			if err != nil {
				return
			}
		}
	}()

	// 1. Send SNRM to connect
	log.Println("Client sending: SNRM")
	snrmFrame, err := hdlcConn.Connect()
	if err != nil {
		log.Fatalf("Client failed to generate SNRM: %v", err)
	}
	_, err = conn.Write(snrmFrame)
	if err != nil {
		log.Fatalf("Failed to write SNRM: %v", err)
	}

	// 2. Wait for UA response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	log.Printf("Client received %d bytes: %x", n, buf[:n])
	_, err = hdlcConn.Receive(buf[:n])
	if err != nil {
		log.Fatalf("Client error handling UA response: %v", err)
	}
	if !hdlcConn.IsConnected() {
		log.Fatalf("Client failed to connect.")
	}
	log.Println("Client is connected.")

	// 3. Send a large, segmented I-frame PDU
	log.Println("Client sending: Large segmented PDU")
	largePDU := bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 10) // 260 bytes > 128 byte maxFrameSize
	frames, err := hdlcConn.Send(largePDU)
	if err != nil {
		log.Fatalf("Client failed to generate segmented I-frames: %v", err)
	}
	for _, frame := range frames {
		_, err = conn.Write(frame)
		if err != nil {
			log.Fatalf("Failed to write I-frame segment: %v", err)
		}
	}

	// 4. Wait for the final RR response
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	log.Printf("Client received %d bytes: %x", n, buf[:n])
	_, err = hdlcConn.Receive(buf[:n])
	if err != nil {
		log.Fatalf("Client error handling RR response: %v", err)
	}
	log.Println("Client received RR acknowledgment.")

	// 5. Send DISC to disconnect
	log.Println("Client sending: DISC")
	time.Sleep(1 * time.Second) // Give a moment before disconnecting
	discFrame, err := hdlcConn.Disconnect()
	if err != nil {
		log.Fatalf("Client failed to generate DISC: %v", err)
	}
	_, err = conn.Write(discFrame)
	if err != nil {
		log.Fatalf("Failed to write DISC: %v", err)
	}

	// 6. Wait for UA response
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	log.Printf("Client received %d bytes: %x", n, buf[:n])
	_, err = hdlcConn.Receive(buf[:n])
	if err != nil {
		log.Fatalf("Client error handling UA for DISC: %v", err)
	}
	if hdlcConn.IsConnected() {
		log.Fatalf("Client failed to disconnect.")
	}
	log.Println("Client is disconnected.")
}

func runWrapperClient(conn net.Conn) {
	wrapperConn := wrapper.NewConnection(conn, nil)

	// Send a message
	payload := []byte("hello, world!")
	frames, err := wrapperConn.Send(payload)
	if err != nil {
		log.Fatalf("Failed to create frame: %v", err)
	}

	log.Printf("Client sending frame: %x", frames[0])
	_, err = conn.Write(frames[0])
	if err != nil {
		log.Fatalf("Failed to send frame: %v", err)
	}

	// Receive a response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	log.Printf("Client received %d bytes: %x", n, buf[:n])
	_, err = wrapperConn.Receive(buf[:n])
	if err != nil {
		log.Fatalf("Client error handling response: %v", err)
	}
	respPDU, _, err := wrapperConn.Read()
	if err != nil {
		log.Fatalf("Failed to read PDU: %v", err)
	}
	log.Printf("Client received PDU: %s", string(respPDU))

	time.Sleep(1 * time.Second)
}
