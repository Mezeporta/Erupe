package signserver

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// SignUserRepository implements SignUserRepo with PostgreSQL.
type SignUserRepository struct {
	db *sqlx.DB
}

// NewSignUserRepository creates a new SignUserRepository.
func NewSignUserRepository(db *sqlx.DB) *SignUserRepository {
	return &SignUserRepository{db: db}
}

func (r *SignUserRepository) GetCredentials(username string) (uint32, string, error) {
	var uid uint32
	var passwordHash string
	err := r.db.QueryRow(`SELECT id, password FROM users WHERE username = $1`, username).Scan(&uid, &passwordHash)
	return uid, passwordHash, err
}

func (r *SignUserRepository) Register(username, passwordHash string, returnExpires time.Time) (uint32, error) {
	var uid uint32
	err := r.db.QueryRow(
		"INSERT INTO users (username, password, return_expires) VALUES ($1, $2, $3) RETURNING id",
		username, passwordHash, returnExpires,
	).Scan(&uid)
	return uid, err
}

func (r *SignUserRepository) GetRights(uid uint32) (uint32, error) {
	var rights uint32
	err := r.db.QueryRow("SELECT rights FROM users WHERE id=$1", uid).Scan(&rights)
	return rights, err
}

func (r *SignUserRepository) GetLastCharacter(uid uint32) (uint32, error) {
	var lastPlayed uint32
	err := r.db.QueryRow("SELECT last_character FROM users WHERE id=$1", uid).Scan(&lastPlayed)
	return lastPlayed, err
}

func (r *SignUserRepository) GetLastLogin(uid uint32) (time.Time, error) {
	var lastLogin time.Time
	err := r.db.Get(&lastLogin, "SELECT COALESCE(last_login, now()) FROM users WHERE id=$1", uid)
	return lastLogin, err
}

func (r *SignUserRepository) GetReturnExpiry(uid uint32) (time.Time, error) {
	var expiry time.Time
	err := r.db.Get(&expiry, "SELECT return_expires FROM users WHERE id=$1", uid)
	return expiry, err
}

func (r *SignUserRepository) UpdateReturnExpiry(uid uint32, expiry time.Time) error {
	_, err := r.db.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", expiry, uid)
	return err
}

func (r *SignUserRepository) UpdateLastLogin(uid uint32, loginTime time.Time) error {
	_, err := r.db.Exec("UPDATE users SET last_login=$1 WHERE id=$2", loginTime, uid)
	return err
}

func (r *SignUserRepository) CountPermanentBans(uid uint32) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires IS NULL`, uid).Scan(&count)
	return count, err
}

func (r *SignUserRepository) CountActiveBans(uid uint32) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires > now()`, uid).Scan(&count)
	return count, err
}

func (r *SignUserRepository) GetByWiiUKey(wiiuKey string) (uint32, error) {
	var uid uint32
	err := r.db.QueryRow(`SELECT id FROM users WHERE wiiu_key = $1`, wiiuKey).Scan(&uid)
	return uid, err
}

func (r *SignUserRepository) GetByPSNID(psnID string) (uint32, error) {
	var uid uint32
	err := r.db.QueryRow(`SELECT id FROM users WHERE psn_id = $1`, psnID).Scan(&uid)
	return uid, err
}

func (r *SignUserRepository) CountByPSNID(psnID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(*) FROM users WHERE psn_id = $1`, psnID).Scan(&count)
	return count, err
}

func (r *SignUserRepository) GetPSNIDForUsername(username string) (string, error) {
	var psnID string
	err := r.db.QueryRow(`SELECT COALESCE(psn_id, '') FROM users WHERE username = $1`, username).Scan(&psnID)
	return psnID, err
}

func (r *SignUserRepository) SetPSNID(username, psnID string) error {
	_, err := r.db.Exec(`UPDATE users SET psn_id = $1 WHERE username = $2`, psnID, username)
	return err
}

func (r *SignUserRepository) GetPSNIDForUser(uid uint32) (string, error) {
	var psnID string
	err := r.db.QueryRow("SELECT psn_id FROM users WHERE id = $1", uid).Scan(&psnID)
	return psnID, err
}
