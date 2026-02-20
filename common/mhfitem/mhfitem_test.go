package mhfitem

import (
	"bytes"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/token"
	_config "erupe-ce/config"
	"testing"
)

func TestReadWarehouseItem(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(12345)  // WarehouseID
	bf.WriteUint16(100)    // ItemID
	bf.WriteUint16(5)      // Quantity
	bf.WriteUint32(999999) // Unk0

	_, _ = bf.Seek(0, 0)
	item := ReadWarehouseItem(bf)

	if item.WarehouseID != 12345 {
		t.Errorf("WarehouseID = %d, want 12345", item.WarehouseID)
	}
	if item.Item.ItemID != 100 {
		t.Errorf("ItemID = %d, want 100", item.Item.ItemID)
	}
	if item.Quantity != 5 {
		t.Errorf("Quantity = %d, want 5", item.Quantity)
	}
	if item.Unk0 != 999999 {
		t.Errorf("Unk0 = %d, want 999999", item.Unk0)
	}
}

func TestReadWarehouseItem_ZeroWarehouseID(t *testing.T) {
	// When WarehouseID is 0, it should be replaced with a random value
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)   // WarehouseID = 0
	bf.WriteUint16(100) // ItemID
	bf.WriteUint16(5)   // Quantity
	bf.WriteUint32(0)   // Unk0

	_, _ = bf.Seek(0, 0)
	item := ReadWarehouseItem(bf)

	if item.WarehouseID == 0 {
		t.Error("WarehouseID should be replaced with random value when input is 0")
	}
}

func TestMHFItemStack_ToBytes(t *testing.T) {
	item := MHFItemStack{
		WarehouseID: 12345,
		Item:        MHFItem{ItemID: 100},
		Quantity:    5,
		Unk0:        999999,
	}

	data := item.ToBytes()
	if len(data) != 12 { // 4 + 2 + 2 + 4
		t.Errorf("ToBytes() length = %d, want 12", len(data))
	}

	// Read it back
	bf := byteframe.NewByteFrameFromBytes(data)
	readItem := ReadWarehouseItem(bf)

	if readItem.WarehouseID != item.WarehouseID {
		t.Errorf("WarehouseID = %d, want %d", readItem.WarehouseID, item.WarehouseID)
	}
	if readItem.Item.ItemID != item.Item.ItemID {
		t.Errorf("ItemID = %d, want %d", readItem.Item.ItemID, item.Item.ItemID)
	}
	if readItem.Quantity != item.Quantity {
		t.Errorf("Quantity = %d, want %d", readItem.Quantity, item.Quantity)
	}
	if readItem.Unk0 != item.Unk0 {
		t.Errorf("Unk0 = %d, want %d", readItem.Unk0, item.Unk0)
	}
}

func TestSerializeWarehouseItems(t *testing.T) {
	items := []MHFItemStack{
		{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5, Unk0: 0},
		{WarehouseID: 2, Item: MHFItem{ItemID: 200}, Quantity: 10, Unk0: 0},
	}

	data := SerializeWarehouseItems(items)
	bf := byteframe.NewByteFrameFromBytes(data)

	count := bf.ReadUint16()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	bf.ReadUint16() // Skip unused

	for i := 0; i < 2; i++ {
		item := ReadWarehouseItem(bf)
		if item.WarehouseID != items[i].WarehouseID {
			t.Errorf("item[%d] WarehouseID = %d, want %d", i, item.WarehouseID, items[i].WarehouseID)
		}
		if item.Item.ItemID != items[i].Item.ItemID {
			t.Errorf("item[%d] ItemID = %d, want %d", i, item.Item.ItemID, items[i].Item.ItemID)
		}
	}
}

