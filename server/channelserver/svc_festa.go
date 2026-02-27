package channelserver

import (
	"time"

	"go.uber.org/zap"
)

// FestaService encapsulates festa business logic, sitting between handlers and repos.
type FestaService struct {
	festaRepo FestaRepo
	logger    *zap.Logger
}

// NewFestaService creates a new FestaService.
func NewFestaService(fr FestaRepo, log *zap.Logger) *FestaService {
	return &FestaService{
		festaRepo: fr,
		logger:    log,
	}
}

// EnsureActiveEvent checks whether the current festa event is still active.
// If it has expired or none exists, all festa state is cleaned up and a new
// event is created starting at the next midnight. Returns the (possibly new)
// start time.
func (svc *FestaService) EnsureActiveEvent(currentStart uint32, now time.Time, nextMidnight time.Time) (uint32, error) {
	if currentStart != 0 && now.Unix() <= int64(currentStart)+festaEventLifespan {
		return currentStart, nil
	}

	if err := svc.festaRepo.CleanupAll(); err != nil {
		svc.logger.Error("Failed to cleanup festa", zap.Error(err))
		return 0, err
	}

	newStart := uint32(nextMidnight.Unix())
	if err := svc.festaRepo.InsertEvent(newStart); err != nil {
		svc.logger.Error("Failed to insert festa event", zap.Error(err))
		return 0, err
	}

	return newStart, nil
}

// SubmitSouls filters out zero-value soul entries and records the remaining
// submissions for the character. Returns nil if all entries are zero.
func (svc *FestaService) SubmitSouls(charID, guildID uint32, souls []uint16) error {
	hasNonZero := false
	for _, s := range souls {
		if s != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		return nil
	}
	return svc.festaRepo.SubmitSouls(charID, guildID, souls)
}
