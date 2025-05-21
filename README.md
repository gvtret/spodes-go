[![Build Status](https://github.com/gvtret/spodes-go/actions/workflows/test.yml/badge.svg)](https://github.com/gvtret/spodes-go/actions)
[![Coverage Status](https://coveralls.io/repos/github/gvtret/spodes-go/badge.svg?branch=main)](https://coveralls.io/github/gvtret/spodes-go?branch=main)
![Go Version](https://img.shields.io/github/go-mod/go-version/gvtret/spodes-go)
# Spodes-Go: DLMS/COSEM Toolkit

`spodes-go` is a Go module providing tools for working with energy metering data exchange protocols, focusing on DLMS/COSEM standards, including A-XDR encoding/decoding and HDLC framing. This library is designed with a focus on correctness and performance, featuring optimizations to minimize reflection overhead and memory allocations.

## Project Structure

The library is organized into several packages:

-   `pkg/axdr/`: Implements the Abstract Syntax Description Rules (A-XDR) as per IEC 62056-6-2 for encoding and decoding data types used in DLMS/COSEM. It supports various primitive types, custom date/time structures, arrays, and complex structures.
-   `pkg/hdlc/`: Provides tools for High-Level Data Link Control (HDLC) framing, essential for communication in many metering systems. This includes frame encoding, decoding, and support for different frame types.
-   `pkg/cosem/`: (Planned) Will contain implementations for COSEM interface classes and object modeling.
-   `examples/`: Contains runnable examples demonstrating how to use the `axdr` and `hdlc` packages.

## Installation

```bash
go get github.com/gvtret/spodes-go
```

Then, import the necessary packages in your Go code:

```go
import (
    "github.com/gvtret/spodes-go/pkg/axdr"
    "github.com/gvtret/spodes-go/pkg/hdlc"
    // ... and so on
)
```

## Features

*   **A-XDR Encoding/Decoding:** Comprehensive support for various A-XDR data types including:
    *   Primitive types (boolean, integers, unsigned integers, floats)
    *   String types (octet-string, visible-string)
    *   Date, Time, and DateTime structures
    *   BitString and BCD (Binary Coded Decimal)
    *   Complex types: Array, Structure, CompactArray
*   **HDLC Framing:** Encoding and decoding of HDLC frames, supporting I-frames, S-frames, and U-frames (including UI, SNRM, UA, DISC, DM, FRMR).
*   **Performance:** Optimized for speed and low memory allocations, particularly in A-XDR encoding/decoding and HDLC bit stuffing/unstuffing routines.
*   **Modularity:** Designed with separate packages for distinct functionalities.

## Examples

Runnable examples are provided in the `examples/` directory.

### Running Examples

To run an example, navigate to the root of the repository and use the `go run` command:

```bash
# For the A-XDR example
go run ./examples/axdr_example/main.go

# For the HDLC example
go run ./examples/hdlc_example/main.go
```

### A-XDR Encoding Snippet

Here's a quick look at encoding a simple A-XDR structure:

```go
package main

import (
	"fmt"
	"log"
	"github.com/gvtret/spodes-go/pkg/axdr"
)

func main() {
	myStructure := axdr.Structure{
		int32(1024),             // A double-long integer
		"hello axdr",            // A visible-string
		true,                    // A boolean
		uint8(250),              // An unsigned integer
		axdr.Array{uint16(1), uint16(2), uint16(3)}, // An array of long-unsigned
	}

	encodedBytes, err := axdr.Encode(myStructure)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Printf("Encoded A-XDR: %X\n", encodedBytes)
    // Example Output: Encoded A-XDR: 02051E000004000A0A68656C6C6F206178647203011FFA0103200001200002200003
    // (Note: Actual byte output may vary based on specific values and library version details)

	decodedData, err := axdr.Decode(encodedBytes)
	if err != nil {
		log.Fatalf("Decoding failed: %v", err)
	}
	fmt.Printf("Decoded Data: %+v\n", decodedData)
}
```
*(For more detailed A-XDR and HDLC usage, please refer to the files in the `examples/` directory.)*

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs, feature requests, or improvements.

## License

This project is licensed under the MIT License. (A `LICENSE` file should be added to the repository if one doesn't exist).