package channelserver

import (
	"fmt"

	"erupe-ce/common/stringsupport"

	"github.com/jmoiron/sqlx"
)

// TowerRepository centralizes all database access for tower-related tables
// (tower, guilds tower columns, guild_characters tower columns).
type TowerRepository struct {
	db *sqlx.DB
}

// NewTowerRepository creates a new TowerRepository.
func NewTowerRepository(db *sqlx.DB) *TowerRepository {
	return &TowerRepository{db: db}
}

// TowerData holds the core tower stats for a character.
type TowerData struct {
	TR     int32
	TRP    int32
	TSP    int32
	Block1 int32
	Block2 int32
	Skills string
}

// GetTowerData returns tower stats for a character, creating the row if it doesn't exist.
func (r *TowerRepository) GetTowerData(charID uint32) (TowerData, error) {
	var td TowerData
	err := r.db.QueryRow(
		`SELECT COALESCE(tr, 0), COALESCE(trp, 0), COALESCE(tsp, 0), COALESCE(block1, 0), COALESCE(block2, 0), COALESCE(skills, $1) FROM tower WHERE char_id=$2`,
		EmptyTowerCSV(64), charID,
	).Scan(&td.TR, &td.TRP, &td.TSP, &td.Block1, &td.Block2, &td.Skills)
	if err != nil {
		_, err = r.db.Exec(`INSERT INTO tower (char_id) VALUES ($1)`, charID)
		return TowerData{Skills: EmptyTowerCSV(64)}, err
	}
	return td, nil
}

// GetSkills returns the skills CSV string for a character.
func (r *TowerRepository) GetSkills(charID uint32) (string, error) {
	var skills string
	err := r.db.QueryRow(`SELECT COALESCE(skills, $1) FROM tower WHERE char_id=$2`, EmptyTowerCSV(64), charID).Scan(&skills)
	return skills, err
}

// UpdateSkills updates a single skill and deducts TSP cost.
func (r *TowerRepository) UpdateSkills(charID uint32, skills string, cost int32) error {
	_, err := r.db.Exec(`UPDATE tower SET skills=$1, tsp=tsp-$2 WHERE char_id=$3`, skills, cost, charID)
	return err
}

// UpdateProgress updates tower progress (TR, TRP, TSP, block1).
func (r *TowerRepository) UpdateProgress(charID uint32, tr, trp, cost, block1 int32) error {
	_, err := r.db.Exec(
		`UPDATE tower SET tr=$1, trp=COALESCE(trp, 0)+$2, tsp=COALESCE(tsp, 0)+$3, block1=COALESCE(block1, 0)+$4 WHERE char_id=$5`,
		tr, trp, cost, block1, charID,
	)
	return err
}

// GetGems returns the gems CSV string for a character.
func (r *TowerRepository) GetGems(charID uint32) (string, error) {
	var gems string
	err := r.db.QueryRow(`SELECT COALESCE(gems, $1) FROM tower WHERE char_id=$2`, EmptyTowerCSV(30), charID).Scan(&gems)
	return gems, err
}

// UpdateGems saves the gems CSV string for a character.
func (r *TowerRepository) UpdateGems(charID uint32, gems string) error {
	_, err := r.db.Exec(`UPDATE tower SET gems=$1 WHERE char_id=$2`, gems, charID)
	return err
}

// AddGem adds quantity to a specific gem index.
func (r *TowerRepository) AddGem(charID uint32, gemIndex int, quantity int) error {
	gems, err := r.GetGems(charID)
	if err != nil {
		return err
	}
	newGems := stringsupport.CSVSetIndex(gems, gemIndex, stringsupport.CSVGetIndex(gems, gemIndex)+quantity)
	return r.UpdateGems(charID, newGems)
}

// TenrouiraiProgressData holds the guild's tenrouirai (sky corridor) progress.
type TenrouiraiProgressData struct {
	Page     uint8
	Mission1 uint16
	Mission2 uint16
	Mission3 uint16
}

// GetTenrouiraiProgress returns the guild's tower mission page and aggregated mission scores.
func (r *TowerRepository) GetTenrouiraiProgress(guildID uint32) (TenrouiraiProgressData, error) {
	var p TenrouiraiProgressData
	if err := r.db.QueryRow(`SELECT tower_mission_page FROM guilds WHERE id=$1`, guildID).Scan(&p.Page); err != nil {
		return p, err
	}
	_ = r.db.QueryRow(
		`SELECT SUM(tower_mission_1) AS _, SUM(tower_mission_2) AS _, SUM(tower_mission_3) AS _ FROM guild_characters WHERE guild_id=$1`,
		guildID,
	).Scan(&p.Mission1, &p.Mission2, &p.Mission3)
	return p, nil
}

// GetTenrouiraiMissionScores returns per-character scores for a specific mission index (1-3).
func (r *TowerRepository) GetTenrouiraiMissionScores(guildID uint32, missionIndex uint8) ([]TenrouiraiCharScore, error) {
	if missionIndex < 1 || missionIndex > 3 {
		missionIndex = (missionIndex % 3) + 1
	}
	rows, err := r.db.Query(
		fmt.Sprintf(
			`SELECT name, tower_mission_%d FROM guild_characters gc INNER JOIN characters c ON gc.character_id = c.id WHERE guild_id=$1 AND tower_mission_%d IS NOT NULL ORDER BY tower_mission_%d DESC`,
			missionIndex, missionIndex, missionIndex,
		),
		guildID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var scores []TenrouiraiCharScore
	for rows.Next() {
		var cs TenrouiraiCharScore
		if err := rows.Scan(&cs.Name, &cs.Score); err == nil {
			scores = append(scores, cs)
		}
	}
	return scores, nil
}

// GetGuildTowerRP returns the guild's tower RP.
func (r *TowerRepository) GetGuildTowerRP(guildID uint32) (uint32, error) {
	var rp uint32
	err := r.db.QueryRow(`SELECT tower_rp FROM guilds WHERE id=$1`, guildID).Scan(&rp)
	return rp, err
}

// GetGuildTowerPageAndRP returns the guild's tower mission page and donated RP.
func (r *TowerRepository) GetGuildTowerPageAndRP(guildID uint32) (page int, donated int, err error) {
	err = r.db.QueryRow(`SELECT tower_mission_page, tower_rp FROM guilds WHERE id=$1`, guildID).Scan(&page, &donated)
	return
}

// AdvanceTenrouiraiPage increments the guild's tower mission page and resets member mission progress.
func (r *TowerRepository) AdvanceTenrouiraiPage(guildID uint32) error {
	if _, err := r.db.Exec(`UPDATE guilds SET tower_mission_page=tower_mission_page+1 WHERE id=$1`, guildID); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE guild_characters SET tower_mission_1=NULL, tower_mission_2=NULL, tower_mission_3=NULL WHERE guild_id=$1`, guildID)
	return err
}

// DonateGuildTowerRP adds RP to the guild's tower total.
func (r *TowerRepository) DonateGuildTowerRP(guildID uint32, rp uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET tower_rp=tower_rp+$1 WHERE id=$2`, rp, guildID)
	return err
}
