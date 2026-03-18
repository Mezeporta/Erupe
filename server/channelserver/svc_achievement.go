package channelserver

import (
	"fmt"

	"go.uber.org/zap"
)

// AchievementService encapsulates business logic for the achievement system.
type AchievementService struct {
	achievementRepo AchievementRepo
	logger          *zap.Logger
}

// NewAchievementService creates a new AchievementService.
func NewAchievementService(ar AchievementRepo, log *zap.Logger) *AchievementService {
	return &AchievementService{achievementRepo: ar, logger: log}
}

const achievementEntryCount = uint8(33)

// AchievementSummary holds the computed achievements and total points for a character.
type AchievementSummary struct {
	Points       uint32
	Achievements [33]Achievement
	Notify       [33]bool
}

// GetAll ensures the achievement record exists, fetches all scores, and computes
// the achievement state for every category. Returns the total accumulated points
// and per-category Achievement data.
func (svc *AchievementService) GetAll(charID uint32) (*AchievementSummary, error) {
	if err := svc.achievementRepo.EnsureExists(charID); err != nil {
		svc.logger.Error("Failed to ensure achievements record", zap.Error(err))
	}

	scores, err := svc.achievementRepo.GetAllScores(charID)
	if err != nil {
		return nil, err
	}

	displayed, err := svc.achievementRepo.GetDisplayedLevels(charID)
	if err != nil {
		svc.logger.Debug("No displayed levels found, all rank-ups will notify", zap.Error(err))
	}

	var summary AchievementSummary
	for id := uint8(0); id < achievementEntryCount; id++ {
		ach := GetAchData(id, scores[id])
		summary.Points += ach.Value
		summary.Achievements[id] = ach

		// Notify if current level exceeds the last-displayed level.
		if ach.Level > 0 {
			if displayed == nil || int(id) >= len(displayed) {
				summary.Notify[id] = true
			} else if ach.Level > displayed[id] {
				summary.Notify[id] = true
			}
		}
	}
	return &summary, nil
}

// MarkDisplayed snapshots the current achievement levels so that future
// GET_ACHIEVEMENT responses only notify on new rank-ups since this point.
func (svc *AchievementService) MarkDisplayed(charID uint32) error {
	if err := svc.achievementRepo.EnsureExists(charID); err != nil {
		svc.logger.Error("Failed to ensure achievements record", zap.Error(err))
	}

	scores, err := svc.achievementRepo.GetAllScores(charID)
	if err != nil {
		return err
	}

	levels := make([]byte, achievementEntryCount)
	for id := uint8(0); id < achievementEntryCount; id++ {
		ach := GetAchData(id, scores[id])
		levels[id] = ach.Level
	}
	return svc.achievementRepo.SaveDisplayedLevels(charID, levels)
}

// Increment validates the achievement ID, ensures the record exists, and bumps
// the score for the given achievement category.
func (svc *AchievementService) Increment(charID uint32, achievementID uint8) error {
	if achievementID > 32 {
		return fmt.Errorf("achievement ID %d out of range [0, 32]", achievementID)
	}

	if err := svc.achievementRepo.EnsureExists(charID); err != nil {
		svc.logger.Error("Failed to ensure achievements record", zap.Error(err))
	}

	return svc.achievementRepo.IncrementScore(charID, achievementID)
}
