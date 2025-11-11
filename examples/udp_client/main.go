package main

import (
	"bytes"
	"log"
	"net"
	"time"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

func main() {
	serverAddr := "127.0.0.1:4060"
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Failed to connect to UDP server: %v", err)
	}
	defer conn.Close()
	log.Printf("UDP client connected to %s", serverAddr)

	config := hdlc.DefaultConfig()
	config.SrcAddr = []byte{0x02}  // Client address
	config.DestAddr = []byte{0x01} // Server address
	hdlcConn := hdlc.NewHDLCConnection(config)

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
	buf := make([]byte, 2048)
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
	largePDU := bytes.Repeat([]byte("udp test "), 20)
	frames, err := hdlcConn.Send(largePDU)
	if err != nil {
		log.Fatalf("Client failed to generate segmented I-frames: %v", err)
	}
	for _, frame := range frames {
		_, err = conn.Write(frame)
		if err != nil {
			log.Fatalf("Failed to write I-frame segment: %v", err)
		}
		// Small delay to allow server to process each frame
		time.Sleep(50 * time.Millisecond)
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
	time.Sleep(1 * time.Second)
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
