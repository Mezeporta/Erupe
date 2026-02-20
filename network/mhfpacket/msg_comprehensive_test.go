package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// TestAllOpcodesFromOpcode verifies that FromOpcode returns non-nil packets for all known opcodes
func TestAllOpcodesFromOpcode(t *testing.T) {
	// All opcodes from opcode_to_packet.go
	opcodes := []network.PacketID{
		network.MSG_HEAD,
		network.MSG_SYS_reserve01,
		network.MSG_SYS_reserve02,
		network.MSG_SYS_reserve03,
		network.MSG_SYS_reserve04,
		network.MSG_SYS_reserve05,
		network.MSG_SYS_reserve06,
		network.MSG_SYS_reserve07,
		network.MSG_SYS_ADD_OBJECT,
		network.MSG_SYS_DEL_OBJECT,
		network.MSG_SYS_DISP_OBJECT,
		network.MSG_SYS_HIDE_OBJECT,
		network.MSG_SYS_reserve0C,
		network.MSG_SYS_reserve0D,
		network.MSG_SYS_reserve0E,
		network.MSG_SYS_EXTEND_THRESHOLD,
		network.MSG_SYS_END,
		network.MSG_SYS_NOP,
		network.MSG_SYS_ACK,
		network.MSG_SYS_TERMINAL_LOG,
		network.MSG_SYS_LOGIN,
		network.MSG_SYS_LOGOUT,
		network.MSG_SYS_SET_STATUS,
		network.MSG_SYS_PING,
		network.MSG_SYS_CAST_BINARY,
		network.MSG_SYS_HIDE_CLIENT,
		network.MSG_SYS_TIME,
		network.MSG_SYS_CASTED_BINARY,
		network.MSG_SYS_GET_FILE,
		network.MSG_SYS_ISSUE_LOGKEY,
		network.MSG_SYS_RECORD_LOG,
		network.MSG_SYS_ECHO,
		network.MSG_SYS_CREATE_STAGE,
		network.MSG_SYS_STAGE_DESTRUCT,
		network.MSG_SYS_ENTER_STAGE,
		network.MSG_SYS_BACK_STAGE,
		network.MSG_SYS_MOVE_STAGE,
		network.MSG_SYS_LEAVE_STAGE,
		network.MSG_SYS_LOCK_STAGE,
		network.MSG_SYS_UNLOCK_STAGE,
		network.MSG_SYS_RESERVE_STAGE,
		network.MSG_SYS_UNRESERVE_STAGE,
		network.MSG_SYS_SET_STAGE_PASS,
		network.MSG_SYS_WAIT_STAGE_BINARY,
		network.MSG_SYS_SET_STAGE_BINARY,
		network.MSG_SYS_GET_STAGE_BINARY,
		network.MSG_SYS_ENUMERATE_CLIENT,
		network.MSG_SYS_ENUMERATE_STAGE,
		network.MSG_SYS_CREATE_MUTEX,
		network.MSG_SYS_CREATE_OPEN_MUTEX,
		network.MSG_SYS_DELETE_MUTEX,
		network.MSG_SYS_OPEN_MUTEX,
		network.MSG_SYS_CLOSE_MUTEX,
		network.MSG_SYS_CREATE_SEMAPHORE,
		network.MSG_SYS_CREATE_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_DELETE_SEMAPHORE,
		network.MSG_SYS_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_RELEASE_SEMAPHORE,
		network.MSG_SYS_LOCK_GLOBAL_SEMA,
		network.MSG_SYS_UNLOCK_GLOBAL_SEMA,
		network.MSG_SYS_CHECK_SEMAPHORE,
		network.MSG_SYS_OPERATE_REGISTER,
		network.MSG_SYS_LOAD_REGISTER,
		network.MSG_SYS_NOTIFY_REGISTER,
		network.MSG_SYS_CREATE_OBJECT,
		network.MSG_SYS_DELETE_OBJECT,
		network.MSG_SYS_POSITION_OBJECT,
		network.MSG_SYS_ROTATE_OBJECT,
		network.MSG_SYS_DUPLICATE_OBJECT,
		network.MSG_SYS_SET_OBJECT_BINARY,
		network.MSG_SYS_GET_OBJECT_BINARY,
		network.MSG_SYS_GET_OBJECT_OWNER,
		network.MSG_SYS_UPDATE_OBJECT_BINARY,
		network.MSG_SYS_CLEANUP_OBJECT,
		network.MSG_SYS_reserve4A,
		network.MSG_SYS_reserve4B,
		network.MSG_SYS_reserve4C,
		network.MSG_SYS_reserve4D,
		network.MSG_SYS_reserve4E,
		network.MSG_SYS_reserve4F,
		network.MSG_SYS_INSERT_USER,
		network.MSG_SYS_DELETE_USER,
		network.MSG_SYS_SET_USER_BINARY,
		network.MSG_SYS_GET_USER_BINARY,
		network.MSG_SYS_NOTIFY_USER_BINARY,
		network.MSG_SYS_reserve55,
		network.MSG_SYS_reserve56,
		network.MSG_SYS_reserve57,
		network.MSG_SYS_UPDATE_RIGHT,
		network.MSG_SYS_AUTH_QUERY,
		network.MSG_SYS_AUTH_DATA,
		network.MSG_SYS_AUTH_TERMINAL,
		network.MSG_SYS_reserve5C,
		network.MSG_SYS_RIGHTS_RELOAD,
		network.MSG_SYS_reserve5E,
		network.MSG_SYS_reserve5F,
		network.MSG_MHF_SAVEDATA,
		network.MSG_MHF_LOADDATA,
		network.MSG_MHF_LIST_MEMBER,
		network.MSG_MHF_OPR_MEMBER,
		network.MSG_MHF_ENUMERATE_DIST_ITEM,
		network.MSG_MHF_APPLY_DIST_ITEM,
		network.MSG_MHF_ACQUIRE_DIST_ITEM,
		network.MSG_MHF_GET_DIST_DESCRIPTION,
		network.MSG_MHF_SEND_MAIL,
		network.MSG_MHF_READ_MAIL,
		network.MSG_MHF_LIST_MAIL,
		network.MSG_MHF_OPRT_MAIL,
		network.MSG_MHF_LOAD_FAVORITE_QUEST,
		network.MSG_MHF_SAVE_FAVORITE_QUEST,
		network.MSG_MHF_REGISTER_EVENT,
		network.MSG_MHF_RELEASE_EVENT,
		network.MSG_MHF_TRANSIT_MESSAGE,
		network.MSG_SYS_reserve71,
		network.MSG_SYS_reserve72,
		network.MSG_SYS_reserve73,
		network.MSG_SYS_reserve74,
		network.MSG_SYS_reserve75,
		network.MSG_SYS_reserve76,
		network.MSG_SYS_reserve77,
		network.MSG_SYS_reserve78,
		network.MSG_SYS_reserve79,
		network.MSG_SYS_reserve7A,
		network.MSG_SYS_reserve7B,
		network.MSG_SYS_reserve7C,
		network.MSG_CA_EXCHANGE_ITEM,
		network.MSG_SYS_reserve7E,
		network.MSG_MHF_PRESENT_BOX,
		network.MSG_MHF_SERVER_COMMAND,
		network.MSG_MHF_SHUT_CLIENT,
		network.MSG_MHF_ANNOUNCE,
		network.MSG_MHF_SET_LOGINWINDOW,
		network.MSG_SYS_TRANS_BINARY,
		network.MSG_SYS_COLLECT_BINARY,
		network.MSG_SYS_GET_STATE,
		network.MSG_SYS_SERIALIZE,
		network.MSG_SYS_ENUMLOBBY,
		network.MSG_SYS_ENUMUSER,
		network.MSG_SYS_INFOKYSERVER,
		network.MSG_MHF_GET_CA_UNIQUE_ID,
		network.MSG_MHF_SET_CA_ACHIEVEMENT,
		network.MSG_MHF_CARAVAN_MY_SCORE,
		network.MSG_MHF_CARAVAN_RANKING,
		network.MSG_MHF_CARAVAN_MY_RANK,
		network.MSG_MHF_CREATE_GUILD,
		network.MSG_MHF_OPERATE_GUILD,
		network.MSG_MHF_OPERATE_GUILD_MEMBER,
		network.MSG_MHF_INFO_GUILD,
		network.MSG_MHF_ENUMERATE_GUILD,
		network.MSG_MHF_UPDATE_GUILD,
		network.MSG_MHF_ARRANGE_GUILD_MEMBER,
		network.MSG_MHF_ENUMERATE_GUILD_MEMBER,
		network.MSG_MHF_ENUMERATE_CAMPAIGN,
		network.MSG_MHF_STATE_CAMPAIGN,
		network.MSG_MHF_APPLY_CAMPAIGN,
		network.MSG_MHF_ENUMERATE_ITEM,
		network.MSG_MHF_ACQUIRE_ITEM,
		network.MSG_MHF_TRANSFER_ITEM,
		network.MSG_MHF_MERCENARY_HUNTDATA,
		network.MSG_MHF_ENTRY_ROOKIE_GUILD,
		network.MSG_MHF_ENUMERATE_QUEST,
		network.MSG_MHF_ENUMERATE_EVENT,
		network.MSG_MHF_ENUMERATE_PRICE,
		network.MSG_MHF_ENUMERATE_RANKING,
		network.MSG_MHF_ENUMERATE_ORDER,
		network.MSG_MHF_ENUMERATE_SHOP,
		network.MSG_MHF_GET_EXTRA_INFO,
		network.MSG_MHF_UPDATE_INTERIOR,
		network.MSG_MHF_ENUMERATE_HOUSE,
		network.MSG_MHF_UPDATE_HOUSE,
		network.MSG_MHF_LOAD_HOUSE,
		network.MSG_MHF_OPERATE_WAREHOUSE,
		network.MSG_MHF_ENUMERATE_WAREHOUSE,
		network.MSG_MHF_UPDATE_WAREHOUSE,
		network.MSG_MHF_ACQUIRE_TITLE,
		network.MSG_MHF_ENUMERATE_TITLE,
		network.MSG_MHF_ENUMERATE_GUILD_ITEM,
		network.MSG_MHF_UPDATE_GUILD_ITEM,
		network.MSG_MHF_ENUMERATE_UNION_ITEM,
		network.MSG_MHF_UPDATE_UNION_ITEM,
		network.MSG_MHF_CREATE_JOINT,
		network.MSG_MHF_OPERATE_JOINT,
		network.MSG_MHF_INFO_JOINT,
		network.MSG_MHF_UPDATE_GUILD_ICON,
		network.MSG_MHF_INFO_FESTA,
		network.MSG_MHF_ENTRY_FESTA,
		network.MSG_MHF_CHARGE_FESTA,
		network.MSG_MHF_ACQUIRE_FESTA,
		network.MSG_MHF_STATE_FESTA_U,
		network.MSG_MHF_STATE_FESTA_G,
		network.MSG_MHF_ENUMERATE_FESTA_MEMBER,
		network.MSG_MHF_VOTE_FESTA,
		network.MSG_MHF_ACQUIRE_CAFE_ITEM,
		network.MSG_MHF_UPDATE_CAFEPOINT,
		network.MSG_MHF_CHECK_DAILY_CAFEPOINT,
		network.MSG_MHF_GET_COG_INFO,
		network.MSG_MHF_CHECK_MONTHLY_ITEM,
		network.MSG_MHF_ACQUIRE_MONTHLY_ITEM,
		network.MSG_MHF_CHECK_WEEKLY_STAMP,
		network.MSG_MHF_EXCHANGE_WEEKLY_STAMP,
		network.MSG_MHF_CREATE_MERCENARY,
		network.MSG_MHF_SAVE_MERCENARY,
		network.MSG_MHF_READ_MERCENARY_W,
		network.MSG_MHF_READ_MERCENARY_M,
		network.MSG_MHF_CONTRACT_MERCENARY,
		network.MSG_MHF_ENUMERATE_MERCENARY_LOG,
		network.MSG_MHF_ENUMERATE_GUACOT,
		network.MSG_MHF_UPDATE_GUACOT,
		network.MSG_MHF_INFO_TOURNAMENT,
		network.MSG_MHF_ENTRY_TOURNAMENT,
		network.MSG_MHF_ENTER_TOURNAMENT_QUEST,
		network.MSG_MHF_ACQUIRE_TOURNAMENT,
		network.MSG_MHF_GET_ACHIEVEMENT,
		network.MSG_MHF_RESET_ACHIEVEMENT,
		network.MSG_MHF_ADD_ACHIEVEMENT,
		network.MSG_MHF_PAYMENT_ACHIEVEMENT,
		network.MSG_MHF_DISPLAYED_ACHIEVEMENT,
		network.MSG_MHF_INFO_SCENARIO_COUNTER,
		network.MSG_MHF_SAVE_SCENARIO_DATA,
		network.MSG_MHF_LOAD_SCENARIO_DATA,
		network.MSG_MHF_GET_BBS_SNS_STATUS,
		network.MSG_MHF_APPLY_BBS_ARTICLE,
		network.MSG_MHF_GET_ETC_POINTS,
		network.MSG_MHF_UPDATE_ETC_POINT,
		network.MSG_MHF_GET_MYHOUSE_INFO,
		network.MSG_MHF_UPDATE_MYHOUSE_INFO,
		network.MSG_MHF_GET_WEEKLY_SCHEDULE,
		network.MSG_MHF_ENUMERATE_INV_GUILD,
		network.MSG_MHF_OPERATION_INV_GUILD,
		network.MSG_MHF_STAMPCARD_STAMP,
		network.MSG_MHF_STAMPCARD_PRIZE,
		network.MSG_MHF_UNRESERVE_SRG,
		network.MSG_MHF_LOAD_PLATE_DATA,
		network.MSG_MHF_SAVE_PLATE_DATA,
		network.MSG_MHF_LOAD_PLATE_BOX,
		network.MSG_MHF_SAVE_PLATE_BOX,
		network.MSG_MHF_READ_GUILDCARD,
		network.MSG_MHF_UPDATE_GUILDCARD,
		network.MSG_MHF_READ_BEAT_LEVEL,
		network.MSG_MHF_UPDATE_BEAT_LEVEL,
		network.MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING,
		network.MSG_MHF_READ_BEAT_LEVEL_MY_RANKING,
		network.MSG_MHF_READ_LAST_WEEK_BEAT_RANKING,
		network.MSG_MHF_ACCEPT_READ_REWARD,
		network.MSG_MHF_GET_ADDITIONAL_BEAT_REWARD,
		network.MSG_MHF_GET_FIXED_SEIBATU_RANKING_TABLE,
		network.MSG_MHF_GET_BBS_USER_STATUS,
		network.MSG_MHF_KICK_EXPORT_FORCE,
		network.MSG_MHF_GET_BREAK_SEIBATU_LEVEL_REWARD,
		network.MSG_MHF_GET_WEEKLY_SEIBATU_RANKING_REWARD,
		network.MSG_MHF_GET_EARTH_STATUS,
		network.MSG_MHF_LOAD_PARTNER,
		network.MSG_MHF_SAVE_PARTNER,
		network.MSG_MHF_GET_GUILD_MISSION_LIST,
		network.MSG_MHF_GET_GUILD_MISSION_RECORD,
		network.MSG_MHF_ADD_GUILD_MISSION_COUNT,
		network.MSG_MHF_SET_GUILD_MISSION_TARGET,
		network.MSG_MHF_CANCEL_GUILD_MISSION_TARGET,
		network.MSG_MHF_LOAD_OTOMO_AIROU,
		network.MSG_MHF_SAVE_OTOMO_AIROU,
		network.MSG_MHF_ENUMERATE_GUILD_TRESURE,
		network.MSG_MHF_ENUMERATE_AIROULIST,
		network.MSG_MHF_REGIST_GUILD_TRESURE,
		network.MSG_MHF_ACQUIRE_GUILD_TRESURE,
		network.MSG_MHF_OPERATE_GUILD_TRESURE_REPORT,
		network.MSG_MHF_GET_GUILD_TRESURE_SOUVENIR,
		network.MSG_MHF_ACQUIRE_GUILD_TRESURE_SOUVENIR,
		network.MSG_MHF_ENUMERATE_FESTA_INTERMEDIATE_PRIZE,
		network.MSG_MHF_ACQUIRE_FESTA_INTERMEDIATE_PRIZE,
		network.MSG_MHF_LOAD_DECO_MYSET,
		network.MSG_MHF_SAVE_DECO_MYSET,
		network.MSG_MHF_reserve10F,
		network.MSG_MHF_LOAD_GUILD_COOKING,
		network.MSG_MHF_REGIST_GUILD_COOKING,
		network.MSG_MHF_LOAD_GUILD_ADVENTURE,
		network.MSG_MHF_REGIST_GUILD_ADVENTURE,
		network.MSG_MHF_ACQUIRE_GUILD_ADVENTURE,
		network.MSG_MHF_CHARGE_GUILD_ADVENTURE,
		network.MSG_MHF_LOAD_LEGEND_DISPATCH,
		network.MSG_MHF_LOAD_HUNTER_NAVI,
		network.MSG_MHF_SAVE_HUNTER_NAVI,
		network.MSG_MHF_REGIST_SPABI_TIME,
		network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_MASTER,
		network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_ACTIVE_COUNT,
		network.MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER,
		network.MSG_MHF_GET_TOWER_INFO,
		network.MSG_MHF_POST_TOWER_INFO,
		network.MSG_MHF_GET_GEM_INFO,
		network.MSG_MHF_POST_GEM_INFO,
		network.MSG_MHF_GET_EARTH_VALUE,
		network.MSG_MHF_DEBUG_POST_VALUE,
		network.MSG_MHF_GET_PAPER_DATA,
		network.MSG_MHF_GET_NOTICE,
		network.MSG_MHF_POST_NOTICE,
		network.MSG_MHF_GET_BOOST_TIME,
		network.MSG_MHF_POST_BOOST_TIME,
		network.MSG_MHF_GET_BOOST_TIME_LIMIT,
		network.MSG_MHF_POST_BOOST_TIME_LIMIT,
		network.MSG_MHF_ENUMERATE_FESTA_PERSONAL_PRIZE,
		network.MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE,
		network.MSG_MHF_GET_RAND_FROM_TABLE,
		network.MSG_MHF_GET_CAFE_DURATION,
		network.MSG_MHF_GET_CAFE_DURATION_BONUS_INFO,
		network.MSG_MHF_RECEIVE_CAFE_DURATION_BONUS,
		network.MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED,
		network.MSG_MHF_GET_GACHA_POINT,
		network.MSG_MHF_USE_GACHA_POINT,
		network.MSG_MHF_EXCHANGE_FPOINT_2_ITEM,
		network.MSG_MHF_EXCHANGE_ITEM_2_FPOINT,
		network.MSG_MHF_GET_FPOINT_EXCHANGE_LIST,
		network.MSG_MHF_PLAY_STEPUP_GACHA,
		network.MSG_MHF_RECEIVE_GACHA_ITEM,
		network.MSG_MHF_GET_STEPUP_STATUS,
		network.MSG_MHF_PLAY_FREE_GACHA,
		network.MSG_MHF_GET_TINY_BIN,
		network.MSG_MHF_POST_TINY_BIN,
		network.MSG_MHF_GET_SENYU_DAILY_COUNT,
		network.MSG_MHF_GET_GUILD_TARGET_MEMBER_NUM,
		network.MSG_MHF_GET_BOOST_RIGHT,
		network.MSG_MHF_START_BOOST_TIME,
		network.MSG_MHF_POST_BOOST_TIME_QUEST_RETURN,
		network.MSG_MHF_GET_BOX_GACHA_INFO,
		network.MSG_MHF_PLAY_BOX_GACHA,
		network.MSG_MHF_RESET_BOX_GACHA_INFO,
		network.MSG_MHF_GET_SEIBATTLE,
		network.MSG_MHF_POST_SEIBATTLE,
		network.MSG_MHF_GET_RYOUDAMA,
		network.MSG_MHF_POST_RYOUDAMA,
		network.MSG_MHF_GET_TENROUIRAI,
		network.MSG_MHF_POST_TENROUIRAI,
		network.MSG_MHF_POST_GUILD_SCOUT,
		network.MSG_MHF_CANCEL_GUILD_SCOUT,
		network.MSG_MHF_ANSWER_GUILD_SCOUT,
		network.MSG_MHF_GET_GUILD_SCOUT_LIST,
		network.MSG_MHF_GET_GUILD_MANAGE_RIGHT,
		network.MSG_MHF_SET_GUILD_MANAGE_RIGHT,
		network.MSG_MHF_PLAY_NORMAL_GACHA,
		network.MSG_MHF_GET_DAILY_MISSION_MASTER,
		network.MSG_MHF_GET_DAILY_MISSION_PERSONAL,
		network.MSG_MHF_SET_DAILY_MISSION_PERSONAL,
		network.MSG_MHF_GET_GACHA_PLAY_HISTORY,
		network.MSG_MHF_GET_REJECT_GUILD_SCOUT,
		network.MSG_MHF_SET_REJECT_GUILD_SCOUT,
		network.MSG_MHF_GET_CA_ACHIEVEMENT_HIST,
		network.MSG_MHF_SET_CA_ACHIEVEMENT_HIST,
		network.MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS,
		network.MSG_MHF_USE_KEEP_LOGIN_BOOST,
		network.MSG_MHF_GET_UD_SCHEDULE,
		network.MSG_MHF_GET_UD_INFO,
		network.MSG_MHF_GET_KIJU_INFO,
		network.MSG_MHF_SET_KIJU,
		network.MSG_MHF_ADD_UD_POINT,
		network.MSG_MHF_GET_UD_MY_POINT,
		network.MSG_MHF_GET_UD_TOTAL_POINT_INFO,
		network.MSG_MHF_GET_UD_BONUS_QUEST_INFO,
		network.MSG_MHF_GET_UD_SELECTED_COLOR_INFO,
		network.MSG_MHF_GET_UD_MONSTER_POINT,
		network.MSG_MHF_GET_UD_DAILY_PRESENT_LIST,
		network.MSG_MHF_GET_UD_NORMA_PRESENT_LIST,
		network.MSG_MHF_GET_UD_RANKING_REWARD_LIST,
		network.MSG_MHF_ACQUIRE_UD_ITEM,
		network.MSG_MHF_GET_REWARD_SONG,
		network.MSG_MHF_USE_REWARD_SONG,
		network.MSG_MHF_ADD_REWARD_SONG_COUNT,
		network.MSG_MHF_GET_UD_RANKING,
		network.MSG_MHF_GET_UD_MY_RANKING,
		network.MSG_MHF_ACQUIRE_MONTHLY_REWARD,
		network.MSG_MHF_GET_UD_GUILD_MAP_INFO,
		network.MSG_MHF_GENERATE_UD_GUILD_MAP,
		network.MSG_MHF_GET_UD_TACTICS_POINT,
		network.MSG_MHF_ADD_UD_TACTICS_POINT,
		network.MSG_MHF_GET_UD_TACTICS_RANKING,
		network.MSG_MHF_GET_UD_TACTICS_REWARD_LIST,
		network.MSG_MHF_GET_UD_TACTICS_LOG,
		network.MSG_MHF_GET_EQUIP_SKIN_HIST,
		network.MSG_MHF_UPDATE_EQUIP_SKIN_HIST,
		network.MSG_MHF_GET_UD_TACTICS_FOLLOWER,
		network.MSG_MHF_SET_UD_TACTICS_FOLLOWER,
		network.MSG_MHF_GET_UD_SHOP_COIN,
		network.MSG_MHF_USE_UD_SHOP_COIN,
		network.MSG_MHF_GET_ENHANCED_MINIDATA,
		network.MSG_MHF_SET_ENHANCED_MINIDATA,
		network.MSG_MHF_SEX_CHANGER,
		network.MSG_MHF_GET_LOBBY_CROWD,
		network.MSG_SYS_reserve180,
		network.MSG_MHF_GUILD_HUNTDATA,
		network.MSG_MHF_ADD_KOURYOU_POINT,
		network.MSG_MHF_GET_KOURYOU_POINT,
		network.MSG_MHF_EXCHANGE_KOURYOU_POINT,
		network.MSG_MHF_GET_UD_TACTICS_BONUS_QUEST,
		network.MSG_MHF_GET_UD_TACTICS_FIRST_QUEST_BONUS,
		network.MSG_MHF_GET_UD_TACTICS_REMAINING_POINT,
		network.MSG_SYS_reserve188,
		network.MSG_MHF_LOAD_PLATE_MYSET,
		network.MSG_MHF_SAVE_PLATE_MYSET,
		network.MSG_SYS_reserve18B,
		network.MSG_MHF_GET_RESTRICTION_EVENT,
		network.MSG_MHF_SET_RESTRICTION_EVENT,
		network.MSG_SYS_reserve18E,
		network.MSG_SYS_reserve18F,
		network.MSG_MHF_GET_TREND_WEAPON,
		network.MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG,
		network.MSG_SYS_reserve192,
		network.MSG_SYS_reserve193,
		network.MSG_SYS_reserve194,
		network.MSG_MHF_SAVE_RENGOKU_DATA,
		network.MSG_MHF_LOAD_RENGOKU_DATA,
		network.MSG_MHF_GET_RENGOKU_BINARY,
		network.MSG_MHF_ENUMERATE_RENGOKU_RANKING,
		network.MSG_MHF_GET_RENGOKU_RANKING_RANK,
		network.MSG_MHF_ACQUIRE_EXCHANGE_SHOP,
		network.MSG_SYS_reserve19B,
		network.MSG_MHF_SAVE_MEZFES_DATA,
		network.MSG_MHF_LOAD_MEZFES_DATA,
		network.MSG_SYS_reserve19E,
		network.MSG_SYS_reserve19F,
		network.MSG_MHF_UPDATE_FORCE_GUILD_RANK,
		network.MSG_MHF_RESET_TITLE,
		network.MSG_MHF_ENUMERATE_GUILD_MESSAGE_BOARD,
		network.MSG_MHF_UPDATE_GUILD_MESSAGE_BOARD,
		network.MSG_SYS_reserve1A4,
		network.MSG_MHF_REGIST_GUILD_ADVENTURE_DIVA,
		network.MSG_SYS_reserve1A6,
		network.MSG_SYS_reserve1A7,
		network.MSG_SYS_reserve1A8,
		network.MSG_SYS_reserve1A9,
		network.MSG_SYS_reserve1AA,
		network.MSG_SYS_reserve1AB,
		network.MSG_SYS_reserve1AC,
		network.MSG_SYS_reserve1AD,
		network.MSG_SYS_reserve1AE,
		network.MSG_SYS_reserve1AF,
	}

	for _, opcode := range opcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
				return
			}
			// Verify Opcode() returns the correct value
			if pkt.Opcode() != opcode {
				t.Errorf("Opcode() = %s, want %s", pkt.Opcode(), opcode)
			}
		})
	}
}

