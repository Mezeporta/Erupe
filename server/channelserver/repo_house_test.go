package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupHouseRepo(t *testing.T) (*HouseRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "house_test_user")
	charID := CreateTestCharacter(t, db, userID, "HouseChar")
	CreateTestUserBinary(t, db, charID)
	repo := NewHouseRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestRepoHouseGetHouseByCharID(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	house, err := repo.GetHouseByCharID(charID)
	if err != nil {
		t.Fatalf("GetHouseByCharID failed: %v", err)
	}
	if house.CharID != charID {
		t.Errorf("Expected charID=%d, got: %d", charID, house.CharID)
	}
	if house.Name != "HouseChar" {
		t.Errorf("Expected name='HouseChar', got: %q", house.Name)
	}
	// Default house_state is 2 (password-protected) via COALESCE
	if house.HouseState != 2 {
		t.Errorf("Expected default house_state=2, got: %d", house.HouseState)
	}
}

func TestRepoHouseSearchHousesByName(t *testing.T) {
	repo, db, _ := setupHouseRepo(t)

	user2 := CreateTestUser(t, db, "house_user2")
	charID2 := CreateTestCharacter(t, db, user2, "HouseAlpha")
	CreateTestUserBinary(t, db, charID2)
	user3 := CreateTestUser(t, db, "house_user3")
	charID3 := CreateTestCharacter(t, db, user3, "BetaHouse")
	CreateTestUserBinary(t, db, charID3)

	houses, err := repo.SearchHousesByName("House")
	if err != nil {
		t.Fatalf("SearchHousesByName failed: %v", err)
	}
	if len(houses) < 2 {
		t.Errorf("Expected at least 2 matches for 'House', got: %d", len(houses))
	}
}

func TestRepoHouseSearchHousesByNameNoMatch(t *testing.T) {
	repo, _, _ := setupHouseRepo(t)

	houses, err := repo.SearchHousesByName("ZZZnonexistent")
	if err != nil {
		t.Fatalf("SearchHousesByName failed: %v", err)
	}
	if len(houses) != 0 {
		t.Errorf("Expected 0 matches, got: %d", len(houses))
	}
}

func TestRepoHouseUpdateHouseState(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	if err := repo.UpdateHouseState(charID, 1, "secret"); err != nil {
		t.Fatalf("UpdateHouseState failed: %v", err)
	}

	state, password, err := repo.GetHouseAccess(charID)
	if err != nil {
		t.Fatalf("GetHouseAccess failed: %v", err)
	}
	if state != 1 {
		t.Errorf("Expected state=1, got: %d", state)
	}
	if password != "secret" {
		t.Errorf("Expected password='secret', got: %q", password)
	}
}

func TestRepoHouseGetHouseAccessDefault(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	state, password, err := repo.GetHouseAccess(charID)
	if err != nil {
		t.Fatalf("GetHouseAccess failed: %v", err)
	}
	if state != 2 {
		t.Errorf("Expected default state=2, got: %d", state)
	}
	if password != "" {
		t.Errorf("Expected empty password, got: %q", password)
	}
}

func TestRepoHouseUpdateInterior(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	furniture := []byte{0x01, 0x02, 0x03}
	if err := repo.UpdateInterior(charID, furniture); err != nil {
		t.Fatalf("UpdateInterior failed: %v", err)
	}

	var got []byte
	if err := db.QueryRow("SELECT house_furniture FROM user_binary WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0x01 {
		t.Errorf("Expected furniture data, got: %x", got)
	}
}

func TestRepoHouseGetHouseContents(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	tier := []byte{0x01}
	data := []byte{0x02}
	furniture := []byte{0x03}
	bookshelf := []byte{0x04}
	gallery := []byte{0x05}
	tore := []byte{0x06}
	garden := []byte{0x07}
	if _, err := db.Exec(
		"UPDATE user_binary SET house_tier=$1, house_data=$2, house_furniture=$3, bookshelf=$4, gallery=$5, tore=$6, garden=$7 WHERE id=$8",
		tier, data, furniture, bookshelf, gallery, tore, garden, charID,
	); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	gotTier, gotData, gotFurniture, gotBookshelf, gotGallery, gotTore, gotGarden, err := repo.GetHouseContents(charID)
	if err != nil {
		t.Fatalf("GetHouseContents failed: %v", err)
	}
	if len(gotTier) != 1 || gotTier[0] != 0x01 {
		t.Errorf("Unexpected tier: %x", gotTier)
	}
	if len(gotData) != 1 || gotData[0] != 0x02 {
		t.Errorf("Unexpected data: %x", gotData)
	}
	if len(gotFurniture) != 1 || gotFurniture[0] != 0x03 {
		t.Errorf("Unexpected furniture: %x", gotFurniture)
	}
	if len(gotBookshelf) != 1 || gotBookshelf[0] != 0x04 {
		t.Errorf("Unexpected bookshelf: %x", gotBookshelf)
	}
	if len(gotGallery) != 1 || gotGallery[0] != 0x05 {
		t.Errorf("Unexpected gallery: %x", gotGallery)
	}
	if len(gotTore) != 1 || gotTore[0] != 0x06 {
		t.Errorf("Unexpected tore: %x", gotTore)
	}
	if len(gotGarden) != 1 || gotGarden[0] != 0x07 {
		t.Errorf("Unexpected garden: %x", gotGarden)
	}
}

