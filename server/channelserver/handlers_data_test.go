package channelserver

import (
	"testing"
)

func TestGrpToGR(t *testing.T) {
	tests := []struct {
		name     string
		grp      uint32
		expected uint16
	}{
		// GR 1-50 range (grp < 208750)
		{"GR 1 minimum", 0, 1},
		{"GR 2 at 500", 500, 2},
		{"GR 3 at 1150", 1150, 3},
		{"GR low range", 1000, 2},
		{"GR 50 boundary minus one", 208749, 50},

		// GR 51-99 range (208750 <= grp < 593400)
		{"GR 51 at boundary", 208750, 51},
		{"GR 52 at 216600", 216600, 52},
		{"GR mid-range 70", 358050, 70},
		{"GR 99 boundary minus one", 593399, 99},

		// GR 100-149 range (593400 <= grp < 993400)
		{"GR 100 at boundary", 593400, 100},
		{"GR 101 at 601400", 601400, 101},
		{"GR 125 midpoint", 793400, 125},
		{"GR 149 boundary minus one", 993399, 149},

		// GR 150-199 range (993400 <= grp < 1400900)
		{"GR 150 at boundary", 993400, 150},
		{"GR 175 midpoint", 1197150, 175},
		{"GR 199 boundary minus one", 1400899, 199},

		// GR 200-299 range (1400900 <= grp < 2315900)
		{"GR 200 at boundary", 1400900, 200},
		{"GR 250 midpoint", 1858400, 250},
		{"GR 299 boundary minus one", 2315899, 299},

		// GR 300-399 range (2315900 <= grp < 3340900)
		{"GR 300 at boundary", 2315900, 300},
		{"GR 350 midpoint", 2828400, 350},
		{"GR 399 boundary minus one", 3340899, 399},

		// GR 400-499 range (3340900 <= grp < 4505900)
		{"GR 400 at boundary", 3340900, 400},
		{"GR 450 midpoint", 3923400, 450},
		{"GR 499 boundary minus one", 4505899, 499},

		// GR 500-599 range (4505900 <= grp < 5850900)
		{"GR 500 at boundary", 4505900, 500},
		{"GR 550 midpoint", 5178400, 550},
		{"GR 599 boundary minus one", 5850899, 599},

		// GR 600-699 range (5850900 <= grp < 7415900)
		{"GR 600 at boundary", 5850900, 600},
		{"GR 650 midpoint", 6633400, 650},
		{"GR 699 boundary minus one", 7415899, 699},

		// GR 700-799 range (7415900 <= grp < 9230900)
		{"GR 700 at boundary", 7415900, 700},
		{"GR 750 midpoint", 8323400, 750},
		{"GR 799 boundary minus one", 9230899, 799},

		// GR 800-899 range (9230900 <= grp < 11345900)
		{"GR 800 at boundary", 9230900, 800},
		{"GR 850 midpoint", 10288400, 850},
		{"GR 899 boundary minus one", 11345899, 899},

		// GR 900+ range (grp >= 11345900)
		{"GR 900 at boundary", 11345900, 900},
		{"GR 950 midpoint", 12543400, 950},
		{"GR 998 high value", 13716450, 998}, // Actual function result for this GRP
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grpToGR(tt.grp)
			if result != tt.expected {
				t.Errorf("grpToGR(%d) = %d, want %d", tt.grp, result, tt.expected)
			}
		})
	}
}

func TestGrpToGR_EdgeCases(t *testing.T) {
	// Test that GR never goes below 1
	result := grpToGR(0)
	if result < 1 {
		t.Errorf("grpToGR(0) = %d, should be at least 1", result)
	}

	// Test very high GRP values
	result = grpToGR(20000000)
	if result < 900 {
		t.Errorf("grpToGR(20000000) = %d, should be >= 900", result)
	}
}

func TestGrpToGR_RangeBoundaries(t *testing.T) {
	// Test that boundary transitions work correctly
	boundaries := []struct {
		grp      uint32
		minGR    uint16
		maxGR    uint16
		rangeEnd uint32
	}{
		{208749, 1, 50, 208750},
		{208750, 51, 99, 593400},
		{593399, 51, 99, 593400},
		{593400, 100, 149, 993400},
		{993399, 100, 149, 993400},
		{993400, 150, 199, 1400900},
	}

	for _, b := range boundaries {
		result := grpToGR(b.grp)
		if result < b.minGR || result > b.maxGR {
			t.Errorf("grpToGR(%d) = %d, expected range [%d, %d]", b.grp, result, b.minGR, b.maxGR)
		}
	}
}
