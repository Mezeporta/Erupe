package channelserver

import (
	"testing"
)

func TestPlayerStruct(t *testing.T) {
	player := Player{
		CharName: "TestPlayer",
		QuestID:  5,
	}

	if player.CharName != "TestPlayer" {
		t.Errorf("CharName = %s, want TestPlayer", player.CharName)
	}
	if player.QuestID != 5 {
		t.Errorf("QuestID = %d, want 5", player.QuestID)
	}
}

func TestGetPlayerSlice_EmptyServer(t *testing.T) {
	server := createMockServer()
	server.Channels = []*Server{}

	players := getPlayerSlice(server)

	if len(players) != 0 {
		t.Errorf("Expected 0 players, got %d", len(players))
	}
}

func TestGetPlayerSlice_WithChannel(t *testing.T) {
	server := createMockServer()

	// Create a channel with stages
	channel := &Server{
		stages: make(map[string]*Stage),
	}

	// Create a stage with clients
	stage := NewStage("test_stage")
	session := createMockSession(1, server)
	session.Name = "Player1"
	stage.clients[session] = session.charID

	channel.stages["test_stage"] = stage
	server.Channels = []*Server{channel}

	players := getPlayerSlice(server)

	if len(players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(players))
	}
	if len(players) > 0 && players[0].CharName != "Player1" {
		t.Errorf("Expected CharName Player1, got %s", players[0].CharName)
	}
}

func TestGetPlayerSlice_MultiplePlayersMultipleStages(t *testing.T) {
	server := createMockServer()

	channel := &Server{
		stages: make(map[string]*Stage),
	}

	// Stage 1 with one player
	stage1 := NewStage("stage1")
	session1 := createMockSession(1, server)
	session1.Name = "Player1"
	stage1.clients[session1] = session1.charID
	channel.stages["stage1"] = stage1

	// Stage 2 with two players
	stage2 := NewStage("stage2")
	session2 := createMockSession(2, server)
	session2.Name = "Player2"
	session3 := createMockSession(3, server)
	session3.Name = "Player3"
	stage2.clients[session2] = session2.charID
	stage2.clients[session3] = session3.charID
	channel.stages["stage2"] = stage2

	server.Channels = []*Server{channel}

	players := getPlayerSlice(server)

	if len(players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(players))
	}
}

func TestGetPlayerSlice_EmptyStage(t *testing.T) {
	server := createMockServer()

	channel := &Server{
		stages: make(map[string]*Stage),
	}

	// Empty stage (no clients)
	emptyStage := NewStage("empty_stage")
	channel.stages["empty_stage"] = emptyStage

	server.Channels = []*Server{channel}

	players := getPlayerSlice(server)

	if len(players) != 0 {
		t.Errorf("Expected 0 players from empty stage, got %d", len(players))
	}
}

func TestGetCharacterList_EmptyServer(t *testing.T) {
	server := createMockServer()
	server.Channels = []*Server{}

	result := getCharacterList(server)

	expected := "===== Online: 0 =====\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetCharacterList_WithPlayers(t *testing.T) {
	server := createMockServer()

	channel := &Server{
		stages: make(map[string]*Stage),
	}

	stage := NewStage("lobby")
	session := createMockSession(1, server)
	session.Name = "Hunter1"
	stage.clients[session] = session.charID
	channel.stages["lobby"] = stage

	server.Channels = []*Server{channel}

	result := getCharacterList(server)

	// Should contain the online count
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	// Should contain "Online: 1"
	if !contains(result, "Online: 1") {
		t.Errorf("Expected result to contain 'Online: 1', got %q", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
