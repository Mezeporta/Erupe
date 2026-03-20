package channelserver

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserRepository centralizes all database access for the users table.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Gacha/Currency methods

// GetGachaPoints returns the user's frontier points, premium gacha coins, and trial gacha coins.
func (r *UserRepository) GetGachaPoints(userID uint32) (fp, premium, trial uint32, err error) {
	err = r.db.QueryRow(
		`SELECT COALESCE(frontier_points, 0), COALESCE(gacha_premium, 0), COALESCE(gacha_trial, 0) FROM users WHERE id=$1`,
		userID,
	).Scan(&fp, &premium, &trial)
	return
}

// GetTrialCoins returns the user's trial gacha coin balance.
func (r *UserRepository) GetTrialCoins(userID uint32) (uint16, error) {
	var balance uint16
	err := r.db.QueryRow(`SELECT COALESCE(gacha_trial, 0) FROM users WHERE id=$1`, userID).Scan(&balance)
	return balance, err
}

// DeductTrialCoins subtracts the given amount from the user's trial gacha coins.
func (r *UserRepository) DeductTrialCoins(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET gacha_trial=gacha_trial-$1 WHERE id=$2`, amount, userID)
	return err
}

// DeductPremiumCoins subtracts the given amount from the user's premium gacha coins.
func (r *UserRepository) DeductPremiumCoins(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET gacha_premium=gacha_premium-$1 WHERE id=$2`, amount, userID)
	return err
}

// AddPremiumCoins adds the given amount to the user's premium gacha coins.
func (r *UserRepository) AddPremiumCoins(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET gacha_premium=gacha_premium+$1 WHERE id=$2`, amount, userID)
	return err
}

// AddTrialCoins adds the given amount to the user's trial gacha coins.
func (r *UserRepository) AddTrialCoins(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET gacha_trial=gacha_trial+$1 WHERE id=$2`, amount, userID)
	return err
}

// DeductFrontierPoints subtracts the given amount from the user's frontier points.
func (r *UserRepository) DeductFrontierPoints(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET frontier_points=frontier_points-$1 WHERE id=$2`, amount, userID)
	return err
}

// AddFrontierPoints adds the given amount to the user's frontier points.
func (r *UserRepository) AddFrontierPoints(userID uint32, amount uint32) error {
	_, err := r.db.Exec(`UPDATE users SET frontier_points=frontier_points+$1 WHERE id=$2`, amount, userID)
	return err
}

// AdjustFrontierPointsDeduct atomically deducts frontier points and returns the new balance.
func (r *UserRepository) AdjustFrontierPointsDeduct(userID uint32, amount int) (uint32, error) {
	var balance uint32
	err := r.db.QueryRow(
		`UPDATE users SET frontier_points=frontier_points::int - $1 WHERE id=$2 RETURNING frontier_points`,
		amount, userID,
	).Scan(&balance)
	return balance, err
}

// AdjustFrontierPointsCredit atomically credits frontier points and returns the new balance.
func (r *UserRepository) AdjustFrontierPointsCredit(userID uint32, amount int) (uint32, error) {
	var balance uint32
	err := r.db.QueryRow(
		`UPDATE users SET frontier_points=COALESCE(frontier_points::int + $1, $1) WHERE id=$2 RETURNING frontier_points`,
		amount, userID,
	).Scan(&balance)
	return balance, err
}

// AddFrontierPointsFromGacha awards frontier points from a gacha entry's defined value.
func (r *UserRepository) AddFrontierPointsFromGacha(userID uint32, gachaID uint32, entryType uint8) error {
	_, err := r.db.Exec(
		`UPDATE users SET frontier_points=frontier_points+(SELECT frontier_points FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2) WHERE id=$3`,
		gachaID, entryType, userID,
	)
	return err
}

// Rights/Permissions methods

// GetRights returns the user's rights bitmask.
func (r *UserRepository) GetRights(userID uint32) (uint32, error) {
	var rights uint32
	err := r.db.QueryRow(`SELECT rights FROM users WHERE id=$1`, userID).Scan(&rights)
	return rights, err
}

// SetRights sets the user's rights bitmask.
func (r *UserRepository) SetRights(userID uint32, rights uint32) error {
	_, err := r.db.Exec(`UPDATE users SET rights=$1 WHERE id=$2`, rights, userID)
	return err
}

// IsOp returns whether the user has operator privileges.
func (r *UserRepository) IsOp(userID uint32) (bool, error) {
	var op bool
	err := r.db.QueryRow(`SELECT op FROM users WHERE id=$1`, userID).Scan(&op)
	if err != nil {
		return false, err
	}
	return op, nil
}

// User metadata methods

// SetLastCharacter records the last-played character for a user.
func (r *UserRepository) SetLastCharacter(userID uint32, charID uint32) error {
	_, err := r.db.Exec(`UPDATE users SET last_character=$1 WHERE id=$2`, charID, userID)
	return err
}

// GetTimer returns whether the user has the quest timer display enabled.
func (r *UserRepository) GetTimer(userID uint32) (bool, error) {
	var timer bool
	err := r.db.QueryRow(`SELECT COALESCE(timer, false) FROM users WHERE id=$1`, userID).Scan(&timer)
	return timer, err
}

// SetTimer sets the user's quest timer display preference.
func (r *UserRepository) SetTimer(userID uint32, value bool) error {
	_, err := r.db.Exec(`UPDATE users SET timer=$1 WHERE id=$2`, value, userID)
	return err
}

// CountByPSNID returns the number of users with the given PSN ID.
func (r *UserRepository) CountByPSNID(psnID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(*) FROM users WHERE psn_id = $1`, psnID).Scan(&count)
	return count, err
}

