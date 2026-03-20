package channelserver

import (
	"erupe-ce/common/stringsupport"

	"go.uber.org/zap"
)

// DonateRPResult holds the outcome of a guild tower RP donation.
type DonateRPResult struct {
	ActualDonated uint16
	Advanced      bool
}

// TowerService encapsulates tower business logic, sitting between handlers and repos.
type TowerService struct {
	towerRepo TowerRepo
	logger    *zap.Logger
}

// NewTowerService creates a new TowerService.
func NewTowerService(tr TowerRepo, log *zap.Logger) *TowerService {
	return &TowerService{
		towerRepo: tr,
		logger:    log,
	}
}

// AddGem adds quantity to a specific gem index for a character.
// This is a fetch-transform-save operation that reads the current gems CSV,
// updates the value at the given index, and writes back.
func (svc *TowerService) AddGem(charID uint32, gemIndex int, quantity int) error {
	gems, err := svc.towerRepo.GetGems(charID)
	if err != nil {
		return err
	}
	newGems := stringsupport.CSVSetIndex(gems, gemIndex, stringsupport.CSVGetIndex(gems, gemIndex)+quantity)
	return svc.towerRepo.UpdateGems(charID, newGems)
}

// GetTenrouiraiProgressCapped returns the guild's tenrouirai progress with
// mission scores capped to their respective goals.
func (svc *TowerService) GetTenrouiraiProgressCapped(guildID uint32) (TenrouiraiProgressData, error) {
	progress, err := svc.towerRepo.GetTenrouiraiProgress(guildID)
	if err != nil {
		return progress, err
	}

	if progress.Page < 1 {
		progress.Page = 1
	}

	idx := int(progress.Page*3) - 3
	if idx >= 0 && idx+2 < len(tenrouiraiData) {
		if progress.Mission1 > tenrouiraiData[idx].Goal {
			progress.Mission1 = tenrouiraiData[idx].Goal
		}
		if progress.Mission2 > tenrouiraiData[idx+1].Goal {
			progress.Mission2 = tenrouiraiData[idx+1].Goal
		}
		if progress.Mission3 > tenrouiraiData[idx+2].Goal {
			progress.Mission3 = tenrouiraiData[idx+2].Goal
		}
	}

	return progress, nil
}

// DonateGuildTowerRP processes a tower RP donation, advancing the mission page
// if the cumulative donation meets the requirement. Returns the actual RP consumed
// and whether the page was advanced.
func (svc *TowerService) DonateGuildTowerRP(guildID uint32, donatedRP uint16) (*DonateRPResult, error) {
	page, donated, err := svc.towerRepo.GetGuildTowerPageAndRP(guildID)
	if err != nil {
		return nil, err
	}

	var requirement int
	for i := 0; i < (page*3)+1 && i < len(tenrouiraiData); i++ {
		requirement += int(tenrouiraiData[i].Cost)
	}

	result := &DonateRPResult{
		ActualDonated: donatedRP,
	}

	if donated+int(donatedRP) >= requirement {
		if err := svc.towerRepo.AdvanceTenrouiraiPage(guildID); err != nil {
			svc.logger.Error("Failed to advance tower mission page", zap.Error(err))
			return nil, err
		}
		result.ActualDonated = uint16(requirement - donated)
		result.Advanced = true
	}

	if err := svc.towerRepo.DonateGuildTowerRP(guildID, result.ActualDonated); err != nil {
		svc.logger.Error("Failed to update guild tower RP", zap.Error(err))
		return nil, err
	}

	return result, nil
}
