package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// ShopRepository centralizes all database access for shop-related tables.
type ShopRepository struct {
	db *sqlx.DB
}

// NewShopRepository creates a new ShopRepository.
func NewShopRepository(db *sqlx.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

// GetShopItems returns shop items with per-character purchase counts.
func (r *ShopRepository) GetShopItems(shopType uint8, shopID uint32, charID uint32) (*sqlx.Rows, error) {
	return r.db.Queryx(`SELECT id, item_id, cost, quantity, min_hr, min_sr, min_gr, store_level, max_quantity,
       		COALESCE((SELECT bought FROM shop_items_bought WHERE shop_item_id=si.id AND character_id=$3), 0) as used_quantity,
       		road_floors, road_fatalis FROM shop_items si WHERE shop_type=$1 AND shop_id=$2
       		`, shopType, shopID, charID)
}

// RecordPurchase upserts a purchase record, adding to the bought count.
func (r *ShopRepository) RecordPurchase(charID, shopItemID, quantity uint32) error {
	_, err := r.db.Exec(`INSERT INTO shop_items_bought (character_id, shop_item_id, bought)
		VALUES ($1,$2,$3) ON CONFLICT (character_id, shop_item_id)
		DO UPDATE SET bought = bought + $3
		WHERE EXCLUDED.character_id=$1 AND EXCLUDED.shop_item_id=$2
	`, charID, shopItemID, quantity)
	return err
}

// GetFpointItem returns the quantity and fpoints cost for a frontier point item.
func (r *ShopRepository) GetFpointItem(tradeID uint32) (quantity, fpoints int, err error) {
	err = r.db.QueryRow("SELECT quantity, fpoints FROM fpoint_items WHERE id=$1", tradeID).Scan(&quantity, &fpoints)
	return
}

// GetFpointExchangeList returns all frontier point exchange items ordered by buyable status.
func (r *ShopRepository) GetFpointExchangeList() (*sqlx.Rows, error) {
	return r.db.Queryx(`SELECT id, item_type, item_id, quantity, fpoints, buyable FROM fpoint_items ORDER BY buyable DESC`)
}
