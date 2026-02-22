package api

import (
	"context"
	"time"
)

// Repository interfaces decouple API server business logic from concrete
// PostgreSQL implementations, enabling mock/stub injection for unit tests.

// APIUserRepo defines the contract for user-related data access.
type APIUserRepo interface {
	// Register creates a new user and returns their ID and rights.
	Register(ctx context.Context, username, passwordHash string, returnExpires time.Time) (id uint32, rights uint32, err error)
	// GetCredentials returns the user's ID, password hash, and rights.
	GetCredentials(ctx context.Context, username string) (id uint32, passwordHash string, rights uint32, err error)
	// GetLastLogin returns the user's last login time.
	GetLastLogin(uid uint32) (time.Time, error)
	// GetReturnExpiry returns the user's return expiry time.
	GetReturnExpiry(uid uint32) (time.Time, error)
	// UpdateReturnExpiry sets the user's return expiry time.
	UpdateReturnExpiry(uid uint32, expiry time.Time) error
	// UpdateLastLogin sets the user's last login time.
	UpdateLastLogin(uid uint32, loginTime time.Time) error
}

// APICharacterRepo defines the contract for character-related data access.
type APICharacterRepo interface {
	// GetNewCharacter returns an existing new (unfinished) character for a user.
	GetNewCharacter(ctx context.Context, userID uint32) (Character, error)
	// CountForUser returns the total number of characters for a user.
	CountForUser(ctx context.Context, userID uint32) (int, error)
	// Create inserts a new character and returns it.
	Create(ctx context.Context, userID uint32, lastLogin uint32) (Character, error)
	// IsNew returns whether a character is a new (unfinished) character.
	IsNew(charID uint32) (bool, error)
	// HardDelete permanently removes a character.
	HardDelete(charID uint32) error
	// SoftDelete marks a character as deleted.
	SoftDelete(charID uint32) error
	// GetForUser returns all finalized (non-deleted) characters for a user.
	GetForUser(ctx context.Context, userID uint32) ([]Character, error)
	// ExportSave returns the full character row as a map.
	ExportSave(ctx context.Context, userID, charID uint32) (map[string]interface{}, error)
}

// APISessionRepo defines the contract for session/token data access.
type APISessionRepo interface {
	// CreateToken inserts a new sign session and returns its ID and token.
	CreateToken(ctx context.Context, uid uint32, token string) (tokenID uint32, err error)
	// GetUserIDByToken returns the user ID for a given session token.
	GetUserIDByToken(ctx context.Context, token string) (uint32, error)
}
