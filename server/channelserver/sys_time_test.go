package channelserver

import "testing"

func TestTimeGameAbsolute(t *testing.T) {
	result := TimeGameAbsolute()

	// TimeGameAbsolute returns (adjustedUnix - 2160) % 5760
	// Result should be in range [0, 5760)
	if result >= 5760 {
		t.Errorf("TimeGameAbsolute() = %d, should be < 5760", result)
	}
}
