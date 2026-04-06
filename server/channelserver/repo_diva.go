package channelserver

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

// DivaRepository centralizes all database access for diva defense events.
type DivaRepository struct {
	db *sqlx.DB
}

// NewDivaRepository creates a new DivaRepository.
func NewDivaRepository(db *sqlx.DB) *DivaRepository {
	return &DivaRepository{db: db}
}

// DeleteEvents removes all diva events.
func (r *DivaRepository) DeleteEvents() error {
	_, err := r.db.Exec("DELETE FROM events WHERE event_type='diva'")
	return err
}

// InsertEvent creates a new diva event with the given start epoch.
func (r *DivaRepository) InsertEvent(startEpoch uint32) error {
	_, err := r.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('diva', to_timestamp($1)::timestamp without time zone)", startEpoch)
	return err
}

// DivaEvent represents a diva event row with ID and start_time epoch.
type DivaEvent struct {
	ID        uint32 `db:"id"`
	StartTime uint32 `db:"start_time"`
}

// GetEvents returns all diva events with their ID and start_time epoch.
func (r *DivaRepository) GetEvents() ([]DivaEvent, error) {
	var result []DivaEvent
	err := r.db.Select(&result, "SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='diva'")
	return result, err
}

// AddPoints atomically adds quest and bonus points for a character in a diva event.
func (r *DivaRepository) AddPoints(charID, eventID, questPoints, bonusPoints uint32) error {
	_, err := r.db.Exec(`
		INSERT INTO diva_points (char_id, event_id, quest_points, bonus_points, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (char_id, event_id) DO UPDATE
		SET quest_points = diva_points.quest_points + EXCLUDED.quest_points,
		    bonus_points = diva_points.bonus_points + EXCLUDED.bonus_points,
		    updated_at = now()`,
		charID, eventID, questPoints, bonusPoints)
	return err
}

// GetPoints returns the accumulated quest and bonus points for a character in an event.
func (r *DivaRepository) GetPoints(charID, eventID uint32) (int64, int64, error) {
	var qp, bp int64
	err := r.db.QueryRow(
		"SELECT quest_points, bonus_points FROM diva_points WHERE char_id=$1 AND event_id=$2",
		charID, eventID).Scan(&qp, &bp)
	if err != nil {
		return 0, 0, err
	}
	return qp, bp, nil
}

// GetTotalPoints returns the sum of all players' quest and bonus points for an event.
func (r *DivaRepository) GetTotalPoints(eventID uint32) (int64, int64, error) {
	var qp, bp int64
	err := r.db.QueryRow(
		"SELECT COALESCE(SUM(quest_points),0), COALESCE(SUM(bonus_points),0) FROM diva_points WHERE event_id=$1",
		eventID).Scan(&qp, &bp)
	if err != nil {
		return 0, 0, err
	}
	return qp, bp, nil
}

// GetBeads returns all active bead types from the diva_beads table.
func (r *DivaRepository) GetBeads() ([]int, error) {
	var types []int
	err := r.db.Select(&types, "SELECT type FROM diva_beads ORDER BY id")
	return types, err
}

// AssignBead inserts a bead assignment for a character, replacing any existing one for that bead slot.
func (r *DivaRepository) AssignBead(characterID uint32, beadIndex int, expiry time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO diva_beads_assignment (character_id, bead_index, expiry)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		characterID, beadIndex, expiry)
	return err
}

// AddBeadPoints records a bead point contribution for a character.
func (r *DivaRepository) AddBeadPoints(characterID uint32, beadIndex int, points int) error {
	_, err := r.db.Exec(
		"INSERT INTO diva_beads_points (character_id, bead_index, points) VALUES ($1, $2, $3)",
		characterID, beadIndex, points)
	return err
}

