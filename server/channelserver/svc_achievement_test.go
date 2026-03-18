package channelserver

import (
	"testing"

	"go.uber.org/zap"
)

func newTestAchievementService(repo AchievementRepo) *AchievementService {
	logger, _ := zap.NewDevelopment()
	return NewAchievementService(repo, logger)
}

func TestAchievementService_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		scores     [33]int32
		scoresErr  error
		wantErr    bool
		wantPoints uint32
	}{
		{
			name:       "all zeros",
			scores:     [33]int32{},
			wantPoints: 0,
		},
		{
			name:       "some scores",
			scores:     [33]int32{5, 0, 20},
			wantPoints: 5 + 0 + 15, // id0: level1=5pts, id1: level0=0pts, id2: level1(5)+level2(10)=15pts (score=20, curve[0]={5,15,...}: 20-5=15, 15-15=0 → level2=15pts)
		},
		{
			name:      "db error",
			scoresErr: errNotFound,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAchievementRepo{
				scores:       tt.scores,
				getScoresErr: tt.scoresErr,
			}
			svc := newTestAchievementService(mock)

			summary, err := svc.GetAll(1)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !mock.ensureCalled {
				t.Error("EnsureExists should have been called")
			}
			if summary.Points != tt.wantPoints {
				t.Errorf("Points = %d, want %d", summary.Points, tt.wantPoints)
			}
		})
	}
}

func TestAchievementService_GetAll_NotifyOnNewRankUp(t *testing.T) {
	// Scores: ach0=5 (level 1), ach1=0 (level 0), ach2=20 (level 2)
	// Displayed: ach0 was level 0, ach1 was level 0, ach2 was level 2
	// Expected: ach0 notifies (1 > 0), ach1 does not (level 0), ach2 does not (2 == 2)
	mock := &mockAchievementRepo{
		scores:          [33]int32{5, 0, 20},
		displayedLevels: make([]byte, 33), // all zeros
	}
	mock.displayedLevels[2] = 2 // ach2 was already displayed at level 2

	svc := newTestAchievementService(mock)
	summary, err := svc.GetAll(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !summary.Notify[0] {
		t.Error("ach0: expected notify=true (level 1, displayed 0)")
	}
	if summary.Notify[1] {
		t.Error("ach1: expected notify=false (level 0)")
	}
	if summary.Notify[2] {
		t.Error("ach2: expected notify=false (level 2, displayed 2)")
	}
}

func TestAchievementService_GetAll_NotifyAllWhenNeverDisplayed(t *testing.T) {
	// No displayed levels (nil) — all achievements with level > 0 should notify.
	mock := &mockAchievementRepo{
		scores:       [33]int32{5, 15},
		displayedErr: errNotFound,
	}
	svc := newTestAchievementService(mock)
	summary, err := svc.GetAll(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !summary.Notify[0] {
		t.Error("ach0: expected notify=true (never displayed)")
	}
	if !summary.Notify[1] {
		t.Error("ach1: expected notify=true (never displayed)")
	}
	if summary.Notify[2] {
		t.Error("ach2: expected notify=false (level 0)")
	}
}

func TestAchievementService_MarkDisplayed(t *testing.T) {
	mock := &mockAchievementRepo{
		scores: [33]int32{5, 0, 20}, // ach0=level1, ach1=level0, ach2=level2
	}
	svc := newTestAchievementService(mock)

	err := svc.MarkDisplayed(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !mock.ensureCalled {
		t.Error("EnsureExists should have been called")
	}
	if mock.savedLevels == nil {
		t.Fatal("SaveDisplayedLevels should have been called")
	}
	if len(mock.savedLevels) != 33 {
		t.Fatalf("Expected 33 bytes, got %d", len(mock.savedLevels))
	}
	if mock.savedLevels[0] != 1 {
		t.Errorf("ach0 level: got %d, want 1", mock.savedLevels[0])
	}
	if mock.savedLevels[1] != 0 {
		t.Errorf("ach1 level: got %d, want 0", mock.savedLevels[1])
	}
	if mock.savedLevels[2] != 2 {
		t.Errorf("ach2 level: got %d, want 2", mock.savedLevels[2])
	}
}

func TestAchievementService_GetAll_EnsureErrorNonFatal(t *testing.T) {
	mock := &mockAchievementRepo{
		ensureErr: errNotFound,
		scores:    [33]int32{},
	}
	svc := newTestAchievementService(mock)

	summary, err := svc.GetAll(1)
	if err != nil {
		t.Fatalf("EnsureExists error should not propagate: %v", err)
	}
	if summary == nil {
		t.Fatal("Summary should not be nil")
	}
}

func TestAchievementService_GetAll_AchievementCount(t *testing.T) {
	mock := &mockAchievementRepo{scores: [33]int32{}}
	svc := newTestAchievementService(mock)

	summary, err := svc.GetAll(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify all 33 entries are populated
	for id := uint8(0); id < 33; id++ {
		// At score 0, every achievement should be level 0
		if summary.Achievements[id].Level != 0 {
			t.Errorf("Achievement[%d].Level = %d, want 0", id, summary.Achievements[id].Level)
		}
	}
}

func TestAchievementService_Increment(t *testing.T) {
	tests := []struct {
		name          string
		achievementID uint8
		incrementErr  error
		wantErr       bool
		wantEnsure    bool
		wantIncID     uint8
	}{
		{
			name:          "valid ID",
			achievementID: 5,
			wantEnsure:    true,
			wantIncID:     5,
		},
		{
			name:          "boundary ID 0",
			achievementID: 0,
			wantEnsure:    true,
			wantIncID:     0,
		},
		{
			name:          "boundary ID 32",
			achievementID: 32,
			wantEnsure:    true,
			wantIncID:     32,
		},
		{
			name:          "out of range",
			achievementID: 33,
			wantErr:       true,
		},
		{
			name:          "repo error",
			achievementID: 5,
			incrementErr:  errNotFound,
			wantErr:       true,
			wantEnsure:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAchievementRepo{
				incrementErr: tt.incrementErr,
			}
			svc := newTestAchievementService(mock)

			err := svc.Increment(1, tt.achievementID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if mock.ensureCalled != tt.wantEnsure {
				t.Errorf("EnsureExists called = %v, want %v", mock.ensureCalled, tt.wantEnsure)
			}
			if mock.incrementedID != tt.wantIncID {
				t.Errorf("IncrementScore ID = %d, want %d", mock.incrementedID, tt.wantIncID)
			}
		})
	}
}
