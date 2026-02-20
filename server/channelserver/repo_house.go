package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// HouseRepository centralizes all database access for house-related tables
// (user_binary house columns, warehouse, titles).
type HouseRepository struct {
	db *sqlx.DB
}

// NewHouseRepository creates a new HouseRepository.
func NewHouseRepository(db *sqlx.DB) *HouseRepository {
	return &HouseRepository{db: db}
}

// user_binary house columns

// UpdateInterior saves the house furniture layout.
func (r *HouseRepository) UpdateInterior(charID uint32, data []byte) error {
	_, err := r.db.Exec(`UPDATE user_binary SET house_furniture=$1 WHERE id=$2`, data, charID)
	return err
}

const houseQuery = `SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
	FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE c.id=$1`

// GetHouseByCharID returns house data for a single character.
func (r *HouseRepository) GetHouseByCharID(charID uint32) (HouseData, error) {
	var house HouseData
	err := r.db.QueryRowx(houseQuery, charID).StructScan(&house)
	return house, err
}

// SearchHousesByName returns houses matching a name pattern (case-insensitive).
func (r *HouseRepository) SearchHousesByName(name string) ([]HouseData, error) {
	var houses []HouseData
	rows, err := r.db.Queryx(
		`SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
		FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE name ILIKE $1`,
		fmt.Sprintf(`%%%s%%`, name),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var house HouseData
		if err := rows.StructScan(&house); err == nil {
			houses = append(houses, house)
		}
	}
	return houses, nil
}

// UpdateHouseState sets the house visibility state and password.
func (r *HouseRepository) UpdateHouseState(charID uint32, state uint8, password string) error {
	_, err := r.db.Exec(`UPDATE user_binary SET house_state=$1, house_password=$2 WHERE id=$3`, state, password, charID)
	return err
}

// GetHouseAccess returns the house state and password for access control checks.
func (r *HouseRepository) GetHouseAccess(charID uint32) (state uint8, password string, err error) {
	state = 2 // default to password-protected
	err = r.db.QueryRow(
		`SELECT COALESCE(house_state, 2) as house_state, COALESCE(house_password, '') as house_password FROM user_binary WHERE id=$1`,
		charID,
	).Scan(&state, &password)
	return
}

// GetHouseContents returns all house content columns for rendering a house visit.
func (r *HouseRepository) GetHouseContents(charID uint32) (houseTier, houseData, houseFurniture, bookshelf, gallery, tore, garden []byte, err error) {
	err = r.db.QueryRow(
		`SELECT house_tier, house_data, house_furniture, bookshelf, gallery, tore, garden FROM user_binary WHERE id=$1`,
		charID,
	).Scan(&houseTier, &houseData, &houseFurniture, &bookshelf, &gallery, &tore, &garden)
	return
}

// GetMission returns the myhouse mission data.
func (r *HouseRepository) GetMission(charID uint32) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(`SELECT mission FROM user_binary WHERE id=$1`, charID).Scan(&data)
	return data, err
}

// UpdateMission saves the myhouse mission data.
func (r *HouseRepository) UpdateMission(charID uint32, data []byte) error {
	_, err := r.db.Exec(`UPDATE user_binary SET mission=$1 WHERE id=$2`, data, charID)
	return err
}

// Warehouse methods

// InitializeWarehouse ensures a warehouse row exists for the character.
func (r *HouseRepository) InitializeWarehouse(charID uint32) {
	var t int
	err := r.db.QueryRow(`SELECT character_id FROM warehouse WHERE character_id=$1`, charID).Scan(&t)
	if err != nil {
		_, _ = r.db.Exec(`INSERT INTO warehouse (character_id) VALUES ($1)`, charID)
	}
}

const warehouseNamesSQL = `
SELECT
COALESCE(item0name, ''),
COALESCE(item1name, ''),
COALESCE(item2name, ''),
COALESCE(item3name, ''),
COALESCE(item4name, ''),
COALESCE(item5name, ''),
COALESCE(item6name, ''),
COALESCE(item7name, ''),
COALESCE(item8name, ''),
COALESCE(item9name, ''),
COALESCE(equip0name, ''),
COALESCE(equip1name, ''),
COALESCE(equip2name, ''),
COALESCE(equip3name, ''),
COALESCE(equip4name, ''),
COALESCE(equip5name, ''),
COALESCE(equip6name, ''),
COALESCE(equip7name, ''),
COALESCE(equip8name, ''),
COALESCE(equip9name, '')
FROM warehouse WHERE character_id=$1`

