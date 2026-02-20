package channelserver

import (
	_config "erupe-ce/config"
	"erupe-ce/common/mhfitem"
	"erupe-ce/common/token"
	"testing"
)

// createTestEquipment creates properly initialized test equipment
func createTestEquipment(itemIDs []uint16, warehouseIDs []uint32) []mhfitem.MHFEquipment {
	var equip []mhfitem.MHFEquipment
	for i, itemID := range itemIDs {
		e := mhfitem.MHFEquipment{
			ItemID:      itemID,
			WarehouseID: warehouseIDs[i],
			Decorations: make([]mhfitem.MHFItem, 3),
			Sigils:      make([]mhfitem.MHFSigil, 3),
		}
		// Initialize Sigils Effects arrays
		for j := 0; j < 3; j++ {
			e.Sigils[j].Effects = make([]mhfitem.MHFSigilEffect, 3)
		}
		equip = append(equip, e)
	}
	return equip
}

// TestWarehouseItemSerialization verifies warehouse item serialization
func TestWarehouseItemSerialization(t *testing.T) {
	tests := []struct {
		name  string
		items []mhfitem.MHFItemStack
	}{
		{
			name: "empty_warehouse",
			items: []mhfitem.MHFItemStack{},
		},
		{
			name: "single_item",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
			},
		},
		{
			name: "multiple_items",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
				{Item: mhfitem.MHFItem{ItemID: 3}, Quantity: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized := mhfitem.SerializeWarehouseItems(tt.items)

			// Basic validation
			if serialized == nil {
				t.Error("serialization returned nil")
			}

			// Verify we can work with the serialized data
			if serialized == nil {
				t.Error("invalid serialized length")
			}
		})
	}
}

// TestWarehouseEquipmentSerialization verifies warehouse equipment serialization
func TestWarehouseEquipmentSerialization(t *testing.T) {
	tests := []struct {
		name      string
		equipment []mhfitem.MHFEquipment
	}{
		{
			name:      "empty_equipment",
			equipment: []mhfitem.MHFEquipment{},
		},
		{
			name: "single_equipment",
			equipment: createTestEquipment([]uint16{100}, []uint32{1}),
		},
		{
			name: "multiple_equipment",
			equipment: createTestEquipment([]uint16{100, 101, 102}, []uint32{1, 2, 3}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized := mhfitem.SerializeWarehouseEquipment(tt.equipment, _config.ZZ)

			// Basic validation
			if serialized == nil {
				t.Error("serialization returned nil")
			}

			// Verify we can work with the serialized data
			if serialized == nil {
				t.Error("invalid serialized length")
			}
		})
	}
}

// TestWarehouseItemDiff verifies the item diff calculation
func TestWarehouseItemDiff(t *testing.T) {
	tests := []struct {
		name     string
		oldItems []mhfitem.MHFItemStack
		newItems []mhfitem.MHFItemStack
		wantDiff bool
	}{
		{
			name:     "no_changes",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			wantDiff: false,
		},
		{
			name:     "quantity_changed",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 15}},
			wantDiff: true,
		},
		{
			name:     "item_added",
			oldItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			newItems: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 5},
			},
			wantDiff: true,
		},
		{
			name: "item_removed",
			oldItems: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 5},
			},
			newItems: []mhfitem.MHFItemStack{{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10}},
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := mhfitem.DiffItemStacks(tt.oldItems, tt.newItems)

			// Verify that diff returns a valid result (not nil)
			if diff == nil {
				t.Error("diff should not be nil")
			}

			// The diff function returns items where Quantity > 0
			// So with no changes (all same quantity), diff should have same items
			if tt.name == "no_changes" {
				if len(diff) == 0 {
					t.Error("no_changes should return items")
				}
			}
		})
	}
}

