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
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("Failed to close log file: %v", err)
		}
	}()
	log.SetOutput(logFile)

	listenAddr := "127.0.0.1:4059"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Failed to close listener: %v", err)
		}
	}()
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
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()
	log.Printf("Accepted HDLC connection from %s", conn.RemoteAddr())

	config := hdlc.DefaultConfig()
	config.SrcAddr = []byte{0x01}  // Server address
	config.DestAddr = []byte{0x02} // Client address
	hdlcConn := hdlc.NewHDLCConnection(config)

	app, err := setupApplication(hdlcConn)
	if err != nil {
		log.Printf("Failed to set up application: %v", err)
		return
	}

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
		responses, err := hdlcConn.Receive(buf[:n])
		if err != nil {
			log.Printf("Error handling HDLC data: %v", err)
			return
		}
		for _, resp := range responses {
			log.Printf("Server sending frame: %x", resp)
			if _, err := conn.Write(resp); err != nil {
				log.Printf("Error writing HDLC response: %v", err)
				return
			}
		}
	}
}

func handleWrapperConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()
	log.Printf("Accepted WRAPPER connection from %s", conn.RemoteAddr())

	wrapperConn := wrapper.NewConnection(conn, nil)
	app, err := setupApplication(wrapperConn)
	if err != nil {
		log.Printf("Failed to set up application: %v", err)
		return
	}

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
		responses, err := wrapperConn.Receive(buf[:n])
		if err != nil {
			log.Printf("Error handling WRAPPER data: %v", err)
			return
		}
		for _, frame := range responses {
			log.Printf("Server sending frame: %x", frame)
			if _, err := conn.Write(frame); err != nil {
				log.Printf("Error writing WRAPPER response: %v", err)
				return
			}
		}
	}
}

func setupApplication(tp transport.Transport) (*cosem.Application, error) {
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
	if err := app.PopulateObjectList(assocPub, []cosem.ObisCode{*obisClock}); err != nil {
		return nil, err
	} // Only clock is public

	// Private/Admin Client (address 0x20)
	addrPriv, _ := cosem.NewObisCodeFromString("0.0.40.0.1.255")
	assocPriv, _ := cosem.NewAssociationLN(*addrPriv)
	app.AddAssociation("20", assocPriv)
	if err := app.PopulateObjectList(assocPriv, []cosem.ObisCode{*obisClock, *obisData}); err != nil {
		return nil, err
	} // Both objects are visible

	return app, nil
}