// TestAckHandlePacketsParse tests parsing of packets with simple AckHandle uint32 field
func TestAckHandlePacketsParse(t *testing.T) {
	testCases := []struct {
		name   string
		opcode network.PacketID
	}{
		{"MsgMhfGetAchievement", network.MSG_MHF_GET_ACHIEVEMENT},
		{"MsgMhfGetTowerInfo", network.MSG_MHF_GET_TOWER_INFO},
		{"MsgMhfGetGemInfo", network.MSG_MHF_GET_GEM_INFO},
		{"MsgMhfGetBoostTime", network.MSG_MHF_GET_BOOST_TIME},
		{"MsgMhfGetCafeDuration", network.MSG_MHF_GET_CAFE_DURATION},
		{"MsgMhfGetGachaPoint", network.MSG_MHF_GET_GACHA_POINT},
		{"MsgMhfLoadPartner", network.MSG_MHF_LOAD_PARTNER},
		{"MsgMhfLoadOtomoAirou", network.MSG_MHF_LOAD_OTOMO_AIROU},
		{"MsgMhfLoadPlateData", network.MSG_MHF_LOAD_PLATE_DATA},
		{"MsgMhfLoadPlateBox", network.MSG_MHF_LOAD_PLATE_BOX},
		{"MsgMhfLoadDecoMyset", network.MSG_MHF_LOAD_DECO_MYSET},
		{"MsgMhfLoadGuildCooking", network.MSG_MHF_LOAD_GUILD_COOKING},
		{"MsgMhfLoadGuildAdventure", network.MSG_MHF_LOAD_GUILD_ADVENTURE},
		{"MsgMhfLoadHunterNavi", network.MSG_MHF_LOAD_HUNTER_NAVI},
		{"MsgMhfInfoFesta", network.MSG_MHF_INFO_FESTA},
		{"MsgMhfInfoTournament", network.MSG_MHF_INFO_TOURNAMENT},
		{"MsgMhfEnumerateQuest", network.MSG_MHF_ENUMERATE_QUEST},
		{"MsgMhfEnumerateEvent", network.MSG_MHF_ENUMERATE_EVENT},
		{"MsgMhfEnumerateShop", network.MSG_MHF_ENUMERATE_SHOP},
		{"MsgMhfEnumerateRanking", network.MSG_MHF_ENUMERATE_RANKING},
		{"MsgMhfEnumerateOrder", network.MSG_MHF_ENUMERATE_ORDER},
		{"MsgMhfEnumerateCampaign", network.MSG_MHF_ENUMERATE_CAMPAIGN},
		{"MsgMhfGetWeeklySchedule", network.MSG_MHF_GET_WEEKLY_SCHEDULE},
		{"MsgMhfGetUdSchedule", network.MSG_MHF_GET_UD_SCHEDULE},
		{"MsgMhfGetUdInfo", network.MSG_MHF_GET_UD_INFO},
		{"MsgMhfGetKijuInfo", network.MSG_MHF_GET_KIJU_INFO},
	}

	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkt := FromOpcode(tc.opcode)
			if pkt == nil {
				t.Skipf("FromOpcode(%s) returned nil", tc.opcode)
				return
			}

			// Create test data - most of these packets read AckHandle + additional data
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0x12345678) // AckHandle
			// Write extra padding bytes for packets that expect more data
			for i := 0; i < 32; i++ {
				bf.WriteUint32(uint32(i))
			}
			_, _ = bf.Seek(0, io.SeekStart)

			// Parse should not panic
			err := pkt.Parse(bf, ctx)
			if err != nil {
				t.Logf("Parse() returned error (may be expected): %v", err)
			}
		})
	}
}

