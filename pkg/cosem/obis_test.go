package cosem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObisCodeCreate(t *testing.T) {
	str_oc := "1.0.0.0.0.255"
	oc, err := CreateObisFromString(str_oc)
	assert.NoError(t, err)
	assert.Equal(t, oc.Bytes(), [6]byte{1, 0, 0, 0, 0, 255})
	assert.Equal(t, oc.String(), str_oc)

	oc, err = CreateObisFromString(str_oc)
	assert.NoError(t, err)
	assert.Equal(t, oc.Bytes(), [6]byte{1, 0, 0, 0, 0, 255})
	assert.Equal(t, oc.String(), str_oc)

	oc, err = CreateObisFromString("1:0:0:0:0:255")
	assert.NoError(t, err)
	assert.Equal(t, oc.Bytes(), [6]byte{1, 0, 0, 0, 0, 255})
	assert.Equal(t, oc.String(), str_oc)

	oc, err = CreateObisFromString("1 : 0 -  0  . 0  :  0.255")
	assert.NoError(t, err)
	assert.Equal(t, oc.Bytes(), [6]byte{1, 0, 0, 0, 0, 255})
	assert.Equal(t, oc.String(), str_oc)
}

func TestObisCodeInvalid(t *testing.T) {
	_, err := CreateObisFromString("1.2.3.4.5")
	assert.Error(t, err)

	_, err = CreateObisFromString("1.2.3.4.5.")
	assert.Error(t, err)

	_, err = CreateObisFromString("1.2.3.4.5.6.7")
	assert.Error(t, err)

	_, err = CreateObisFromString("1.2.3.4.5.A")
	assert.Error(t, err)

	_, err = CreateObisFromString("1.2.3.4.5.256")
	assert.Error(t, err)
}