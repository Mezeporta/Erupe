package channelserver

import (
	"erupe-ce/common/gametime"
	"time"
)

// TimeAdjusted, TimeMidnight, TimeWeekStart, TimeWeekNext, and TimeGameAbsolute
// are package-level wrappers around the gametime utility functions, providing
// convenient access to adjusted server time, daily/weekly boundaries, and the
// absolute game timestamp used by the MHF client.

func TimeAdjusted() time.Time    { return gametime.Adjusted() }
func TimeMidnight() time.Time    { return gametime.Midnight() }
func TimeWeekStart() time.Time   { return gametime.WeekStart() }
func TimeWeekNext() time.Time    { return gametime.WeekNext() }
func TimeMonthStart() time.Time  { return gametime.MonthStart() }
func TimeGameAbsolute() uint32   { return gametime.GameAbsolute() }