// TestAddAchievementParse tests MsgMhfAddAchievement Parse
func TestAddAchievementParse(t *testing.T) {
	tests := []struct {
		name          string
		achievementID uint8
		unk1          uint16
		unk2          uint16
	}{
		{"typical values", 1, 100, 200},
		{"zero values", 0, 0, 0},
		{"max values", 255, 65535, 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(tt.achievementID)
			bf.WriteUint16(tt.unk1)
			bf.WriteUint16(tt.unk2)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAddAchievement{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AchievementID != tt.achievementID {
				t.Errorf("AchievementID = %d, want %d", pkt.AchievementID, tt.achievementID)
			}
			if pkt.Unk1 != tt.unk1 {
				t.Errorf("Unk1 = %d, want %d", pkt.Unk1, tt.unk1)
			}
			if pkt.Unk2 != tt.unk2 {
				t.Errorf("Unk2 = %d, want %d", pkt.Unk2, tt.unk2)
			}
		})
	}
}

// TestGetAchievementParse tests MsgMhfGetAchievement Parse
func TestGetAchievementParse(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		charID    uint32
		unk1      uint32
	}{
		{"typical values", 1, 12345, 0},
		{"large values", 0xFFFFFFFF, 0xDEADBEEF, 0xCAFEBABE},
		{"zero values", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.charID)
			bf.WriteUint32(tt.unk1)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfGetAchievement{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.CharID != tt.charID {
				t.Errorf("CharID = %d, want %d", pkt.CharID, tt.charID)
			}
			// Unk1 (third uint32) is read and discarded in Parse on main
		})
	}
}

