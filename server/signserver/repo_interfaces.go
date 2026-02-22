package signserver

import "time"

// Repository interfaces decouple sign server business logic from concrete
// PostgreSQL implementations, enabling mock/stub injection for unit tests.

// character represents a player character record from the characters table.
type character struct {
	ID             uint32 `db:"id"`
	IsFemale       bool   `db:"is_female"`
	IsNewCharacter bool   `db:"is_new_character"`
	Name           string `db:"name"`
	UnkDescString  string `db:"unk_desc_string"`
	HR             uint16 `db:"hr"`
	GR             uint16 `db:"gr"`
	WeaponType     uint16 `db:"weapon_type"`
	LastLogin      uint32 `db:"last_login"`
}

// members represents a friend or guildmate entry used in the sign response.
type members struct {
	CID  uint32 // Local character ID
	ID   uint32 `db:"id"`
	Name string `db:"name"`
}

// SignUserRepo defines the contract for user-related data access (users, bans tables).
type SignUserRepo interface {
	GetCredentials(username string) (uid uint32, passwordHash string, err error)
	Register(username, passwordHash string, returnExpires time.Time) (uint32, error)
	GetRights(uid uint32) (uint32, error)
	GetLastCharacter(uid uint32) (uint32, error)
	GetLastLogin(uid uint32) (time.Time, error)
	GetReturnExpiry(uid uint32) (time.Time, error)
	UpdateReturnExpiry(uid uint32, expiry time.Time) error
	UpdateLastLogin(uid uint32, loginTime time.Time) error
	CountPermanentBans(uid uint32) (int, error)
	CountActiveBans(uid uint32) (int, error)
	GetByWiiUKey(wiiuKey string) (uint32, error)
	GetByPSNID(psnID string) (uint32, error)
	CountByPSNID(psnID string) (int, error)
	GetPSNIDForUsername(username string) (string, error)
	SetPSNID(username, psnID string) error
	GetPSNIDForUser(uid uint32) (string, error)
}

// SignCharacterRepo defines the contract for character data access.
type SignCharacterRepo interface {
	CountNewCharacters(uid uint32) (int, error)
	CreateCharacter(uid uint32, lastLogin uint32) error
	GetForUser(uid uint32) ([]character, error)
	IsNewCharacter(cid int) (bool, error)
	HardDelete(cid int) error
	SoftDelete(cid int) error
	GetFriends(charID uint32) ([]members, error)
	GetGuildmates(charID uint32) ([]members, error)
}

// SignSessionRepo defines the contract for sign session/token data access.
type SignSessionRepo interface {
	RegisterUID(uid uint32, token string) (tokenID uint32, err error)
	RegisterPSN(psnID, token string) (tokenID uint32, err error)
	Validate(token string, tokenID uint32) (bool, error)
	GetPSNIDByToken(token string) (string, error)
}
