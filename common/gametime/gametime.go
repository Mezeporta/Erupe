package gametime

import (
	"time"
)

// Adjusted returns the current time in JST (UTC+9), the timezone used by MHF.
func Adjusted() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), baseTime.Nanosecond(), baseTime.Location())
}

// Midnight returns today's midnight (00:00) in JST.
func Midnight() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
}

// WeekStart returns the most recent Monday at midnight in JST.
func WeekStart() time.Time {
	midnight := Midnight()
	offset := int(midnight.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return midnight.Add(-time.Duration(offset) * 24 * time.Hour)
}

// WeekNext returns the next Monday at midnight in JST.
func WeekNext() time.Time {
	return WeekStart().Add(time.Hour * 24 * 7)
}

// GameAbsolute returns the current position within the 5760-second (96-minute)
// in-game day/night cycle, offset by 2160 seconds.
func GameAbsolute() uint32 {
	return uint32((Adjusted().Unix() - 2160) % 5760)
}