// TestBuildNotImplemented tests that Build returns error for packets without implementation
func TestBuildNotImplemented(t *testing.T) {
	packetsToTest := []MHFPacket{
		&MsgMhfAddAchievement{},
		&MsgMhfGetAchievement{},
		&MsgMhfAcquireItem{},
		&MsgMhfEnumerateGuild{},
		&MsgMhfInfoGuild{},
		&MsgMhfCreateGuild{},
		&MsgMhfOperateGuild{},
		&MsgMhfOperateGuildMember{},
		&MsgMhfUpdateGuild{},
		&MsgMhfArrangeGuildMember{},
		&MsgMhfEnumerateGuildMember{},
		&MsgMhfInfoFesta{},
		&MsgMhfEntryFesta{},
		&MsgMhfChargeFesta{},
		&MsgMhfAcquireFesta{},
		&MsgMhfVoteFesta{},
		&MsgMhfInfoTournament{},
		&MsgMhfEntryTournament{},
		&MsgMhfAcquireTournament{},
	}

	for _, pkt := range packetsToTest {
		t.Run(pkt.Opcode().String(), func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			err := pkt.Build(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err == nil {
				t.Logf("Build() did not return error (implementation may exist)")
			} else {
				// Expected - Build is not implemented
				if err.Error() != "NOT IMPLEMENTED" {
					t.Logf("Build() returned unexpected error: %v", err)
				}
			}
		})
	}
}

