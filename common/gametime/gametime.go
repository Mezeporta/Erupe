package gametime

import (
	"time"
)

func Adjusted() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), baseTime.Nanosecond(), baseTime.Location())
}

func Midnight() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
}

func WeekStart() time.Time {
	midnight := Midnight()
	offset := int(midnight.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return midnight.Add(-time.Duration(offset) * 24 * time.Hour)
}

func WeekNext() time.Time {
	return WeekStart().Add(time.Hour * 24 * 7)
}

func GameAbsolute() uint32 {
	return uint32((Adjusted().Unix() - 2160) % 5760)
}
