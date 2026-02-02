package timeserver

import (
	"testing"
	"time"
)

func TestPFaddTime(t *testing.T) {
	// Save original state
	originalPnewtime := Pnewtime
	defer func() { Pnewtime = originalPnewtime }()

	Pnewtime = 0

	// First call should return 24
	result := PFadd_time()
	if result != time.Duration(24) {
		t.Errorf("PFadd_time() = %v, want 24", result)
	}

	// Second call should return 48
	result = PFadd_time()
	if result != time.Duration(48) {
		t.Errorf("PFadd_time() second call = %v, want 48", result)
	}

	// Check Pfixtimer is updated
	if Pfixtimer != time.Duration(48) {
		t.Errorf("Pfixtimer = %v, want 48", Pfixtimer)
	}
}

func TestTimeCurrent(t *testing.T) {
	result := TimeCurrent()
	if result.IsZero() {
		t.Error("TimeCurrent() returned zero time")
	}

	// Result should be in the past (7 years ago)
	now := time.Now()
	diff := now.Year() - result.Year()
	if diff != 7 {
		t.Errorf("TimeCurrent() year diff = %d, want 7", diff)
	}
}

func TestTimeMidnight(t *testing.T) {
	result := Time_midnight()
	if result.IsZero() {
		t.Error("Time_midnight() returned zero time")
	}

	// Result should be midnight (with hour added, so 1:00 AM)
	if result.Minute() != 0 {
		t.Errorf("Time_midnight() minute = %d, want 0", result.Minute())
	}
	if result.Second() != 0 {
		t.Errorf("Time_midnight() second = %d, want 0", result.Second())
	}
}

func TestTimeStatic(t *testing.T) {
	// Reset state for testing
	DoOnce_t = false
	Fix_t = time.Time{}

	result := Time_static()
	if result.IsZero() {
		t.Error("Time_static() returned zero time")
	}

	// Calling again should return same time (static)
	result2 := Time_static()
	if !result.Equal(result2) {
		t.Error("Time_static() should return same time on second call")
	}
}

func TestTstaticMidnight(t *testing.T) {
	// Reset state for testing
	DoOnce_midnight = false
	Fix_midnight = time.Time{}

	result := Tstatic_midnight()
	if result.IsZero() {
		t.Error("Tstatic_midnight() returned zero time")
	}

	// Calling again should return same time (static)
	result2 := Tstatic_midnight()
	if !result.Equal(result2) {
		t.Error("Tstatic_midnight() should return same time on second call")
	}
}

func TestTimeCurrentWeekUint8(t *testing.T) {
	result := Time_Current_Week_uint8()

	// Week of month should be 1-5
	if result < 1 || result > 5 {
		t.Errorf("Time_Current_Week_uint8() = %d, expected 1-5", result)
	}
}

func TestTimeCurrentWeekUint32(t *testing.T) {
	result := Time_Current_Week_uint32()

	// Week of month should be 1-5
	if result < 1 || result > 5 {
		t.Errorf("Time_Current_Week_uint32() = %d, expected 1-5", result)
	}
}

func TestDetectDay(t *testing.T) {
	result := Detect_Day()

	// Result should be bool, and true only on Wednesday
	isWednesday := time.Now().Weekday() == time.Wednesday
	if result != isWednesday {
		t.Errorf("Detect_Day() = %v, expected %v (today is %v)", result, isWednesday, time.Now().Weekday())
	}
}

func TestGlobalVariables(t *testing.T) {
	// Test that global variables exist and have expected default types
	_ = DoOnce_midnight
	_ = DoOnce_t2
	_ = DoOnce_t
	_ = Fix_midnight
	_ = Fix_t2
	_ = Fix_t
	_ = Pfixtimer
	_ = Pnewtime
}

func TestTimeConsistency(t *testing.T) {
	// Test that TimeCurrent and Time_midnight are on the same day
	current := TimeCurrent()
	midnight := Time_midnight()

	if current.Year() != midnight.Year() {
		t.Errorf("Year mismatch: current=%d, midnight=%d", current.Year(), midnight.Year())
	}
	if current.Month() != midnight.Month() {
		t.Errorf("Month mismatch: current=%v, midnight=%v", current.Month(), midnight.Month())
	}
	if current.Day() != midnight.Day() {
		t.Errorf("Day mismatch: current=%d, midnight=%d", current.Day(), midnight.Day())
	}
}