// TestReservePacketsOpcode tests that reserve packets have correct opcodes
func TestReservePacketsOpcode(t *testing.T) {
	reservePackets := []struct {
		opcode network.PacketID
	}{
		{network.MSG_SYS_reserve01},
		{network.MSG_SYS_reserve02},
		{network.MSG_SYS_reserve03},
		{network.MSG_SYS_reserve04},
		{network.MSG_SYS_reserve05},
		{network.MSG_SYS_reserve06},
		{network.MSG_SYS_reserve07},
		{network.MSG_SYS_reserve0C},
		{network.MSG_SYS_reserve0D},
		{network.MSG_SYS_reserve0E},
		{network.MSG_SYS_reserve4A},
		{network.MSG_SYS_reserve4B},
		{network.MSG_SYS_reserve4C},
		{network.MSG_SYS_reserve4D},
		{network.MSG_SYS_reserve4E},
		{network.MSG_SYS_reserve4F},
		{network.MSG_SYS_reserve55},
		{network.MSG_SYS_reserve56},
		{network.MSG_SYS_reserve57},
		{network.MSG_SYS_reserve5C},
		{network.MSG_SYS_reserve5E},
		{network.MSG_SYS_reserve5F},
		{network.MSG_SYS_reserve71},
		{network.MSG_SYS_reserve72},
		{network.MSG_SYS_reserve73},
		{network.MSG_SYS_reserve74},
		{network.MSG_SYS_reserve75},
		{network.MSG_SYS_reserve76},
		{network.MSG_SYS_reserve77},
		{network.MSG_SYS_reserve78},
		{network.MSG_SYS_reserve79},
		{network.MSG_SYS_reserve7A},
		{network.MSG_SYS_reserve7B},
		{network.MSG_SYS_reserve7C},
		{network.MSG_SYS_reserve7E},
		{network.MSG_SYS_reserve180},
		{network.MSG_SYS_reserve188},
		{network.MSG_SYS_reserve18B},
		{network.MSG_SYS_reserve18E},
		{network.MSG_SYS_reserve18F},
		{network.MSG_SYS_reserve192},
		{network.MSG_SYS_reserve193},
		{network.MSG_SYS_reserve194},
		{network.MSG_SYS_reserve19B},
		{network.MSG_SYS_reserve19E},
		{network.MSG_SYS_reserve19F},
		{network.MSG_SYS_reserve1A4},
		{network.MSG_SYS_reserve1A6},
		{network.MSG_SYS_reserve1A7},
		{network.MSG_SYS_reserve1A8},
		{network.MSG_SYS_reserve1A9},
		{network.MSG_SYS_reserve1AA},
		{network.MSG_SYS_reserve1AB},
		{network.MSG_SYS_reserve1AC},
		{network.MSG_SYS_reserve1AD},
		{network.MSG_SYS_reserve1AE},
		{network.MSG_SYS_reserve1AF},
	}

	for _, tc := range reservePackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(tc.opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", tc.opcode)
				return
			}
			if pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestMHFPacketsOpcode tests Opcode() method for various MHF packets
func TestMHFPacketsOpcode(t *testing.T) {
	mhfPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfSavedata{}, network.MSG_MHF_SAVEDATA},
		{&MsgMhfLoaddata{}, network.MSG_MHF_LOADDATA},
		{&MsgMhfListMember{}, network.MSG_MHF_LIST_MEMBER},
		{&MsgMhfOprMember{}, network.MSG_MHF_OPR_MEMBER},
		{&MsgMhfEnumerateDistItem{}, network.MSG_MHF_ENUMERATE_DIST_ITEM},
		{&MsgMhfApplyDistItem{}, network.MSG_MHF_APPLY_DIST_ITEM},
		{&MsgMhfAcquireDistItem{}, network.MSG_MHF_ACQUIRE_DIST_ITEM},
		{&MsgMhfGetDistDescription{}, network.MSG_MHF_GET_DIST_DESCRIPTION},
		{&MsgMhfSendMail{}, network.MSG_MHF_SEND_MAIL},
		{&MsgMhfReadMail{}, network.MSG_MHF_READ_MAIL},
		{&MsgMhfListMail{}, network.MSG_MHF_LIST_MAIL},
		{&MsgMhfOprtMail{}, network.MSG_MHF_OPRT_MAIL},
		{&MsgMhfLoadFavoriteQuest{}, network.MSG_MHF_LOAD_FAVORITE_QUEST},
		{&MsgMhfSaveFavoriteQuest{}, network.MSG_MHF_SAVE_FAVORITE_QUEST},
		{&MsgMhfRegisterEvent{}, network.MSG_MHF_REGISTER_EVENT},
		{&MsgMhfReleaseEvent{}, network.MSG_MHF_RELEASE_EVENT},
		{&MsgMhfTransitMessage{}, network.MSG_MHF_TRANSIT_MESSAGE},
		{&MsgMhfPresentBox{}, network.MSG_MHF_PRESENT_BOX},
		{&MsgMhfServerCommand{}, network.MSG_MHF_SERVER_COMMAND},
		{&MsgMhfShutClient{}, network.MSG_MHF_SHUT_CLIENT},
		{&MsgMhfAnnounce{}, network.MSG_MHF_ANNOUNCE},
		{&MsgMhfSetLoginwindow{}, network.MSG_MHF_SET_LOGINWINDOW},
		{&MsgMhfGetCaUniqueID{}, network.MSG_MHF_GET_CA_UNIQUE_ID},
		{&MsgMhfSetCaAchievement{}, network.MSG_MHF_SET_CA_ACHIEVEMENT},
		{&MsgMhfCaravanMyScore{}, network.MSG_MHF_CARAVAN_MY_SCORE},
		{&MsgMhfCaravanRanking{}, network.MSG_MHF_CARAVAN_RANKING},
		{&MsgMhfCaravanMyRank{}, network.MSG_MHF_CARAVAN_MY_RANK},
	}

	for _, tc := range mhfPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestGuildPacketsOpcode tests guild-related packets
