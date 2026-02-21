package channelserver

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository interfaces decouple handlers from concrete PostgreSQL implementations,
// enabling mock/stub injection for unit tests and alternative storage backends.

// CharacterRepo defines the contract for character data access.
type CharacterRepo interface {
	LoadColumn(charID uint32, column string) ([]byte, error)
	SaveColumn(charID uint32, column string, data []byte) error
	ReadInt(charID uint32, column string) (int, error)
	AdjustInt(charID uint32, column string, delta int) (int, error)
	GetName(charID uint32) (string, error)
	GetUserID(charID uint32) (uint32, error)
	UpdateLastLogin(charID uint32, timestamp int64) error
	UpdateTimePlayed(charID uint32, timePlayed int) error
	GetCharIDsByUserID(userID uint32) ([]uint32, error)
	ReadTime(charID uint32, column string, defaultVal time.Time) (time.Time, error)
	SaveTime(charID uint32, column string, value time.Time) error
	SaveInt(charID uint32, column string, value int) error
	SaveBool(charID uint32, column string, value bool) error
	SaveString(charID uint32, column string, value string) error
	ReadBool(charID uint32, column string) (bool, error)
	ReadString(charID uint32, column string) (string, error)
	LoadColumnWithDefault(charID uint32, column string, defaultVal []byte) ([]byte, error)
	SetDeleted(charID uint32) error
	UpdateDailyCafe(charID uint32, dailyTime time.Time, bonusQuests, dailyQuests uint32) error
	ResetDailyQuests(charID uint32) error
	ReadEtcPoints(charID uint32) (bonusQuests, dailyQuests, promoPoints uint32, err error)
	ResetCafeTime(charID uint32, cafeReset time.Time) error
	UpdateGuildPostChecked(charID uint32) error
	ReadGuildPostChecked(charID uint32) (time.Time, error)
	SaveMercenary(charID uint32, data []byte, rastaID uint32) error
	UpdateGCPAndPact(charID uint32, gcp uint32, pactID uint32) error
	FindByRastaID(rastaID int) (charID uint32, name string, err error)
	SaveCharacterData(charID uint32, compSave []byte, hr, gr uint16, isFemale bool, weaponType uint8, weaponID uint16) error
	SaveHouseData(charID uint32, houseTier []byte, houseData, bookshelf, gallery, tore, garden []byte) error
	LoadSaveData(charID uint32) (uint32, []byte, bool, string, error)
}

