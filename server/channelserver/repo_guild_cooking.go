package channelserver

import "time"

// ListMeals returns all meals for a guild.
func (r *GuildRepository) ListMeals(guildID uint32) ([]*GuildMeal, error) {
	rows, err := r.db.Queryx("SELECT id, meal_id, level, created_at FROM guild_meals WHERE guild_id = $1", guildID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var meals []*GuildMeal
	for rows.Next() {
		meal := &GuildMeal{}
		if err := rows.StructScan(meal); err != nil {
			continue
		}
		meals = append(meals, meal)
	}
	return meals, nil
}

// CreateMeal inserts a new guild meal and returns the new ID.
func (r *GuildRepository) CreateMeal(guildID, mealID, level uint32, createdAt time.Time) (uint32, error) {
	var id uint32
	err := r.db.QueryRow(
		"INSERT INTO guild_meals (guild_id, meal_id, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		guildID, mealID, level, createdAt).Scan(&id)
	return id, err
}

// UpdateMeal updates an existing guild meal's fields.
func (r *GuildRepository) UpdateMeal(mealID, newMealID, level uint32, createdAt time.Time) error {
	_, err := r.db.Exec("UPDATE guild_meals SET meal_id = $1, level = $2, created_at = $3 WHERE id = $4",
		newMealID, level, createdAt, mealID)
	return err
}

// ClaimHuntBox updates the box_claimed timestamp for a guild character.
func (r *GuildRepository) ClaimHuntBox(charID uint32, claimedAt time.Time) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET box_claimed=$1 WHERE character_id=$2`, claimedAt, charID)
	return err
}