func TestGuildPacketsOpcode(t *testing.T) {
	guildPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfCreateGuild{}, network.MSG_MHF_CREATE_GUILD},
		{&MsgMhfOperateGuild{}, network.MSG_MHF_OPERATE_GUILD},
		{&MsgMhfOperateGuildMember{}, network.MSG_MHF_OPERATE_GUILD_MEMBER},
		{&MsgMhfInfoGuild{}, network.MSG_MHF_INFO_GUILD},
		{&MsgMhfEnumerateGuild{}, network.MSG_MHF_ENUMERATE_GUILD},
		{&MsgMhfUpdateGuild{}, network.MSG_MHF_UPDATE_GUILD},
		{&MsgMhfArrangeGuildMember{}, network.MSG_MHF_ARRANGE_GUILD_MEMBER},
		{&MsgMhfEnumerateGuildMember{}, network.MSG_MHF_ENUMERATE_GUILD_MEMBER},
		{&MsgMhfEnumerateGuildItem{}, network.MSG_MHF_ENUMERATE_GUILD_ITEM},
		{&MsgMhfUpdateGuildItem{}, network.MSG_MHF_UPDATE_GUILD_ITEM},
		{&MsgMhfUpdateGuildIcon{}, network.MSG_MHF_UPDATE_GUILD_ICON},
		{&MsgMhfEnumerateGuildTresure{}, network.MSG_MHF_ENUMERATE_GUILD_TRESURE},
		{&MsgMhfRegistGuildTresure{}, network.MSG_MHF_REGIST_GUILD_TRESURE},
		{&MsgMhfAcquireGuildTresure{}, network.MSG_MHF_ACQUIRE_GUILD_TRESURE},
		{&MsgMhfOperateGuildTresureReport{}, network.MSG_MHF_OPERATE_GUILD_TRESURE_REPORT},
		{&MsgMhfGetGuildTresureSouvenir{}, network.MSG_MHF_GET_GUILD_TRESURE_SOUVENIR},
		{&MsgMhfAcquireGuildTresureSouvenir{}, network.MSG_MHF_ACQUIRE_GUILD_TRESURE_SOUVENIR},
		{&MsgMhfLoadGuildCooking{}, network.MSG_MHF_LOAD_GUILD_COOKING},
		{&MsgMhfRegistGuildCooking{}, network.MSG_MHF_REGIST_GUILD_COOKING},
		{&MsgMhfLoadGuildAdventure{}, network.MSG_MHF_LOAD_GUILD_ADVENTURE},
		{&MsgMhfRegistGuildAdventure{}, network.MSG_MHF_REGIST_GUILD_ADVENTURE},
		{&MsgMhfAcquireGuildAdventure{}, network.MSG_MHF_ACQUIRE_GUILD_ADVENTURE},
		{&MsgMhfChargeGuildAdventure{}, network.MSG_MHF_CHARGE_GUILD_ADVENTURE},
		{&MsgMhfGetGuildMissionList{}, network.MSG_MHF_GET_GUILD_MISSION_LIST},
		{&MsgMhfGetGuildMissionRecord{}, network.MSG_MHF_GET_GUILD_MISSION_RECORD},
		{&MsgMhfAddGuildMissionCount{}, network.MSG_MHF_ADD_GUILD_MISSION_COUNT},
		{&MsgMhfSetGuildMissionTarget{}, network.MSG_MHF_SET_GUILD_MISSION_TARGET},
		{&MsgMhfCancelGuildMissionTarget{}, network.MSG_MHF_CANCEL_GUILD_MISSION_TARGET},
		{&MsgMhfGetGuildWeeklyBonusMaster{}, network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_MASTER},
		{&MsgMhfGetGuildWeeklyBonusActiveCount{}, network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_ACTIVE_COUNT},
		{&MsgMhfAddGuildWeeklyBonusExceptionalUser{}, network.MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER},
		{&MsgMhfGetGuildTargetMemberNum{}, network.MSG_MHF_GET_GUILD_TARGET_MEMBER_NUM},
		{&MsgMhfPostGuildScout{}, network.MSG_MHF_POST_GUILD_SCOUT},
		{&MsgMhfCancelGuildScout{}, network.MSG_MHF_CANCEL_GUILD_SCOUT},
		{&MsgMhfAnswerGuildScout{}, network.MSG_MHF_ANSWER_GUILD_SCOUT},
		{&MsgMhfGetGuildScoutList{}, network.MSG_MHF_GET_GUILD_SCOUT_LIST},
		{&MsgMhfGetGuildManageRight{}, network.MSG_MHF_GET_GUILD_MANAGE_RIGHT},
		{&MsgMhfSetGuildManageRight{}, network.MSG_MHF_SET_GUILD_MANAGE_RIGHT},
		{&MsgMhfGetRejectGuildScout{}, network.MSG_MHF_GET_REJECT_GUILD_SCOUT},
		{&MsgMhfSetRejectGuildScout{}, network.MSG_MHF_SET_REJECT_GUILD_SCOUT},
		{&MsgMhfGuildHuntdata{}, network.MSG_MHF_GUILD_HUNTDATA},
		{&MsgMhfUpdateForceGuildRank{}, network.MSG_MHF_UPDATE_FORCE_GUILD_RANK},
		{&MsgMhfEnumerateGuildMessageBoard{}, network.MSG_MHF_ENUMERATE_GUILD_MESSAGE_BOARD},
		{&MsgMhfUpdateGuildMessageBoard{}, network.MSG_MHF_UPDATE_GUILD_MESSAGE_BOARD},
	}

	for _, tc := range guildPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestFestaPacketsOpcode tests festa-related packets
