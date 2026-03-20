package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RengokuRepository centralizes all database access for the rengoku_score table.
type RengokuRepository struct {
	db *sqlx.DB
}

// NewRengokuRepository creates a new RengokuRepository.
func NewRengokuRepository(db *sqlx.DB) *RengokuRepository {
	return &RengokuRepository{db: db}
}

// UpsertScore ensures a rengoku_score row exists for the character and updates it.
func (r *RengokuRepository) UpsertScore(charID uint32, maxStagesMp, maxPointsMp, maxStagesSp, maxPointsSp uint32) error {
	var t int
	err := r.db.QueryRow("SELECT character_id FROM rengoku_score WHERE character_id=$1", charID).Scan(&t)
	if err != nil {
		if _, err := r.db.Exec("INSERT INTO rengoku_score (character_id) VALUES ($1)", charID); err != nil {
			return fmt.Errorf("insert rengoku_score: %w", err)
		}
	}
	if _, err := r.db.Exec(
		"UPDATE rengoku_score SET max_stages_mp=$1, max_points_mp=$2, max_stages_sp=$3, max_points_sp=$4 WHERE character_id=$5",
		maxStagesMp, maxPointsMp, maxStagesSp, maxPointsSp, charID,
	); err != nil {
		return fmt.Errorf("update rengoku_score: %w", err)
	}
	return nil
}

// rengokuScoreQuery is the shared FROM/JOIN clause for ranking queries.
const rengokuScoreQueryRepo = `, c.name FROM rengoku_score rs
LEFT JOIN characters c ON c.id = rs.character_id
LEFT JOIN guild_characters gc ON gc.character_id = rs.character_id `

// rengokuColumnForLeaderboard maps a leaderboard index to the score column name.
func rengokuColumnForLeaderboard(leaderboard uint32) string {
	switch leaderboard {
	case 0, 2:
		return "max_stages_mp"
	case 1, 3:
		return "max_points_mp"
	case 4, 6:
		return "max_stages_sp"
	case 5, 7:
		return "max_points_sp"
	default:
		return "max_stages_mp"
	}
}

// rengokuIsGuildFiltered returns true if the leaderboard index is guild-scoped.
func rengokuIsGuildFiltered(leaderboard uint32) bool {
	return leaderboard == 2 || leaderboard == 3 || leaderboard == 6 || leaderboard == 7
}

// GetRanking returns rengoku scores for the given leaderboard.
// For guild-scoped leaderboards (2,3,6,7), guildID filters the results.
func (r *RengokuRepository) GetRanking(leaderboard uint32, guildID uint32) ([]RengokuScore, error) {
	col := rengokuColumnForLeaderboard(leaderboard)
	var result []RengokuScore
	var err error
	if rengokuIsGuildFiltered(leaderboard) {
		err = r.db.Select(&result,
			fmt.Sprintf("SELECT %s AS score %s WHERE guild_id=$1 ORDER BY %s DESC", col, rengokuScoreQueryRepo, col),
			guildID,
		)
	} else {
		err = r.db.Select(&result,
			fmt.Sprintf("SELECT %s AS score %s ORDER BY %s DESC", col, rengokuScoreQueryRepo, col),
		)
	}
	return result, err
}
