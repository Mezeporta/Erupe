package timeserver

import (
	"time"
)

var DoOnce_midnight = false
var DoOnce_t2 = false
var DoOnce_t = false
var Fix_midnight = time.Time{}
var Fix_t2 = time.Time{}
var Fix_t = time.Time{}
var Pfixtimer time.Duration
var Pnewtime = 0
var yearsFixed = -7

func PFadd_time() time.Duration {
	Pnewtime = Pnewtime + 24
	Pfixtimer = time.Duration(Pnewtime)
	return Pfixtimer
}

func Time_static() time.Time {
	if !DoOnce_t {
		DoOnce_t = true
		// Force to 201x
		tFix1 := time.Now()
		tFix2 := tFix1.AddDate(yearsFixed, 0, 0)
		Fix_t = tFix2.In(time.FixedZone("UTC+1", 1*60*60))
	}
	return Fix_t
}

func Tstatic_midnight() time.Time {
	if !DoOnce_midnight {
		DoOnce_midnight = true
		// Force to 201x
		tFix1 := time.Now()
		tFix2 := tFix1.AddDate(yearsFixed, 0, 0)
		var tFix = tFix2.In(time.FixedZone("UTC+1", 1*60*60))
		yearFix, monthFix, dayFix := tFix2.Date()
		Fix_midnight = time.Date(yearFix, monthFix, dayFix, 0, 0, 0, 0, tFix.Location()).Add(time.Hour)
	}
	return Fix_midnight
}

func Time_midnight() time.Time {
	// Force to 201x
	t1 := time.Now()
	t2 := t1.AddDate(yearsFixed, 0, 0)
	var t = t2.In(time.FixedZone("UTC+1", 1*60*60))
	year, month, day := t2.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Add(time.Hour)
	return midnight
}

func TimeCurrent() time.Time {
	// Force to 201x
	t1 := time.Now()
	t2 := t1.AddDate(yearsFixed, 0, 0)
	var t = t2.In(time.FixedZone("UTC+1", 1*60*60))
	return t
}

func Time_Current_Week_uint8() uint8 {
	beginningOfTheMonth := time.Date(TimeCurrent().Year(), TimeCurrent().Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := TimeCurrent().ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()

	return uint8(1 + thisWeek - beginningWeek)
}

func Time_Current_Week_uint32() uint32 {
	beginningOfTheMonth := time.Date(TimeCurrent().Year(), TimeCurrent().Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := TimeCurrent().ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	result := 1 + thisWeek - beginningWeek
	return uint32(result)
}

func Detect_Day() bool {
	switch time.Now().Weekday() {
	case time.Wednesday:
		return true
	}
	return false
}
