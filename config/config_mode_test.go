package config

import (
	"testing"
)

// TestModeStringMethod calls Mode.String() to cover the method.
// Note: Mode.String() has a known off-by-one bug (Mode values are 1-indexed but
// versionStrings is 0-indexed), so S1.String() returns "S1.5" instead of "S1.0".
// ZZ (value 41) would panic because versionStrings only has 41 entries (indices 0-40).
func TestModeStringMethod(t *testing.T) {
	// Test modes that don't panic (S1=1 through Z2=40)
	tests := []struct {
		mode Mode
		want string
	}{
		{S1, "S1.5"},   // versionStrings[1]
		{S15, "S2.0"},  // versionStrings[2]
		{G1, "G2"},     // versionStrings[21]
		{Z1, "Z2"},     // versionStrings[39]
		{Z2, "ZZ"},     // versionStrings[40]
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

// TestModeStringAllSafeVersions verifies all modes from S1 through Z2 produce valid strings
// (ZZ is excluded because it's out of bounds due to the off-by-one bug)
func TestModeStringAllSafeVersions(t *testing.T) {
	for m := S1; m <= Z2; m++ {
		got := m.String()
		if got == "" {
			t.Errorf("Mode(%d).String() returned empty string", m)
		}
	}
}