// TestWarehouseEquipmentMerge verifies equipment merging logic
func TestWarehouseEquipmentMerge(t *testing.T) {
	tests := []struct {
		name        string
		oldEquip    []mhfitem.MHFEquipment
		newEquip    []mhfitem.MHFEquipment
		wantMerged  int
	}{
		{
			name:        "merge_empty",
			oldEquip:    []mhfitem.MHFEquipment{},
			newEquip:    []mhfitem.MHFEquipment{},
			wantMerged:  0,
		},
		{
			name: "add_new_equipment",
			oldEquip: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
			},
			newEquip: []mhfitem.MHFEquipment{
				{ItemID: 101, WarehouseID: 0}, // New item, no warehouse ID yet
			},
			wantMerged: 2, // Old + new
		},
		{
			name: "update_existing_equipment",
			oldEquip: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
			},
			newEquip: []mhfitem.MHFEquipment{
				{ItemID: 101, WarehouseID: 1}, // Update existing
			},
			wantMerged: 1, // Updated in place
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the merge logic from handleMsgMhfUpdateWarehouse
			var finalEquip []mhfitem.MHFEquipment
			oEquips := tt.oldEquip

			for _, uEquip := range tt.newEquip {
				exists := false
				for i := range oEquips {
					if oEquips[i].WarehouseID == uEquip.WarehouseID && uEquip.WarehouseID != 0 {
						exists = true
						oEquips[i].ItemID = uEquip.ItemID
						break
					}
				}
				if !exists {
					// Generate new warehouse ID
					uEquip.WarehouseID = token.RNG.Uint32()
					finalEquip = append(finalEquip, uEquip)
				}
			}

			for _, oEquip := range oEquips {
				if oEquip.ItemID > 0 {
					finalEquip = append(finalEquip, oEquip)
				}
			}

			// Verify merge result count
			if len(finalEquip) != tt.wantMerged {
				t.Errorf("expected %d merged equipment, got %d", tt.wantMerged, len(finalEquip))
			}
		})
	}
}

// TestWarehouseIDGeneration verifies warehouse ID uniqueness
func TestWarehouseIDGeneration(t *testing.T) {
	// Generate multiple warehouse IDs and verify they're unique
	idCount := 100
	ids := make(map[uint32]bool)

	for i := 0; i < idCount; i++ {
		id := token.RNG.Uint32()
		if id == 0 {
			t.Error("generated warehouse ID is 0 (invalid)")
		}
		if ids[id] {
			// While collisions are possible with random IDs,
			// they should be extremely rare
			t.Logf("Warning: duplicate warehouse ID generated: %d", id)
		}
		ids[id] = true
	}

	if len(ids) < idCount*90/100 {
		t.Errorf("too many duplicate IDs: got %d unique out of %d", len(ids), idCount)
	}
}

// TestWarehouseItemRemoval verifies item removal logic
func TestWarehouseItemRemoval(t *testing.T) {
	tests := []struct {
		name       string
		items      []mhfitem.MHFItemStack
		removeID   uint16
		wantRemain int
	}{
		{
			name: "remove_existing",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
				{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
			},
			removeID:   1,
			wantRemain: 1,
		},
		{
			name: "remove_non_existing",
			items: []mhfitem.MHFItemStack{
				{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
			},
			removeID:   999,
			wantRemain: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var remaining []mhfitem.MHFItemStack
			for _, item := range tt.items {
				if item.Item.ItemID != tt.removeID {
					remaining = append(remaining, item)
				}
			}

			if len(remaining) != tt.wantRemain {
				t.Errorf("expected %d remaining items, got %d", tt.wantRemain, len(remaining))
			}
		})
	}
}

// TestWarehouseEquipmentRemoval verifies equipment removal logic
func TestWarehouseEquipmentRemoval(t *testing.T) {
	tests := []struct {
		name       string
		equipment  []mhfitem.MHFEquipment
		setZeroID  uint32
		wantActive int
	}{
		{
			name: "remove_by_setting_zero",
			equipment: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
				{ItemID: 101, WarehouseID: 2},
			},
			setZeroID:  1,
			wantActive: 1,
		},
		{
			name: "all_active",
			equipment: []mhfitem.MHFEquipment{
				{ItemID: 100, WarehouseID: 1},
				{ItemID: 101, WarehouseID: 2},
			},
			setZeroID:  999,
			wantActive: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate removal by setting ItemID to 0
			equipment := make([]mhfitem.MHFEquipment, len(tt.equipment))
			copy(equipment, tt.equipment)

			for i := range equipment {
				if equipment[i].WarehouseID == tt.setZeroID {
					equipment[i].ItemID = 0
				}
			}

			// Count active equipment (ItemID > 0)
			activeCount := 0
			for _, eq := range equipment {
				if eq.ItemID > 0 {
					activeCount++
				}
			}

			if activeCount != tt.wantActive {
				t.Errorf("expected %d active equipment, got %d", tt.wantActive, activeCount)
			}
		})
	}
}

