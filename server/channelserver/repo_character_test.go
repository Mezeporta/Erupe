package channelserver

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupCharRepo(t *testing.T) (*CharacterRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "repo_test_user")
	charID := CreateTestCharacter(t, db, userID, "RepoChar")
	repo := NewCharacterRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestLoadColumn(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	// Write a known blob to a column
	blob := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	_, err := db.Exec("UPDATE characters SET otomoairou=$1 WHERE id=$2", blob, charID)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	data, err := repo.LoadColumn(charID, "otomoairou")
	if err != nil {
		t.Fatalf("LoadColumn failed: %v", err)
	}
	if len(data) != 4 || data[0] != 0xDE || data[3] != 0xEF {
		t.Errorf("LoadColumn returned unexpected data: %x", data)
	}
}

func TestLoadColumnNil(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// Column should be NULL by default
	data, err := repo.LoadColumn(charID, "otomoairou")
	if err != nil {
		t.Fatalf("LoadColumn failed: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil for NULL column, got: %x", data)
	}
}

func TestSaveColumn(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	blob := []byte{0x01, 0x02, 0x03}
	if err := repo.SaveColumn(charID, "otomoairou", blob); err != nil {
		t.Fatalf("SaveColumn failed: %v", err)
	}

	// Verify via direct SELECT
	var got []byte
	if err := db.QueryRow("SELECT otomoairou FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0x01 || got[2] != 0x03 {
		t.Errorf("SaveColumn wrote unexpected data: %x", got)
	}
}

func TestReadInt(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// time_played defaults to 0 via COALESCE
	val, err := repo.ReadInt(charID, "time_played")
	if err != nil {
		t.Fatalf("ReadInt failed: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected 0 for default time_played, got: %d", val)
	}
}

func TestReadIntWithValue(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	_, err := db.Exec("UPDATE characters SET time_played=42 WHERE id=$1", charID)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	val, err := repo.ReadInt(charID, "time_played")
	if err != nil {
		t.Fatalf("ReadInt failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got: %d", val)
	}
}

func TestAdjustInt(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// First adjustment from NULL (COALESCE makes it 0 + 10 = 10)
	val, err := repo.AdjustInt(charID, "time_played", 10)
	if err != nil {
		t.Fatalf("AdjustInt failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10 after first adjust, got: %d", val)
	}

	// Second adjustment: 10 + 5 = 15
	val, err = repo.AdjustInt(charID, "time_played", 5)
	if err != nil {
		t.Fatalf("AdjustInt failed: %v", err)
	}
	if val != 15 {
		t.Errorf("Expected 15 after second adjust, got: %d", val)
	}
}

func TestGetName(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	name, err := repo.GetName(charID)
	if err != nil {
		t.Fatalf("GetName failed: %v", err)
	}
	if name != "RepoChar" {
		t.Errorf("Expected 'RepoChar', got: %q", name)
	}
}

