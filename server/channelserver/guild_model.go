package channelserver

import (
	"database/sql/driver"
	"encoding/json"
	_config "erupe-ce/config"
	"time"
)

// FestivalColor is a festival color identifier string.
type FestivalColor string

const (
	FestivalColorNone FestivalColor = "none"
	FestivalColorBlue FestivalColor = "blue"
	FestivalColorRed  FestivalColor = "red"
)

// FestivalColorCodes maps festival colors to their numeric codes.
var FestivalColorCodes = map[FestivalColor]int16{
	FestivalColorNone: -1,
	FestivalColorBlue: 0,
	FestivalColorRed:  1,
}

// GuildApplicationType is the type of a guild application (applied or invited).
type GuildApplicationType string

const (
	GuildApplicationTypeApplied GuildApplicationType = "applied"
	GuildApplicationTypeInvited GuildApplicationType = "invited"
)

// Guild represents a guild with all its metadata.
type Guild struct {
	ID            uint32        `db:"id"`
	Name          string        `db:"name"`
	MainMotto     uint8         `db:"main_motto"`
	SubMotto      uint8         `db:"sub_motto"`
	CreatedAt     time.Time     `db:"created_at"`
	MemberCount   uint16        `db:"member_count"`
	RankRP        uint32        `db:"rank_rp"`
	EventRP       uint32        `db:"event_rp"`
	RoomRP        uint16        `db:"room_rp"`
	RoomExpiry    time.Time     `db:"room_expiry"`
	Comment       string        `db:"comment"`
	PugiName1     string        `db:"pugi_name_1"`
	PugiName2     string        `db:"pugi_name_2"`
	PugiName3     string        `db:"pugi_name_3"`
	PugiOutfit1   uint8         `db:"pugi_outfit_1"`
	PugiOutfit2   uint8         `db:"pugi_outfit_2"`
	PugiOutfit3   uint8         `db:"pugi_outfit_3"`
	PugiOutfits   uint32        `db:"pugi_outfits"`
	Recruiting    bool          `db:"recruiting"`
	FestivalColor FestivalColor `db:"festival_color"`
	Souls         uint32        `db:"souls"`
	AllianceID    uint32        `db:"alliance_id"`
	Icon          *GuildIcon    `db:"icon"`
	RPResetAt     time.Time     `db:"rp_reset_at"`

	GuildLeader
}

// GuildLeader holds the character ID and name of a guild's leader.
type GuildLeader struct {
	LeaderCharID uint32 `db:"leader_id"`
	LeaderName   string `db:"leader_name"`
}

// GuildIconPart represents one graphical part of a guild icon.
type GuildIconPart struct {
	Index    uint16
	ID       uint16
	Page     uint8
	Size     uint8
	Rotation uint8
	Red      uint8
	Green    uint8
	Blue     uint8
	PosX     uint16
	PosY     uint16
}

// GuildApplication represents a pending guild application or invitation.
type GuildApplication struct {
	ID              int                  `db:"id"`
	GuildID         uint32               `db:"guild_id"`
	CharID          uint32               `db:"character_id"`
	ActorID         uint32               `db:"actor_id"`
	ApplicationType GuildApplicationType `db:"application_type"`
	CreatedAt       time.Time            `db:"created_at"`
}

// GuildIcon is a composite guild icon made up of multiple parts.
type GuildIcon struct {
	Parts []GuildIconPart
}

func (gi *GuildIcon) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &gi)
	case string:
		err = json.Unmarshal([]byte(v), &gi)
	}

	return
}

func (gi *GuildIcon) Value() (valuer driver.Value, err error) {
	return json.Marshal(gi)
}

func (g *Guild) Rank(mode _config.Mode) uint16 {
	rpMap := []uint32{
		24, 48, 96, 144, 192, 240, 288, 360, 432,
		504, 600, 696, 792, 888, 984, 1080, 1200,
	}
	if mode <= _config.Z2 {
		rpMap = []uint32{
			3500, 6000, 8500, 11000, 13500, 16000, 20000, 24000, 28000,
			33000, 38000, 43000, 48000, 55000, 70000, 90000, 120000,
		}
	}
	for i, u := range rpMap {
		if g.RankRP < u {
			if mode <= _config.S6 && i >= 12 {
				return 12
			} else if mode <= _config.F5 && i >= 13 {
				return 13
			} else if mode <= _config.G32 && i >= 14 {
				return 14
			}
			return uint16(i)
		}
	}
	if mode <= _config.S6 {
		return 12
	} else if mode <= _config.F5 {
		return 13
	} else if mode <= _config.G32 {
		return 14
	}
	return 17
}
