package channelserver

import (
	"net"
	"sync"
	"testing"
)

func createTestChannels(count int) []*Server {
	channels := make([]*Server, count)
	for i := 0; i < count; i++ {
		s := createTestServer()
		s.ID = uint16(0x1010 + i)
		s.IP = "10.0.0.1"
		s.Port = uint16(54001 + i)
		s.GlobalID = "0101"
		s.userBinaryParts = make(map[userBinaryPartID][]byte)
		channels[i] = s
	}
	return channels
}

func TestLocalRegistryFindSessionByCharID(t *testing.T) {
	channels := createTestChannels(2)
	reg := NewLocalChannelRegistry(channels)

	conn1 := &mockConn{}
	sess1 := createTestSessionForServer(channels[0], conn1, 100, "Alice")
	channels[0].Lock()
	channels[0].sessions[conn1] = sess1
	channels[0].Unlock()

	conn2 := &mockConn{}
	sess2 := createTestSessionForServer(channels[1], conn2, 200, "Bob")
	channels[1].Lock()
	channels[1].sessions[conn2] = sess2
	channels[1].Unlock()

	// Find on first channel
	found := reg.FindSessionByCharID(100)
	if found == nil || found.charID != 100 {
		t.Errorf("FindSessionByCharID(100) = %v, want session with charID 100", found)
	}

	// Find on second channel
	found = reg.FindSessionByCharID(200)
	if found == nil || found.charID != 200 {
		t.Errorf("FindSessionByCharID(200) = %v, want session with charID 200", found)
	}

	// Not found
	found = reg.FindSessionByCharID(999)
	if found != nil {
		t.Errorf("FindSessionByCharID(999) = %v, want nil", found)
	}
}

func TestLocalRegistryFindChannelForStage(t *testing.T) {
	channels := createTestChannels(2)
	channels[0].GlobalID = "0101"
	channels[1].GlobalID = "0102"
	reg := NewLocalChannelRegistry(channels)

	channels[1].stagesLock.Lock()
	channels[1].stages["sl2Qs123p0a0u42"] = NewStage("sl2Qs123p0a0u42")
	channels[1].stagesLock.Unlock()

	gid := reg.FindChannelForStage("u42")
	if gid != "0102" {
		t.Errorf("FindChannelForStage(u42) = %q, want %q", gid, "0102")
	}

	gid = reg.FindChannelForStage("u999")
	if gid != "" {
		t.Errorf("FindChannelForStage(u999) = %q, want empty", gid)
	}
}

func TestLocalRegistryDisconnectUser(t *testing.T) {
	channels := createTestChannels(1)
	reg := NewLocalChannelRegistry(channels)

	conn := &mockConn{}
	sess := createTestSessionForServer(channels[0], conn, 42, "Target")
	channels[0].Lock()
	channels[0].sessions[conn] = sess
	channels[0].Unlock()

	reg.DisconnectUser([]uint32{42})

	if !conn.WasClosed() {
		t.Error("DisconnectUser should have closed the connection for charID 42")
	}
}

func TestLocalRegistrySearchSessions(t *testing.T) {
	channels := createTestChannels(2)
	reg := NewLocalChannelRegistry(channels)

	// Add 3 sessions across 2 channels
	for i, ch := range channels {
		conn := &mockConn{}
		sess := createTestSessionForServer(ch, conn, uint32(i+1), "Player")
		sess.stage = NewStage("sl1Ns200p0a0u0")
		ch.Lock()
		ch.sessions[conn] = sess
		ch.Unlock()
	}
	conn3 := &mockConn{}
	sess3 := createTestSessionForServer(channels[0], conn3, 3, "Player")
	sess3.stage = NewStage("sl1Ns200p0a0u0")
	channels[0].Lock()
	channels[0].sessions[conn3] = sess3
	channels[0].Unlock()

	// Search all
	results := reg.SearchSessions(func(s SessionSnapshot) bool { return true }, 10)
	if len(results) != 3 {
		t.Errorf("SearchSessions(all) returned %d results, want 3", len(results))
	}

	// Search with max
	results = reg.SearchSessions(func(s SessionSnapshot) bool { return true }, 2)
	if len(results) != 2 {
		t.Errorf("SearchSessions(max=2) returned %d results, want 2", len(results))
	}

	// Search with predicate
	results = reg.SearchSessions(func(s SessionSnapshot) bool { return s.CharID == 1 }, 10)
	if len(results) != 1 {
		t.Errorf("SearchSessions(charID==1) returned %d results, want 1", len(results))
	}
}

func TestLocalRegistrySearchStages(t *testing.T) {
	channels := createTestChannels(1)
	reg := NewLocalChannelRegistry(channels)

	channels[0].stagesLock.Lock()
	channels[0].stages["sl2Ls210test1"] = NewStage("sl2Ls210test1")
	channels[0].stages["sl2Ls210test2"] = NewStage("sl2Ls210test2")
	channels[0].stages["sl1Ns200other"] = NewStage("sl1Ns200other")
	channels[0].stagesLock.Unlock()

	results := reg.SearchStages("sl2Ls210", 10)
	if len(results) != 2 {
		t.Errorf("SearchStages(sl2Ls210) returned %d results, want 2", len(results))
	}

	results = reg.SearchStages("sl2Ls210", 1)
	if len(results) != 1 {
		t.Errorf("SearchStages(sl2Ls210, max=1) returned %d results, want 1", len(results))
	}
}

func TestLocalRegistryConcurrentAccess(t *testing.T) {
	channels := createTestChannels(2)
	reg := NewLocalChannelRegistry(channels)

	// Populate some sessions
	for _, ch := range channels {
		for i := 0; i < 10; i++ {
			conn := &mockConn{remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 50000 + i}}
			sess := createTestSessionForServer(ch, conn, uint32(i+1), "Player")
			sess.stage = NewStage("sl1Ns200p0a0u0")
			ch.Lock()
			ch.sessions[conn] = sess
			ch.Unlock()
		}
	}

	// Run concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(id int) {
			defer wg.Done()
			_ = reg.FindSessionByCharID(uint32(id%10 + 1))
		}(i)
		go func() {
			defer wg.Done()
			_ = reg.FindChannelForStage("u0")
		}()
		go func() {
			defer wg.Done()
			_ = reg.SearchSessions(func(s SessionSnapshot) bool { return true }, 5)
		}()
	}
	wg.Wait()
}
