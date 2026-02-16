package channelserver

import (
	"testing"
	"time"
)

func TestTimeAdjusted(t *testing.T) {
	result := TimeAdjusted()

	// Should return a time in UTC+9 timezone
	_, offset := result.Zone()
	expectedOffset := 9 * 60 * 60 // 9 hours in seconds
	if offset != expectedOffset {
		t.Errorf("TimeAdjusted() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}

	// The time should be close to current time (within a few seconds)
	now := time.Now()
	diff := result.Sub(now.In(time.FixedZone("UTC+9", 9*60*60)))
	if diff < -time.Second || diff > time.Second {
		t.Errorf("TimeAdjusted() time differs from expected by %v", diff)
	}
}

func TestTimeMidnight(t *testing.T) {
	midnight := TimeMidnight()

	// Should be at midnight (hour=0, minute=0, second=0, nanosecond=0)
	if midnight.Hour() != 0 {
		t.Errorf("TimeMidnight() hour = %d, want 0", midnight.Hour())
	}
	if midnight.Minute() != 0 {
		t.Errorf("TimeMidnight() minute = %d, want 0", midnight.Minute())
	}
	if midnight.Second() != 0 {
		t.Errorf("TimeMidnight() second = %d, want 0", midnight.Second())
	}
	if midnight.Nanosecond() != 0 {
		t.Errorf("TimeMidnight() nanosecond = %d, want 0", midnight.Nanosecond())
	}

	// Should be in UTC+9 timezone
	_, offset := midnight.Zone()
	expectedOffset := 9 * 60 * 60
	if offset != expectedOffset {
		t.Errorf("TimeMidnight() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}
}

func TestTimeWeekStart(t *testing.T) {
	weekStart := TimeWeekStart()

	// Should be on Monday (weekday = 1)
	if weekStart.Weekday() != time.Monday {
		t.Errorf("TimeWeekStart() weekday = %v, want Monday", weekStart.Weekday())
	}

	// Should be at midnight
	if weekStart.Hour() != 0 || weekStart.Minute() != 0 || weekStart.Second() != 0 {
		t.Errorf("TimeWeekStart() should be at midnight, got %02d:%02d:%02d",
			weekStart.Hour(), weekStart.Minute(), weekStart.Second())
	}

	// Should be in UTC+9 timezone
	_, offset := weekStart.Zone()
	expectedOffset := 9 * 60 * 60
	if offset != expectedOffset {
		t.Errorf("TimeWeekStart() zone offset = %d, want %d (UTC+9)", offset, expectedOffset)
	}

	// Week start should be before or equal to current midnight
	midnight := TimeMidnight()
	if weekStart.After(midnight) {
		t.Errorf("TimeWeekStart() %v should be <= current midnight %v", weekStart, midnight)
	}
}

func TestTimeWeekNext(t *testing.T) {
	weekStart := TimeWeekStart()
	weekNext := TimeWeekNext()

	// TimeWeekNext should be exactly 7 days after TimeWeekStart
	expectedNext := weekStart.Add(time.Hour * 24 * 7)
	if !weekNext.Equal(expectedNext) {
		t.Errorf("TimeWeekNext() = %v, want %v (7 days after WeekStart)", weekNext, expectedNext)
	}

	// Should also be on Monday
	if weekNext.Weekday() != time.Monday {
		t.Errorf("TimeWeekNext() weekday = %v, want Monday", weekNext.Weekday())
	}

	// Should be at midnight
	if weekNext.Hour() != 0 || weekNext.Minute() != 0 || weekNext.Second() != 0 {
		t.Errorf("TimeWeekNext() should be at midnight, got %02d:%02d:%02d",
			weekNext.Hour(), weekNext.Minute(), weekNext.Second())
	}

	// Should be in the future relative to week start
	if !weekNext.After(weekStart) {
		t.Errorf("TimeWeekNext() %v should be after TimeWeekStart() %v", weekNext, weekStart)
	}
}

func TestTimeWeekStartSundayEdge(t *testing.T) {
	// When today is Sunday, the calculation should go back to last Monday
	// This is tested indirectly by verifying the weekday is always Monday
	weekStart := TimeWeekStart()

	// Regardless of what day it is now, week start should be Monday
	if weekStart.Weekday() != time.Monday {
		t.Errorf("TimeWeekStart() on any day should return Monday, got %v", weekStart.Weekday())
	}
}

func TestTimeMidnightSameDay(t *testing.T) {
	adjusted := TimeAdjusted()
	midnight := TimeMidnight()

	// Midnight should be on the same day (year, month, day)
	if midnight.Year() != adjusted.Year() ||
		midnight.Month() != adjusted.Month() ||
		midnight.Day() != adjusted.Day() {
		t.Errorf("TimeMidnight() date = %v, want same day as TimeAdjusted() %v",
			midnight.Format("2006-01-02"), adjusted.Format("2006-01-02"))
	}
}

func TestTimeWeekDuration(t *testing.T) {
	weekStart := TimeWeekStart()
	weekNext := TimeWeekNext()

	// Duration between week boundaries should be exactly 7 days
	duration := weekNext.Sub(weekStart)
	expectedDuration := time.Hour * 24 * 7

	if duration != expectedDuration {
		t.Errorf("Duration between WeekStart and WeekNext = %v, want %v", duration, expectedDuration)
	}
}

func TestTimeZoneConsistency(t *testing.T) {
	adjusted := TimeAdjusted()
	midnight := TimeMidnight()
	weekStart := TimeWeekStart()
	weekNext := TimeWeekNext()

	// All times should be in the same timezone (UTC+9)
	times := []struct {
		name string
		time time.Time
	}{
		{"TimeAdjusted", adjusted},
		{"TimeMidnight", midnight},
		{"TimeWeekStart", weekStart},
		{"TimeWeekNext", weekNext},
	}

	expectedOffset := 9 * 60 * 60
	for _, tt := range times {
		_, offset := tt.time.Zone()
		if offset != expectedOffset {
			t.Errorf("%s() zone offset = %d, want %d (UTC+9)", tt.name, offset, expectedOffset)
		}
	}
}
