package cosem

import (
	"fmt"
	"strconv"
	"strings"
)

// ObisCode представляет OBIS-код.
type ObisCode struct {
	bValue [6]byte
	sValue string
}

func CreateObisFromBytes(value [6]byte) (*ObisCode) {
	oc := &ObisCode{}
	oc.SetFromBytes(value)
	return oc
}

func (oc *ObisCode) SetFromBytes(value [6]byte) {
	var parts []string

	for _, b := range value {
		parts = append(parts, strconv.Itoa(int(b)))
	}

	oc.bValue = value
	oc.sValue = strings.Join(parts, ".")
}

func CreateObisFromString(value string) (*ObisCode, error) {
	oc := &ObisCode{}
	err := oc.SetFromString(value)

	if err != nil {
		return nil, err
	}

	return oc, nil
}

// ParseObisCode парсит строку в ObisCode.
func (oc *ObisCode) SetFromString(value string) error {
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, ":", ".")
	value = strings.ReplaceAll(value, "-", ".")
	chunks := strings.Split(value, ".")

	if len(chunks) != 6 {
		return fmt.Errorf("invalid OBIS code format: %s. Must have exactly 6 parts", value)
	}

	var bValue [6]byte

	for i, chunk := range chunks {
		b, err := strconv.ParseUint(chunk, 10, 8)
		if err != nil {
			return fmt.Errorf("failed to parse chunk as byte for %w", err)
		}

		bValue[i] = byte(b)
	}

	oc.sValue = value
	oc.bValue = bValue
	return nil
}

// String возвращает строковое представление OBIS-кода с разделителями ".".
func (oc ObisCode) String() string {
	return oc.sValue
}

// Bytes возвращает массив из 6 байт, представляющих OBIS-код.
func (oc ObisCode) Bytes() [6]byte {
	return oc.bValue
}