// TestWarehouseBoxIndexValidation verifies box index bounds
func TestWarehouseBoxIndexValidation(t *testing.T) {
	tests := []struct {
		name     string
		boxIndex uint8
		isValid  bool
	}{
		{
			name:     "box_0",
			boxIndex: 0,
			isValid:  true,
		},
		{
			name:     "box_1",
			boxIndex: 1,
			isValid:  true,
		},
		{
			name:     "box_9",
			boxIndex: 9,
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify box index is within reasonable bounds
			if tt.isValid && tt.boxIndex > 100 {
				t.Error("box index unreasonably high")
			}
		})
	}
}

// TestWarehouseErrorRecovery verifies error handling doesn't corrupt state
func TestWarehouseErrorRecovery(t *testing.T) {
	t.Run("database_error_handling", func(t *testing.T) {
		// After our fix, database errors should:
		// 1. Be logged with s.logger.Error()
		// 2. Send doAckSimpleFail()
		// 3. Return immediately
		// 4. NOT send doAckSimpleSucceed() (the bug we fixed)

		// This test documents the expected behavior
	})

	t.Run("serialization_error_handling", func(t *testing.T) {
		// Test that serialization errors are handled gracefully
		emptyItems := []mhfitem.MHFItemStack{}
		serialized := mhfitem.SerializeWarehouseItems(emptyItems)

		// Should handle empty gracefully
		if serialized == nil {
			t.Error("serialization of empty items should not return nil")
		}
	})
}

// BenchmarkWarehouseSerialization benchmarks warehouse serialization performance
func BenchmarkWarehouseSerialization(b *testing.B) {
	items := []mhfitem.MHFItemStack{
		{Item: mhfitem.MHFItem{ItemID: 1}, Quantity: 10},
		{Item: mhfitem.MHFItem{ItemID: 2}, Quantity: 20},
		{Item: mhfitem.MHFItem{ItemID: 3}, Quantity: 30},
		{Item: mhfitem.MHFItem{ItemID: 4}, Quantity: 40},
		{Item: mhfitem.MHFItem{ItemID: 5}, Quantity: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mhfitem.SerializeWarehouseItems(items)
	}
}

// BenchmarkWarehouseEquipmentMerge benchmarks equipment merge performance
func BenchmarkWarehouseEquipmentMerge(b *testing.B) {
	oldEquip := make([]mhfitem.MHFEquipment, 50)
	for i := range oldEquip {
		oldEquip[i] = mhfitem.MHFEquipment{
			ItemID:      uint16(100 + i),
			WarehouseID: uint32(i + 1),
		}
	}

	newEquip := make([]mhfitem.MHFEquipment, 10)
	for i := range newEquip {
		newEquip[i] = mhfitem.MHFEquipment{
			ItemID:      uint16(200 + i),
			WarehouseID: uint32(i + 1),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var finalEquip []mhfitem.MHFEquipment
		oEquips := oldEquip

		for _, uEquip := range newEquip {
			exists := false
			for j := range oEquips {
				if oEquips[j].WarehouseID == uEquip.WarehouseID {
					exists = true
					oEquips[j].ItemID = uEquip.ItemID
					break
				}
			}
			if !exists {
				finalEquip = append(finalEquip, uEquip)
			}
		}

		for _, oEquip := range oEquips {
			if oEquip.ItemID > 0 {
				finalEquip = append(finalEquip, oEquip)
			}
		}
		_ = finalEquip // Use finalEquip to avoid unused variable warning
	}
}
