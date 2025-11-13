package cosem

import (
	"errors"
	"reflect"
	"testing"
)

func newTestProfileGeneric(t *testing.T) *ProfileGeneric {
	t.Helper()

	obis, err := NewObisCodeFromString("1.0.0.2.0.255")
	if err != nil {
		t.Fatalf("failed to create obis code: %v", err)
	}

	pg, err := NewProfileGeneric(*obis, [][]byte{}, []CaptureObjectDefinition{}, 60, 0, CosemAttributeDescriptor{})
	if err != nil {
		t.Fatalf("failed to create profile generic: %v", err)
	}

	return pg
}

func TestProfileGenericSetAttributeCapturePeriod(t *testing.T) {
	pg := newTestProfileGeneric(t)

	newPeriod := uint32(120)
	if err := pg.SetAttribute(4, newPeriod); err != nil {
		t.Fatalf("expected capture period update to succeed, got error: %v", err)
	}

	value, err := pg.GetAttribute(4)
	if err != nil {
		t.Fatalf("failed to retrieve capture period: %v", err)
	}
	if got := value.(uint32); got != newPeriod {
		t.Fatalf("unexpected capture period value: got %d, want %d", got, newPeriod)
	}

	if err := pg.SetAttribute(4, uint32(0)); err == nil || !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid capture period error, got: %v", err)
	}

	value, err = pg.GetAttribute(4)
	if err != nil {
		t.Fatalf("failed to retrieve capture period after invalid update: %v", err)
	}
	if got := value.(uint32); got != newPeriod {
		t.Fatalf("capture period changed after invalid update: got %d, want %d", got, newPeriod)
	}
}

func TestProfileGenericSetAttributeSortMethod(t *testing.T) {
	pg := newTestProfileGeneric(t)

	newMethod := uint8(2)
	if err := pg.SetAttribute(5, newMethod); err != nil {
		t.Fatalf("expected sort method update to succeed, got error: %v", err)
	}

	value, err := pg.GetAttribute(5)
	if err != nil {
		t.Fatalf("failed to retrieve sort method: %v", err)
	}
	if got := value.(uint8); got != newMethod {
		t.Fatalf("unexpected sort method value: got %d, want %d", got, newMethod)
	}

	if err := pg.SetAttribute(5, uint8(10)); err == nil || !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid sort method error, got: %v", err)
	}

	value, err = pg.GetAttribute(5)
	if err != nil {
		t.Fatalf("failed to retrieve sort method after invalid update: %v", err)
	}
	if got := value.(uint8); got != newMethod {
		t.Fatalf("sort method changed after invalid update: got %d, want %d", got, newMethod)
	}
}

func TestProfileGenericSetAttributeSortObject(t *testing.T) {
	pg := newTestProfileGeneric(t)

	sortTarget, err := NewObisCodeFromString("1.0.1.8.0.255")
	if err != nil {
		t.Fatalf("failed to create sort target obis code: %v", err)
	}

	validSortObject := CosemAttributeDescriptor{
		ClassID:     RegisterClassID,
		InstanceID:  *sortTarget,
		AttributeID: 2,
	}

	if err := pg.SetAttribute(5, uint8(1)); err != nil {
		t.Fatalf("failed to set sort method prerequisite: %v", err)
	}

	if err := pg.SetAttribute(6, validSortObject); err != nil {
		t.Fatalf("expected sort object update to succeed, got error: %v", err)
	}

	value, err := pg.GetAttribute(6)
	if err != nil {
		t.Fatalf("failed to retrieve sort object: %v", err)
	}

	if got := value.(CosemAttributeDescriptor); !reflect.DeepEqual(got, validSortObject) {
		t.Fatalf("unexpected sort object value: got %+v, want %+v", got, validSortObject)
	}

	invalidSortObject := CosemAttributeDescriptor{
		ClassID:     0,
		InstanceID:  *sortTarget,
		AttributeID: 2,
	}

	if err := pg.SetAttribute(6, invalidSortObject); err == nil || !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid sort object error, got: %v", err)
	}

	value, err = pg.GetAttribute(6)
	if err != nil {
		t.Fatalf("failed to retrieve sort object after invalid update: %v", err)
	}

	if got := value.(CosemAttributeDescriptor); !reflect.DeepEqual(got, validSortObject) {
		t.Fatalf("sort object changed after invalid update: got %+v, want %+v", got, validSortObject)
	}
}

func TestProfileGenericResetMethod(t *testing.T) {
	pg := newTestProfileGeneric(t)

	first := []byte{0x01}
	second := []byte{0x02}

	if _, err := pg.Invoke(2, []interface{}{first}); err != nil {
		t.Fatalf("capture invocation failed: %v", err)
	}
	if _, err := pg.Invoke(2, []interface{}{second}); err != nil {
		t.Fatalf("capture invocation failed: %v", err)
	}

	if _, err := pg.Invoke(1, nil); err != nil {
		t.Fatalf("reset invocation failed: %v", err)
	}

	bufferValue, err := pg.GetAttribute(2)
	if err != nil {
		t.Fatalf("failed to get buffer: %v", err)
	}
	records := bufferValue.([][]byte)
	if len(records) != 0 {
		t.Fatalf("expected empty buffer after reset, got %d records", len(records))
	}

	entriesValue, err := pg.GetAttribute(7)
	if err != nil {
		t.Fatalf("failed to get entries_in_use: %v", err)
	}
	if entries := entriesValue.(uint32); entries != 0 {
		t.Fatalf("expected entries_in_use to be 0 after reset, got %d", entries)
	}

	if _, err := pg.Invoke(1, []interface{}{uint8(1)}); !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid parameter error for reset with arguments, got %v", err)
	}
}

func TestProfileGenericCaptureMethod(t *testing.T) {
	pg := newTestProfileGeneric(t)

	first := []byte{0x01, 0x02}
	second := []byte{0x03, 0x04}

	if _, err := pg.Invoke(2, []interface{}{first}); err != nil {
		t.Fatalf("capture invocation failed: %v", err)
	}
	if _, err := pg.Invoke(2, []interface{}{second}); err != nil {
		t.Fatalf("capture invocation failed: %v", err)
	}

	bufferValue, err := pg.GetAttribute(2)
	if err != nil {
		t.Fatalf("failed to get buffer: %v", err)
	}
	records := bufferValue.([][]byte)
	if len(records) != 2 {
		t.Fatalf("unexpected number of records, got %d want 2", len(records))
	}
	if got, want := records[0], first; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected first record: got %v want %v", got, want)
	}
	if got, want := records[1], second; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected second record: got %v want %v", got, want)
	}

	entriesValue, err := pg.GetAttribute(7)
	if err != nil {
		t.Fatalf("failed to get entries_in_use: %v", err)
	}
	if entries := entriesValue.(uint32); entries != 2 {
		t.Fatalf("unexpected entries_in_use value: got %d want 2", entries)
	}

	if _, err := pg.Invoke(2, nil); !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid parameter error for capture without arguments, got %v", err)
	}
	if _, err := pg.Invoke(2, []interface{}{uint32(5)}); !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected invalid parameter error for capture with wrong argument type, got %v", err)
	}
}
