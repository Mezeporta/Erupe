package mhfcourse

import (
	"math"
	"testing"
	"time"
)

func TestCourse_Aliases(t *testing.T) {
	tests := []struct {
		id      uint16
		wantLen int
		want    []string
	}{
		{1, 2, []string{"Trial", "TL"}},
		{2, 2, []string{"HunterLife", "HL"}},
		{3, 3, []string{"Extra", "ExtraA", "EX"}},
		{8, 4, []string{"Assist", "***ist", "Legend", "Rasta"}},
		{26, 4, []string{"NetCafe", "Cafe", "OfficialCafe", "Official"}},
		{13, 0, nil}, // Unknown course
		{99, 0, nil}, // Unknown course
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.id)), func(t *testing.T) {
			c := Course{ID: tt.id}
			got := c.Aliases()
			if len(got) != tt.wantLen {
				t.Errorf("Course{ID: %d}.Aliases() length = %d, want %d", tt.id, len(got), tt.wantLen)
			}
			if tt.want != nil {
				for i, alias := range tt.want {
					if i >= len(got) || got[i] != alias {
						t.Errorf("Course{ID: %d}.Aliases()[%d] = %q, want %q", tt.id, i, got[i], alias)
					}
				}
			}
		})
	}
}

func TestCourses(t *testing.T) {
	courses := Courses()
	if len(courses) != 32 {
		t.Errorf("Courses() length = %d, want 32", len(courses))
	}

	// Verify IDs are sequential from 0 to 31
	for i, course := range courses {
		if course.ID != uint16(i) {
			t.Errorf("Courses()[%d].ID = %d, want %d", i, course.ID, i)
		}
	}
}

func TestCourse_Value(t *testing.T) {
	tests := []struct {
		id       uint16
		expected uint32
	}{
		{0, 1},           // 2^0
		{1, 2},           // 2^1
		{2, 4},           // 2^2
		{3, 8},           // 2^3
		{4, 16},          // 2^4
		{5, 32},          // 2^5
		{10, 1024},       // 2^10
		{15, 32768},      // 2^15
		{20, 1048576},    // 2^20
		{31, 2147483648}, // 2^31
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.id)), func(t *testing.T) {
			c := Course{ID: tt.id}
			got := c.Value()
			if got != tt.expected {
				t.Errorf("Course{ID: %d}.Value() = %d, want %d", tt.id, got, tt.expected)
			}
		})
	}
}

func TestCourseExists(t *testing.T) {
	courses := []Course{
		{ID: 1},
		{ID: 5},
		{ID: 10},
		{ID: 15},
	}

	tests := []struct {
		name     string
		id       uint16
		expected bool
	}{
		{"exists first", 1, true},
		{"exists middle", 5, true},
		{"exists last", 15, true},
		{"not exists", 3, false},
		{"not exists 0", 0, false},
		{"not exists 20", 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CourseExists(tt.id, courses)
			if got != tt.expected {
				t.Errorf("CourseExists(%d, courses) = %v, want %v", tt.id, got, tt.expected)
			}
		})
	}
}

func TestCourseExists_EmptySlice(t *testing.T) {
	var courses []Course
	if CourseExists(1, courses) {
		t.Error("CourseExists(1, []) should return false for empty slice")
	}
}

