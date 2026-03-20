// Package channelserver implements the gameplay channel server (TCP port
// 54001+) that handles all in-game multiplayer functionality. It manages
// player sessions, stage (lobby/quest room) state, guild operations, item
// management, event systems, and binary state relay between clients.
//
// # Handler Organization
//
// Packet handlers are organized by game system into separate files
// (handlers_quest.go, handlers_guild.go, etc.) and registered via
// [buildHandlerTable] in handlers_table.go. Each handler has the signature:
//
//	func(s *Session, p mhfpacket.MHFPacket)
//
// To add a new handler:
//  1. Define the packet struct in network/mhfpacket/msg_*.go
//  2. Add an entry in [buildHandlerTable] mapping the opcode to the handler
//  3. Implement the handler in the appropriate handlers_*.go file
//
// # Repository Pattern
//
// All database access goes through interface-based repositories defined in
// repo_interfaces.go. The [Server] struct holds interface types, not concrete
// implementations. Concrete PostgreSQL implementations live in repo_*.go
// files. Mock implementations in repo_mocks_test.go enable unit tests
// without a database.
//
// Handler code must never contain inline SQL — use the appropriate repo
// method. If a query doesn't exist yet, add it to the relevant repo file
// and its interface.
//
// # Testing
//
// Tests use mock repositories (repo_mocks_test.go) and shared test helpers
// (test_helpers_test.go). The standard pattern is table-driven tests; see
// handlers_achievement_test.go for a typical example. Always run tests with
// the race detector: go test -race ./...
//
// # Concurrency
//
// Lock ordering: Server.Mutex → Stage.RWMutex → semaphoreLock.
// The stage map uses sync.Map for lock-free concurrent access; individual
// Stage structs have their own sync.RWMutex. Cross-channel operations go
// exclusively through the [ChannelRegistry] interface.
package channelserver