// GuildRepo defines the contract for guild data access.
type GuildRepo interface {
	GetByID(guildID uint32) (*Guild, error)
	GetByCharID(charID uint32) (*Guild, error)
	ListAll() ([]*Guild, error)
	Create(leaderCharID uint32, guildName string) (int32, error)
	Save(guild *Guild) error
	Disband(guildID uint32) error
	RemoveCharacter(charID uint32) error
	AcceptApplication(guildID, charID uint32) error
	CreateApplication(guildID, charID, actorID uint32, appType GuildApplicationType, tx *sql.Tx) error
	CancelInvitation(guildID, charID uint32) error
	RejectApplication(guildID, charID uint32) error
	ArrangeCharacters(charIDs []uint32) error
	GetApplication(guildID, charID uint32, appType GuildApplicationType) (*GuildApplication, error)
	HasApplication(guildID, charID uint32) (bool, error)
	GetItemBox(guildID uint32) ([]byte, error)
	SaveItemBox(guildID uint32, data []byte) error
	GetMembers(guildID uint32, applicants bool) ([]*GuildMember, error)
	GetCharacterMembership(charID uint32) (*GuildMember, error)
	SaveMember(member *GuildMember) error
	SetRecruiting(guildID uint32, recruiting bool) error
	SetPugiOutfits(guildID uint32, outfits uint32) error
	SetRecruiter(charID uint32, allowed bool) error
	AddMemberDailyRP(charID uint32, amount uint16) error
	ExchangeEventRP(guildID uint32, amount uint16) (uint32, error)
	AddRankRP(guildID uint32, amount uint16) error
	AddEventRP(guildID uint32, amount uint16) error
	GetRoomRP(guildID uint32) (uint16, error)
	SetRoomRP(guildID uint32, rp uint16) error
	AddRoomRP(guildID uint32, amount uint16) error
	SetRoomExpiry(guildID uint32, expiry time.Time) error
	ListPosts(guildID uint32, postType int) ([]*MessageBoardPost, error)
	CreatePost(guildID, authorID, stampID uint32, postType int, title, body string, maxPosts int) error
	DeletePost(postID uint32) error
	UpdatePost(postID uint32, title, body string) error
	UpdatePostStamp(postID, stampID uint32) error
	GetPostLikedBy(postID uint32) (string, error)
	SetPostLikedBy(postID uint32, likedBy string) error
	CountNewPosts(guildID uint32, since time.Time) (int, error)
	GetAllianceByID(allianceID uint32) (*GuildAlliance, error)
	ListAlliances() ([]*GuildAlliance, error)
	CreateAlliance(name string, parentGuildID uint32) error
	DeleteAlliance(allianceID uint32) error
	RemoveGuildFromAlliance(allianceID, guildID, subGuild1ID, subGuild2ID uint32) error
	ListAdventures(guildID uint32) ([]*GuildAdventure, error)
	CreateAdventure(guildID, destination uint32, depart, returnTime int64) error
	CreateAdventureWithCharge(guildID, destination, charge uint32, depart, returnTime int64) error
	CollectAdventure(adventureID uint32, charID uint32) error
	ChargeAdventure(adventureID uint32, amount uint32) error
	GetPendingHunt(charID uint32) (*TreasureHunt, error)
	ListGuildHunts(guildID, charID uint32) ([]*TreasureHunt, error)
	CreateHunt(guildID, hostID, destination, level uint32, huntData []byte, catsUsed string) error
	AcquireHunt(huntID uint32) error
	RegisterHuntReport(huntID, charID uint32) error
	CollectHunt(huntID uint32) error
	ClaimHuntReward(huntID, charID uint32) error
	ListMeals(guildID uint32) ([]*GuildMeal, error)
	CreateMeal(guildID, mealID, level uint32, createdAt time.Time) (uint32, error)
	UpdateMeal(mealID, newMealID, level uint32, createdAt time.Time) error
	ClaimHuntBox(charID uint32, claimedAt time.Time) error
	ListGuildKills(guildID, charID uint32) ([]*GuildKill, error)
	CountGuildKills(guildID, charID uint32) (int, error)
	ClearTreasureHunt(charID uint32) error
	InsertKillLog(charID uint32, monster int, quantity uint8, timestamp time.Time) error
	ListInvitedCharacters(guildID uint32) ([]*ScoutedCharacter, error)
	RolloverDailyRP(guildID uint32, noon time.Time) error
	AddWeeklyBonusUsers(guildID uint32, numUsers uint8) error
}

// UserRepo defines the contract for user account data access.
type UserRepo interface {
	GetGachaPoints(userID uint32) (fp, premium, trial uint32, err error)
	GetTrialCoins(userID uint32) (uint16, error)
	DeductTrialCoins(userID uint32, amount uint32) error
	DeductPremiumCoins(userID uint32, amount uint32) error
	AddPremiumCoins(userID uint32, amount uint32) error
	AddTrialCoins(userID uint32, amount uint32) error
	DeductFrontierPoints(userID uint32, amount uint32) error
	AddFrontierPoints(userID uint32, amount uint32) error
	AdjustFrontierPointsDeduct(userID uint32, amount int) (uint32, error)
	AdjustFrontierPointsCredit(userID uint32, amount int) (uint32, error)
	AddFrontierPointsFromGacha(userID uint32, gachaID uint32, entryType uint8) error
	GetRights(userID uint32) (uint32, error)
	SetRights(userID uint32, rights uint32) error
	IsOp(userID uint32) (bool, error)
	SetLastCharacter(userID uint32, charID uint32) error
	GetTimer(userID uint32) (bool, error)
	SetTimer(userID uint32, value bool) error
	CountByPSNID(psnID string) (int, error)
	SetPSNID(userID uint32, psnID string) error
	GetDiscordToken(userID uint32) (string, error)
	SetDiscordToken(userID uint32, token string) error
	GetItemBox(userID uint32) ([]byte, error)
	SetItemBox(userID uint32, data []byte) error
	LinkDiscord(discordID string, token string) (string, error)
	SetPasswordByDiscordID(discordID string, hash []byte) error
	GetByIDAndUsername(charID uint32) (userID uint32, username string, err error)
	BanUser(userID uint32, expires *time.Time) error
}

