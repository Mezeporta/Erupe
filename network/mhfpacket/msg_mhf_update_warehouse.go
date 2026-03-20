package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfUpdateWarehouse represents the MSG_MHF_UPDATE_WAREHOUSE
type MsgMhfUpdateWarehouse struct {
	AckHandle        uint32
	BoxType          uint8
	BoxIndex         uint8
	UpdatedItems     []mhfitem.MHFItemStack
	UpdatedEquipment []mhfitem.MHFEquipment
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateWarehouse) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_WAREHOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateWarehouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.BoxType = bf.ReadUint8()
	m.BoxIndex = bf.ReadUint8()
	changes := int(bf.ReadUint16())
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	for i := 0; i < changes; i++ {
		switch m.BoxType {
		case 0:
			m.UpdatedItems = append(m.UpdatedItems, mhfitem.ReadWarehouseItem(bf))
		case 1:
			m.UpdatedEquipment = append(m.UpdatedEquipment, mhfitem.ReadWarehouseEquipment(bf, ctx.RealClientMode))
		}
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateWarehouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
