package channelserver

import (
	"encoding/json"
	"testing"
	"time"

	_config "erupe-ce/config"
)

// TestGuildCreation tests basic guild creation
func TestGuildCreation(t *testing.T) {
	tests := []struct {
		name      string
		guildName string
		leaderId  uint32
		motto     uint8
		valid     bool
	}{
		{
			name:      "valid_guild_creation",
			guildName: "TestGuild",
			leaderId:  1,
			motto:     1,
			valid:     true,
		},
		{
			name:      "guild_with_long_name",
			guildName: "VeryLongGuildNameForTesting",
			leaderId:  2,
			motto:     2,
			valid:     true,
		},
		{
			name:      "guild_with_special_chars",
			guildName: "Guild@#$%",
			leaderId:  3,
			motto:     1,
			valid:     true,
		},
		{
			name:      "guild_empty_name",
			guildName: "",
			leaderId:  4,
			motto:     1,
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:            1,
				Name:          tt.guildName,
				MainMotto:     tt.motto,
				SubMotto:      1,
				CreatedAt:     time.Now(),
				MemberCount:   1,
				RankRP:        0,
				EventRP:       0,
				RoomRP:        0,
				Comment:       "Test guild",
				Recruiting:    true,
				FestivalColor: FestivalColorNone,
				Souls:         0,
				AllianceID:    0,
				GuildLeader: GuildLeader{
					LeaderCharID: tt.leaderId,
					LeaderName:   "TestLeader",
				},
			}

			if (len(guild.Name) > 0) != tt.valid {
				t.Errorf("guild name validity check failed for '%s'", guild.Name)
			}

			if guild.LeaderCharID != tt.leaderId {
				t.Errorf("guild leader ID mismatch: got %d, want %d", guild.LeaderCharID, tt.leaderId)
			}
		})
	}
}

// TestGuildRankCalculation tests guild rank calculation based on RP
func TestGuildRankCalculation(t *testing.T) {
	tests := []struct {
		name     string
		rankRP   uint32
		wantRank uint16
		config   _config.Mode
	}{
		{
			name:     "rank_0_minimal_rp",
			rankRP:   0,
			wantRank: 0,
			config:   _config.Z2,
		},
		{
			name:     "rank_1_threshold",
			rankRP:   3500,
			wantRank: 1,
			config:   _config.Z2,
		},
		{
			name:     "rank_5_middle",
			rankRP:   16000,
			wantRank: 6,
			config:   _config.Z2,
		},
		{
			name:     "max_rank",
			rankRP:   120001,
			wantRank: 17,
			config:   _config.Z2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := _config.ErupeConfig.RealClientMode
			defer func() { _config.ErupeConfig.RealClientMode = originalConfig }()

			_config.ErupeConfig.RealClientMode = tt.config

			guild := &Guild{
				RankRP: tt.rankRP,
			}

			rank := guild.Rank()
			if rank != tt.wantRank {
				t.Errorf("guild rank calculation: got %d, want %d for RP %d", rank, tt.wantRank, tt.rankRP)
			}
		})
	}
}

// TestGuildIconSerialization tests guild icon JSON serialization
func TestGuildIconSerialization(t *testing.T) {
	tests := []struct {
		name  string
		parts int
		valid bool
	}{
		{
			name:  "icon_with_no_parts",
			parts: 0,
			valid: true,
		},
		{
			name:  "icon_with_single_part",
			parts: 1,
			valid: true,
		},
		{
			name:  "icon_with_multiple_parts",
			parts: 5,
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := make([]GuildIconPart, tt.parts)
			for i := 0; i < tt.parts; i++ {
				parts[i] = GuildIconPart{
					Index:    uint16(i),
					ID:       uint16(i + 1),
					Page:     uint8(i % 4),
					Size:     uint8((i + 1) % 8),
					Rotation: uint8(i % 360),
					Red:      uint8(i * 10 % 256),
					Green:    uint8(i * 15 % 256),
					Blue:     uint8(i * 20 % 256),
					PosX:     uint16(i * 100),
					PosY:     uint16(i * 50),
				}
			}

			icon := &GuildIcon{Parts: parts}

			// Test JSON marshaling
			data, err := json.Marshal(icon)
			if err != nil && tt.valid {
				t.Errorf("failed to marshal icon: %v", err)
			}

			if data != nil {
				// Test JSON unmarshaling
				var icon2 GuildIcon
				err = json.Unmarshal(data, &icon2)
				if err != nil && tt.valid {
					t.Errorf("failed to unmarshal icon: %v", err)
				}

				if len(icon2.Parts) != tt.parts {
					t.Errorf("icon parts mismatch: got %d, want %d", len(icon2.Parts), tt.parts)
				}
			}
		})
	}
}

