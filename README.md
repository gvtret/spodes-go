# Spodes

A Go module implementing the СПОДЭС (Russian standard for energy metering data exchange) based on DLMS/COSEM.

## Structure

- `internal/ber/`: BER encoding/decoding logic.
- `pkg/spodes/`: COSEM interface classes (`Data`, `Register`) and common interface (`SpodesObject`).
- `pkg/examples/`: Usage examples for `Data` and `Register`.

## Installation

```bash
go get github.com/gvtret/spodes-go