func TestSerializeWarehouseItems_Empty(t *testing.T) {
	items := []MHFItemStack{}
	data := SerializeWarehouseItems(items)
	bf := byteframe.NewByteFrameFromBytes(data)

	count := bf.ReadUint16()
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestDiffItemStacks(t *testing.T) {
	tests := []struct {
		name    string
		old     []MHFItemStack
		update  []MHFItemStack
		wantLen int
		checkFn func(t *testing.T, result []MHFItemStack)
	}{
		{
			name: "update existing quantity",
			old: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5},
			},
			update: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 10},
			},
			wantLen: 1,
			checkFn: func(t *testing.T, result []MHFItemStack) {
				if result[0].Quantity != 10 {
					t.Errorf("Quantity = %d, want 10", result[0].Quantity)
				}
			},
		},
		{
			name: "add new item",
			old: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5},
			},
			update: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5},
				{WarehouseID: 0, Item: MHFItem{ItemID: 200}, Quantity: 3}, // WarehouseID 0 = new
			},
			wantLen: 2,
			checkFn: func(t *testing.T, result []MHFItemStack) {
				hasNewItem := false
				for _, item := range result {
					if item.Item.ItemID == 200 {
						hasNewItem = true
						if item.WarehouseID == 0 {
							t.Error("New item should have generated WarehouseID")
						}
					}
				}
				if !hasNewItem {
					t.Error("New item should be in result")
				}
			},
		},
		{
			name: "remove item (quantity 0)",
			old: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5},
				{WarehouseID: 2, Item: MHFItem{ItemID: 200}, Quantity: 10},
			},
			update: []MHFItemStack{
				{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 0}, // Removed
			},
			wantLen: 1,
			checkFn: func(t *testing.T, result []MHFItemStack) {
				for _, item := range result {
					if item.WarehouseID == 1 {
						t.Error("Item with quantity 0 should be removed")
					}
				}
			},
		},
		{
			name:    "empty old, add new",
			old:     []MHFItemStack{},
			update:  []MHFItemStack{{WarehouseID: 0, Item: MHFItem{ItemID: 100}, Quantity: 5}},
			wantLen: 1,
			checkFn: func(t *testing.T, result []MHFItemStack) {
				if len(result) != 1 || result[0].Item.ItemID != 100 {
					t.Error("Should add new item to empty list")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DiffItemStacks(tt.old, tt.update)
			if len(result) != tt.wantLen {
				t.Errorf("DiffItemStacks() length = %d, want %d", len(result), tt.wantLen)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestReadWarehouseEquipment(t *testing.T) {
	mode := _config.Z1

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(12345) // WarehouseID
	bf.WriteUint8(1)      // ItemType
	bf.WriteUint8(2)      // Unk0
	bf.WriteUint16(100)   // ItemID
	bf.WriteUint16(5)     // Level

	// Write 3 decorations
	bf.WriteUint16(201)
	bf.WriteUint16(202)
	bf.WriteUint16(203)

	// Write 3 sigils (G1+)
	for i := 0; i < 3; i++ {
		// 3 effects per sigil
		for j := 0; j < 3; j++ {
			bf.WriteUint16(uint16(300 + i*10 + j)) // Effect ID
		}
		for j := 0; j < 3; j++ {
			bf.WriteUint16(uint16(1 + j)) // Effect Level
		}
		bf.WriteUint8(10)
		bf.WriteUint8(11)
		bf.WriteUint8(12)
		bf.WriteUint8(13)
	}

	// Unk1 (Z1+)
	bf.WriteUint16(9999)

	_, _ = bf.Seek(0, 0)
	equipment := ReadWarehouseEquipment(bf, mode)

	if equipment.WarehouseID != 12345 {
		t.Errorf("WarehouseID = %d, want 12345", equipment.WarehouseID)
	}
	if equipment.ItemType != 1 {
		t.Errorf("ItemType = %d, want 1", equipment.ItemType)
	}
	if equipment.ItemID != 100 {
		t.Errorf("ItemID = %d, want 100", equipment.ItemID)
	}
	if equipment.Level != 5 {
		t.Errorf("Level = %d, want 5", equipment.Level)
	}
	if equipment.Decorations[0].ItemID != 201 {
		t.Errorf("Decoration[0] = %d, want 201", equipment.Decorations[0].ItemID)
	}
	if equipment.Sigils[0].Effects[0].ID != 300 {
		t.Errorf("Sigil[0].Effect[0].ID = %d, want 300", equipment.Sigils[0].Effects[0].ID)
	}
	if equipment.Unk1 != 9999 {
		t.Errorf("Unk1 = %d, want 9999", equipment.Unk1)
	}
}

func TestReadWarehouseEquipment_ZeroWarehouseID(t *testing.T) {
	mode := _config.Z1

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // WarehouseID = 0
	bf.WriteUint8(1)
	bf.WriteUint8(2)
	bf.WriteUint16(100)
	bf.WriteUint16(5)
	// Write decorations
	for i := 0; i < 3; i++ {
		bf.WriteUint16(0)
	}
	// Write sigils
	for i := 0; i < 3; i++ {
		for j := 0; j < 6; j++ {
			bf.WriteUint16(0)
		}
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
	}
	bf.WriteUint16(0)

	_, _ = bf.Seek(0, 0)
	equipment := ReadWarehouseEquipment(bf, mode)

	if equipment.WarehouseID == 0 {
		t.Error("WarehouseID should be replaced with random value when input is 0")
	}
}

func TestMHFEquipment_ToBytes(t *testing.T) {
	mode := _config.Z1

	equipment := MHFEquipment{
		WarehouseID: 12345,
		ItemType:    1,
		Unk0:        2,
		ItemID:      100,
		Level:       5,
		Decorations: []MHFItem{{ItemID: 201}, {ItemID: 202}, {ItemID: 203}},
		Sigils:      make([]MHFSigil, 3),
		Unk1:        9999,
	}
	for i := 0; i < 3; i++ {
		equipment.Sigils[i].Effects = make([]MHFSigilEffect, 3)
	}

	data := equipment.ToBytes(mode)
	bf := byteframe.NewByteFrameFromBytes(data)
	readEquipment := ReadWarehouseEquipment(bf, mode)

	if readEquipment.WarehouseID != equipment.WarehouseID {
		t.Errorf("WarehouseID = %d, want %d", readEquipment.WarehouseID, equipment.WarehouseID)
	}
	if readEquipment.ItemID != equipment.ItemID {
		t.Errorf("ItemID = %d, want %d", readEquipment.ItemID, equipment.ItemID)
	}
	if readEquipment.Level != equipment.Level {
		t.Errorf("Level = %d, want %d", readEquipment.Level, equipment.Level)
	}
	if readEquipment.Unk1 != equipment.Unk1 {
		t.Errorf("Unk1 = %d, want %d", readEquipment.Unk1, equipment.Unk1)
	}
}

func TestSerializeWarehouseEquipment(t *testing.T) {
	mode := _config.Z1

	equipment := []MHFEquipment{
		{
			WarehouseID: 1,
			ItemType:    1,
			ItemID:      100,
			Level:       5,
			Decorations: []MHFItem{{ItemID: 0}, {ItemID: 0}, {ItemID: 0}},
			Sigils:      make([]MHFSigil, 3),
		},
		{
			WarehouseID: 2,
			ItemType:    2,
			ItemID:      200,
			Level:       10,
			Decorations: []MHFItem{{ItemID: 0}, {ItemID: 0}, {ItemID: 0}},
			Sigils:      make([]MHFSigil, 3),
		},
	}
	for i := range equipment {
		for j := 0; j < 3; j++ {
			equipment[i].Sigils[j].Effects = make([]MHFSigilEffect, 3)
		}
	}

	data := SerializeWarehouseEquipment(equipment, mode)
	bf := byteframe.NewByteFrameFromBytes(data)

	count := bf.ReadUint16()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestMHFEquipment_RoundTrip(t *testing.T) {
	mode := _config.Z1

	original := MHFEquipment{
		WarehouseID: 99999,
		ItemType:    5,
		Unk0:        10,
		ItemID:      500,
		Level:       25,
		Decorations: []MHFItem{{ItemID: 1}, {ItemID: 2}, {ItemID: 3}},
		Sigils:      make([]MHFSigil, 3),
		Unk1:        12345,
	}
	for i := 0; i < 3; i++ {
		original.Sigils[i].Effects = []MHFSigilEffect{
			{ID: uint16(100 + i), Level: 1},
			{ID: uint16(200 + i), Level: 2},
			{ID: uint16(300 + i), Level: 3},
		}
	}

	// Write to bytes
	data := original.ToBytes(mode)

	// Read back
	bf := byteframe.NewByteFrameFromBytes(data)
	recovered := ReadWarehouseEquipment(bf, mode)

	// Compare
	if recovered.WarehouseID != original.WarehouseID {
		t.Errorf("WarehouseID = %d, want %d", recovered.WarehouseID, original.WarehouseID)
	}
	if recovered.ItemType != original.ItemType {
		t.Errorf("ItemType = %d, want %d", recovered.ItemType, original.ItemType)
	}
	if recovered.ItemID != original.ItemID {
		t.Errorf("ItemID = %d, want %d", recovered.ItemID, original.ItemID)
	}
	if recovered.Level != original.Level {
		t.Errorf("Level = %d, want %d", recovered.Level, original.Level)
	}
	for i := 0; i < 3; i++ {
		if recovered.Decorations[i].ItemID != original.Decorations[i].ItemID {
			t.Errorf("Decoration[%d] = %d, want %d", i, recovered.Decorations[i].ItemID, original.Decorations[i].ItemID)
		}
	}
}

func BenchmarkReadWarehouseItem(b *testing.B) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(12345)
	bf.WriteUint16(100)
	bf.WriteUint16(5)
	bf.WriteUint32(0)
	data := bf.Data()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := byteframe.NewByteFrameFromBytes(data)
		_ = ReadWarehouseItem(bf)
	}
}

func BenchmarkDiffItemStacks(b *testing.B) {
	old := []MHFItemStack{
		{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 5},
		{WarehouseID: 2, Item: MHFItem{ItemID: 200}, Quantity: 10},
		{WarehouseID: 3, Item: MHFItem{ItemID: 300}, Quantity: 15},
	}
	update := []MHFItemStack{
		{WarehouseID: 1, Item: MHFItem{ItemID: 100}, Quantity: 8},
		{WarehouseID: 0, Item: MHFItem{ItemID: 400}, Quantity: 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DiffItemStacks(old, update)
	}
}

func BenchmarkSerializeWarehouseItems(b *testing.B) {
	items := make([]MHFItemStack, 100)
	for i := range items {
		items[i] = MHFItemStack{
			WarehouseID: uint32(i),
			Item:        MHFItem{ItemID: uint16(i)},
			Quantity:    uint16(i % 99),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SerializeWarehouseItems(items)
	}
}

func TestMHFItemStack_ToBytes_RoundTrip(t *testing.T) {
	original := MHFItemStack{
		WarehouseID: 12345,
		Item:        MHFItem{ItemID: 999},
		Quantity:    42,
		Unk0:        777,
	}

	data := original.ToBytes()
	bf := byteframe.NewByteFrameFromBytes(data)
	recovered := ReadWarehouseItem(bf)

	if !bytes.Equal(original.ToBytes(), recovered.ToBytes()) {
		t.Error("Round-trip serialization failed")
	}
}

func TestDiffItemStacks_PreserveOldWarehouseID(t *testing.T) {
	// Verify that when updating existing items, the old WarehouseID is preserved
	old := []MHFItemStack{
		{WarehouseID: 555, Item: MHFItem{ItemID: 100}, Quantity: 5},
	}
	update := []MHFItemStack{
		{WarehouseID: 555, Item: MHFItem{ItemID: 100}, Quantity: 10},
	}

	result := DiffItemStacks(old, update)
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}
	if result[0].WarehouseID != 555 {
		t.Errorf("WarehouseID = %d, want 555", result[0].WarehouseID)
	}
	if result[0].Quantity != 10 {
		t.Errorf("Quantity = %d, want 10", result[0].Quantity)
	}
}

func TestDiffItemStacks_GeneratesNewWarehouseID(t *testing.T) {
	// Verify that new items get a generated WarehouseID
	old := []MHFItemStack{}
	update := []MHFItemStack{
		{WarehouseID: 0, Item: MHFItem{ItemID: 100}, Quantity: 5},
	}

	// Reset RNG for consistent test
	token.RNG = token.NewSafeRand()

	result := DiffItemStacks(old, update)
	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}
	if result[0].WarehouseID == 0 {
		t.Error("New item should have generated WarehouseID, got 0")
	}
}