// GetWarehouseNames returns item and equipment box names.
func (r *HouseRepository) GetWarehouseNames(charID uint32) (itemNames, equipNames [10]string, err error) {
	err = r.db.QueryRow(warehouseNamesSQL, charID).Scan(
		&itemNames[0], &itemNames[1], &itemNames[2], &itemNames[3], &itemNames[4],
		&itemNames[5], &itemNames[6], &itemNames[7], &itemNames[8], &itemNames[9],
		&equipNames[0], &equipNames[1], &equipNames[2], &equipNames[3], &equipNames[4],
		&equipNames[5], &equipNames[6], &equipNames[7], &equipNames[8], &equipNames[9],
	)
	return
}

// RenameWarehouseBox renames an item or equipment warehouse box.
// boxType 0 = items, 1 = equipment. boxIndex must be 0-9.
func (r *HouseRepository) RenameWarehouseBox(charID uint32, boxType uint8, boxIndex uint8, name string) error {
	var col string
	switch boxType {
	case 0:
		col = fmt.Sprintf("item%dname", boxIndex)
	case 1:
		col = fmt.Sprintf("equip%dname", boxIndex)
	default:
		return fmt.Errorf("invalid box type: %d", boxType)
	}
	_, err := r.db.Exec(fmt.Sprintf("UPDATE warehouse SET %s=$1 WHERE character_id=$2", col), name, charID)
	return err
}

// GetWarehouseItemData returns raw serialized item data for a warehouse box.
// index 0-10 (10 = gift box).
func (r *HouseRepository) GetWarehouseItemData(charID uint32, index uint8) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(fmt.Sprintf(`SELECT item%d FROM warehouse WHERE character_id=$1`, index), charID).Scan(&data)
	return data, err
}

// SetWarehouseItemData saves raw serialized item data for a warehouse box.
func (r *HouseRepository) SetWarehouseItemData(charID uint32, index uint8, data []byte) error {
	_, err := r.db.Exec(fmt.Sprintf(`UPDATE warehouse SET item%d=$1 WHERE character_id=$2`, index), data, charID)
	return err
}

// GetWarehouseEquipData returns raw serialized equipment data for a warehouse box.
func (r *HouseRepository) GetWarehouseEquipData(charID uint32, index uint8) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(fmt.Sprintf(`SELECT equip%d FROM warehouse WHERE character_id=$1`, index), charID).Scan(&data)
	return data, err
}

// SetWarehouseEquipData saves raw serialized equipment data for a warehouse box.
func (r *HouseRepository) SetWarehouseEquipData(charID uint32, index uint8, data []byte) error {
	_, err := r.db.Exec(fmt.Sprintf(`UPDATE warehouse SET equip%d=$1 WHERE character_id=$2`, index), data, charID)
	return err
}

// Title methods

// GetTitles returns all titles for a character.
func (r *HouseRepository) GetTitles(charID uint32) ([]Title, error) {
	var titles []Title
	rows, err := r.db.Queryx(`SELECT id, unlocked_at, updated_at FROM titles WHERE char_id=$1`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var title Title
		if err := rows.StructScan(&title); err == nil {
			titles = append(titles, title)
		}
	}
	return titles, nil
}

// AcquireTitle inserts a new title or updates its timestamp if it already exists.
func (r *HouseRepository) AcquireTitle(titleID uint16, charID uint32) error {
	var exists int
	err := r.db.QueryRow(`SELECT count(*) FROM titles WHERE id=$1 AND char_id=$2`, titleID, charID).Scan(&exists)
	if err != nil || exists == 0 {
		_, err = r.db.Exec(`INSERT INTO titles VALUES ($1, $2, now(), now())`, titleID, charID)
	} else {
		_, err = r.db.Exec(`UPDATE titles SET updated_at=now() WHERE id=$1 AND char_id=$2`, titleID, charID)
	}
	return err
}