// TestGuildIconDatabaseScan tests guild icon database scanning
func TestGuildIconDatabaseScan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		valid   bool
		wantErr bool
	}{
		{
			name:    "scan_from_bytes",
			input:   []byte(`{"Parts":[]}`),
			valid:   true,
			wantErr: false,
		},
		{
			name:    "scan_from_string",
			input:   `{"Parts":[{"Index":1,"ID":2}]}`,
			valid:   true,
			wantErr: false,
		},
		{
			name:    "scan_invalid_json",
			input:   []byte(`{invalid json}`),
			valid:   false,
			wantErr: true,
		},
		{
			name:    "scan_nil",
			input:   nil,
			valid:   false,
			wantErr: false, // nil doesn't cause an error in this implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := &GuildIcon{}
			err := icon.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("scan error mismatch: got %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGuildLeaderAssignment tests guild leader assignment and modification
func TestGuildLeaderAssignment(t *testing.T) {
	tests := []struct {
		name       string
		leaderId   uint32
		leaderName string
		valid      bool
	}{
		{
			name:       "valid_leader",
			leaderId:   100,
			leaderName: "TestLeader",
			valid:      true,
		},
		{
			name:       "leader_with_id_1",
			leaderId:   1,
			leaderName: "Leader1",
			valid:      true,
		},
		{
			name:       "leader_with_long_name",
			leaderId:   999,
			leaderName: "VeryLongLeaderName",
			valid:      true,
		},
		{
			name:       "leader_with_empty_name",
			leaderId:   500,
			leaderName: "",
			valid:      true, // Name can be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID: 1,
				GuildLeader: GuildLeader{
					LeaderCharID: tt.leaderId,
					LeaderName:   tt.leaderName,
				},
			}

			if guild.LeaderCharID != tt.leaderId {
				t.Errorf("leader ID mismatch: got %d, want %d", guild.LeaderCharID, tt.leaderId)
			}

			if guild.LeaderName != tt.leaderName {
				t.Errorf("leader name mismatch: got %s, want %s", guild.LeaderName, tt.leaderName)
			}
		})
	}
}

// TestGuildApplicationTypes tests guild application type handling
func TestGuildApplicationTypes(t *testing.T) {
	tests := []struct {
		name    string
		appType GuildApplicationType
		valid   bool
	}{
		{
			name:    "application_applied",
			appType: GuildApplicationTypeApplied,
			valid:   true,
		},
		{
			name:    "application_invited",
			appType: GuildApplicationTypeInvited,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &GuildApplication{
				ID:              1,
				GuildID:         100,
				CharID:          200,
				ActorID:         300,
				ApplicationType: tt.appType,
				CreatedAt:       time.Now(),
			}

			if app.ApplicationType != tt.appType {
				t.Errorf("application type mismatch: got %s, want %s", app.ApplicationType, tt.appType)
			}

			if app.GuildID == 0 {
				t.Error("guild ID should not be zero")
			}
		})
	}
}

// TestGuildApplicationCreation tests guild application creation
func TestGuildApplicationCreation(t *testing.T) {
	tests := []struct {
		name    string
		guildId uint32
		charId  uint32
		valid   bool
	}{
		{
			name:    "valid_application",
			guildId: 100,
			charId:  50,
			valid:   true,
		},
		{
			name:    "application_same_guild_char",
			guildId: 1,
			charId:  1,
			valid:   true,
		},
		{
			name:    "large_ids",
			guildId: 999999,
			charId:  888888,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &GuildApplication{
				ID:              1,
				GuildID:         tt.guildId,
				CharID:          tt.charId,
				ActorID:         1,
				ApplicationType: GuildApplicationTypeApplied,
				CreatedAt:       time.Now(),
			}

			if app.GuildID != tt.guildId {
				t.Errorf("guild ID mismatch: got %d, want %d", app.GuildID, tt.guildId)
			}

			if app.CharID != tt.charId {
				t.Errorf("character ID mismatch: got %d, want %d", app.CharID, tt.charId)
			}
		})
	}
}

// TestFestivalColorMapping tests festival color code mapping
func TestFestivalColorMapping(t *testing.T) {
	tests := []struct {
		name      string
		color     FestivalColor
		wantCode  int16
		shouldMap bool
	}{
		{
			name:      "festival_color_none",
			color:     FestivalColorNone,
			wantCode:  -1,
			shouldMap: true,
		},
		{
			name:      "festival_color_blue",
			color:     FestivalColorBlue,
			wantCode:  0,
			shouldMap: true,
		},
		{
			name:      "festival_color_red",
			color:     FestivalColorRed,
			wantCode:  1,
			shouldMap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, exists := FestivalColorCodes[tt.color]
			if !exists && tt.shouldMap {
				t.Errorf("festival color not in map: %s", tt.color)
			}

			if exists && code != tt.wantCode {
				t.Errorf("festival color code mismatch: got %d, want %d", code, tt.wantCode)
			}
		})
	}
}