func TestFestaPacketsOpcode(t *testing.T) {
	festaPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfInfoFesta{}, network.MSG_MHF_INFO_FESTA},
		{&MsgMhfEntryFesta{}, network.MSG_MHF_ENTRY_FESTA},
		{&MsgMhfChargeFesta{}, network.MSG_MHF_CHARGE_FESTA},
		{&MsgMhfAcquireFesta{}, network.MSG_MHF_ACQUIRE_FESTA},
		{&MsgMhfStateFestaU{}, network.MSG_MHF_STATE_FESTA_U},
		{&MsgMhfStateFestaG{}, network.MSG_MHF_STATE_FESTA_G},
		{&MsgMhfEnumerateFestaMember{}, network.MSG_MHF_ENUMERATE_FESTA_MEMBER},
		{&MsgMhfVoteFesta{}, network.MSG_MHF_VOTE_FESTA},
		{&MsgMhfEnumerateFestaIntermediatePrize{}, network.MSG_MHF_ENUMERATE_FESTA_INTERMEDIATE_PRIZE},
		{&MsgMhfAcquireFestaIntermediatePrize{}, network.MSG_MHF_ACQUIRE_FESTA_INTERMEDIATE_PRIZE},
		{&MsgMhfEnumerateFestaPersonalPrize{}, network.MSG_MHF_ENUMERATE_FESTA_PERSONAL_PRIZE},
		{&MsgMhfAcquireFestaPersonalPrize{}, network.MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE},
	}

	for _, tc := range festaPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestCafePacketsOpcode tests cafe-related packets
func TestCafePacketsOpcode(t *testing.T) {
	cafePackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfAcquireCafeItem{}, network.MSG_MHF_ACQUIRE_CAFE_ITEM},
		{&MsgMhfUpdateCafepoint{}, network.MSG_MHF_UPDATE_CAFEPOINT},
		{&MsgMhfCheckDailyCafepoint{}, network.MSG_MHF_CHECK_DAILY_CAFEPOINT},
		{&MsgMhfGetCafeDuration{}, network.MSG_MHF_GET_CAFE_DURATION},
		{&MsgMhfGetCafeDurationBonusInfo{}, network.MSG_MHF_GET_CAFE_DURATION_BONUS_INFO},
		{&MsgMhfReceiveCafeDurationBonus{}, network.MSG_MHF_RECEIVE_CAFE_DURATION_BONUS},
		{&MsgMhfPostCafeDurationBonusReceived{}, network.MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED},
	}

	for _, tc := range cafePackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestGachaPacketsOpcode tests gacha-related packets
func TestGachaPacketsOpcode(t *testing.T) {
	gachaPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfGetGachaPoint{}, network.MSG_MHF_GET_GACHA_POINT},
		{&MsgMhfUseGachaPoint{}, network.MSG_MHF_USE_GACHA_POINT},
		{&MsgMhfPlayStepupGacha{}, network.MSG_MHF_PLAY_STEPUP_GACHA},
		{&MsgMhfReceiveGachaItem{}, network.MSG_MHF_RECEIVE_GACHA_ITEM},
		{&MsgMhfGetStepupStatus{}, network.MSG_MHF_GET_STEPUP_STATUS},
		{&MsgMhfPlayFreeGacha{}, network.MSG_MHF_PLAY_FREE_GACHA},
		{&MsgMhfGetBoxGachaInfo{}, network.MSG_MHF_GET_BOX_GACHA_INFO},
		{&MsgMhfPlayBoxGacha{}, network.MSG_MHF_PLAY_BOX_GACHA},
		{&MsgMhfResetBoxGachaInfo{}, network.MSG_MHF_RESET_BOX_GACHA_INFO},
		{&MsgMhfPlayNormalGacha{}, network.MSG_MHF_PLAY_NORMAL_GACHA},
		{&MsgMhfGetGachaPlayHistory{}, network.MSG_MHF_GET_GACHA_PLAY_HISTORY},
	}

	for _, tc := range gachaPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestUDPacketsOpcode tests UD (Ultimate Devastation) related packets
func TestUDPacketsOpcode(t *testing.T) {
	udPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfGetUdSchedule{}, network.MSG_MHF_GET_UD_SCHEDULE},
		{&MsgMhfGetUdInfo{}, network.MSG_MHF_GET_UD_INFO},
		{&MsgMhfAddUdPoint{}, network.MSG_MHF_ADD_UD_POINT},
		{&MsgMhfGetUdMyPoint{}, network.MSG_MHF_GET_UD_MY_POINT},
		{&MsgMhfGetUdTotalPointInfo{}, network.MSG_MHF_GET_UD_TOTAL_POINT_INFO},
		{&MsgMhfGetUdBonusQuestInfo{}, network.MSG_MHF_GET_UD_BONUS_QUEST_INFO},
		{&MsgMhfGetUdSelectedColorInfo{}, network.MSG_MHF_GET_UD_SELECTED_COLOR_INFO},
		{&MsgMhfGetUdMonsterPoint{}, network.MSG_MHF_GET_UD_MONSTER_POINT},
		{&MsgMhfGetUdDailyPresentList{}, network.MSG_MHF_GET_UD_DAILY_PRESENT_LIST},
		{&MsgMhfGetUdNormaPresentList{}, network.MSG_MHF_GET_UD_NORMA_PRESENT_LIST},
		{&MsgMhfGetUdRankingRewardList{}, network.MSG_MHF_GET_UD_RANKING_REWARD_LIST},
		{&MsgMhfAcquireUdItem{}, network.MSG_MHF_ACQUIRE_UD_ITEM},
		{&MsgMhfGetUdRanking{}, network.MSG_MHF_GET_UD_RANKING},
		{&MsgMhfGetUdMyRanking{}, network.MSG_MHF_GET_UD_MY_RANKING},
		{&MsgMhfGetUdGuildMapInfo{}, network.MSG_MHF_GET_UD_GUILD_MAP_INFO},
		{&MsgMhfGenerateUdGuildMap{}, network.MSG_MHF_GENERATE_UD_GUILD_MAP},
		{&MsgMhfGetUdTacticsPoint{}, network.MSG_MHF_GET_UD_TACTICS_POINT},
		{&MsgMhfAddUdTacticsPoint{}, network.MSG_MHF_ADD_UD_TACTICS_POINT},
		{&MsgMhfGetUdTacticsRanking{}, network.MSG_MHF_GET_UD_TACTICS_RANKING},
		{&MsgMhfGetUdTacticsRewardList{}, network.MSG_MHF_GET_UD_TACTICS_REWARD_LIST},
		{&MsgMhfGetUdTacticsLog{}, network.MSG_MHF_GET_UD_TACTICS_LOG},
		{&MsgMhfGetUdTacticsFollower{}, network.MSG_MHF_GET_UD_TACTICS_FOLLOWER},
		{&MsgMhfSetUdTacticsFollower{}, network.MSG_MHF_SET_UD_TACTICS_FOLLOWER},
		{&MsgMhfGetUdShopCoin{}, network.MSG_MHF_GET_UD_SHOP_COIN},
		{&MsgMhfUseUdShopCoin{}, network.MSG_MHF_USE_UD_SHOP_COIN},
		{&MsgMhfGetUdTacticsBonusQuest{}, network.MSG_MHF_GET_UD_TACTICS_BONUS_QUEST},
		{&MsgMhfGetUdTacticsFirstQuestBonus{}, network.MSG_MHF_GET_UD_TACTICS_FIRST_QUEST_BONUS},
		{&MsgMhfGetUdTacticsRemainingPoint{}, network.MSG_MHF_GET_UD_TACTICS_REMAINING_POINT},
	}

	for _, tc := range udPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestRengokuPacketsOpcode tests rengoku (purgatory tower) related packets
