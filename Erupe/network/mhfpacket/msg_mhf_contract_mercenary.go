package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfContractMercenary represents the MSG_MHF_CONTRACT_MERCENARY
type MsgMhfContractMercenary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfContractMercenary) Opcode() network.PacketID {
	return network.MSG_MHF_CONTRACT_MERCENARY
}

// Parse parses the packet from binary
func (m *MsgMhfContractMercenary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfContractMercenary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