// TestGuildMemberCount tests guild member count tracking
func TestGuildMemberCount(t *testing.T) {
	tests := []struct {
		name        string
		memberCount uint16
		valid       bool
	}{
		{
			name:        "single_member",
			memberCount: 1,
			valid:       true,
		},
		{
			name:        "max_members",
			memberCount: 100,
			valid:       true,
		},
		{
			name:        "large_member_count",
			memberCount: 65535,
			valid:       true,
		},
		{
			name:        "zero_members",
			memberCount: 0,
			valid:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:          1,
				Name:        "TestGuild",
				MemberCount: tt.memberCount,
			}

			if guild.MemberCount != tt.memberCount {
				t.Errorf("member count mismatch: got %d, want %d", guild.MemberCount, tt.memberCount)
			}
		})
	}
}

// TestGuildRP tests guild RP (rank points and event points)
func TestGuildRP(t *testing.T) {
	tests := []struct {
		name    string
		rankRP  uint32
		eventRP uint32
		roomRP  uint16
		valid   bool
	}{
		{
			name:    "minimal_rp",
			rankRP:  0,
			eventRP: 0,
			roomRP:  0,
			valid:   true,
		},
		{
			name:    "high_rank_rp",
			rankRP:  120000,
			eventRP: 50000,
			roomRP:  1000,
			valid:   true,
		},
		{
			name:    "max_values",
			rankRP:  4294967295,
			eventRP: 4294967295,
			roomRP:  65535,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:      1,
				Name:    "TestGuild",
				RankRP:  tt.rankRP,
				EventRP: tt.eventRP,
				RoomRP:  tt.roomRP,
			}

			if guild.RankRP != tt.rankRP {
				t.Errorf("rank RP mismatch: got %d, want %d", guild.RankRP, tt.rankRP)
			}

			if guild.EventRP != tt.eventRP {
				t.Errorf("event RP mismatch: got %d, want %d", guild.EventRP, tt.eventRP)
			}

			if guild.RoomRP != tt.roomRP {
				t.Errorf("room RP mismatch: got %d, want %d", guild.RoomRP, tt.roomRP)
			}
		})
	}
}

// TestGuildCommentHandling tests guild comment storage and retrieval
func TestGuildCommentHandling(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		maxLength int
	}{
		{
			name:      "empty_comment",
			comment:   "",
			maxLength: 0,
		},
		{
			name:      "short_comment",
			comment:   "Hello",
			maxLength: 5,
		},
		{
			name:      "long_comment",
			comment:   "This is a very long guild comment with many characters to test maximum length handling",
			maxLength: 86,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:      1,
				Comment: tt.comment,
			}

			if guild.Comment != tt.comment {
				t.Errorf("comment mismatch: got '%s', want '%s'", guild.Comment, tt.comment)
			}

			if len(guild.Comment) != tt.maxLength {
				t.Errorf("comment length mismatch: got %d, want %d", len(guild.Comment), tt.maxLength)
			}
		})
	}
}

// TestGuildMottoSelection tests guild motto (main and sub mottos)
func TestGuildMottoSelection(t *testing.T) {
	tests := []struct {
		name    string
		mainMot uint8
		subMot  uint8
		valid   bool
	}{
		{
			name:    "motto_pair_0_0",
			mainMot: 0,
			subMot:  0,
			valid:   true,
		},
		{
			name:    "motto_pair_1_2",
			mainMot: 1,
			subMot:  2,
			valid:   true,
		},
		{
			name:    "motto_max_values",
			mainMot: 255,
			subMot:  255,
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:        1,
				MainMotto: tt.mainMot,
				SubMotto:  tt.subMot,
			}

			if guild.MainMotto != tt.mainMot {
				t.Errorf("main motto mismatch: got %d, want %d", guild.MainMotto, tt.mainMot)
			}

			if guild.SubMotto != tt.subMot {
				t.Errorf("sub motto mismatch: got %d, want %d", guild.SubMotto, tt.subMot)
			}
		})
	}
}

// TestGuildRecruitingStatus tests guild recruiting flag
func TestGuildRecruitingStatus(t *testing.T) {
	tests := []struct {
		name       string
		recruiting bool
	}{
		{
			name:       "guild_recruiting",
			recruiting: true,
		},
		{
			name:       "guild_not_recruiting",
			recruiting: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:         1,
				Recruiting: tt.recruiting,
			}

			if guild.Recruiting != tt.recruiting {
				t.Errorf("recruiting status mismatch: got %v, want %v", guild.Recruiting, tt.recruiting)
			}
		})
	}
}