func TestGetUserID(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	// Look up the expected user_id
	var expectedUID uint32
	if err := db.QueryRow("SELECT user_id FROM characters WHERE id=$1", charID).Scan(&expectedUID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	uid, err := repo.GetUserID(charID)
	if err != nil {
		t.Fatalf("GetUserID failed: %v", err)
	}
	if uid != expectedUID {
		t.Errorf("Expected user_id %d, got: %d", expectedUID, uid)
	}
}

func TestUpdateLastLogin(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	ts := int64(1700000000)
	if err := repo.UpdateLastLogin(charID, ts); err != nil {
		t.Fatalf("UpdateLastLogin failed: %v", err)
	}

	var got int64
	if err := db.QueryRow("SELECT last_login FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != ts {
		t.Errorf("Expected last_login %d, got: %d", ts, got)
	}
}

func TestUpdateTimePlayed(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.UpdateTimePlayed(charID, 999); err != nil {
		t.Fatalf("UpdateTimePlayed failed: %v", err)
	}

	var got int
	if err := db.QueryRow("SELECT time_played FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != 999 {
		t.Errorf("Expected time_played 999, got: %d", got)
	}
}

func TestGetCharIDsByUserID(t *testing.T) {
	repo, db, _ := setupCharRepo(t)

	// Create a second user with multiple characters
	uid2 := CreateTestUser(t, db, "multi_char_user")
	cid1 := CreateTestCharacter(t, db, uid2, "Char1")
	cid2 := CreateTestCharacter(t, db, uid2, "Char2")

	ids, err := repo.GetCharIDsByUserID(uid2)
	if err != nil {
		t.Fatalf("GetCharIDsByUserID failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("Expected 2 character IDs, got: %d", len(ids))
	}

	// Check both IDs are present (order may vary)
	found := map[uint32]bool{cid1: false, cid2: false}
	for _, id := range ids {
		found[id] = true
	}
	if !found[cid1] || !found[cid2] {
		t.Errorf("Expected IDs %d and %d, got: %v", cid1, cid2, ids)
	}
}

func TestGetCharIDsByUserIDEmpty(t *testing.T) {
	repo, db, _ := setupCharRepo(t)

	// Create a user with no characters
	uid := CreateTestUser(t, db, "no_chars_user")

	ids, err := repo.GetCharIDsByUserID(uid)
	if err != nil {
		t.Fatalf("GetCharIDsByUserID failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected 0 character IDs for user with no chars, got: %d", len(ids))
	}
}

func TestReadTimeNull(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	defaultTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := repo.ReadTime(charID, "daily_time", defaultTime)
	if err != nil {
		t.Fatalf("ReadTime failed: %v", err)
	}
	if !got.Equal(defaultTime) {
		t.Errorf("Expected default time %v, got: %v", defaultTime, got)
	}
}

func TestReadTimeWithValue(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	expected := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if _, err := db.Exec("UPDATE characters SET daily_time=$1 WHERE id=$2", expected, charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.ReadTime(charID, "daily_time", time.Time{})
	if err != nil {
		t.Fatalf("ReadTime failed: %v", err)
	}
	if !got.Equal(expected) {
		t.Errorf("Expected %v, got: %v", expected, got)
	}
}

func TestSaveTime(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	expected := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := repo.SaveTime(charID, "daily_time", expected); err != nil {
		t.Fatalf("SaveTime failed: %v", err)
	}

	var got time.Time
	if err := db.QueryRow("SELECT daily_time FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !got.Equal(expected) {
		t.Errorf("Expected %v, got: %v", expected, got)
	}
}

func TestSaveInt(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.SaveInt(charID, "netcafe_points", 500); err != nil {
		t.Fatalf("SaveInt failed: %v", err)
	}

	var got int
	if err := db.QueryRow("SELECT netcafe_points FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != 500 {
		t.Errorf("Expected 500, got: %d", got)
	}
}

func TestSaveBool(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.SaveBool(charID, "restrict_guild_scout", true); err != nil {
		t.Fatalf("SaveBool failed: %v", err)
	}

	var got bool
	if err := db.QueryRow("SELECT restrict_guild_scout FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !got {
		t.Errorf("Expected true, got false")
	}
}

func TestReadBool(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET restrict_guild_scout=true WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.ReadBool(charID, "restrict_guild_scout")
	if err != nil {
		t.Fatalf("ReadBool failed: %v", err)
	}
	if !got {
		t.Errorf("Expected true, got false")
	}
}

func TestSaveString(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.SaveString(charID, "friends", "1,2,3"); err != nil {
		t.Fatalf("SaveString failed: %v", err)
	}

	var got string
	if err := db.QueryRow("SELECT friends FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != "1,2,3" {
		t.Errorf("Expected '1,2,3', got: %q", got)
	}
}

func TestReadString(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET friends='4,5,6' WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.ReadString(charID, "friends")
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if got != "4,5,6" {
		t.Errorf("Expected '4,5,6', got: %q", got)
	}
}

func TestReadStringNull(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	got, err := repo.ReadString(charID, "friends")
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if got != "" {
		t.Errorf("Expected empty string for NULL, got: %q", got)
	}
}

func TestLoadColumnWithDefault(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	defaultVal := []byte{0x00, 0x01, 0x02}
	got, err := repo.LoadColumnWithDefault(charID, "skin_hist", defaultVal)
	if err != nil {
		t.Fatalf("LoadColumnWithDefault failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0x00 || got[2] != 0x02 {
		t.Errorf("Expected default value, got: %x", got)
	}
}

func TestLoadColumnWithDefaultExistingData(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	blob := []byte{0xAA, 0xBB}
	if _, err := db.Exec("UPDATE characters SET skin_hist=$1 WHERE id=$2", blob, charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.LoadColumnWithDefault(charID, "skin_hist", []byte{0x00})
	if err != nil {
		t.Fatalf("LoadColumnWithDefault failed: %v", err)
	}
	if len(got) != 2 || got[0] != 0xAA || got[1] != 0xBB {
		t.Errorf("Expected stored data, got: %x", got)
	}
}

func TestSetDeleted(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.SetDeleted(charID); err != nil {
		t.Fatalf("SetDeleted failed: %v", err)
	}

	var deleted bool
	if err := db.QueryRow("SELECT deleted FROM characters WHERE id=$1", charID).Scan(&deleted); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !deleted {
		t.Errorf("Expected deleted=true")
	}
}

func TestUpdateDailyCafe(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	dailyTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := repo.UpdateDailyCafe(charID, dailyTime, 5, 10); err != nil {
		t.Fatalf("UpdateDailyCafe failed: %v", err)
	}

	var gotTime time.Time
	var bonus, daily uint32
	if err := db.QueryRow("SELECT daily_time, bonus_quests, daily_quests FROM characters WHERE id=$1", charID).Scan(&gotTime, &bonus, &daily); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !gotTime.Equal(dailyTime) {
		t.Errorf("Expected daily_time %v, got: %v", dailyTime, gotTime)
	}
	if bonus != 5 || daily != 10 {
		t.Errorf("Expected bonus=5 daily=10, got bonus=%d daily=%d", bonus, daily)
	}
}

func TestResetDailyQuests(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET bonus_quests=5, daily_quests=10 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.ResetDailyQuests(charID); err != nil {
		t.Fatalf("ResetDailyQuests failed: %v", err)
	}

	var bonus, daily uint32
	if err := db.QueryRow("SELECT bonus_quests, daily_quests FROM characters WHERE id=$1", charID).Scan(&bonus, &daily); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if bonus != 0 || daily != 0 {
		t.Errorf("Expected bonus=0 daily=0, got bonus=%d daily=%d", bonus, daily)
	}
}

func TestReadEtcPoints(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET bonus_quests=3, daily_quests=7, promo_points=100 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	bonus, daily, promo, err := repo.ReadEtcPoints(charID)
	if err != nil {
		t.Fatalf("ReadEtcPoints failed: %v", err)
	}
	if bonus != 3 || daily != 7 || promo != 100 {
		t.Errorf("Expected 3/7/100, got %d/%d/%d", bonus, daily, promo)
	}
}

func TestResetCafeTime(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET cafe_time=999 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	cafeReset := time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)
	if err := repo.ResetCafeTime(charID, cafeReset); err != nil {
		t.Fatalf("ResetCafeTime failed: %v", err)
	}

	var cafeTime int
	var gotReset time.Time
	if err := db.QueryRow("SELECT cafe_time, cafe_reset FROM characters WHERE id=$1", charID).Scan(&cafeTime, &gotReset); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if cafeTime != 0 {
		t.Errorf("Expected cafe_time=0, got: %d", cafeTime)
	}
	if !gotReset.Equal(cafeReset) {
		t.Errorf("Expected cafe_reset %v, got: %v", cafeReset, gotReset)
	}
}

func TestUpdateGuildPostChecked(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	before := time.Now().Add(-time.Second)
	if err := repo.UpdateGuildPostChecked(charID); err != nil {
		t.Fatalf("UpdateGuildPostChecked failed: %v", err)
	}

	var got time.Time
	if err := db.QueryRow("SELECT guild_post_checked FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got.Before(before) {
		t.Errorf("Expected guild_post_checked to be recent, got: %v", got)
	}
}

func TestReadGuildPostChecked(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	expected := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if _, err := db.Exec("UPDATE characters SET guild_post_checked=$1 WHERE id=$2", expected, charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.ReadGuildPostChecked(charID)
	if err != nil {
		t.Fatalf("ReadGuildPostChecked failed: %v", err)
	}
	if !got.Equal(expected) {
		t.Errorf("Expected %v, got: %v", expected, got)
	}
}

func TestSaveMercenary(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	data := []byte{0x01, 0x02, 0x03, 0x04}
	if err := repo.SaveMercenary(charID, data, 42); err != nil {
		t.Fatalf("SaveMercenary failed: %v", err)
	}

	var gotData []byte
	var gotRastaID uint32
	if err := db.QueryRow("SELECT savemercenary, rasta_id FROM characters WHERE id=$1", charID).Scan(&gotData, &gotRastaID); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if len(gotData) != 4 || gotData[0] != 0x01 {
		t.Errorf("Expected mercenary data, got: %x", gotData)
	}
	if gotRastaID != 42 {
		t.Errorf("Expected rasta_id=42, got: %d", gotRastaID)
	}
}

func TestUpdateGCPAndPact(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.UpdateGCPAndPact(charID, 100, 55); err != nil {
		t.Fatalf("UpdateGCPAndPact failed: %v", err)
	}

	var gcp, pactID uint32
	if err := db.QueryRow("SELECT gcp, pact_id FROM characters WHERE id=$1", charID).Scan(&gcp, &pactID); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if gcp != 100 || pactID != 55 {
		t.Errorf("Expected gcp=100 pact_id=55, got gcp=%d pact_id=%d", gcp, pactID)
	}
}

func TestFindByRastaID(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if _, err := db.Exec("UPDATE characters SET rasta_id=999 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	gotID, gotName, err := repo.FindByRastaID(999)
	if err != nil {
		t.Fatalf("FindByRastaID failed: %v", err)
	}
	if gotID != charID {
		t.Errorf("Expected charID %d, got: %d", charID, gotID)
	}
	if gotName != "RepoChar" {
		t.Errorf("Expected 'RepoChar', got: %q", gotName)
	}
}
