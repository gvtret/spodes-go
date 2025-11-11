package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/gvtret/spodes-go/pkg/cosem"
	"github.com/gvtret/spodes-go/pkg/hdlc"
	"github.com/gvtret/spodes-go/pkg/transport"
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
	config.SrcAddr = []byte{0x01}  // Server address
	config.DestAddr = []byte{0x02} // Client address
	hdlcConn := hdlc.NewHDLCConnection(config)

	app := setupApplication(hdlcConn)

	go func() {
		for {
			pdu, clientAddr, err := hdlcConn.Read()
			if err != nil {
				log.Printf("Error reading PDU: %v", err)
				return
			}
			log.Printf("Server received PDU from %s: %x", clientAddr, pdu)

			responsePDU, err := app.HandleAPDU(pdu, clientAddr)
			if err != nil {
				log.Printf("Error handling APDU: %v", err)
				continue
			}

			if responsePDU != nil {
				frames, err := hdlcConn.Send(responsePDU)
				if err != nil {
					log.Printf("Error sending response: %v", err)
					continue
				}
				for _, frame := range frames {
					log.Printf("Server sending response frame: %x", frame)
					_, err := conn.Write(frame)
					if err != nil {
						log.Printf("Error writing to connection: %v", err)
						return
					}
				}
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		log.Printf("Server received raw data: %x", buf[:n])
		hdlcConn.Receive(buf[:n])
	}
}

func handleWrapperConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Accepted WRAPPER connection from %s", conn.RemoteAddr())

	wrapperConn := wrapper.NewConnection(conn, nil)
	app := setupApplication(wrapperConn)

	go func() {
		for {
			pdu, clientAddr, err := wrapperConn.Read()
			if err != nil {
				log.Printf("Error reading PDU: %v", err)
				return
			}
			log.Printf("Server received PDU from %s: %x", clientAddr, pdu)

			responsePDU, err := app.HandleAPDU(pdu, clientAddr)
			if err != nil {
				log.Printf("Error handling APDU: %v", err)
				continue
			}

			if responsePDU != nil {
				frames, err := wrapperConn.Send(responsePDU)
				if err != nil {
					log.Printf("Error sending response: %v", err)
					continue
				}
				log.Printf("Server sending response frame: %x", frames[0])
				_, err = conn.Write(frames[0])
				if err != nil {
					log.Printf("Error sending wrapper frame: %v", err)
					return
				}
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}
		log.Printf("Server received raw data: %x", buf[:n])
		wrapperConn.Receive(buf[:n])
	}
}

func setupApplication(tp transport.Transport) *cosem.Application {
	// 1. Initialize SecuritySetup
	obisSecurity, _ := cosem.NewObisCodeFromString("0.0.43.0.0.255")
	securitySetup, _ := cosem.NewSecuritySetup(*obisSecurity, nil, nil, nil, nil, nil)

	// 2. Create the Application instance
	app := cosem.NewApplication(tp, securitySetup)

	// 3. Create COSEM objects
	obisClock, _ := cosem.NewObisCodeFromString("0.0.1.0.0.255")
	clockObj, _ := cosem.NewClock(*obisClock)
	app.RegisterObject(clockObj)

	obisData, _ := cosem.NewObisCodeFromString("1.0.0.3.0.255")
	dataObj, _ := cosem.NewData(*obisData, uint32(12345))
	app.RegisterObject(dataObj)

	// 4. Create Associations for different clients
	// Public Client (address 0x10)
	addrPub, _ := cosem.NewObisCodeFromString("0.0.40.0.0.255")
	assocPub, _ := cosem.NewAssociationLN(*addrPub)
	app.AddAssociation("10", assocPub)
	app.PopulateObjectList(assocPub, []cosem.ObisCode{*obisClock}) // Only clock is public

	// Private/Admin Client (address 0x20)
	addrPriv, _ := cosem.NewObisCodeFromString("0.0.40.0.1.255")
	assocPriv, _ := cosem.NewAssociationLN(*addrPriv)
	app.AddAssociation("20", assocPriv)
	app.PopulateObjectList(assocPriv, []cosem.ObisCode{*obisClock, *obisData}) // Both objects are visible

	return app
}
