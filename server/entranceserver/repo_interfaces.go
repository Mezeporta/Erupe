package entranceserver

// Repository interfaces decouple entrance server business logic from concrete
// PostgreSQL implementations, enabling mock/stub injection for unit tests.

// EntranceServerRepo defines the contract for server-related data access
// used by the entrance server when building server list responses.
type EntranceServerRepo interface {
	// GetCurrentPlayers returns the current player count for a given server ID.
	GetCurrentPlayers(serverID int) (uint16, error)
}

// EntranceSessionRepo defines the contract for session-related data access
// used by the entrance server when resolving user locations.
type EntranceSessionRepo interface {
	// GetServerIDForCharacter returns the server ID where the given character
	// is currently signed in, or 0 if not found.
	GetServerIDForCharacter(charID uint32) (uint16, error)
}
