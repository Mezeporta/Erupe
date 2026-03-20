package channelserver

import (
	"database/sql"
	"errors"
	"time"

	"erupe-ce/common/byteframe"

	"go.uber.org/zap"
)

// GachaService encapsulates business logic for the gacha lottery system.
type GachaService struct {
	gachaRepo        GachaRepo
	userRepo         UserRepo
	charRepo         CharacterRepo
	logger           *zap.Logger
	maxNetcafePoints int
}

// NewGachaService creates a new GachaService.
func NewGachaService(gr GachaRepo, ur UserRepo, cr CharacterRepo, log *zap.Logger, maxNP int) *GachaService {
	return &GachaService{
		gachaRepo:        gr,
		userRepo:         ur,
		charRepo:         cr,
		logger:           log,
		maxNetcafePoints: maxNP,
	}
}

// GachaReward represents a single gacha reward item with rarity.
type GachaReward struct {
	ItemType uint8
	ItemID   uint16
	Quantity uint16
	Rarity   uint8
}

// GachaPlayResult holds the outcome of a normal or box gacha play.
type GachaPlayResult struct {
	Rewards []GachaReward
}

// StepupPlayResult holds the outcome of a stepup gacha play.
type StepupPlayResult struct {
	RandomRewards     []GachaReward
	GuaranteedRewards []GachaReward
}

// StepupStatus holds the current stepup state for a character on a gacha.
type StepupStatus struct {
	Step uint8
}

// transact processes the cost for a gacha roll, deducting the appropriate currency.
func (svc *GachaService) transact(userID, charID, gachaID uint32, rollID uint8) (int, error) {
	itemType, itemNumber, rolls, err := svc.gachaRepo.GetEntryForTransaction(gachaID, rollID)
	if err != nil {
		return 0, err
	}
	switch itemType {
	case 17:
		svc.deductNetcafePoints(charID, int(itemNumber))
	case 19, 20:
		svc.spendGachaCoin(userID, itemNumber)
	case 21:
		if err := svc.userRepo.DeductFrontierPoints(userID, uint32(itemNumber)); err != nil {
			svc.logger.Error("Failed to deduct frontier points for gacha", zap.Error(err))
		}
	}
	return rolls, nil
}

// deductNetcafePoints removes netcafe points from a character's save data.
func (svc *GachaService) deductNetcafePoints(charID uint32, amount int) {
	points, err := svc.charRepo.ReadInt(charID, "netcafe_points")
	if err != nil {
		svc.logger.Error("Failed to read netcafe points", zap.Error(err))
		return
	}
	points = min(points-amount, svc.maxNetcafePoints)
	if err := svc.charRepo.SaveInt(charID, "netcafe_points", points); err != nil {
		svc.logger.Error("Failed to update netcafe points", zap.Error(err))
	}
}

// spendGachaCoin deducts gacha coins, preferring trial coins over premium.
func (svc *GachaService) spendGachaCoin(userID uint32, quantity uint16) {
	gt, _ := svc.userRepo.GetTrialCoins(userID)
	if quantity <= gt {
		if err := svc.userRepo.DeductTrialCoins(userID, uint32(quantity)); err != nil {
			svc.logger.Error("Failed to deduct gacha trial coins", zap.Error(err))
		}
	} else {
		if err := svc.userRepo.DeductPremiumCoins(userID, uint32(quantity)); err != nil {
			svc.logger.Error("Failed to deduct gacha premium coins", zap.Error(err))
		}
	}
}

// resolveRewards selects random entries and resolves them into rewards.
func (svc *GachaService) resolveRewards(entries []GachaEntry, rolls int, isBox bool) []GachaReward {
	rewardEntries, err := getRandomEntries(entries, rolls, isBox)
	if err != nil {
		svc.logger.Warn("Failed to select gacha entries", zap.Error(err))
		return nil
	}
	var rewards []GachaReward
	for i := range rewardEntries {
		entryItems, err := svc.gachaRepo.GetItemsForEntry(rewardEntries[i].ID)
		if err != nil {
			svc.logger.Warn("Gacha entry has no items",
				zap.Uint32("entryID", rewardEntries[i].ID), zap.Error(err))
			continue
		}
		for _, item := range entryItems {
			rewards = append(rewards, GachaReward{
				ItemType: item.ItemType,
				ItemID:   item.ItemID,
				Quantity: item.Quantity,
				Rarity:   rewardEntries[i].Rarity,
			})
		}
	}
	return rewards
}

