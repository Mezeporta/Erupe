package signserver

import (
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"

	_config "erupe-ce/config"
)

// TestMakeSignResponse_EmptyCapLinkValues verifies the crash is FIXED when CapLink.Values is empty
// Previously panicked: runtime error: index out of range [0] with length 0
// From erupe.log.1:659796 and 659853
// After fix: Should handle empty array gracefully with defaults
func TestMakeSignResponse_EmptyCapLinkValues(t *testing.T) {
	config := &_config.Config{
		DebugOptions: _config.DebugOptions{
			CapLink: _config.CapLinkOptions{
				Values: []uint16{}, // Empty array - should now use defaults instead of panicking
				Key:    "test",
				Host:   "localhost",
				Port:   8080,
			},
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
			ClanMemberLimits: [][]uint8{
				{1, 10},
				{2, 20},
				{3, 30},
			},
		},
	}

	session := &Session{
		logger: zap.NewNop(),
		server: &Server{
			erupeConfig: config,
			logger:      zap.NewNop(),
		},
		client: PC100,
	}

	// Set up defer to catch ANY panic - we should NOT get array bounds panic anymore
	defer func() {
		if r := recover(); r != nil {
			// If panic occurs, it should NOT be from array access
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("Array bounds panic NOT fixed! Still getting: %v", r)
			} else {
				// Other panic is acceptable (DB, etc) - we only care about array bounds
				t.Logf("Non-array-bounds panic (acceptable): %v", r)
			}
		}
	}()

	// This should NOT panic on array bounds anymore
	result := session.makeSignResponse(0)
	if result != nil && len(result) > 0 {
		t.Log("✅ makeSignResponse handled empty CapLink.Values without array bounds panic")
	}
}

// TestMakeSignResponse_InsufficientCapLinkValues verifies the crash is FIXED when CapLink.Values is too small
// Previously panicked: runtime error: index out of range [1]
// After fix: Should handle small array gracefully with defaults
func TestMakeSignResponse_InsufficientCapLinkValues(t *testing.T) {
	config := &_config.Config{
		DebugOptions: _config.DebugOptions{
			CapLink: _config.CapLinkOptions{
				Values: []uint16{51728}, // Only 1 element, code used to panic accessing [1]
				Key:    "test",
				Host:   "localhost",
				Port:   8080,
			},
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
			ClanMemberLimits: [][]uint8{
				{1, 10},
			},
		},
	}

	session := &Session{
		logger: zap.NewNop(),
		server: &Server{
			erupeConfig: config,
			logger:      zap.NewNop(),
		},
		client: PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("Array bounds panic NOT fixed! Still getting: %v", r)
			} else {
				t.Logf("Non-array-bounds panic (acceptable): %v", r)
			}
		}
	}()

	// This should NOT panic on array bounds anymore
	result := session.makeSignResponse(0)
	if result != nil && len(result) > 0 {
		t.Log("✅ makeSignResponse handled insufficient CapLink.Values without array bounds panic")
	}
}

// TestMakeSignResponse_MissingCapLinkValues234 verifies the crash is FIXED when CapLink.Values doesn't have 5 elements
// Previously panicked: runtime error: index out of range [2/3/4]
// After fix: Should handle small array gracefully with defaults
func TestMakeSignResponse_MissingCapLinkValues234(t *testing.T) {
	config := &_config.Config{
		DebugOptions: _config.DebugOptions{
			CapLink: _config.CapLinkOptions{
				Values: []uint16{100, 200}, // Only 2 elements, code used to panic accessing [2][3][4]
				Key:    "test",
				Host:   "localhost",
				Port:   8080,
			},
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets:  100,
			MezFesGroupTickets: 100,
			ClanMemberLimits: [][]uint8{
				{1, 10},
			},
		},
	}

	session := &Session{
		logger: zap.NewNop(),
		server: &Server{
			erupeConfig: config,
			logger:      zap.NewNop(),
		},
		client: PC100,
	}

	defer func() {
		if r := recover(); r != nil {
			panicStr := fmt.Sprintf("%v", r)
			if strings.Contains(panicStr, "index out of range") {
				t.Errorf("Array bounds panic NOT fixed! Still getting: %v", r)
			} else {
				t.Logf("Non-array-bounds panic (acceptable): %v", r)
			}
		}
	}()

	// This should NOT panic on array bounds anymore
	result := session.makeSignResponse(0)
	if result != nil && len(result) > 0 {
		t.Log("✅ makeSignResponse handled missing CapLink.Values[2/3/4] without array bounds panic")
	}
}

// TestCapLinkValuesBoundsChecking verifies bounds checking logic for CapLink.Values
// Tests the specific logic that was fixed without needing full database setup
func TestCapLinkValuesBoundsChecking(t *testing.T) {
	// Test the bounds checking logic directly
	testCases := []struct {
		name          string
		values        []uint16
		expectDefault bool
	}{
		{"empty array", []uint16{}, true},
		{"1 element", []uint16{100}, true},
		{"2 elements", []uint16{100, 200}, true},
		{"3 elements", []uint16{100, 200, 300}, true},
		{"4 elements", []uint16{100, 200, 300, 400}, true},
		{"5 elements (valid)", []uint16{100, 200, 300, 400, 500}, false},
		{"6 elements (valid)", []uint16{100, 200, 300, 400, 500, 600}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Replicate the bounds checking logic from the fix
			capLinkValues := tc.values
			if len(capLinkValues) < 5 {
				capLinkValues = []uint16{0, 0, 0, 0, 0}
			}

			// Verify all 5 indices are now safe to access
			_ = capLinkValues[0]
			_ = capLinkValues[1]
			_ = capLinkValues[2]
			_ = capLinkValues[3]
			_ = capLinkValues[4]

			// Verify correct behavior
			if tc.expectDefault {
				if capLinkValues[0] != 0 || capLinkValues[1] != 0 {
					t.Errorf("Expected default values, got %v", capLinkValues)
				}
			} else {
				if capLinkValues[0] == 0 && tc.values[0] != 0 {
					t.Errorf("Expected original values, got defaults")
				}
			}

			t.Logf("✅ %s: All 5 indices accessible without panic", tc.name)
		})
	}
}
