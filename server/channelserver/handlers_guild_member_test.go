package channelserver

import (
	"testing"
)

func TestGuildMember_CanRecruit(t *testing.T) {
	tests := []struct {
		name     string
		member   GuildMember
		expected bool
	}{
		{
			name: "recruiter flag true",
			member: GuildMember{
				Recruiter:  true,
				OrderIndex: 10,
				IsLeader:   false,
			},
			expected: true,
		},
		{
			name: "order index 1",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 1,
				IsLeader:   false,
			},
			expected: true,
		},
		{
			name: "order index 2",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 2,
				IsLeader:   false,
			},
			expected: true,
		},
		{
			name: "order index 3",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 3,
				IsLeader:   false,
			},
			expected: true,
		},
		{
			name: "order index 0 (sub-leader)",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 0,
				IsLeader:   false,
			},
			expected: true,
		},
		{
			name: "order index 4 cannot recruit",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 4,
				IsLeader:   false,
			},
			expected: false,
		},
		{
			name: "order index 5 cannot recruit",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 5,
				IsLeader:   false,
			},
			expected: false,
		},
		{
			name: "is leader can recruit",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 100,
				IsLeader:   true,
			},
			expected: true,
		},
		{
			name: "regular member cannot recruit",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 10,
				IsLeader:   false,
			},
			expected: false,
		},
		{
			name: "all flags true",
			member: GuildMember{
				Recruiter:  true,
				OrderIndex: 1,
				IsLeader:   true,
			},
			expected: true,
		},
		{
			name: "high order index with leader",
			member: GuildMember{
				Recruiter:  false,
				OrderIndex: 255,
				IsLeader:   true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.member.CanRecruit()
			if result != tt.expected {
				t.Errorf("CanRecruit() = %v, expected %v (Recruiter=%v, OrderIndex=%d, IsLeader=%v)",
					result, tt.expected, tt.member.Recruiter, tt.member.OrderIndex, tt.member.IsLeader)
			}
		})
	}
}

func TestGuildMember_IsSubLeader(t *testing.T) {
	tests := []struct {
		name       string
		orderIndex uint8
		expected   bool
	}{
		{
			name:       "order index 0",
			orderIndex: 0,
			expected:   true,
		},
		{
			name:       "order index 1",
			orderIndex: 1,
			expected:   true,
		},
		{
			name:       "order index 2",
			orderIndex: 2,
			expected:   true,
		},
		{
			name:       "order index 3",
			orderIndex: 3,
			expected:   true,
		},
		{
			name:       "order index 4",
			orderIndex: 4,
			expected:   false,
		},
		{
			name:       "order index 5",
			orderIndex: 5,
			expected:   false,
		},
		{
			name:       "order index 100",
			orderIndex: 100,
			expected:   false,
		},
		{
			name:       "order index 255",
			orderIndex: 255,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := GuildMember{OrderIndex: tt.orderIndex}
			result := member.IsSubLeader()
			if result != tt.expected {
				t.Errorf("IsSubLeader() with OrderIndex=%d = %v, expected %v",
					tt.orderIndex, result, tt.expected)
			}
		})
	}
}

func TestGuildMember_CanRecruit_Priority(t *testing.T) {
	// Test that Recruiter flag takes priority (short-circuit)
	member := GuildMember{
		Recruiter:  true,
		OrderIndex: 100, // Would fail OrderIndex check
		IsLeader:   false,
	}

	if !member.CanRecruit() {
		t.Error("Recruiter flag should allow recruiting regardless of OrderIndex")
	}
}

func TestGuildMember_CanRecruit_OrderIndexBoundary(t *testing.T) {
	// Test the exact boundary at OrderIndex == 3 vs 4
	member3 := GuildMember{Recruiter: false, OrderIndex: 3, IsLeader: false}
	member4 := GuildMember{Recruiter: false, OrderIndex: 4, IsLeader: false}

	if !member3.CanRecruit() {
		t.Error("OrderIndex 3 should be able to recruit")
	}
	if member4.CanRecruit() {
		t.Error("OrderIndex 4 should NOT be able to recruit")
	}
}
