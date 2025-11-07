# Examples

This directory contains examples of how to use the transport-agnostic HDLC package with a TCP transport layer.

## Server

To run the server, navigate to the `server` directory and run:

```
go run .
```

The server will listen on `localhost:8080`.

## Client

To run the client, navigate to the `client` directory and run:

```
go run .
```

The client will connect to the server, send a test message, and print the server's response.
