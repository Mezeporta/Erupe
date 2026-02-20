package scenario

import (
	"fmt"

	"erupe-ce/cmd/protbot/protocol"
)

// Logout sends MSG_SYS_LOGOUT and closes the channel connection.
func Logout(ch *protocol.ChannelConn) error {
	fmt.Println("[logout] Sending MSG_SYS_LOGOUT...")
	if err := ch.SendPacket(protocol.BuildLogoutPacket()); err != nil {
		_ = ch.Close()
		return fmt.Errorf("logout send: %w", err)
	}
	return ch.Close()
}