// GachaRepo defines the contract for gacha system data access.
type GachaRepo interface {
	GetEntryForTransaction(gachaID uint32, rollID uint8) (itemType uint8, itemNumber uint16, rolls int, err error)
	GetRewardPool(gachaID uint32) ([]GachaEntry, error)
	GetItemsForEntry(entryID uint32) ([]GachaItem, error)
	GetGuaranteedItems(rollType uint8, gachaID uint32) ([]GachaItem, error)
	GetStepupStep(gachaID uint32, charID uint32) (uint8, error)
	GetStepupWithTime(gachaID uint32, charID uint32) (uint8, time.Time, error)
	HasEntryType(gachaID uint32, entryType uint8) (bool, error)
	DeleteStepup(gachaID uint32, charID uint32) error
	InsertStepup(gachaID uint32, step uint8, charID uint32) error
	GetBoxEntryIDs(gachaID uint32, charID uint32) ([]uint32, error)
	InsertBoxEntry(gachaID uint32, entryID uint32, charID uint32) error
	DeleteBoxEntries(gachaID uint32, charID uint32) error
	ListShop() ([]Gacha, error)
	GetShopType(shopID uint32) (int, error)
	GetAllEntries(gachaID uint32) ([]GachaEntry, error)
	GetWeightDivisor(gachaID uint32) (float64, error)
}

// HouseRepo defines the contract for house/housing data access.
type HouseRepo interface {
	UpdateInterior(charID uint32, data []byte) error
	GetHouseByCharID(charID uint32) (HouseData, error)
	SearchHousesByName(name string) ([]HouseData, error)
	UpdateHouseState(charID uint32, state uint8, password string) error
	GetHouseAccess(charID uint32) (state uint8, password string, err error)
	GetHouseContents(charID uint32) (houseTier, houseData, houseFurniture, bookshelf, gallery, tore, garden []byte, err error)
	GetMission(charID uint32) ([]byte, error)
	UpdateMission(charID uint32, data []byte) error
	InitializeWarehouse(charID uint32) error
	GetWarehouseNames(charID uint32) (itemNames, equipNames [10]string, err error)
	RenameWarehouseBox(charID uint32, boxType uint8, boxIndex uint8, name string) error
	GetWarehouseItemData(charID uint32, index uint8) ([]byte, error)
	SetWarehouseItemData(charID uint32, index uint8, data []byte) error
	GetWarehouseEquipData(charID uint32, index uint8) ([]byte, error)
	SetWarehouseEquipData(charID uint32, index uint8, data []byte) error
	GetTitles(charID uint32) ([]Title, error)
	AcquireTitle(titleID uint16, charID uint32) error
}

// FestaRepo defines the contract for festa event data access.
type FestaRepo interface {
	CleanupAll() error
	InsertEvent(startTime uint32) error
	GetFestaEvents() ([]FestaEvent, error)
	GetTeamSouls(team string) (uint32, error)
	GetTrialsWithMonopoly() ([]FestaTrial, error)
	GetTopGuildForTrial(trialType uint16) (FestaGuildRanking, error)
	GetTopGuildInWindow(start, end uint32) (FestaGuildRanking, error)
	GetCharSouls(charID uint32) (uint32, error)
	HasClaimedMainPrize(charID uint32) bool
	VoteTrial(charID uint32, trialID uint32) error
	RegisterGuild(guildID uint32, team string) error
	SubmitSouls(charID, guildID uint32, souls []uint16) error
	ClaimPrize(prizeID uint32, charID uint32) error
	ListPrizes(charID uint32, prizeType string) ([]Prize, error)
}

// TowerRepo defines the contract for tower/tenrouirai data access.
type TowerRepo interface {
	GetTowerData(charID uint32) (TowerData, error)
	GetSkills(charID uint32) (string, error)
	UpdateSkills(charID uint32, skills string, cost int32) error
	UpdateProgress(charID uint32, tr, trp, cost, block1 int32) error
	GetGems(charID uint32) (string, error)
	UpdateGems(charID uint32, gems string) error
	AddGem(charID uint32, gemIndex int, quantity int) error
	GetTenrouiraiProgress(guildID uint32) (TenrouiraiProgressData, error)
	GetTenrouiraiMissionScores(guildID uint32, missionIndex uint8) ([]TenrouiraiCharScore, error)
	GetGuildTowerRP(guildID uint32) (uint32, error)
	GetGuildTowerPageAndRP(guildID uint32) (page int, donated int, err error)
	AdvanceTenrouiraiPage(guildID uint32) error
	DonateGuildTowerRP(guildID uint32, rp uint16) error
}

// RengokuRepo defines the contract for rengoku score/ranking data access.
type RengokuRepo interface {
	UpsertScore(charID uint32, maxStagesMp, maxPointsMp, maxStagesSp, maxPointsSp uint32) error
	GetRanking(leaderboard uint32, guildID uint32) (*sqlx.Rows, error)
}

// MailRepo defines the contract for in-game mail data access.
type MailRepo interface {
	SendMail(senderID, recipientID uint32, subject, body string, itemID, itemAmount uint16, isGuildInvite, isSystemMessage bool) error
	SendMailTx(tx *sql.Tx, senderID, recipientID uint32, subject, body string, itemID, itemAmount uint16, isGuildInvite, isSystemMessage bool) error
	GetListForCharacter(charID uint32) ([]Mail, error)
	GetByID(id int) (*Mail, error)
	MarkRead(id int) error
	MarkDeleted(id int) error
	SetLocked(id int, locked bool) error
	MarkItemReceived(id int) error
}

