package channelserver

import (
	"fmt"
	"time"
)

var (
	Offset      = 9
	YearAdjust  = -7
	MonthAdjust = 0
	DayAdjust   = 0
)

var (
	TimeStatic = time.Time{}
)

func Time_Current() time.Time {
	baseTime := time.Now().In(time.FixedZone(fmt.Sprintf("UTC+%d", Offset), Offset*60*60))
	return baseTime
}

func Time_Current_Adjusted() time.Time {
	baseTime := time.Now().In(time.FixedZone(fmt.Sprintf("UTC+%d", Offset), Offset*60*60)).AddDate(YearAdjust, MonthAdjust, DayAdjust)
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), baseTime.Nanosecond(), baseTime.Location())
}

func Time_Current_Midnight() time.Time {
	baseTime := time.Now().In(time.FixedZone(fmt.Sprintf("UTC+%d", Offset), Offset*60*60)).AddDate(YearAdjust, MonthAdjust, DayAdjust)
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
}

func Time_Current_Week_uint8() uint8 {
	baseTime := time.Now().In(time.FixedZone(fmt.Sprintf("UTC+%d", Offset), Offset*60*60)).AddDate(YearAdjust, MonthAdjust, DayAdjust)

	_, thisWeek := baseTime.ISOWeek()
	_, beginningOfTheMonth := time.Date(baseTime.Year(), baseTime.Month(), 1, 0, 0, 0, 0, baseTime.Location()).ISOWeek()

	return uint8(1 + thisWeek - beginningOfTheMonth)
}

func Time_static() time.Time {
	if TimeStatic.IsZero() {
		TimeStatic = Time_Current_Adjusted()
	}
	return TimeStatic
}
