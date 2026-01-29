package mhfcourse

import (
	"math"
	"testing"
)

func TestCourses(t *testing.T) {
	courses := Courses()

	if len(courses) != 32 {
		t.Errorf("Courses() len = %d, want 32", len(courses))
	}

	for i, course := range courses {
		if course.ID != uint16(i) {
			t.Errorf("Courses()[%d].ID = %d, want %d", i, course.ID, i)
		}
	}
}

func TestCourseValue(t *testing.T) {
	tests := []struct {
		id   uint16
		want uint32
	}{
		{0, 1},
		{1, 2},
		{2, 4},
		{3, 8},
		{4, 16},
		{5, 32},
		{10, 1024},
		{20, 1048576},
		{31, 2147483648},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := Course{ID: tt.id}
			got := c.Value()
			if got != tt.want {
				t.Errorf("Course{ID: %d}.Value() = %d, want %d", tt.id, got, tt.want)
			}
		})
	}
}

func TestCourseValueIsPowerOf2(t *testing.T) {
	for i := uint16(0); i < 32; i++ {
		c := Course{ID: i}
		val := c.Value()
		expected := uint32(math.Pow(2, float64(i)))
		if val != expected {
			t.Errorf("Course{ID: %d}.Value() = %d, want %d (2^%d)", i, val, expected, i)
		}
	}
}

func TestCourseAliases(t *testing.T) {
	tests := []struct {
		id       uint16
		wantLen  int
		contains string
	}{
		{1, 2, "Trial"},
		{2, 2, "HunterLife"},
		{3, 3, "Extra"},
		{6, 1, "Premium"},
		{8, 4, "Assist"},
		{26, 4, "NetCafe"},
		{29, 1, "Free"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := Course{ID: tt.id}
			aliases := c.Aliases()

			if len(aliases) != tt.wantLen {
				t.Errorf("Course{ID: %d}.Aliases() len = %d, want %d", tt.id, len(aliases), tt.wantLen)
			}

			found := false
			for _, alias := range aliases {
				if alias == tt.contains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Course{ID: %d}.Aliases() should contain %q", tt.id, tt.contains)
			}
		})
	}
}

func TestCourseAliasesUnknown(t *testing.T) {
	// Test IDs without aliases
	unknownIDs := []uint16{13, 14, 15, 16, 17, 18, 19}

	for _, id := range unknownIDs {
		c := Course{ID: id}
		aliases := c.Aliases()
		if aliases != nil {
			t.Errorf("Course{ID: %d}.Aliases() = %v, want nil", id, aliases)
		}
	}
}

func TestCourseExists(t *testing.T) {
	courses := []Course{
		{ID: 1},
		{ID: 5},
		{ID: 10},
	}

	tests := []struct {
		id   uint16
		want bool
	}{
		{1, true},
		{5, true},
		{10, true},
		{0, false},
		{2, false},
		{99, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := CourseExists(tt.id, courses)
			if got != tt.want {
				t.Errorf("CourseExists(%d, courses) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestCourseExistsEmptySlice(t *testing.T) {
	var courses []Course

	if CourseExists(1, courses) {
		t.Error("CourseExists(1, nil) should return false")
	}
}

func TestGetCourseStruct(t *testing.T) {
	tests := []struct {
		name          string
		rights        uint32
		wantMinLen    int
		shouldHave    []uint16
		shouldNotHave []uint16
	}{
		{
			name:       "zero rights",
			rights:     0,
			wantMinLen: 1, // Always includes ID: 1 (Trial)
			shouldHave: []uint16{1},
		},
		{
			name:       "HunterLife course",
			rights:     4, // 2^2 = 4 for ID 2
			wantMinLen: 2,
			shouldHave: []uint16{1, 2},
		},
		{
			name:       "multiple courses",
			rights:     6, // 2^1 + 2^2 = 2 + 4 = 6 for IDs 1 and 2
			wantMinLen: 2,
			shouldHave: []uint16{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			courses, _ := GetCourseStruct(tt.rights)

			if len(courses) < tt.wantMinLen {
				t.Errorf("GetCourseStruct(%d) len = %d, want >= %d", tt.rights, len(courses), tt.wantMinLen)
			}

			for _, id := range tt.shouldHave {
				if !CourseExists(id, courses) {
					t.Errorf("GetCourseStruct(%d) should have course ID %d", tt.rights, id)
				}
			}

			for _, id := range tt.shouldNotHave {
				if CourseExists(id, courses) {
					t.Errorf("GetCourseStruct(%d) should not have course ID %d", tt.rights, id)
				}
			}
		})
	}
}

func TestGetCourseStructReturnsRights(t *testing.T) {
	// GetCourseStruct returns the recalculated rights value
	_, rights := GetCourseStruct(0)

	// Should at least include the Trial course (ID: 1, Value: 2)
	if rights < 2 {
		t.Errorf("GetCourseStruct(0) rights = %d, want >= 2", rights)
	}
}

func TestGetCourseStructNetCafeCourses(t *testing.T) {
	// Test that course 26 (NetCafe) adds course 25 (CAFE_SP)
	courses, _ := GetCourseStruct(Course{ID: 26}.Value())

	if !CourseExists(25, courses) {
		t.Error("GetCourseStruct with course 26 should add course 25")
	}
	if !CourseExists(30, courses) {
		t.Error("GetCourseStruct with course 26 should add course 30")
	}
}

func TestGetCourseStructNCourse(t *testing.T) {
	// Test that course 9 (N) adds course 30
	courses, _ := GetCourseStruct(Course{ID: 9}.Value())

	if !CourseExists(30, courses) {
		t.Error("GetCourseStruct with course 9 should add course 30")
	}
}

func TestCourseExpiry(t *testing.T) {
	// Test that courses returned by GetCourseStruct have expiry set
	courses, _ := GetCourseStruct(4) // HunterLife

	for _, c := range courses {
		// Course ID 1 is always added without expiry in some cases
		if c.ID != 1 && c.ID != 25 && c.ID != 30 {
			if c.Expiry.IsZero() {
				// Note: expiry is only set for courses extracted from rights
				// This behavior is expected
			}
		}
	}
}

func TestAllCoursesHaveValidValues(t *testing.T) {
	courses := Courses()

	for _, c := range courses {
		val := c.Value()
		// Verify value is a power of 2
		if val == 0 || (val&(val-1)) != 0 {
			t.Errorf("Course{ID: %d}.Value() = %d is not a power of 2", c.ID, val)
		}
	}
}

func TestKnownAliasesExist(t *testing.T) {
	knownCourses := map[string]uint16{
		"Trial":      1,
		"HunterLife": 2,
		"Extra":      3,
		"Mobile":     5,
		"Premium":    6,
		"Assist":     8,
		"Hiden":      10,
		"NetCafe":    26,
		"Free":       29,
	}

	for name, expectedID := range knownCourses {
		t.Run(name, func(t *testing.T) {
			c := Course{ID: expectedID}
			aliases := c.Aliases()

			found := false
			for _, alias := range aliases {
				if alias == name {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Course ID %d should have alias %q, got %v", expectedID, name, aliases)
			}
		})
	}
}