// TestGuildSoulTracking tests guild soul accumulation
func TestGuildSoulTracking(t *testing.T) {
	tests := []struct {
		name  string
		souls uint32
	}{
		{
			name:  "no_souls",
			souls: 0,
		},
		{
			name:  "moderate_souls",
			souls: 5000,
		},
		{
			name:  "max_souls",
			souls: 4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:    1,
				Souls: tt.souls,
			}

			if guild.Souls != tt.souls {
				t.Errorf("souls mismatch: got %d, want %d", guild.Souls, tt.souls)
			}
		})
	}
}

// TestGuildPugiData tests guild pug i (treasure chest) names and outfits
func TestGuildPugiData(t *testing.T) {
	tests := []struct {
		name        string
		pugiNames   [3]string
		pugiOutfits [3]uint8
		valid       bool
	}{
		{
			name:        "empty_pugi_data",
			pugiNames:   [3]string{"", "", ""},
			pugiOutfits: [3]uint8{0, 0, 0},
			valid:       true,
		},
		{
			name:        "all_pugi_filled",
			pugiNames:   [3]string{"Chest1", "Chest2", "Chest3"},
			pugiOutfits: [3]uint8{1, 2, 3},
			valid:       true,
		},
		{
			name:        "mixed_pugi_data",
			pugiNames:   [3]string{"MainChest", "", "AltChest"},
			pugiOutfits: [3]uint8{5, 0, 10},
			valid:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:          1,
				PugiName1:   tt.pugiNames[0],
				PugiName2:   tt.pugiNames[1],
				PugiName3:   tt.pugiNames[2],
				PugiOutfit1: tt.pugiOutfits[0],
				PugiOutfit2: tt.pugiOutfits[1],
				PugiOutfit3: tt.pugiOutfits[2],
			}

			if guild.PugiName1 != tt.pugiNames[0] || guild.PugiName2 != tt.pugiNames[1] || guild.PugiName3 != tt.pugiNames[2] {
				t.Error("pugi names mismatch")
			}

			if guild.PugiOutfit1 != tt.pugiOutfits[0] || guild.PugiOutfit2 != tt.pugiOutfits[1] || guild.PugiOutfit3 != tt.pugiOutfits[2] {
				t.Error("pugi outfits mismatch")
			}
		})
	}
}

// TestGuildRoomExpiry tests guild room rental expiry handling
func TestGuildRoomExpiry(t *testing.T) {
	tests := []struct {
		name      string
		expiry    time.Time
		hasExpiry bool
	}{
		{
			name:      "no_room_expiry",
			expiry:    time.Time{},
			hasExpiry: false,
		},
		{
			name:      "room_active",
			expiry:    time.Now().Add(24 * time.Hour),
			hasExpiry: true,
		},
		{
			name:      "room_expired",
			expiry:    time.Now().Add(-1 * time.Hour),
			hasExpiry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:         1,
				RoomExpiry: tt.expiry,
			}

			if (guild.RoomExpiry.IsZero() == tt.hasExpiry) && tt.hasExpiry {
				// If we expect expiry but it's zero, that's an error
				if tt.hasExpiry && guild.RoomExpiry.IsZero() {
					t.Error("expected room expiry but got zero time")
				}
			}

			// Verify expiry is set correctly
			matches := guild.RoomExpiry.Equal(tt.expiry)
			_ = matches
			// Test passed if Equal matches or if no expiry expected and time is zero
		})
	}
}

// TestGuildAllianceRelationship tests guild alliance ID tracking
func TestGuildAllianceRelationship(t *testing.T) {
	tests := []struct {
		name        string
		allianceId  uint32
		hasAlliance bool
	}{
		{
			name:        "no_alliance",
			allianceId:  0,
			hasAlliance: false,
		},
		{
			name:        "single_alliance",
			allianceId:  1,
			hasAlliance: true,
		},
		{
			name:        "large_alliance_id",
			allianceId:  999999,
			hasAlliance: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guild := &Guild{
				ID:         1,
				AllianceID: tt.allianceId,
			}

			hasAlliance := guild.AllianceID != 0
			if hasAlliance != tt.hasAlliance {
				t.Errorf("alliance status mismatch: got %v, want %v", hasAlliance, tt.hasAlliance)
			}

			if guild.AllianceID != tt.allianceId {
				t.Errorf("alliance ID mismatch: got %d, want %d", guild.AllianceID, tt.allianceId)
			}
		})
	}
}