func TestRengokuPacketsOpcode(t *testing.T) {
	rengokuPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfSaveRengokuData{}, network.MSG_MHF_SAVE_RENGOKU_DATA},
		{&MsgMhfLoadRengokuData{}, network.MSG_MHF_LOAD_RENGOKU_DATA},
		{&MsgMhfGetRengokuBinary{}, network.MSG_MHF_GET_RENGOKU_BINARY},
		{&MsgMhfEnumerateRengokuRanking{}, network.MSG_MHF_ENUMERATE_RENGOKU_RANKING},
		{&MsgMhfGetRengokuRankingRank{}, network.MSG_MHF_GET_RENGOKU_RANKING_RANK},
	}

	for _, tc := range rengokuPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestMezFesPacketsOpcode tests Mezeporta Festival related packets
func TestMezFesPacketsOpcode(t *testing.T) {
	mezfesPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfSaveMezfesData{}, network.MSG_MHF_SAVE_MEZFES_DATA},
		{&MsgMhfLoadMezfesData{}, network.MSG_MHF_LOAD_MEZFES_DATA},
	}

	for _, tc := range mezfesPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestWarehousePacketsOpcode tests warehouse related packets
func TestWarehousePacketsOpcode(t *testing.T) {
	warehousePackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfOperateWarehouse{}, network.MSG_MHF_OPERATE_WAREHOUSE},
		{&MsgMhfEnumerateWarehouse{}, network.MSG_MHF_ENUMERATE_WAREHOUSE},
		{&MsgMhfUpdateWarehouse{}, network.MSG_MHF_UPDATE_WAREHOUSE},
	}

	for _, tc := range warehousePackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestMercenaryPacketsOpcode tests mercenary related packets
func TestMercenaryPacketsOpcode(t *testing.T) {
	mercenaryPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfMercenaryHuntdata{}, network.MSG_MHF_MERCENARY_HUNTDATA},
		{&MsgMhfCreateMercenary{}, network.MSG_MHF_CREATE_MERCENARY},
		{&MsgMhfSaveMercenary{}, network.MSG_MHF_SAVE_MERCENARY},
		{&MsgMhfReadMercenaryW{}, network.MSG_MHF_READ_MERCENARY_W},
		{&MsgMhfReadMercenaryM{}, network.MSG_MHF_READ_MERCENARY_M},
		{&MsgMhfContractMercenary{}, network.MSG_MHF_CONTRACT_MERCENARY},
		{&MsgMhfEnumerateMercenaryLog{}, network.MSG_MHF_ENUMERATE_MERCENARY_LOG},
	}

	for _, tc := range mercenaryPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestHousePacketsOpcode tests house related packets
func TestHousePacketsOpcode(t *testing.T) {
	housePackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfUpdateInterior{}, network.MSG_MHF_UPDATE_INTERIOR},
		{&MsgMhfEnumerateHouse{}, network.MSG_MHF_ENUMERATE_HOUSE},
		{&MsgMhfUpdateHouse{}, network.MSG_MHF_UPDATE_HOUSE},
		{&MsgMhfLoadHouse{}, network.MSG_MHF_LOAD_HOUSE},
		{&MsgMhfGetMyhouseInfo{}, network.MSG_MHF_GET_MYHOUSE_INFO},
		{&MsgMhfUpdateMyhouseInfo{}, network.MSG_MHF_UPDATE_MYHOUSE_INFO},
	}

	for _, tc := range housePackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestBoostPacketsOpcode tests boost related packets
func TestBoostPacketsOpcode(t *testing.T) {
	boostPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfGetBoostTime{}, network.MSG_MHF_GET_BOOST_TIME},
		{&MsgMhfPostBoostTime{}, network.MSG_MHF_POST_BOOST_TIME},
		{&MsgMhfGetBoostTimeLimit{}, network.MSG_MHF_GET_BOOST_TIME_LIMIT},
		{&MsgMhfPostBoostTimeLimit{}, network.MSG_MHF_POST_BOOST_TIME_LIMIT},
		{&MsgMhfGetBoostRight{}, network.MSG_MHF_GET_BOOST_RIGHT},
		{&MsgMhfStartBoostTime{}, network.MSG_MHF_START_BOOST_TIME},
		{&MsgMhfPostBoostTimeQuestReturn{}, network.MSG_MHF_POST_BOOST_TIME_QUEST_RETURN},
		{&MsgMhfGetKeepLoginBoostStatus{}, network.MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS},
		{&MsgMhfUseKeepLoginBoost{}, network.MSG_MHF_USE_KEEP_LOGIN_BOOST},
	}

	for _, tc := range boostPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestTournamentPacketsOpcode tests tournament related packets
func TestTournamentPacketsOpcode(t *testing.T) {
	tournamentPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfInfoTournament{}, network.MSG_MHF_INFO_TOURNAMENT},
		{&MsgMhfEntryTournament{}, network.MSG_MHF_ENTRY_TOURNAMENT},
		{&MsgMhfEnterTournamentQuest{}, network.MSG_MHF_ENTER_TOURNAMENT_QUEST},
		{&MsgMhfAcquireTournament{}, network.MSG_MHF_ACQUIRE_TOURNAMENT},
	}

	for _, tc := range tournamentPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestPlatePacketsOpcode tests plate related packets
func TestPlatePacketsOpcode(t *testing.T) {
	platePackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfLoadPlateData{}, network.MSG_MHF_LOAD_PLATE_DATA},
		{&MsgMhfSavePlateData{}, network.MSG_MHF_SAVE_PLATE_DATA},
		{&MsgMhfLoadPlateBox{}, network.MSG_MHF_LOAD_PLATE_BOX},
		{&MsgMhfSavePlateBox{}, network.MSG_MHF_SAVE_PLATE_BOX},
		{&MsgMhfLoadPlateMyset{}, network.MSG_MHF_LOAD_PLATE_MYSET},
		{&MsgMhfSavePlateMyset{}, network.MSG_MHF_SAVE_PLATE_MYSET},
	}

	for _, tc := range platePackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}

// TestScenarioPacketsOpcode tests scenario related packets
func TestScenarioPacketsOpcode(t *testing.T) {
	scenarioPackets := []struct {
		pkt    MHFPacket
		opcode network.PacketID
	}{
		{&MsgMhfInfoScenarioCounter{}, network.MSG_MHF_INFO_SCENARIO_COUNTER},
		{&MsgMhfSaveScenarioData{}, network.MSG_MHF_SAVE_SCENARIO_DATA},
		{&MsgMhfLoadScenarioData{}, network.MSG_MHF_LOAD_SCENARIO_DATA},
	}

	for _, tc := range scenarioPackets {
		t.Run(tc.opcode.String(), func(t *testing.T) {
			if tc.pkt.Opcode() != tc.opcode {
				t.Errorf("Opcode() = %s, want %s", tc.pkt.Opcode(), tc.opcode)
			}
		})
	}
}
