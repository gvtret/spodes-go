# HDLC TCP Examples

This directory contains simple TCP client and server examples that demonstrate how to use the transport-agnostic HDLC package.

## Server

The server listens for incoming TCP connections on port `4059`. For each connection, it creates an HDLC connection and processes incoming frames.

To run the server:
```sh
go run examples/server/main.go
```

## Client

The client connects to the server on port `4059`, establishes an HDLC connection, sends a large, segmented PDU, and then disconnects.

To run the client:
```sh
go run examples/client/main.go
```
