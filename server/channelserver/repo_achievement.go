package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// AchievementRepository centralizes all database access for the achievements table.
type AchievementRepository struct {
	db *sqlx.DB
}

// NewAchievementRepository creates a new AchievementRepository.
func NewAchievementRepository(db *sqlx.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

// EnsureExists creates an achievements record for the character if one doesn't exist.
func (r *AchievementRepository) EnsureExists(charID uint32) error {
	_, err := r.db.Exec("INSERT INTO achievements (id) VALUES ($1) ON CONFLICT DO NOTHING", charID)
	return err
}

// GetAllScores returns all 33 achievement scores for a character.
func (r *AchievementRepository) GetAllScores(charID uint32) ([33]int32, error) {
	var scores [33]int32
	err := r.db.QueryRow("SELECT * FROM achievements WHERE id=$1", charID).Scan(&scores[0],
		&scores[0], &scores[1], &scores[2], &scores[3], &scores[4], &scores[5], &scores[6], &scores[7], &scores[8],
		&scores[9], &scores[10], &scores[11], &scores[12], &scores[13], &scores[14], &scores[15], &scores[16],
		&scores[17], &scores[18], &scores[19], &scores[20], &scores[21], &scores[22], &scores[23], &scores[24],
		&scores[25], &scores[26], &scores[27], &scores[28], &scores[29], &scores[30], &scores[31], &scores[32])
	return scores, err
}

// IncrementScore increments the score for a specific achievement column.
// achievementID must be in the range [0, 32] to prevent SQL injection.
func (r *AchievementRepository) IncrementScore(charID uint32, achievementID uint8) error {
	if achievementID > 32 {
		return fmt.Errorf("achievement ID %d out of range [0, 32]", achievementID)
	}
	_, err := r.db.Exec(fmt.Sprintf("UPDATE achievements SET ach%d=ach%d+1 WHERE id=$1", achievementID, achievementID), charID)
	return err
}