// GetCharacterBeadPoints returns the summed points per bead_index for a character.
func (r *DivaRepository) GetCharacterBeadPoints(characterID uint32) (map[int]int, error) {
	rows, err := r.db.Query(
		"SELECT bead_index, COALESCE(SUM(points),0) FROM diva_beads_points WHERE character_id=$1 GROUP BY bead_index",
		characterID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make(map[int]int)
	for rows.Next() {
		var idx, pts int
		if err := rows.Scan(&idx, &pts); err != nil {
			return nil, err
		}
		result[idx] = pts
	}
	return result, rows.Err()
}

// GetTotalBeadPoints returns the sum of all points across all characters and bead slots.
func (r *DivaRepository) GetTotalBeadPoints() (int64, error) {
	var total int64
	err := r.db.QueryRow("SELECT COALESCE(SUM(points),0) FROM diva_beads_points").Scan(&total)
	return total, err
}

// GetTopBeadPerDay returns the bead_index with the most points contributed on day offset `day`
// (0 = today, 1 = yesterday, etc.). Returns 0 if no data exists for that day.
func (r *DivaRepository) GetTopBeadPerDay(day int) (int, error) {
	var beadIndex int
	err := r.db.QueryRow(`
		SELECT bead_index
		FROM diva_beads_points
		WHERE timestamp >= (NOW() - ($1 + 1) * INTERVAL '1 day')
		  AND timestamp <  (NOW() - $1 * INTERVAL '1 day')
		GROUP BY bead_index
		ORDER BY SUM(points) DESC
		LIMIT 1`,
		day).Scan(&beadIndex)
	if err != nil {
		return 0, nil // no data for this day is not an error
	}
	return beadIndex, nil
}

// CleanupBeads deletes all rows from diva_beads, diva_beads_assignment, and diva_beads_points.
func (r *DivaRepository) CleanupBeads() error {
	if _, err := r.db.Exec("DELETE FROM diva_beads_points"); err != nil {
		return err
	}
	if _, err := r.db.Exec("DELETE FROM diva_beads_assignment"); err != nil {
		return err
	}
	_, err := r.db.Exec("DELETE FROM diva_beads")
	return err
}

// GetPersonalPrizes returns all prize rows with type='personal', ordered by points_req.
func (r *DivaRepository) GetPersonalPrizes() ([]DivaPrize, error) {
	return r.getPrizesByType("personal")
}

// GetGuildPrizes returns all prize rows with type='guild', ordered by points_req.
func (r *DivaRepository) GetGuildPrizes() ([]DivaPrize, error) {
	return r.getPrizesByType("guild")
}

func (r *DivaRepository) getPrizesByType(prizeType string) ([]DivaPrize, error) {
	rows, err := r.db.Query(`
		SELECT id, type, points_req, item_type, item_id, quantity, gr, repeatable
		FROM diva_prizes
		WHERE type=$1
		ORDER BY points_req`,
		prizeType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var prizes []DivaPrize
	for rows.Next() {
		var p DivaPrize
		if err := rows.Scan(&p.ID, &p.Type, &p.PointsReq, &p.ItemType, &p.ItemID, &p.Quantity, &p.GR, &p.Repeatable); err != nil {
			return nil, err
		}
		prizes = append(prizes, p)
	}
	return prizes, rows.Err()
}

// GetCharacterInterceptionPoints returns the interception_points JSON map from guild_characters.
func (r *DivaRepository) GetCharacterInterceptionPoints(characterID uint32) (map[string]int, error) {
	var raw []byte
	err := r.db.QueryRow(
		"SELECT interception_points FROM guild_characters WHERE char_id=$1",
		characterID).Scan(&raw)
	if err != nil {
		return map[string]int{}, nil
	}
	result := make(map[string]int)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &result); err != nil {
			return map[string]int{}, nil
		}
	}
	return result, nil
}

// AddInterceptionPoints increments the interception points for a quest file ID in guild_characters.
func (r *DivaRepository) AddInterceptionPoints(characterID uint32, questFileID int, points int) error {
	_, err := r.db.Exec(`
		UPDATE guild_characters
		SET interception_points = interception_points || jsonb_build_object(
			$2::text,
			COALESCE((interception_points->>$2::text)::int, 0) + $3
		)
		WHERE char_id=$1`,
		characterID, questFileID, points)
	return err
}
