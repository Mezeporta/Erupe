// Package channelserver implements the gameplay channel server (TCP port
// 54001+) that handles all in-game multiplayer functionality. It manages
// player sessions, stage (lobby/quest room) state, guild operations, item
// management, event systems, and binary state relay between clients.
//
// Packet handlers are organized by game system into separate files
// (handlers_quest.go, handlers_guild.go, etc.) and registered in
// handlers_table.go. Each handler has the signature:
//
//	func(s *Session, p mhfpacket.MHFPacket)
package channelserver
