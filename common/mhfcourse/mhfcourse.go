package mhfcourse

import (
	"math"
	"sort"
	"time"
)

// Course represents an active subscription course with its ID and expiry time.
type Course struct {
	ID     uint16
	Expiry time.Time
}

var aliases = map[uint16][]string{
	1:  {"Trial", "TL"},
	2:  {"HunterLife", "HL"},
	3:  {"Extra", "ExtraA", "EX"},
	4:  {"ExtraB"},
	5:  {"Mobile"},
	6:  {"Premium"},
	7:  {"Pallone", "ExtraC"},
	8:  {"Assist", "***ist", "Legend", "Rasta"},
	9:  {"N"},
	10: {"Hiden", "Secret"},
	11: {"HunterSupport", "HunterAid", "Support", "Aid", "Royal"},
	12: {"NBoost", "NetCafeBoost", "Boost"},
	// 13-19 show up as (unknown)
	20: {"DEBUG"},
	21: {"COG_LINK_EXPIRED"},
	22: {"360_GOLD"},
	23: {"PS3_TROP"},
	24: {"COG"},
	25: {"CAFE_SP"},
	26: {"NetCafe", "Cafe", "OfficialCafe", "Official"},
	27: {"HLRenewing", "HLR", "HLRenewal", "HLRenew", "CardHL"},
	28: {"EXRenewing", "EXR", "EXRenewal", "EXRenew", "CardEX"},
	29: {"Free"},
	// 30 = Real NetCafe course
}

// Aliases returns the human-readable names for this course (e.g. "HunterLife", "HL").
func (c Course) Aliases() []string {
	return aliases[c.ID]
}

// Courses returns all 32 possible course slots with zero-value expiry times.
func Courses() []Course {
	courses := make([]Course, 32)
	for i := range courses {
		courses[i].ID = uint16(i)
	}
	return courses
}

// Value returns the bitmask value for this course (2^ID).
func (c Course) Value() uint32 {
	return uint32(math.Pow(2, float64(c.ID)))
}

// CourseExists returns true if the named course exists in the given slice
func CourseExists(ID uint16, c []Course) bool {
	for _, course := range c {
		if course.ID == ID {
			return true
		}
	}
	return false
}

// GetCourseStruct returns a slice of Course(s) from a rights integer
func GetCourseStruct(rights uint32, defaultCourses []uint16) ([]Course, uint32) {
	var resp []Course
	for _, c := range defaultCourses {
		resp = append(resp, Course{ID: c})
	}
	s := Courses()
	sort.Slice(s, func(i, j int) bool {
		return s[i].ID > s[j].ID
	})
	var normalCafeCourseSet, netcafeCourseSet, hidenCourseSet bool
	for _, course := range s {
		if rights-course.Value() < 0x80000000 {
			switch course.ID {
			case 26:
				if normalCafeCourseSet {
					break
				}
				normalCafeCourseSet = true
				resp = append(resp, Course{ID: 25})
				fallthrough
			case 9:
				if netcafeCourseSet {
					break
				}
				netcafeCourseSet = true
				resp = append(resp, Course{ID: 30})
			case 10:
				if hidenCourseSet {
					break
				}
				hidenCourseSet = true
				resp = append(resp, Course{ID: 31})
			}
			course.Expiry = time.Date(2030, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
			resp = append(resp, course)
			rights -= course.Value()
		}
	}
	rights = 0
	for _, course := range resp {
		rights += course.Value()
	}
	return resp, rights
}
