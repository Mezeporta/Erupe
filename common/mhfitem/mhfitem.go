package mhfitem

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/token"
	"erupe-ce/config"
)

// MHFItem represents a single item identified by its in-game item ID.
type MHFItem struct {
	ItemID uint16
}

// MHFSigilEffect represents a single effect slot on a sigil with an ID and level.
type MHFSigilEffect struct {
	ID    uint16
	Level uint16
}

// MHFSigil represents a weapon sigil containing up to three effects.
type MHFSigil struct {
	Effects []MHFSigilEffect
	Unk0    uint8
	Unk1    uint8
	Unk2    uint8
	Unk3    uint8
}

// MHFEquipment represents an equipment piece (weapon or armor) with its
// decorations and sigils as stored in the player's warehouse.
type MHFEquipment struct {
	WarehouseID uint32
	ItemType    uint8
	Unk0        uint8
	ItemID      uint16
	Level       uint16
	Decorations []MHFItem
	Sigils      []MHFSigil
	Unk1        uint16
}

// MHFItemStack represents a stacked item slot in the warehouse with a quantity.
type MHFItemStack struct {
	WarehouseID uint32
	Item        MHFItem
	Quantity    uint16
	Unk0        uint32
}

// ReadWarehouseItem deserializes an MHFItemStack from a ByteFrame, assigning a
// random warehouse ID if the encoded ID is zero.
func ReadWarehouseItem(bf *byteframe.ByteFrame) MHFItemStack {
	var item MHFItemStack
	item.WarehouseID = bf.ReadUint32()
	if item.WarehouseID == 0 {
		item.WarehouseID = token.RNG.Uint32()
	}
	item.Item.ItemID = bf.ReadUint16()
	item.Quantity = bf.ReadUint16()
	item.Unk0 = bf.ReadUint32()
	return item
}

// DiffItemStacks merges an updated item stack list into an existing one,
// matching by warehouse ID. New items receive a random ID; items with zero
// quantity in the old list are removed.
func DiffItemStacks(o []MHFItemStack, u []MHFItemStack) []MHFItemStack {
	// o = old, u = update, f = final
	var f []MHFItemStack
	for _, uItem := range u {
		exists := false
		for i := range o {
			if o[i].WarehouseID == uItem.WarehouseID {
				exists = true
				o[i].Quantity = uItem.Quantity
			}
		}
		if !exists {
			uItem.WarehouseID = token.RNG.Uint32()
			f = append(f, uItem)
		}
	}
	for _, oItem := range o {
		if oItem.Quantity > 0 {
			f = append(f, oItem)
		}
	}
	return f
}

// ToBytes serializes the item stack to its binary protocol representation.
func (is MHFItemStack) ToBytes() []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(is.WarehouseID)
	bf.WriteUint16(is.Item.ItemID)
	bf.WriteUint16(is.Quantity)
	bf.WriteUint32(is.Unk0)
	return bf.Data()
}

// SerializeWarehouseItems serializes a slice of item stacks with a uint16
// count header for transmission in warehouse response packets.
func SerializeWarehouseItems(i []MHFItemStack) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(i)))
	bf.WriteUint16(0) // Unused
	for _, j := range i {
		bf.WriteBytes(j.ToBytes())
	}
	return bf.Data()
}

// ReadWarehouseEquipment deserializes an MHFEquipment from a ByteFrame. The
// binary layout varies by game version: sigils are present from G1 onward and
// an additional field is present from Z1 onward.
func ReadWarehouseEquipment(bf *byteframe.ByteFrame, mode _config.Mode) MHFEquipment {
	var equipment MHFEquipment
	equipment.Decorations = make([]MHFItem, 3)
	equipment.Sigils = make([]MHFSigil, 3)
	for i := 0; i < 3; i++ {
		equipment.Sigils[i].Effects = make([]MHFSigilEffect, 3)
	}
	equipment.WarehouseID = bf.ReadUint32()
	if equipment.WarehouseID == 0 {
		equipment.WarehouseID = token.RNG.Uint32()
	}
	equipment.ItemType = bf.ReadUint8()
	equipment.Unk0 = bf.ReadUint8()
	equipment.ItemID = bf.ReadUint16()
	equipment.Level = bf.ReadUint16()
	for i := 0; i < 3; i++ {
		equipment.Decorations[i].ItemID = bf.ReadUint16()
	}
	if mode >= _config.G1 {
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				equipment.Sigils[i].Effects[j].ID = bf.ReadUint16()
			}
			for j := 0; j < 3; j++ {
				equipment.Sigils[i].Effects[j].Level = bf.ReadUint16()
			}
			equipment.Sigils[i].Unk0 = bf.ReadUint8()
			equipment.Sigils[i].Unk1 = bf.ReadUint8()
			equipment.Sigils[i].Unk2 = bf.ReadUint8()
			equipment.Sigils[i].Unk3 = bf.ReadUint8()
		}
	}
	if mode >= _config.Z1 {
		equipment.Unk1 = bf.ReadUint16()
	}
	return equipment
}

// ToBytes serializes the equipment to its binary protocol representation.
func (e MHFEquipment) ToBytes(mode _config.Mode) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(e.WarehouseID)
	bf.WriteUint8(e.ItemType)
	bf.WriteUint8(e.Unk0)
	bf.WriteUint16(e.ItemID)
	bf.WriteUint16(e.Level)
	for i := 0; i < 3; i++ {
		bf.WriteUint16(e.Decorations[i].ItemID)
	}
	if mode >= _config.G1 {
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				bf.WriteUint16(e.Sigils[i].Effects[j].ID)
			}
			for j := 0; j < 3; j++ {
				bf.WriteUint16(e.Sigils[i].Effects[j].Level)
			}
			bf.WriteUint8(e.Sigils[i].Unk0)
			bf.WriteUint8(e.Sigils[i].Unk1)
			bf.WriteUint8(e.Sigils[i].Unk2)
			bf.WriteUint8(e.Sigils[i].Unk3)
		}
	}
	if mode >= _config.Z1 {
		bf.WriteUint16(e.Unk1)
	}
	return bf.Data()
}

// SerializeWarehouseEquipment serializes a slice of equipment with a uint16
// count header for transmission in warehouse response packets.
func SerializeWarehouseEquipment(i []MHFEquipment, mode _config.Mode) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(i)))
	bf.WriteUint16(0) // Unused
	for _, j := range i {
		bf.WriteBytes(j.ToBytes(mode))
	}
	return bf.Data()
}
