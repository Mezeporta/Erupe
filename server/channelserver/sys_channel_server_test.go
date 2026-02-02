package channelserver

import (
	"net"
	"sync"
	"testing"
	"time"

	"erupe-ce/config"

	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		DB:          nil,
		DiscordBot:  nil,
		ErupeConfig: &config.Config{DevMode: true},
		Name:        "TestServer",
		Enable:      true,
	}

	s := NewServer(cfg)

	if s == nil {
		t.Fatal("NewServer returned nil")
	}

	// Check ID assignment
	if s.ID != 1 {
		t.Errorf("Server ID = %d, want 1", s.ID)
	}

	// Check name assignment
	if s.name != "TestServer" {
		t.Errorf("Server name = %s, want TestServer", s.name)
	}

	// Check channels are created
	if s.acceptConns == nil {
		t.Error("acceptConns channel is nil")
	}
	if s.deleteConns == nil {
		t.Error("deleteConns channel is nil")
	}

	// Check maps are initialized
	if s.sessions == nil {
		t.Error("sessions map is nil")
	}
	if s.stages == nil {
		t.Error("stages map is nil")
	}
	if s.userBinaryParts == nil {
		t.Error("userBinaryParts map is nil")
	}
	if s.semaphore == nil {
		t.Error("semaphore map is nil")
	}

	// Check semaphore index starts at 7 (skips reserved IDs)
	if s.semaphoreIndex != 7 {
		t.Errorf("semaphoreIndex = %d, want 7", s.semaphoreIndex)
	}

	// Check Raviente is initialized
	if s.raviente == nil {
		t.Error("raviente is nil")
	}
}

func TestNewServer_DefaultStages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Check persistent stages are created
	expectedStages := []string{
		"sl1Ns200p0a0u0", // Mezeporta
		"sl1Ns211p0a0u0", // Rasta bar
		"sl1Ns260p0a0u0", // Pallone Caravan
		"sl1Ns262p0a0u0", // Pallone Guest House 1st Floor
		"sl1Ns263p0a0u0", // Pallone Guest House 2nd Floor
		"sl2Ns379p0a0u0", // Diva fountain
		"sl1Ns462p0a0u0", // MezFes
	}

	for _, stageID := range expectedStages {
		if _, ok := s.stages[stageID]; !ok {
			t.Errorf("Expected default stage %s not found", stageID)
		}
	}

	if len(s.stages) != len(expectedStages) {
		t.Errorf("Server has %d stages, expected %d", len(s.stages), len(expectedStages))
	}
}

func TestNewRaviente(t *testing.T) {
	r := NewRaviente()

	if r == nil {
		t.Fatal("NewRaviente returned nil")
	}

	// Check register initialization
	if r.register == nil {
		t.Fatal("Raviente register is nil")
	}
	if r.register.nextTime != 0 {
		t.Errorf("nextTime = %d, want 0", r.register.nextTime)
	}
	if r.register.maxPlayers != 0 {
		t.Errorf("maxPlayers = %d, want 0", r.register.maxPlayers)
	}
	if len(r.register.register) != 5 {
		t.Errorf("register array length = %d, want 5", len(r.register.register))
	}

	// Check state initialization
	if r.state == nil {
		t.Fatal("Raviente state is nil")
	}
	if len(r.state.stateData) != 29 {
		t.Errorf("stateData length = %d, want 29", len(r.state.stateData))
	}

	// Check support initialization
	if r.support == nil {
		t.Fatal("Raviente support is nil")
	}
	if len(r.support.supportData) != 25 {
		t.Errorf("supportData length = %d, want 25", len(r.support.supportData))
	}
}

func TestRavienteRegister_InitialValues(t *testing.T) {
	r := NewRaviente()

	// All register slots should be 0 initially
	for i, v := range r.register.register {
		if v != 0 {
			t.Errorf("register[%d] = %d, want 0", i, v)
		}
	}

	// All state data should be 0 initially
	for i, v := range r.state.stateData {
		if v != 0 {
			t.Errorf("stateData[%d] = %d, want 0", i, v)
		}
	}

	// All support data should be 0 initially
	for i, v := range r.support.supportData {
		if v != 0 {
			t.Errorf("supportData[%d] = %d, want 0", i, v)
		}
	}
}

func TestServerMutex(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Test that mutex works and doesn't deadlock
	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	s.Lock()
	if !s.isShuttingDown {
		t.Error("isShuttingDown should be true")
	}
	s.Unlock()
}

func TestServerStagesLock(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Test RWMutex for stages
	s.stagesLock.RLock()
	count := len(s.stages)
	s.stagesLock.RUnlock()

	if count < 7 {
		t.Errorf("Expected at least 7 default stages, got %d", count)
	}

	// Test write lock
	s.stagesLock.Lock()
	s.stages["test_stage"] = NewStage("test_stage")
	s.stagesLock.Unlock()

	s.stagesLock.RLock()
	if _, ok := s.stages["test_stage"]; !ok {
		t.Error("test_stage not found after adding")
	}
	s.stagesLock.RUnlock()
}

func TestServerConcurrentStageAccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	var wg sync.WaitGroup

	// Multiple concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.stagesLock.RLock()
				_ = len(s.stages)
				s.stagesLock.RUnlock()
			}
		}()
	}

	// Concurrent writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 50; j++ {
			s.stagesLock.Lock()
			stageID := "concurrent_test_" + string(rune('A'+j%26))
			s.stages[stageID] = NewStage(stageID)
			s.stagesLock.Unlock()
		}
	}()

	wg.Wait()
}

func TestNextSemaphoreID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Initial index should be 7
	if s.semaphoreIndex != 7 {
		t.Errorf("Initial semaphoreIndex = %d, want 7", s.semaphoreIndex)
	}

	// Get next IDs
	id1 := s.NextSemaphoreID()
	id2 := s.NextSemaphoreID()
	id3 := s.NextSemaphoreID()

	// IDs should be unique and incrementing
	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Errorf("Semaphore IDs should be unique: %d, %d, %d", id1, id2, id3)
	}

	if id2 <= id1 || id3 <= id2 {
		t.Errorf("Semaphore IDs should be incrementing: %d, %d, %d", id1, id2, id3)
	}
}

func TestNextSemaphoreID_SkipsExisting(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Pre-populate some semaphores
	s.semaphore["test1"] = &Semaphore{id: 8}
	s.semaphore["test2"] = &Semaphore{id: 9}

	id := s.NextSemaphoreID()

	// Should skip 8 and 9 since they exist
	if id == 8 || id == 9 {
		t.Errorf("NextSemaphoreID should skip existing IDs, got %d", id)
	}
}

func TestUserBinaryPartID(t *testing.T) {
	id1 := userBinaryPartID{charID: 100, index: 1}
	id2 := userBinaryPartID{charID: 100, index: 2}
	id3 := userBinaryPartID{charID: 200, index: 1}

	// Same char, different index should be different keys
	if id1 == id2 {
		t.Error("Different indices should produce different keys")
	}

	// Different char, same index should be different keys
	if id1 == id3 {
		t.Error("Different charIDs should produce different keys")
	}

	// Same values should be equal
	id1copy := userBinaryPartID{charID: 100, index: 1}
	if id1 != id1copy {
		t.Error("Same values should be equal")
	}
}

func TestServerUserBinaryParts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	testData := []byte{0x01, 0x02, 0x03}
	partID := userBinaryPartID{charID: 12345, index: 1}

	// Store data
	s.userBinaryPartsLock.Lock()
	s.userBinaryParts[partID] = testData
	s.userBinaryPartsLock.Unlock()

	// Retrieve data
	s.userBinaryPartsLock.RLock()
	data, ok := s.userBinaryParts[partID]
	s.userBinaryPartsLock.RUnlock()

	if !ok {
		t.Error("Failed to retrieve stored binary part")
	}
	if len(data) != 3 || data[0] != 0x01 {
		t.Errorf("Retrieved data doesn't match: %v", data)
	}
}

func TestServerShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Create a test listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	s.listener = listener

	// Shutdown should not panic
	s.Shutdown()

	// Check shutdown flag is set
	s.Lock()
	if !s.isShuttingDown {
		t.Error("isShuttingDown should be true after Shutdown()")
	}
	s.Unlock()
}

func TestServerFindSessionByCharID_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)
	s.Channels = []*Server{s}

	// Search for non-existent character
	session := s.FindSessionByCharID(99999)
	if session != nil {
		t.Error("Expected nil for non-existent character")
	}
}

func TestServerFindObjectByChar_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)

	// Search for non-existent object
	obj := s.FindObjectByChar(99999)
	if obj != nil {
		t.Error("Expected nil for non-existent object owner")
	}
}

func TestServerStartAndShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		ID:          1,
		Logger:      logger,
		ErupeConfig: &config.Config{DevMode: true},
	}

	s := NewServer(cfg)
	s.Port = 0 // Use any available port

	err := s.Start()
	if err != nil {
		t.Fatalf("Server.Start() failed: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(10 * time.Millisecond)

	// Verify listener is created
	if s.listener == nil {
		t.Error("Listener should be created after Start()")
	}

	// Shutdown
	s.Shutdown()

	// Give time for cleanup
	time.Sleep(10 * time.Millisecond)
}

func TestConfigStruct(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	cfg := &Config{
		ID:          42,
		Logger:      logger,
		DB:          nil,
		DiscordBot:  nil,
		ErupeConfig: &config.Config{},
		Name:        "Test Channel",
		Enable:      true,
	}

	if cfg.ID != 42 {
		t.Errorf("Config ID = %d, want 42", cfg.ID)
	}
	if cfg.Name != "Test Channel" {
		t.Errorf("Config Name = %s, want 'Test Channel'", cfg.Name)
	}
	if !cfg.Enable {
		t.Error("Config Enable should be true")
	}
}
