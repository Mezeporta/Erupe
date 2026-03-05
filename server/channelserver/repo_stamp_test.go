package channelserver

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupStampRepo(t *testing.T) (*StampRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "stamp_test_user")
	charID := CreateTestCharacter(t, db, userID, "StampChar")
	repo := NewStampRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func initStamp(t *testing.T, repo *StampRepository, charID uint32) {
	t.Helper()
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := repo.Init(charID, now); err != nil {
		t.Fatalf("Stamp Init failed: %v", err)
	}
}

func TestRepoStampInit(t *testing.T) {
	repo, db, charID := setupStampRepo(t)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := repo.Init(charID, now); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var hlChecked, exChecked time.Time
	if err := db.QueryRow("SELECT hl_checked, ex_checked FROM stamps WHERE character_id=$1", charID).Scan(&hlChecked, &exChecked); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !hlChecked.Equal(now) {
		t.Errorf("Expected hl_checked=%v, got: %v", now, hlChecked)
	}
	if !exChecked.Equal(now) {
		t.Errorf("Expected ex_checked=%v, got: %v", now, exChecked)
	}
}

func TestRepoStampGetChecked(t *testing.T) {
	repo, _, charID := setupStampRepo(t)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := repo.Init(charID, now); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	got, err := repo.GetChecked(charID, "hl")
	if err != nil {
		t.Fatalf("GetChecked failed: %v", err)
	}
	if !got.Equal(now) {
		t.Errorf("Expected %v, got: %v", now, got)
	}
}

func TestRepoStampSetChecked(t *testing.T) {
	repo, _, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	newTime := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.SetChecked(charID, "ex", newTime); err != nil {
		t.Fatalf("SetChecked failed: %v", err)
	}

	got, err := repo.GetChecked(charID, "ex")
	if err != nil {
		t.Fatalf("GetChecked failed: %v", err)
	}
	if !got.Equal(newTime) {
		t.Errorf("Expected %v, got: %v", newTime, got)
	}
}

func TestRepoStampIncrementTotal(t *testing.T) {
	repo, _, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	if err := repo.IncrementTotal(charID, "hl"); err != nil {
		t.Fatalf("First IncrementTotal failed: %v", err)
	}
	if err := repo.IncrementTotal(charID, "hl"); err != nil {
		t.Fatalf("Second IncrementTotal failed: %v", err)
	}

	total, redeemed, err := repo.GetTotals(charID, "hl")
	if err != nil {
		t.Fatalf("GetTotals failed: %v", err)
	}
	if total != 2 {
		t.Errorf("Expected total=2, got: %d", total)
	}
	if redeemed != 0 {
		t.Errorf("Expected redeemed=0, got: %d", redeemed)
	}
}

func TestRepoStampGetTotals(t *testing.T) {
	repo, db, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	if _, err := db.Exec("UPDATE stamps SET hl_total=10, hl_redeemed=3 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	total, redeemed, err := repo.GetTotals(charID, "hl")
	if err != nil {
		t.Fatalf("GetTotals failed: %v", err)
	}
	if total != 10 || redeemed != 3 {
		t.Errorf("Expected total=10 redeemed=3, got total=%d redeemed=%d", total, redeemed)
	}
}

func TestRepoStampExchange(t *testing.T) {
	repo, db, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	if _, err := db.Exec("UPDATE stamps SET hl_total=20, hl_redeemed=0 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	total, redeemed, err := repo.Exchange(charID, "hl")
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}
	if total != 20 {
		t.Errorf("Expected total=20, got: %d", total)
	}
	if redeemed != 8 {
		t.Errorf("Expected redeemed=8, got: %d", redeemed)
	}
}

func TestRepoStampExchangeYearly(t *testing.T) {
	repo, db, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	if _, err := db.Exec("UPDATE stamps SET hl_total=100, hl_redeemed=50 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	total, redeemed, err := repo.ExchangeYearly(charID)
	if err != nil {
		t.Fatalf("ExchangeYearly failed: %v", err)
	}
	if total != 52 {
		t.Errorf("Expected total=52 (100-48), got: %d", total)
	}
	if redeemed != 2 {
		t.Errorf("Expected redeemed=2 (50-48), got: %d", redeemed)
	}
}

func TestRepoStampGetMonthlyClaimed(t *testing.T) {
	repo, db, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	claimedTime := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if _, err := db.Exec("UPDATE stamps SET monthly_claimed=$1 WHERE character_id=$2", claimedTime, charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	got, err := repo.GetMonthlyClaimed(charID, "monthly")
	if err != nil {
		t.Fatalf("GetMonthlyClaimed failed: %v", err)
	}
	if !got.Equal(claimedTime) {
		t.Errorf("Expected %v, got: %v", claimedTime, got)
	}
}

func TestRepoStampSetMonthlyClaimed(t *testing.T) {
	repo, _, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	claimedTime := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.SetMonthlyClaimed(charID, "monthly", claimedTime); err != nil {
		t.Fatalf("SetMonthlyClaimed failed: %v", err)
	}

	got, err := repo.GetMonthlyClaimed(charID, "monthly")
	if err != nil {
		t.Fatalf("GetMonthlyClaimed failed: %v", err)
	}
	if !got.Equal(claimedTime) {
		t.Errorf("Expected %v, got: %v", claimedTime, got)
	}
}

func TestRepoStampExTypes(t *testing.T) {
	repo, db, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	// Verify ex stamp type works too
	if err := repo.IncrementTotal(charID, "ex"); err != nil {
		t.Fatalf("IncrementTotal(ex) failed: %v", err)
	}

	if _, err := db.Exec("UPDATE stamps SET ex_total=16, ex_redeemed=0 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	total, redeemed, err := repo.Exchange(charID, "ex")
	if err != nil {
		t.Fatalf("Exchange(ex) failed: %v", err)
	}
	if total != 16 {
		t.Errorf("Expected ex_total=16, got: %d", total)
	}
	if redeemed != 8 {
		t.Errorf("Expected ex_redeemed=8, got: %d", redeemed)
	}
}

func TestRepoStampMonthlyHlClaimed(t *testing.T) {
	repo, _, charID := setupStampRepo(t)
	initStamp(t, repo, charID)

	claimedTime := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)
	if err := repo.SetMonthlyClaimed(charID, "monthly_hl", claimedTime); err != nil {
		t.Fatalf("SetMonthlyClaimed(monthly_hl) failed: %v", err)
	}

	got, err := repo.GetMonthlyClaimed(charID, "monthly_hl")
	if err != nil {
		t.Fatalf("GetMonthlyClaimed(monthly_hl) failed: %v", err)
	}
	if !got.Equal(claimedTime) {
		t.Errorf("Expected %v, got: %v", claimedTime, got)
	}
}