// saveGachaItems appends reward items to the character's gacha item storage.
func (svc *GachaService) saveGachaItems(charID uint32, items []GachaItem) {
	data, _ := svc.charRepo.LoadColumn(charID, "gacha_items")
	if len(data) > 0 {
		numItems := int(data[0])
		data = data[1:]
		oldItem := byteframe.NewByteFrameFromBytes(data)
		for i := 0; i < numItems; i++ {
			items = append(items, GachaItem{
				ItemType: oldItem.ReadUint8(),
				ItemID:   oldItem.ReadUint16(),
				Quantity: oldItem.ReadUint16(),
			})
		}
	}
	newItem := byteframe.NewByteFrame()
	newItem.WriteUint8(uint8(len(items)))
	for i := range items {
		newItem.WriteUint8(items[i].ItemType)
		newItem.WriteUint16(items[i].ItemID)
		newItem.WriteUint16(items[i].Quantity)
	}
	if err := svc.charRepo.SaveColumn(charID, "gacha_items", newItem.Data()); err != nil {
		svc.logger.Error("Failed to update gacha items", zap.Error(err))
	}
}

// rewardsToItems converts GachaReward slices to GachaItem slices for storage.
func rewardsToItems(rewards []GachaReward) []GachaItem {
	items := make([]GachaItem, len(rewards))
	for i, r := range rewards {
		items[i] = GachaItem{ItemType: r.ItemType, ItemID: r.ItemID, Quantity: r.Quantity}
	}
	return items
}

// PlayNormalGacha processes a normal gacha roll: validates the reward pool,
// deducts cost, selects random rewards, saves items, and returns the result.
func (svc *GachaService) PlayNormalGacha(userID, charID, gachaID uint32, rollType uint8) (*GachaPlayResult, error) {
	entries, err := svc.gachaRepo.GetRewardPool(gachaID)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("gacha has no valid reward entries")
	}
	rolls, err := svc.transact(userID, charID, gachaID, rollType)
	if err != nil {
		return nil, err
	}
	rewards := svc.resolveRewards(entries, rolls, false)
	svc.saveGachaItems(charID, rewardsToItems(rewards))
	return &GachaPlayResult{Rewards: rewards}, nil
}

// PlayStepupGacha processes a stepup gacha roll: validates the reward pool,
// deducts cost, advances step, awards frontier points, selects random +
// guaranteed rewards, and saves items.
func (svc *GachaService) PlayStepupGacha(userID, charID, gachaID uint32, rollType uint8) (*StepupPlayResult, error) {
	entries, err := svc.gachaRepo.GetRewardPool(gachaID)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("gacha has no valid reward entries")
	}
	rolls, err := svc.transact(userID, charID, gachaID, rollType)
	if err != nil {
		return nil, err
	}
	if err := svc.userRepo.AddFrontierPointsFromGacha(userID, gachaID, rollType); err != nil {
		svc.logger.Error("Failed to award stepup gacha frontier points", zap.Error(err))
	}
	if err := svc.gachaRepo.DeleteStepup(gachaID, charID); err != nil {
		svc.logger.Error("Failed to delete gacha stepup state", zap.Error(err))
	}
	if err := svc.gachaRepo.InsertStepup(gachaID, rollType+1, charID); err != nil {
		svc.logger.Error("Failed to insert gacha stepup state", zap.Error(err))
	}

	guaranteedItems, _ := svc.gachaRepo.GetGuaranteedItems(rollType, gachaID)
	randomRewards := svc.resolveRewards(entries, rolls, false)

	var guaranteedRewards []GachaReward
	for _, item := range guaranteedItems {
		guaranteedRewards = append(guaranteedRewards, GachaReward{
			ItemType: item.ItemType,
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
			Rarity:   0,
		})
	}

	svc.saveGachaItems(charID, rewardsToItems(randomRewards))
	svc.saveGachaItems(charID, rewardsToItems(guaranteedRewards))
	return &StepupPlayResult{
		RandomRewards:     randomRewards,
		GuaranteedRewards: guaranteedRewards,
	}, nil
}

