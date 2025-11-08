package main

import (
	"log"
	"net"
	"sync"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

// activeConnections stores the HDLC connection for each client address.
var activeConnections = make(map[string]*hdlc.HDLCConnection)
var mutex = &sync.Mutex{}

func main() {
	listenAddr := "127.0.0.1:4060"
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP: %v", err)
	}
	defer conn.Close()
	log.Printf("UDP HDLC server listening on %s", listenAddr)

	buf := make([]byte, 2048)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go handlePacket(conn, remoteAddr, buf[:n])
	}
}

func handlePacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte) {
	mutex.Lock()
	hdlcConn, exists := activeConnections[remoteAddr.String()]
	if !exists {
		hdlcConn = hdlc.NewHDLCConnection(nil)
		// Assuming server address 0x01, client can be anything.
		// In a real scenario, the address would be part of the initial handshake.
		hdlcConn.SetAddress([]byte{0x01}, []byte{0x02})
		activeConnections[remoteAddr.String()] = hdlcConn
		log.Printf("New client connection from %s", remoteAddr.String())
	}
	mutex.Unlock()

	log.Printf("Server received %d bytes from %s: %x", len(data), remoteAddr.String(), data)

	responses, err := hdlcConn.Handle(data)
	if err != nil {
		log.Printf("Error handling HDLC data from %s: %v", remoteAddr.String(), err)
		return
	}

	for _, resp := range responses {
		log.Printf("Server sending response to %s: %x", remoteAddr.String(), resp)
		_, err := conn.WriteToUDP(resp, remoteAddr)
		if err != nil {
			log.Printf("Error writing to UDP for %s: %v", remoteAddr.String(), err)
		}
	}
}