func TestGetCourseStruct(t *testing.T) {
	defaultCourses := []uint16{1, 2}

	tests := []struct {
		name         string
		rights       uint32
		wantMinLen   int // Minimum expected courses (including defaults)
		checkCourses []uint16
	}{
		{
			name:         "no rights",
			rights:       0,
			wantMinLen:   2, // Just default courses
			checkCourses: []uint16{1, 2},
		},
		{
			name:         "course 3 only",
			rights:       8, // 2^3
			wantMinLen:   3, // defaults + course 3
			checkCourses: []uint16{1, 2, 3},
		},
		{
			name:         "course 1",
			rights:       2, // 2^1
			wantMinLen:   2,
			checkCourses: []uint16{1, 2},
		},
		{
			name:         "multiple courses",
			rights:       2 + 8 + 32, // courses 1, 3, 5
			wantMinLen:   4,
			checkCourses: []uint16{1, 2, 3, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			courses, newRights := GetCourseStruct(tt.rights, defaultCourses)

			if len(courses) < tt.wantMinLen {
				t.Errorf("GetCourseStruct(%d) returned %d courses, want at least %d", tt.rights, len(courses), tt.wantMinLen)
			}

			// Verify expected courses are present
			for _, id := range tt.checkCourses {
				found := false
				for _, c := range courses {
					if c.ID == id {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetCourseStruct(%d) missing expected course ID %d", tt.rights, id)
				}
			}

			// Verify newRights is a valid sum of course values
			if newRights < tt.rights {
				t.Logf("GetCourseStruct(%d) newRights = %d (may include additional courses)", tt.rights, newRights)
			}
		})
	}
}

func TestGetCourseStruct_NetcafeCourse(t *testing.T) {
	// Course 26 (NetCafe) should add course 25
	courses, _ := GetCourseStruct(1<<26, nil)

	hasNetcafe := false
	hasCafeSP := false
	hasRealNetcafe := false
	for _, c := range courses {
		if c.ID == 26 {
			hasNetcafe = true
		}
		if c.ID == 25 {
			hasCafeSP = true
		}
		if c.ID == 30 {
			hasRealNetcafe = true
		}
	}

	if !hasNetcafe {
		t.Error("Course 26 (NetCafe) should be present")
	}
	if !hasCafeSP {
		t.Error("Course 25 should be added when course 26 is present")
	}
	if !hasRealNetcafe {
		t.Error("Course 30 should be added when course 26 is present")
	}
}

func TestGetCourseStruct_NCourse(t *testing.T) {
	// Course 9 should add course 30
	courses, _ := GetCourseStruct(1<<9, nil)

	hasNCourse := false
	hasRealNetcafe := false
	for _, c := range courses {
		if c.ID == 9 {
			hasNCourse = true
		}
		if c.ID == 30 {
			hasRealNetcafe = true
		}
	}

	if !hasNCourse {
		t.Error("Course 9 (N) should be present")
	}
	if !hasRealNetcafe {
		t.Error("Course 30 should be added when course 9 is present")
	}
}

func TestGetCourseStruct_HidenCourse(t *testing.T) {
	// Course 10 (Hiden) should add course 31
	courses, _ := GetCourseStruct(1<<10, nil)

	hasHiden := false
	hasHidenExtra := false
	for _, c := range courses {
		if c.ID == 10 {
			hasHiden = true
		}
		if c.ID == 31 {
			hasHidenExtra = true
		}
	}

	if !hasHiden {
		t.Error("Course 10 (Hiden) should be present")
	}
	if !hasHidenExtra {
		t.Error("Course 31 should be added when course 10 is present")
	}
}

func TestGetCourseStruct_ExpiryDate(t *testing.T) {
	courses, _ := GetCourseStruct(1<<3, nil)

	expectedExpiry := time.Date(2030, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))

	for _, c := range courses {
		if c.ID == 3 && !c.Expiry.IsZero() {
			if !c.Expiry.Equal(expectedExpiry) {
				t.Errorf("Course expiry = %v, want %v", c.Expiry, expectedExpiry)
			}
		}
	}
}

func TestGetCourseStruct_ReturnsRecalculatedRights(t *testing.T) {
	courses, newRights := GetCourseStruct(2+8+32, nil) // courses 1, 3, 5

	// Calculate expected rights from returned courses
	var expectedRights uint32
	for _, c := range courses {
		expectedRights += c.Value()
	}

	if newRights != expectedRights {
		t.Errorf("GetCourseStruct() newRights = %d, want %d (sum of returned course values)", newRights, expectedRights)
	}
}

func TestCourse_ValueMatchesPowerOfTwo(t *testing.T) {
	// Verify that Value() correctly implements 2^ID
	for id := uint16(0); id < 32; id++ {
		c := Course{ID: id}
		expected := uint32(math.Pow(2, float64(id)))
		got := c.Value()
		if got != expected {
			t.Errorf("Course{ID: %d}.Value() = %d, want %d", id, got, expected)
		}
	}
}

func BenchmarkCourse_Value(b *testing.B) {
	c := Course{ID: 15}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Value()
	}
}

func BenchmarkCourseExists(b *testing.B) {
	courses := []Course{
		{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
		{ID: 10}, {ID: 15}, {ID: 20}, {ID: 25}, {ID: 30},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CourseExists(15, courses)
	}
}

func BenchmarkGetCourseStruct(b *testing.B) {
	defaultCourses := []uint16{1, 2}
	rights := uint32(2 + 8 + 32 + 128 + 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetCourseStruct(rights, defaultCourses)
	}
}

func BenchmarkCourses(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Courses()
	}
}
