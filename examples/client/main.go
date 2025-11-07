package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gvtret/spodes-go/pkg/hdlc"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:4059")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	client := hdlc.NewHDLCConnection()
	da, sa := []byte{0x01}, []byte{0x02}
	client.SetAddress(da, sa)

	// Send SNRM
	snrmFrame, err := client.Connect()
	if err != nil {
		fmt.Printf("Failed to create SNRM frame: %v\n", err)
		return
	}
	conn.Write(snrmFrame)

	// Wait for UA
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from server: %v\n", err)
		return
	}
	_, err = client.HandleFrame(buf[:n])
	if err != nil {
		fmt.Printf("Failed to handle UA frame: %v\n", err)
		return
	}

	fmt.Println("HDLC connection established.")
	time.Sleep(1 * time.Second)

	// Send DISC
	discFrame, err := client.Disconnect()
	if err != nil {
		fmt.Printf("Failed to create DISC frame: %v\n", err)
		return
	}
	conn.Write(discFrame)

	// Wait for UA
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Printf("Failed to read from server: %v\n", err)
		return
	}
	_, err = client.HandleFrame(buf[:n])
	if err != nil {
		fmt.Printf("Failed to handle UA frame: %v\n", err)
		return
	}

	fmt.Println("HDLC connection disconnected.")
}
