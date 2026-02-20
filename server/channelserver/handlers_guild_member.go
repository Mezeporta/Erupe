package channelserver

import (
	"time"
)

// GuildMember represents a guild member with role and stats.
type GuildMember struct {
	GuildID         uint32     `db:"guild_id"`
	CharID          uint32     `db:"character_id"`
	JoinedAt        *time.Time `db:"joined_at"`
	Souls           uint32     `db:"souls"`
	RPToday         uint16     `db:"rp_today"`
	RPYesterday     uint16     `db:"rp_yesterday"`
	Name            string     `db:"name"`
	IsApplicant     bool       `db:"is_applicant"`
	OrderIndex      uint16     `db:"order_index"`
	LastLogin       uint32     `db:"last_login"`
	Recruiter       bool       `db:"recruiter"`
	AvoidLeadership bool       `db:"avoid_leadership"`
	IsLeader        bool       `db:"is_leader"`
	HR              uint16     `db:"hr"`
	GR              uint16     `db:"gr"`
	WeaponID        uint16     `db:"weapon_id"`
	WeaponType      uint8      `db:"weapon_type"`
}

func (gm *GuildMember) CanRecruit() bool {
	if gm.Recruiter {
		return true
	}
	if gm.OrderIndex <= 3 {
		return true
	}
	if gm.IsLeader {
		return true
	}
	return false
}

func (gm *GuildMember) IsSubLeader() bool {
	return gm.OrderIndex <= 3
}
