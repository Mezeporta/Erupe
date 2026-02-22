package entranceserver

// mockEntranceServerRepo implements EntranceServerRepo for testing.
type mockEntranceServerRepo struct {
	currentPlayers    uint16
	currentPlayersErr error
}

func (m *mockEntranceServerRepo) GetCurrentPlayers(_ int) (uint16, error) {
	return m.currentPlayers, m.currentPlayersErr
}

// mockEntranceSessionRepo implements EntranceSessionRepo for testing.
type mockEntranceSessionRepo struct {
	serverID    uint16
	serverIDErr error
}

func (m *mockEntranceSessionRepo) GetServerIDForCharacter(_ uint32) (uint16, error) {
	return m.serverID, m.serverIDErr
}