// StampRepo defines the contract for stamp card data access.
type StampRepo interface {
	GetChecked(charID uint32, stampType string) (time.Time, error)
	Init(charID uint32, now time.Time) error
	SetChecked(charID uint32, stampType string, now time.Time) error
	IncrementTotal(charID uint32, stampType string) error
	GetTotals(charID uint32, stampType string) (total, redeemed uint16, err error)
	ExchangeYearly(charID uint32) (total, redeemed uint16, err error)
	Exchange(charID uint32, stampType string) (total, redeemed uint16, err error)
}

// DistributionRepo defines the contract for distribution/event item data access.
type DistributionRepo interface {
	List(charID uint32, distType uint8) ([]Distribution, error)
	GetItems(distributionID uint32) ([]DistributionItem, error)
	RecordAccepted(distributionID, charID uint32) error
	GetDescription(distributionID uint32) (string, error)
}

// SessionRepo defines the contract for session/login token data access.
type SessionRepo interface {
	ValidateLoginToken(token string, sessionID uint32, charID uint32) error
	BindSession(token string, serverID uint16, charID uint32) error
	ClearSession(token string) error
	UpdatePlayerCount(serverID uint16, count int) error
}

// EventRepo defines the contract for event/login boost data access.
type EventRepo interface {
	GetFeatureWeapon(startTime time.Time) (activeFeature, error)
	InsertFeatureWeapon(startTime time.Time, features uint32) error
	GetLoginBoosts(charID uint32) (*sqlx.Rows, error)
	InsertLoginBoost(charID uint32, weekReq uint8, expiration, reset time.Time) error
	UpdateLoginBoost(charID uint32, weekReq uint8, expiration, reset time.Time) error
	GetEventQuests() (*sql.Rows, error)
	UpdateEventQuestStartTime(tx *sql.Tx, id uint32, startTime time.Time) error
	BeginTx() (*sql.Tx, error)
}

// AchievementRepo defines the contract for achievement data access.
type AchievementRepo interface {
	EnsureExists(charID uint32) error
	GetAllScores(charID uint32) ([33]int32, error)
	IncrementScore(charID uint32, achievementID uint8) error
}

// ShopRepo defines the contract for shop data access.
type ShopRepo interface {
	GetShopItems(shopType uint8, shopID uint32, charID uint32) (*sqlx.Rows, error)
	RecordPurchase(charID, shopItemID, quantity uint32) error
	GetFpointItem(tradeID uint32) (quantity, fpoints int, err error)
	GetFpointExchangeList() (*sqlx.Rows, error)
}

// CafeRepo defines the contract for cafe bonus data access.
type CafeRepo interface {
	ResetAccepted(charID uint32) error
	GetBonuses(charID uint32) (*sqlx.Rows, error)
	GetClaimable(charID uint32, elapsedSec int64) (*sqlx.Rows, error)
	GetBonusItem(bonusID uint32) (itemType, quantity uint32, err error)
	AcceptBonus(bonusID, charID uint32) error
}

// GoocooRepo defines the contract for goocoo (pet) data access.
type GoocooRepo interface {
	EnsureExists(charID uint32) error
	GetSlot(charID uint32, slot uint32) ([]byte, error)
	ClearSlot(charID uint32, slot uint32) error
	SaveSlot(charID uint32, slot uint32, data []byte) error
}

// DivaRepo defines the contract for diva event data access.
type DivaRepo interface {
	DeleteEvents() error
	InsertEvent(startEpoch uint32) error
	GetEvents() (*sqlx.Rows, error)
}

// MiscRepo defines the contract for miscellaneous data access.
type MiscRepo interface {
	GetTrendWeapons(weaponType uint8) (*sql.Rows, error)
	UpsertTrendWeapon(weaponID uint16, weaponType uint8) error
}

// ScenarioRepo defines the contract for scenario counter data access.
type ScenarioRepo interface {
	GetCounters() (*sqlx.Rows, error)
}

// MercenaryRepo defines the contract for mercenary/rasta data access.
type MercenaryRepo interface {
	NextRastaID() (uint32, error)
	NextAirouID() (uint32, error)
	GetMercenaryLoans(charID uint32) (*sql.Rows, error)
	GetGuildHuntCatsUsed(charID uint32) (*sql.Rows, error)
	GetGuildAirou(guildID uint32) (*sql.Rows, error)
}
