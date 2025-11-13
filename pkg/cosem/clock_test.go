package cosem

import (
	"testing"
	"time"
)

func TestClockAdjustMethods(t *testing.T) {
	obis, err := NewObisCodeFromString("0.0.1.0.0.255")
	if err != nil {
		t.Fatalf("failed to create OBIS: %v", err)
	}

	clock, err := NewClock(*obis)
	if err != nil {
		t.Fatalf("failed to create clock: %v", err)
	}

	baseTime := time.Date(2025, time.January, 2, 10, 7, 45, 0, time.UTC)
	if err := clock.SetAttribute(2, baseTime); err != nil {
		t.Fatalf("set time: %v", err)
	}

	if _, err := clock.Invoke(3, nil); err != nil {
		t.Fatalf("adjust_to_minute: %v", err)
	}
	got, _ := clock.GetAttribute(2)
	if want := time.Date(2025, time.January, 2, 10, 8, 0, 0, time.UTC); !got.(time.Time).Equal(want) {
		t.Fatalf("adjust_to_minute: want %v got %v", want, got)
	}

	if err := clock.SetAttribute(2, time.Date(2025, time.January, 2, 10, 7, 30, 0, time.UTC)); err != nil {
		t.Fatalf("reset time: %v", err)
	}
	if _, err := clock.Invoke(1, nil); err != nil {
		t.Fatalf("adjust_to_quarter_hour: %v", err)
	}
	got, _ = clock.GetAttribute(2)
	if want := time.Date(2025, time.January, 2, 10, 15, 0, 0, time.UTC); !got.(time.Time).Equal(want) {
		t.Fatalf("adjust_to_quarter_hour: want %v got %v", want, got)
	}

	if err := clock.SetAttribute(2, time.Date(2025, time.January, 2, 10, 7, 0, 0, time.UTC)); err != nil {
		t.Fatalf("reset time: %v", err)
	}
	if _, err := clock.Invoke(2, []interface{}{uint16(20)}); err != nil {
		t.Fatalf("adjust_to_measuring_period: %v", err)
	}
	got, _ = clock.GetAttribute(2)
	if want := time.Date(2025, time.January, 2, 10, 0, 0, 0, time.UTC); !got.(time.Time).Equal(want) {
		t.Fatalf("adjust_to_measuring_period: want %v got %v", want, got)
	}

	if _, err := clock.Invoke(2, []interface{}{uint16(0)}); err != ErrInvalidParameter {
		t.Fatalf("expected ErrInvalidParameter for zero period, got %v", err)
	}

	status, _ := clock.GetAttribute(4)
	if status.(uint8)&clockStatusInvalid != 0 {
		t.Fatalf("status invalid flag should be cleared")
	}
}

func TestClockPresetAndShift(t *testing.T) {
	obis, err := NewObisCodeFromString("0.0.1.0.0.255")
	if err != nil {
		t.Fatalf("failed to create OBIS: %v", err)
	}

	clock, err := NewClock(*obis)
	if err != nil {
		t.Fatalf("failed to create clock: %v", err)
	}

	start := time.Date(2025, time.March, 3, 9, 30, 0, 0, time.UTC)
	if err := clock.SetAttribute(2, start); err != nil {
		t.Fatalf("set time: %v", err)
	}

	preset := time.Date(2025, time.March, 3, 10, 0, 0, 0, time.UTC)
	if _, err := clock.Invoke(4, []interface{}{preset}); err != nil {
		t.Fatalf("preset_adjusting_time: %v", err)
	}
	if clock.presetAdjustTime == nil || !clock.presetAdjustTime.Equal(preset) {
		t.Fatalf("preset time not stored correctly")
	}
	status, _ := clock.GetAttribute(4)
	if status.(uint8)&clockStatusDoubtful == 0 {
		t.Fatalf("doubtful flag not set after preset")
	}

	if _, err := clock.Invoke(5, nil); err != nil {
		t.Fatalf("adjust_to_preset_time: %v", err)
	}
	if clock.presetAdjustTime != nil {
		t.Fatalf("preset time not cleared after adjustment")
	}
	timeAttr, _ := clock.GetAttribute(2)
	if !timeAttr.(time.Time).Equal(preset) {
		t.Fatalf("adjust_to_preset_time did not update time")
	}
	status, _ = clock.GetAttribute(4)
	if status.(uint8)&clockStatusDoubtful != 0 {
		t.Fatalf("doubtful flag should be cleared after adjustment")
	}

	if err := clock.SetAttribute(2, preset); err != nil {
		t.Fatalf("reset time: %v", err)
	}
	if _, err := clock.Invoke(6, []interface{}{int32(90)}); err != nil {
		t.Fatalf("shift_time: %v", err)
	}
	shifted, _ := clock.GetAttribute(2)
	expected := preset.Add(90 * time.Second)
	if !shifted.(time.Time).Equal(expected) {
		t.Fatalf("shift_time expected %v got %v", expected, shifted)
	}

	if _, err := clock.Invoke(5, nil); err != ErrInvalidParameter {
		t.Fatalf("expected ErrInvalidParameter when no preset, got %v", err)
	}
}
