[![Build Status](https://github.com/gvtret/spodes-go/actions/workflows/test.yml/badge.svg)](https://github.com/gvtret/spodes-go/actions)
![Coveralls](https://img.shields.io/coverallsCoverage/github/gvtret/spodes-go)
# Spodes

A Go module implementing the СПОДЭС (Russian standard for energy metering data exchange) based on DLMS/COSEM.

## Structure

- `pkg/axdr/`: A-XDR encoding/decoding logic and types.
- `pkg/cosem/`: COSEM interface classes.
- `examples/`: Usage examples.

## Installation

```bash
go get github.com/gvtret/spodes-go