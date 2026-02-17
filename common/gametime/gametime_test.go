package gametime

import (
	"testing"
	"time"
)

func TestAdjusted(t *testing.T) {
	result := Adjusted()

	_, offset := result.Zone()
	expectedOffset := 9 * 60 * 60
	if offset != expectedOffset {
		t.Errorf("Adjusted() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}

	now := time.Now()
	diff := result.Sub(now.In(time.FixedZone("UTC+9", 9*60*60)))
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Adjusted() time differs from expected by %v", diff)
	}
}

func TestMidnight(t *testing.T) {
	midnight := Midnight()

	if midnight.Hour() != 0 {
		t.Errorf("Midnight() hour = %d, want 0", midnight.Hour())
	}
	if midnight.Minute() != 0 {
		t.Errorf("Midnight() minute = %d, want 0", midnight.Minute())
	}
	if midnight.Second() != 0 {
		t.Errorf("Midnight() second = %d, want 0", midnight.Second())
	}
	if midnight.Nanosecond() != 0 {
		t.Errorf("Midnight() nanosecond = %d, want 0", midnight.Nanosecond())
	}

	_, offset := midnight.Zone()
	expectedOffset := 9 * 60 * 60
	if offset != expectedOffset {
		t.Errorf("Midnight() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}
}

func TestWeekStart(t *testing.T) {
	weekStart := WeekStart()

	if weekStart.Weekday() != time.Monday {
		t.Errorf("WeekStart() weekday = %v, want Monday", weekStart.Weekday())
	}

	if weekStart.Hour() != 0 || weekStart.Minute() != 0 || weekStart.Second() != 0 {
		t.Errorf("WeekStart() should be at midnight, got %02d:%02d:%02d",
			weekStart.Hour(), weekStart.Minute(), weekStart.Second())
	}

	_, offset := weekStart.Zone()
	expectedOffset := 9 * 60 * 60
	if offset != expectedOffset {
		t.Errorf("WeekStart() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}

	midnight := Midnight()
	if weekStart.After(midnight) {
		t.Errorf("WeekStart() %v should be <= current midnight %v", weekStart, midnight)
	}
}

func TestWeekNext(t *testing.T) {
	weekStart := WeekStart()
	weekNext := WeekNext()

	expectedNext := weekStart.Add(time.Hour * 24 * 7)
	if !weekNext.Equal(expectedNext) {
		t.Errorf("WeekNext() = %v, want %v (7 days after WeekStart)", weekNext, expectedNext)
	}

	if weekNext.Weekday() != time.Monday {
		t.Errorf("WeekNext() weekday = %v, want Monday", weekNext.Weekday())
	}

	if weekNext.Hour() != 0 || weekNext.Minute() != 0 || weekNext.Second() != 0 {
		t.Errorf("WeekNext() should be at midnight, got %02d:%02d:%02d",
			weekNext.Hour(), weekNext.Minute(), weekNext.Second())
	}

	if !weekNext.After(weekStart) {
		t.Errorf("WeekNext() %v should be after WeekStart() %v", weekNext, weekStart)
	}
}

func TestWeekStartSundayEdge(t *testing.T) {
	weekStart := WeekStart()

	if weekStart.Weekday() != time.Monday {
		t.Errorf("WeekStart() on any day should return Monday, got %v", weekStart.Weekday())
	}
}

func TestMidnightSameDay(t *testing.T) {
	adjusted := Adjusted()
	midnight := Midnight()

	if midnight.Year() != adjusted.Year() ||
		midnight.Month() != adjusted.Month() ||
		midnight.Day() != adjusted.Day() {
		t.Errorf("Midnight() date = %v, want same day as Adjusted() %v",
			midnight.Format("2006-01-02"), adjusted.Format("2006-01-02"))
	}
}

func TestWeekDuration(t *testing.T) {
	weekStart := WeekStart()
	weekNext := WeekNext()

	duration := weekNext.Sub(weekStart)
	expectedDuration := time.Hour * 24 * 7

	if duration != expectedDuration {
		t.Errorf("Duration between WeekStart and WeekNext = %v, want %v", duration, expectedDuration)
	}
}

func TestTimeZoneConsistency(t *testing.T) {
	adjusted := Adjusted()
	midnight := Midnight()
	weekStart := WeekStart()
	weekNext := WeekNext()

	times := []struct {
		name string
		time time.Time
	}{
		{"Adjusted", adjusted},
		{"Midnight", midnight},
		{"WeekStart", weekStart},
		{"WeekNext", weekNext},
	}

	expectedOffset := 9 * 60 * 60
	for _, tt := range times {
		_, offset := tt.time.Zone()
		if offset != expectedOffset {
			t.Errorf("%s() zone offset = %d, want %d (UTC+9)", tt.name, offset, expectedOffset)
		}
	}
}

func TestGameAbsolute(t *testing.T) {
	result := GameAbsolute()

	if result >= 5760 {
		t.Errorf("GameAbsolute() = %d, should be < 5760", result)
	}
}