func TestRepoHouseGetMission(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	mission := []byte{0xAA, 0xBB}
	if _, err := db.Exec("UPDATE user_binary SET mission=$1 WHERE id=$2", mission, charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.GetMission(charID)
	if err != nil {
		t.Fatalf("GetMission failed: %v", err)
	}
	if len(got) != 2 || got[0] != 0xAA {
		t.Errorf("Expected mission data, got: %x", got)
	}
}

func TestRepoHouseUpdateMission(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	mission := []byte{0xCC, 0xDD, 0xEE}
	if err := repo.UpdateMission(charID, mission); err != nil {
		t.Fatalf("UpdateMission failed: %v", err)
	}

	var got []byte
	if err := db.QueryRow("SELECT mission FROM user_binary WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0xCC {
		t.Errorf("Expected mission data, got: %x", got)
	}
}

func TestRepoHouseInitializeWarehouse(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("InitializeWarehouse failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM warehouse WHERE character_id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 warehouse row, got: %d", count)
	}

	// Calling again should be idempotent
	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("Second InitializeWarehouse failed: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM warehouse WHERE character_id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected still 1 warehouse row after idempotent call, got: %d", count)
	}
}

func TestRepoHouseGetWarehouseNames(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("InitializeWarehouse failed: %v", err)
	}
	if _, err := db.Exec("UPDATE warehouse SET item0name='Items Box 0', equip3name='Equip Box 3' WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	itemNames, equipNames, err := repo.GetWarehouseNames(charID)
	if err != nil {
		t.Fatalf("GetWarehouseNames failed: %v", err)
	}
	if itemNames[0] != "Items Box 0" {
		t.Errorf("Expected item0name='Items Box 0', got: %q", itemNames[0])
	}
	if equipNames[3] != "Equip Box 3" {
		t.Errorf("Expected equip3name='Equip Box 3', got: %q", equipNames[3])
	}
	// Other names should be empty (COALESCE)
	if itemNames[1] != "" {
		t.Errorf("Expected empty item1name, got: %q", itemNames[1])
	}
}

func TestRepoHouseRenameWarehouseBox(t *testing.T) {
	repo, db, charID := setupHouseRepo(t)

	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("InitializeWarehouse failed: %v", err)
	}

	if err := repo.RenameWarehouseBox(charID, 0, 5, "My Items"); err != nil {
		t.Fatalf("RenameWarehouseBox(item) failed: %v", err)
	}
	if err := repo.RenameWarehouseBox(charID, 1, 2, "My Equips"); err != nil {
		t.Fatalf("RenameWarehouseBox(equip) failed: %v", err)
	}

	var item5name, equip2name string
	if err := db.QueryRow("SELECT COALESCE(item5name,''), COALESCE(equip2name,'') FROM warehouse WHERE character_id=$1", charID).Scan(&item5name, &equip2name); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if item5name != "My Items" {
		t.Errorf("Expected item5name='My Items', got: %q", item5name)
	}
	if equip2name != "My Equips" {
		t.Errorf("Expected equip2name='My Equips', got: %q", equip2name)
	}
}

func TestRepoHouseRenameWarehouseBoxInvalidType(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	err := repo.RenameWarehouseBox(charID, 5, 0, "Bad")
	if err == nil {
		t.Fatal("Expected error for invalid box type, got nil")
	}
}

func TestRepoHouseWarehouseItemData(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("InitializeWarehouse failed: %v", err)
	}

	data := []byte{0x01, 0x02, 0x03}
	if err := repo.SetWarehouseItemData(charID, 3, data); err != nil {
		t.Fatalf("SetWarehouseItemData failed: %v", err)
	}

	got, err := repo.GetWarehouseItemData(charID, 3)
	if err != nil {
		t.Fatalf("GetWarehouseItemData failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0x01 {
		t.Errorf("Expected item data, got: %x", got)
	}
}

func TestRepoHouseWarehouseEquipData(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	if err := repo.InitializeWarehouse(charID); err != nil {
		t.Fatalf("InitializeWarehouse failed: %v", err)
	}

	data := []byte{0xAA, 0xBB}
	if err := repo.SetWarehouseEquipData(charID, 7, data); err != nil {
		t.Fatalf("SetWarehouseEquipData failed: %v", err)
	}

	got, err := repo.GetWarehouseEquipData(charID, 7)
	if err != nil {
		t.Fatalf("GetWarehouseEquipData failed: %v", err)
	}
	if len(got) != 2 || got[0] != 0xAA {
		t.Errorf("Expected equip data, got: %x", got)
	}
}

func TestRepoHouseAcquireTitle(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	if err := repo.AcquireTitle(100, charID); err != nil {
		t.Fatalf("AcquireTitle failed: %v", err)
	}

	titles, err := repo.GetTitles(charID)
	if err != nil {
		t.Fatalf("GetTitles failed: %v", err)
	}
	if len(titles) != 1 {
		t.Fatalf("Expected 1 title, got: %d", len(titles))
	}
	if titles[0].ID != 100 {
		t.Errorf("Expected title ID=100, got: %d", titles[0].ID)
	}
}

func TestRepoHouseAcquireTitleIdempotent(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	if err := repo.AcquireTitle(100, charID); err != nil {
		t.Fatalf("First AcquireTitle failed: %v", err)
	}
	if err := repo.AcquireTitle(100, charID); err != nil {
		t.Fatalf("Second AcquireTitle failed: %v", err)
	}

	titles, err := repo.GetTitles(charID)
	if err != nil {
		t.Fatalf("GetTitles failed: %v", err)
	}
	if len(titles) != 1 {
		t.Errorf("Expected 1 title after idempotent acquire, got: %d", len(titles))
	}
}

func TestRepoHouseGetTitlesEmpty(t *testing.T) {
	repo, _, charID := setupHouseRepo(t)

	titles, err := repo.GetTitles(charID)
	if err != nil {
		t.Fatalf("GetTitles failed: %v", err)
	}
	if len(titles) != 0 {
		t.Errorf("Expected 0 titles, got: %d", len(titles))
	}
}
