package channelserver

import (
	"context"
	"time"
)

// AddMemberDailyRP adds RP to a member's daily total.
func (r *GuildRepository) AddMemberDailyRP(charID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET rp_today=rp_today+$1 WHERE character_id=$2`, amount, charID)
	return err
}

// ExchangeEventRP subtracts RP from a guild's event pool and returns the new balance.
func (r *GuildRepository) ExchangeEventRP(guildID uint32, amount uint16) (uint32, error) {
	var balance uint32
	err := r.db.QueryRow(`UPDATE guilds SET event_rp=event_rp-$1 WHERE id=$2 RETURNING event_rp`, amount, guildID).Scan(&balance)
	return balance, err
}

// AddRankRP adds RP to a guild's rank total.
func (r *GuildRepository) AddRankRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET rank_rp = rank_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// AddEventRP adds RP to a guild's event total.
func (r *GuildRepository) AddEventRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET event_rp = event_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// GetRoomRP returns the current room RP for a guild.
func (r *GuildRepository) GetRoomRP(guildID uint32) (uint16, error) {
	var rp uint16
	err := r.db.QueryRow(`SELECT room_rp FROM guilds WHERE id = $1`, guildID).Scan(&rp)
	return rp, err
}

// SetRoomRP sets the room RP for a guild.
func (r *GuildRepository) SetRoomRP(guildID uint32, rp uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_rp = $1 WHERE id = $2`, rp, guildID)
	return err
}

// AddRoomRP atomically adds RP to a guild's room total.
func (r *GuildRepository) AddRoomRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_rp = room_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// SetRoomExpiry sets the room expiry time for a guild.
func (r *GuildRepository) SetRoomExpiry(guildID uint32, expiry time.Time) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_expiry = $1 WHERE id = $2`, expiry, guildID)
	return err
}

// RolloverDailyRP moves rp_today into rp_yesterday for all members of a guild,
// then updates the guild's rp_reset_at timestamp.
// Uses SELECT FOR UPDATE to prevent concurrent rollovers from racing.
func (r *GuildRepository) RolloverDailyRP(guildID uint32, noon time.Time) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock the guild row and re-check whether rollover is still needed.
	var rpResetAt time.Time
	if err := tx.QueryRow(
		`SELECT COALESCE(rp_reset_at, '2000-01-01'::timestamptz) FROM guilds WHERE id = $1 FOR UPDATE`,
		guildID,
	).Scan(&rpResetAt); err != nil {
		return err
	}
	if !rpResetAt.Before(noon) {
		// Another goroutine already rolled over; nothing to do.
		return nil
	}
	if _, err := tx.Exec(
		`UPDATE guild_characters SET rp_yesterday = rp_today, rp_today = 0 WHERE guild_id = $1`,
		guildID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE guilds SET rp_reset_at = $1 WHERE id = $2`,
		noon, guildID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// AddWeeklyBonusUsers atomically adds numUsers to the guild's weekly bonus exceptional user count.
func (r *GuildRepository) AddWeeklyBonusUsers(guildID uint32, numUsers uint8) error {
	_, err := r.db.Exec(
		"UPDATE guilds SET weekly_bonus_users = weekly_bonus_users + $1 WHERE id = $2",
		numUsers, guildID,
	)
	return err
}