// PlayBoxGacha processes a box gacha roll: validates the reward pool, deducts
// cost, selects random entries without replacement, records drawn entries,
// saves items, and returns the result.
func (svc *GachaService) PlayBoxGacha(userID, charID, gachaID uint32, rollType uint8) (*GachaPlayResult, error) {
	entries, err := svc.gachaRepo.GetRewardPool(gachaID)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("gacha has no valid reward entries")
	}
	rolls, err := svc.transact(userID, charID, gachaID, rollType)
	if err != nil {
		return nil, err
	}
	rewardEntries, err := getRandomEntries(entries, rolls, true)
	if err != nil {
		svc.logger.Warn("Failed to select box gacha entries", zap.Error(err))
		return &GachaPlayResult{}, nil
	}
	var rewards []GachaReward
	for i := range rewardEntries {
		entryItems, err := svc.gachaRepo.GetItemsForEntry(rewardEntries[i].ID)
		if err != nil {
			svc.logger.Warn("Box gacha entry has no items",
				zap.Uint32("entryID", rewardEntries[i].ID), zap.Error(err))
			continue
		}
		if err := svc.gachaRepo.InsertBoxEntry(gachaID, rewardEntries[i].ID, charID); err != nil {
			svc.logger.Error("Failed to insert gacha box entry", zap.Error(err))
		}
		for _, item := range entryItems {
			rewards = append(rewards, GachaReward{
				ItemType: item.ItemType,
				ItemID:   item.ItemID,
				Quantity: item.Quantity,
				Rarity:   0,
			})
		}
	}
	svc.saveGachaItems(charID, rewardsToItems(rewards))
	return &GachaPlayResult{Rewards: rewards}, nil
}

// GetStepupStatus returns the current stepup step for a character, resetting
// stale progress based on the noon boundary. The now parameter enables
// deterministic testing.
func (svc *GachaService) GetStepupStatus(gachaID, charID uint32, now time.Time) (*StepupStatus, error) {
	// Compute the most recent noon boundary
	y, m, d := now.Date()
	midday := time.Date(y, m, d, 12, 0, 0, 0, now.Location())
	if now.Before(midday) {
		midday = midday.Add(-24 * time.Hour)
	}

	step, createdAt, err := svc.gachaRepo.GetStepupWithTime(gachaID, charID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		svc.logger.Error("Failed to get gacha stepup state", zap.Error(err))
	}

	if err == nil && createdAt.Before(midday) {
		if err := svc.gachaRepo.DeleteStepup(gachaID, charID); err != nil {
			svc.logger.Error("Failed to reset stale gacha stepup", zap.Error(err))
		}
		step = 0
	} else if err == nil {
		hasEntry, _ := svc.gachaRepo.HasEntryType(gachaID, step)
		if !hasEntry {
			if err := svc.gachaRepo.DeleteStepup(gachaID, charID); err != nil {
				svc.logger.Error("Failed to reset gacha stepup state", zap.Error(err))
			}
			step = 0
		}
	}

	return &StepupStatus{Step: step}, nil
}

// GetBoxInfo returns the entry IDs already drawn for a box gacha.
func (svc *GachaService) GetBoxInfo(gachaID, charID uint32) ([]uint32, error) {
	return svc.gachaRepo.GetBoxEntryIDs(gachaID, charID)
}

// ResetBox clears all drawn entries for a box gacha.
func (svc *GachaService) ResetBox(gachaID, charID uint32) error {
	return svc.gachaRepo.DeleteBoxEntries(gachaID, charID)
}

