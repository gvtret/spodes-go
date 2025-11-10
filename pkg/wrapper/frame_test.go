package wrapper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameEncode(t *testing.T) {
	f := &Frame{
		Version: Version,
		SrcAddr: 1,
		DstAddr: 2,
		Length:  5,
		Payload: []byte("hello"),
	}
	encoded, err := f.Encode()
	assert.NoError(t, err)

	expected := []byte{
		0x00, 0x01, // Version
		0x00, 0x01, // SrcAddr
		0x00, 0x02, // DstAddr
		0x00, 0x05, // Length
		'h', 'e', 'l', 'l', 'o',
	}
	assert.Equal(t, expected, encoded)
}

func TestFrameDecode(t *testing.T) {
	encoded := []byte{
		0x00, 0x01, // Version
		0x00, 0x01, // SrcAddr
		0x00, 0x02, // DstAddr
		0x00, 0x05, // Length
		'h', 'e', 'l', 'l', 'o',
	}
	f := &Frame{}
	err := f.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, uint16(Version), f.Version)
	assert.Equal(t, uint16(1), f.SrcAddr)
	assert.Equal(t, uint16(2), f.DstAddr)
	assert.Equal(t, uint16(5), f.Length)
	assert.Equal(t, []byte("hello"), f.Payload)
}

func TestFrameEncodeDecode(t *testing.T) {
	f1 := &Frame{
		Version: Version,
		SrcAddr: 10,
		DstAddr: 20,
		Length:  12,
		Payload: []byte("test payload"),
	}

	encoded, err := f1.Encode()
	assert.NoError(t, err)

	f2 := &Frame{}
	err = f2.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, f1.Version, f2.Version)
	assert.Equal(t, f1.SrcAddr, f2.SrcAddr)
	assert.Equal(t, f1.DstAddr, f2.DstAddr)
	assert.Equal(t, f1.Length, f2.Length)
	assert.True(t, bytes.Equal(f1.Payload, f2.Payload))
}

func TestFrameDecodeInvalid(t *testing.T) {
	// Insufficient data for header
	encoded1 := []byte{0x00, 0x01, 0x00, 0x01}
	f1 := &Frame{}
	err := f1.Decode(encoded1)
	assert.Error(t, err)

	// Insufficient data for payload
	encoded2 := []byte{
		0x00, 0x01, // Version
		0x00, 0x01, // SrcAddr
		0x00, 0x02, // DstAddr
		0x00, 0x05, // Length
		'h', 'e', 'l', 'l',
	}
	f2 := &Frame{}
	err = f2.Decode(encoded2)
	assert.Error(t, err)
}
