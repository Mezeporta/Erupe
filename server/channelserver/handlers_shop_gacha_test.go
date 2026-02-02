package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
)

func TestWriteShopItems_Empty(t *testing.T) {
	bf := byteframe.NewByteFrame()
	items := []ShopItem{}

	writeShopItems(bf, items)

	result := byteframe.NewByteFrameFromBytes(bf.Data())
	count1 := result.ReadUint16()
	count2 := result.ReadUint16()

	if count1 != 0 {
		t.Errorf("Expected first count 0, got %d", count1)
	}
	if count2 != 0 {
		t.Errorf("Expected second count 0, got %d", count2)
	}
}

func TestWriteShopItems_SingleItem(t *testing.T) {
	bf := byteframe.NewByteFrame()
	items := []ShopItem{
		{
			ID:           1,
			ItemID:       100,
			Cost:         500,
			Quantity:     10,
			MinHR:        1,
			MinSR:        0,
			MinGR:        0,
			StoreLevel:   1,
			MaxQuantity:  99,
			UsedQuantity: 5,
			RoadFloors:   0,
			RoadFatalis:  0,
		},
	}

	writeShopItems(bf, items)

	result := byteframe.NewByteFrameFromBytes(bf.Data())
	count1 := result.ReadUint16()
	count2 := result.ReadUint16()

	if count1 != 1 {
		t.Errorf("Expected first count 1, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("Expected second count 1, got %d", count2)
	}

	// Read the item data
	id := result.ReadUint32()
	_ = result.ReadUint16() // padding
	itemID := result.ReadUint16()
	cost := result.ReadUint32()
	quantity := result.ReadUint16()
	minHR := result.ReadUint16()
	minSR := result.ReadUint16()
	minGR := result.ReadUint16()
	storeLevel := result.ReadUint16()
	maxQuantity := result.ReadUint16()
	usedQuantity := result.ReadUint16()
	roadFloors := result.ReadUint16()
	roadFatalis := result.ReadUint16()

	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}
	if itemID != 100 {
		t.Errorf("Expected itemID 100, got %d", itemID)
	}
	if cost != 500 {
		t.Errorf("Expected cost 500, got %d", cost)
	}
	if quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", quantity)
	}
	if minHR != 1 {
		t.Errorf("Expected minHR 1, got %d", minHR)
	}
	if minSR != 0 {
		t.Errorf("Expected minSR 0, got %d", minSR)
	}
	if minGR != 0 {
		t.Errorf("Expected minGR 0, got %d", minGR)
	}
	if storeLevel != 1 {
		t.Errorf("Expected storeLevel 1, got %d", storeLevel)
	}
	if maxQuantity != 99 {
		t.Errorf("Expected maxQuantity 99, got %d", maxQuantity)
	}
	if usedQuantity != 5 {
		t.Errorf("Expected usedQuantity 5, got %d", usedQuantity)
	}
	if roadFloors != 0 {
		t.Errorf("Expected roadFloors 0, got %d", roadFloors)
	}
	if roadFatalis != 0 {
		t.Errorf("Expected roadFatalis 0, got %d", roadFatalis)
	}
}

func TestWriteShopItems_MultipleItems(t *testing.T) {
	bf := byteframe.NewByteFrame()
	items := []ShopItem{
		{ID: 1, ItemID: 100, Cost: 500, Quantity: 10},
		{ID: 2, ItemID: 200, Cost: 1000, Quantity: 5},
		{ID: 3, ItemID: 300, Cost: 2000, Quantity: 1},
	}

	writeShopItems(bf, items)

	result := byteframe.NewByteFrameFromBytes(bf.Data())
	count1 := result.ReadUint16()
	count2 := result.ReadUint16()

	if count1 != 3 {
		t.Errorf("Expected first count 3, got %d", count1)
	}
	if count2 != 3 {
		t.Errorf("Expected second count 3, got %d", count2)
	}
}

// Test struct definitions
func TestShopItemStruct(t *testing.T) {
	item := ShopItem{
		ID:           42,
		ItemID:       1234,
		Cost:         9999,
		Quantity:     50,
		MinHR:        10,
		MinSR:        5,
		MinGR:        100,
		StoreLevel:   3,
		MaxQuantity:  99,
		UsedQuantity: 10,
		RoadFloors:   50,
		RoadFatalis:  25,
	}

	if item.ID != 42 {
		t.Errorf("ID = %d, want 42", item.ID)
	}
	if item.ItemID != 1234 {
		t.Errorf("ItemID = %d, want 1234", item.ItemID)
	}
	if item.Cost != 9999 {
		t.Errorf("Cost = %d, want 9999", item.Cost)
	}
}

func TestGachaStruct(t *testing.T) {
	gacha := Gacha{
		ID:           1,
		MinGR:        100,
		MinHR:        999,
		Name:         "Test Gacha",
		URLBanner:    "http://example.com/banner.png",
		URLFeature:   "http://example.com/feature.png",
		URLThumbnail: "http://example.com/thumb.png",
		Wide:         true,
		Recommended:  true,
		GachaType:    2,
		Hidden:       false,
	}

	if gacha.ID != 1 {
		t.Errorf("ID = %d, want 1", gacha.ID)
	}
	if gacha.Name != "Test Gacha" {
		t.Errorf("Name = %s, want Test Gacha", gacha.Name)
	}
	if !gacha.Wide {
		t.Error("Wide should be true")
	}
	if !gacha.Recommended {
		t.Error("Recommended should be true")
	}
}

func TestGachaEntryStruct(t *testing.T) {
	entry := GachaEntry{
		EntryType:      1,
		ID:             100,
		ItemType:       0,
		ItemNumber:     1234,
		ItemQuantity:   10,
		Weight:         0.5,
		Rarity:         3,
		Rolls:          1,
		FrontierPoints: 500,
		DailyLimit:     5,
	}

	if entry.EntryType != 1 {
		t.Errorf("EntryType = %d, want 1", entry.EntryType)
	}
	if entry.ID != 100 {
		t.Errorf("ID = %d, want 100", entry.ID)
	}
	if entry.Weight != 0.5 {
		t.Errorf("Weight = %f, want 0.5", entry.Weight)
	}
}

func TestGachaItemStruct(t *testing.T) {
	item := GachaItem{
		ItemType: 0,
		ItemID:   5678,
		Quantity: 20,
	}

	if item.ItemType != 0 {
		t.Errorf("ItemType = %d, want 0", item.ItemType)
	}
	if item.ItemID != 5678 {
		t.Errorf("ItemID = %d, want 5678", item.ItemID)
	}
	if item.Quantity != 20 {
		t.Errorf("Quantity = %d, want 20", item.Quantity)
	}
}
