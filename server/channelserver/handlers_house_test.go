package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func TestBoxToBytes_EmptyItemBox(t *testing.T) {
	stacks := []mhfpacket.WarehouseStack{}
	result := boxToBytes(stacks, "item")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 0 {
		t.Errorf("Expected 0 stacks, got %d", numStacks)
	}

	// Should have trailing uint16(0)
	if len(result) != 4 {
		t.Errorf("Expected 4 bytes for empty box, got %d", len(result))
	}
}

func TestBoxToBytes_SingleItemStack(t *testing.T) {
	stacks := []mhfpacket.WarehouseStack{
		{
			ID:       1,
			Index:    0,
			ItemID:   100,
			Quantity: 50,
		},
	}
	result := boxToBytes(stacks, "item")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 1 {
		t.Errorf("Expected 1 stack, got %d", numStacks)
	}

	// Read first stack
	id := bf.ReadUint32()
	index := bf.ReadUint16()
	itemID := bf.ReadUint16()
	quantity := bf.ReadUint16()
	_ = bf.ReadUint16() // padding

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}
	if index != 1 { // Index is written as i+1
		t.Errorf("Expected index 1, got %d", index)
	}
	if itemID != 100 {
		t.Errorf("Expected itemID 100, got %d", itemID)
	}
	if quantity != 50 {
		t.Errorf("Expected quantity 50, got %d", quantity)
	}
}

func TestBoxToBytes_MultipleItemStacks(t *testing.T) {
	stacks := []mhfpacket.WarehouseStack{
		{ID: 1, Index: 0, ItemID: 100, Quantity: 10},
		{ID: 2, Index: 1, ItemID: 200, Quantity: 20},
		{ID: 3, Index: 2, ItemID: 300, Quantity: 30},
	}
	result := boxToBytes(stacks, "item")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 3 {
		t.Errorf("Expected 3 stacks, got %d", numStacks)
	}
}

func TestBoxToBytes_EmptyEquipBox(t *testing.T) {
	stacks := []mhfpacket.WarehouseStack{}
	result := boxToBytes(stacks, "equip")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 0 {
		t.Errorf("Expected 0 stacks, got %d", numStacks)
	}
}

func TestBoxToBytes_SingleEquipStack(t *testing.T) {
	equipData := make([]byte, 56)
	for i := range equipData {
		equipData[i] = byte(i)
	}

	stacks := []mhfpacket.WarehouseStack{
		{
			ID:        1,
			Index:     0,
			EquipType: 5,
			ItemID:    1000,
			Data:      equipData,
		},
	}
	result := boxToBytes(stacks, "equip")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 1 {
		t.Errorf("Expected 1 stack, got %d", numStacks)
	}

	// Read first equip stack
	id := bf.ReadUint32()
	index := bf.ReadUint16()
	equipType := bf.ReadUint16()
	itemID := bf.ReadUint16()
	data := bf.ReadBytes(56)

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}
	if index != 1 { // Index is written as i+1
		t.Errorf("Expected index 1, got %d", index)
	}
	if equipType != 5 {
		t.Errorf("Expected equipType 5, got %d", equipType)
	}
	if itemID != 1000 {
		t.Errorf("Expected itemID 1000, got %d", itemID)
	}
	if len(data) != 56 {
		t.Errorf("Expected 56 bytes data, got %d", len(data))
	}
}

func TestBoxToBytes_MultipleEquipStacks(t *testing.T) {
	equipData := make([]byte, 56)

	stacks := []mhfpacket.WarehouseStack{
		{ID: 1, Index: 0, EquipType: 1, ItemID: 100, Data: equipData},
		{ID: 2, Index: 1, EquipType: 2, ItemID: 200, Data: equipData},
	}
	result := boxToBytes(stacks, "equip")

	bf := byteframe.NewByteFrameFromBytes(result)
	numStacks := bf.ReadUint16()
	if numStacks != 2 {
		t.Errorf("Expected 2 stacks, got %d", numStacks)
	}
}

// Test HouseData struct
func TestHouseDataStruct(t *testing.T) {
	house := HouseData{
		CharID:        12345,
		HRP:           999,
		GR:            500,
		Name:          "TestPlayer",
		HouseState:    2,
		HousePassword: "pass123",
	}

	if house.CharID != 12345 {
		t.Errorf("CharID = %d, want 12345", house.CharID)
	}
	if house.HRP != 999 {
		t.Errorf("HRP = %d, want 999", house.HRP)
	}
	if house.GR != 500 {
		t.Errorf("GR = %d, want 500", house.GR)
	}
	if house.Name != "TestPlayer" {
		t.Errorf("Name = %s, want TestPlayer", house.Name)
	}
	if house.HouseState != 2 {
		t.Errorf("HouseState = %d, want 2", house.HouseState)
	}
	if house.HousePassword != "pass123" {
		t.Errorf("HousePassword = %s, want pass123", house.HousePassword)
	}
}

// Test Title struct
func TestTitleStruct(t *testing.T) {
	title := Title{
		ID: 42,
	}

	if title.ID != 42 {
		t.Errorf("ID = %d, want 42", title.ID)
	}
}

// Test decoMyset constants
func TestDecoMysetConstants(t *testing.T) {
	if maxDecoMysets != 40 {
		t.Errorf("maxDecoMysets = %d, want 40", maxDecoMysets)
	}
	if decoMysetSize != 78 {
		t.Errorf("decoMysetSize = %d, want 78", decoMysetSize)
	}
}