// SetPSNID associates a PSN ID with the user's account.
func (r *UserRepository) SetPSNID(userID uint32, psnID string) error {
	_, err := r.db.Exec(`UPDATE users SET psn_id=$1 WHERE id=$2`, psnID, userID)
	return err
}

// GetDiscordToken returns the user's discord link token.
func (r *UserRepository) GetDiscordToken(userID uint32) (string, error) {
	var token string
	err := r.db.QueryRow(`SELECT discord_token FROM users WHERE id=$1`, userID).Scan(&token)
	return token, err
}

// SetDiscordToken sets the user's discord link token.
func (r *UserRepository) SetDiscordToken(userID uint32, token string) error {
	_, err := r.db.Exec(`UPDATE users SET discord_token = $1 WHERE id=$2`, token, userID)
	return err
}

// Warehouse methods

// GetItemBox returns the user's serialized warehouse item data.
func (r *UserRepository) GetItemBox(userID uint32) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(`SELECT item_box FROM users WHERE id=$1`, userID).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

// SetItemBox persists the user's warehouse item data.
func (r *UserRepository) SetItemBox(userID uint32, data []byte) error {
	_, err := r.db.Exec(`UPDATE users SET item_box=$1 WHERE id=$2`, data, userID)
	return err
}

// Discord bot methods (Server-level)

// LinkDiscord associates a Discord user ID with the account matching the given token.
// Returns the discord_id on success.
func (r *UserRepository) LinkDiscord(discordID string, token string) (string, error) {
	var result string
	err := r.db.QueryRow(
		`UPDATE users SET discord_id = $1 WHERE discord_token = $2 RETURNING discord_id`,
		discordID, token,
	).Scan(&result)
	return result, err
}

// SetPasswordByDiscordID updates the password for the user linked to the given Discord ID.
func (r *UserRepository) SetPasswordByDiscordID(discordID string, hash []byte) error {
	_, err := r.db.Exec(`UPDATE users SET password = $1 WHERE discord_id = $2`, hash, discordID)
	return err
}

// Auth methods

// GetByIDAndUsername resolves a character ID to the owning user's ID and username.
func (r *UserRepository) GetByIDAndUsername(charID uint32) (userID uint32, username string, err error) {
	err = r.db.QueryRow(
		`SELECT id, username FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`,
		charID,
	).Scan(&userID, &username)
	return
}

// BanUser inserts or updates a ban for the given user.
// A nil expires means a permanent ban; non-nil sets a temporary ban with expiry.
func (r *UserRepository) BanUser(userID uint32, expires *time.Time) error {
	if expires == nil {
		_, err := r.db.Exec(`INSERT INTO bans VALUES ($1)
			ON CONFLICT (user_id) DO UPDATE SET expires=NULL`, userID)
		return err
	}
	_, err := r.db.Exec(`INSERT INTO bans VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET expires=$2`, userID, *expires)
	return err
}
