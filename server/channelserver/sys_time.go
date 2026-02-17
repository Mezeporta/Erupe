package channelserver

import (
	"erupe-ce/common/gametime"
	"time"
)

func TimeAdjusted() time.Time    { return gametime.Adjusted() }
func TimeMidnight() time.Time    { return gametime.Midnight() }
func TimeWeekStart() time.Time   { return gametime.WeekStart() }
func TimeWeekNext() time.Time    { return gametime.WeekNext() }
func TimeGameAbsolute() uint32   { return gametime.GameAbsolute() }
