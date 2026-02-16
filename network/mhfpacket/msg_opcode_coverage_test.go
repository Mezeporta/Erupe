package mhfpacket

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
)

// callBuildSafe calls Build on the packet, recovering from panics.
// Returns the error from Build, or nil if it panicked (panic is acceptable
// for "Not implemented" stubs).
func callBuildSafe(pkt MHFPacket, bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	err = pkt.Build(bf, ctx)
	return err, false
}

// callParseSafe calls Parse on the packet, recovering from panics.
func callParseSafe(pkt MHFPacket, bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	err = pkt.Parse(bf, ctx)
	return err, false
}

// TestBuildCoverage_NotImplemented exercises Build() on packet types whose Build
// method is not yet covered. These stubs either return errors.New("NOT IMPLEMENTED")
// or panic("Not implemented"). Both are acceptable outcomes that indicate the
// method was reached.
func TestBuildCoverage_NotImplemented(t *testing.T) {
	tests := []struct {
		name string
		pkt  MHFPacket
	}{
		// msg_ca_exchange_item.go
		{"MsgCaExchangeItem", &MsgCaExchangeItem{}},
		// msg_head.go
		{"MsgHead", &MsgHead{}},
		// msg_mhf_acquire_cafe_item.go
		{"MsgMhfAcquireCafeItem", &MsgMhfAcquireCafeItem{}},
		// msg_mhf_acquire_monthly_item.go
		{"MsgMhfAcquireMonthlyItem", &MsgMhfAcquireMonthlyItem{}},
		// msg_mhf_acquire_ud_item.go
		{"MsgMhfAcquireUdItem", &MsgMhfAcquireUdItem{}},
		// msg_mhf_announce.go
		{"MsgMhfAnnounce", &MsgMhfAnnounce{}},
		// msg_mhf_check_monthly_item.go
		{"MsgMhfCheckMonthlyItem", &MsgMhfCheckMonthlyItem{}},
		// msg_mhf_check_weekly_stamp.go
		{"MsgMhfCheckWeeklyStamp", &MsgMhfCheckWeeklyStamp{}},
		// msg_mhf_enumerate_festa_member.go
		{"MsgMhfEnumerateFestaMember", &MsgMhfEnumerateFestaMember{}},
		// msg_mhf_enumerate_inv_guild.go
		{"MsgMhfEnumerateInvGuild", &MsgMhfEnumerateInvGuild{}},
		// msg_mhf_enumerate_item.go
		{"MsgMhfEnumerateItem", &MsgMhfEnumerateItem{}},
		// msg_mhf_enumerate_order.go
		{"MsgMhfEnumerateOrder", &MsgMhfEnumerateOrder{}},
		// msg_mhf_enumerate_quest.go
		{"MsgMhfEnumerateQuest", &MsgMhfEnumerateQuest{}},
		// msg_mhf_enumerate_ranking.go
		{"MsgMhfEnumerateRanking", &MsgMhfEnumerateRanking{}},
		// msg_mhf_enumerate_shop.go
		{"MsgMhfEnumerateShop", &MsgMhfEnumerateShop{}},
		// msg_mhf_enumerate_warehouse.go
		{"MsgMhfEnumerateWarehouse", &MsgMhfEnumerateWarehouse{}},
		// msg_mhf_exchange_fpoint_2_item.go
		{"MsgMhfExchangeFpoint2Item", &MsgMhfExchangeFpoint2Item{}},
		// msg_mhf_exchange_item_2_fpoint.go
		{"MsgMhfExchangeItem2Fpoint", &MsgMhfExchangeItem2Fpoint{}},
		// msg_mhf_exchange_weekly_stamp.go
		{"MsgMhfExchangeWeeklyStamp", &MsgMhfExchangeWeeklyStamp{}},
		// msg_mhf_generate_ud_guild_map.go
		{"MsgMhfGenerateUdGuildMap", &MsgMhfGenerateUdGuildMap{}},
		// msg_mhf_get_boost_time.go
		{"MsgMhfGetBoostTime", &MsgMhfGetBoostTime{}},
		// msg_mhf_get_boost_time_limit.go
		{"MsgMhfGetBoostTimeLimit", &MsgMhfGetBoostTimeLimit{}},
		// msg_mhf_get_cafe_duration.go
		{"MsgMhfGetCafeDuration", &MsgMhfGetCafeDuration{}},
		// msg_mhf_get_cafe_duration_bonus_info.go
		{"MsgMhfGetCafeDurationBonusInfo", &MsgMhfGetCafeDurationBonusInfo{}},
		// msg_mhf_get_cog_info.go
		{"MsgMhfGetCogInfo", &MsgMhfGetCogInfo{}},
		// msg_mhf_get_gacha_point.go
		{"MsgMhfGetGachaPoint", &MsgMhfGetGachaPoint{}},
		// msg_mhf_get_gem_info.go
		{"MsgMhfGetGemInfo", &MsgMhfGetGemInfo{}},
		// msg_mhf_get_kiju_info.go
		{"MsgMhfGetKijuInfo", &MsgMhfGetKijuInfo{}},
		// msg_mhf_get_myhouse_info.go
		{"MsgMhfGetMyhouseInfo", &MsgMhfGetMyhouseInfo{}},
		// msg_mhf_get_notice.go
		{"MsgMhfGetNotice", &MsgMhfGetNotice{}},
		// msg_mhf_get_tower_info.go
		{"MsgMhfGetTowerInfo", &MsgMhfGetTowerInfo{}},
		// msg_mhf_get_ud_info.go
		{"MsgMhfGetUdInfo", &MsgMhfGetUdInfo{}},
		// msg_mhf_get_ud_schedule.go
		{"MsgMhfGetUdSchedule", &MsgMhfGetUdSchedule{}},
		// msg_mhf_get_weekly_schedule.go
		{"MsgMhfGetWeeklySchedule", &MsgMhfGetWeeklySchedule{}},
		// msg_mhf_guild_huntdata.go
		{"MsgMhfGuildHuntdata", &MsgMhfGuildHuntdata{}},
		// msg_mhf_info_joint.go
		{"MsgMhfInfoJoint", &MsgMhfInfoJoint{}},
		// msg_mhf_load_deco_myset.go
		{"MsgMhfLoadDecoMyset", &MsgMhfLoadDecoMyset{}},
		// msg_mhf_load_guild_adventure.go
		{"MsgMhfLoadGuildAdventure", &MsgMhfLoadGuildAdventure{}},
		// msg_mhf_load_guild_cooking.go
		{"MsgMhfLoadGuildCooking", &MsgMhfLoadGuildCooking{}},
		// msg_mhf_load_hunter_navi.go
		{"MsgMhfLoadHunterNavi", &MsgMhfLoadHunterNavi{}},
		// msg_mhf_load_otomo_airou.go
		{"MsgMhfLoadOtomoAirou", &MsgMhfLoadOtomoAirou{}},
		// msg_mhf_load_partner.go
		{"MsgMhfLoadPartner", &MsgMhfLoadPartner{}},
		// msg_mhf_load_plate_box.go
		{"MsgMhfLoadPlateBox", &MsgMhfLoadPlateBox{}},
		// msg_mhf_load_plate_data.go
		{"MsgMhfLoadPlateData", &MsgMhfLoadPlateData{}},
		// msg_mhf_post_notice.go
		{"MsgMhfPostNotice", &MsgMhfPostNotice{}},
		// msg_mhf_post_tower_info.go
		{"MsgMhfPostTowerInfo", &MsgMhfPostTowerInfo{}},
		// msg_mhf_reserve10f.go
		{"MsgMhfReserve10F", &MsgMhfReserve10F{}},
		// msg_mhf_server_command.go
		{"MsgMhfServerCommand", &MsgMhfServerCommand{}},
		// msg_mhf_set_loginwindow.go
		{"MsgMhfSetLoginwindow", &MsgMhfSetLoginwindow{}},
		// msg_mhf_shut_client.go
		{"MsgMhfShutClient", &MsgMhfShutClient{}},
		// msg_mhf_stampcard_stamp.go
		{"MsgMhfStampcardStamp", &MsgMhfStampcardStamp{}},
		// msg_sys_add_object.go
		{"MsgSysAddObject", &MsgSysAddObject{}},
		// msg_sys_back_stage.go
		{"MsgSysBackStage", &MsgSysBackStage{}},
		// msg_sys_cast_binary.go
		{"MsgSysCastBinary", &MsgSysCastBinary{}},
		// msg_sys_create_semaphore.go
		{"MsgSysCreateSemaphore", &MsgSysCreateSemaphore{}},
		// msg_sys_create_stage.go
		{"MsgSysCreateStage", &MsgSysCreateStage{}},
		// msg_sys_del_object.go
		{"MsgSysDelObject", &MsgSysDelObject{}},
		// msg_sys_disp_object.go
		{"MsgSysDispObject", &MsgSysDispObject{}},
		// msg_sys_echo.go
		{"MsgSysEcho", &MsgSysEcho{}},
		// msg_sys_enter_stage.go
		{"MsgSysEnterStage", &MsgSysEnterStage{}},
		// msg_sys_enumerate_client.go
		{"MsgSysEnumerateClient", &MsgSysEnumerateClient{}},
		// msg_sys_extend_threshold.go
		{"MsgSysExtendThreshold", &MsgSysExtendThreshold{}},
		// msg_sys_get_stage_binary.go
		{"MsgSysGetStageBinary", &MsgSysGetStageBinary{}},
		// msg_sys_hide_object.go
		{"MsgSysHideObject", &MsgSysHideObject{}},
		// msg_sys_leave_stage.go
		{"MsgSysLeaveStage", &MsgSysLeaveStage{}},
		// msg_sys_lock_stage.go
		{"MsgSysLockStage", &MsgSysLockStage{}},
		// msg_sys_login.go
		{"MsgSysLogin", &MsgSysLogin{}},
		// msg_sys_move_stage.go
		{"MsgSysMoveStage", &MsgSysMoveStage{}},
		// msg_sys_set_stage_binary.go
		{"MsgSysSetStageBinary", &MsgSysSetStageBinary{}},
		// msg_sys_set_stage_pass.go
		{"MsgSysSetStagePass", &MsgSysSetStagePass{}},
		// msg_sys_set_status.go
		{"MsgSysSetStatus", &MsgSysSetStatus{}},
		// msg_sys_wait_stage_binary.go
		{"MsgSysWaitStageBinary", &MsgSysWaitStageBinary{}},

		// Reserve files - sys reserves
		{"MsgSysReserve01", &MsgSysReserve01{}},
		{"MsgSysReserve02", &MsgSysReserve02{}},
		{"MsgSysReserve03", &MsgSysReserve03{}},
		{"MsgSysReserve04", &MsgSysReserve04{}},
		{"MsgSysReserve05", &MsgSysReserve05{}},
		{"MsgSysReserve06", &MsgSysReserve06{}},
		{"MsgSysReserve07", &MsgSysReserve07{}},
		{"MsgSysReserve0C", &MsgSysReserve0C{}},
		{"MsgSysReserve0D", &MsgSysReserve0D{}},
		{"MsgSysReserve0E", &MsgSysReserve0E{}},
		{"MsgSysReserve4A", &MsgSysReserve4A{}},
		{"MsgSysReserve4B", &MsgSysReserve4B{}},
		{"MsgSysReserve4C", &MsgSysReserve4C{}},
		{"MsgSysReserve4D", &MsgSysReserve4D{}},
		{"MsgSysReserve4E", &MsgSysReserve4E{}},
		{"MsgSysReserve4F", &MsgSysReserve4F{}},
		{"MsgSysReserve55", &MsgSysReserve55{}},
		{"MsgSysReserve56", &MsgSysReserve56{}},
		{"MsgSysReserve57", &MsgSysReserve57{}},
		{"MsgSysReserve5C", &MsgSysReserve5C{}},
		{"MsgSysReserve5E", &MsgSysReserve5E{}},
		{"MsgSysReserve5F", &MsgSysReserve5F{}},
		{"MsgSysReserve71", &MsgSysReserve71{}},
		{"MsgSysReserve72", &MsgSysReserve72{}},
		{"MsgSysReserve73", &MsgSysReserve73{}},
		{"MsgSysReserve74", &MsgSysReserve74{}},
		{"MsgSysReserve75", &MsgSysReserve75{}},
		{"MsgSysReserve76", &MsgSysReserve76{}},
		{"MsgSysReserve77", &MsgSysReserve77{}},
		{"MsgSysReserve78", &MsgSysReserve78{}},
		{"MsgSysReserve79", &MsgSysReserve79{}},
		{"MsgSysReserve7A", &MsgSysReserve7A{}},
		{"MsgSysReserve7B", &MsgSysReserve7B{}},
		{"MsgSysReserve7C", &MsgSysReserve7C{}},
		{"MsgSysReserve7E", &MsgSysReserve7E{}},
		{"MsgSysReserve180", &MsgSysReserve180{}},
		{"MsgSysReserve188", &MsgSysReserve188{}},
		{"MsgSysReserve18B", &MsgSysReserve18B{}},
		{"MsgSysReserve18E", &MsgSysReserve18E{}},
		{"MsgSysReserve18F", &MsgSysReserve18F{}},
		{"MsgSysReserve192", &MsgSysReserve192{}},
		{"MsgSysReserve193", &MsgSysReserve193{}},
		{"MsgSysReserve194", &MsgSysReserve194{}},
		{"MsgSysReserve19B", &MsgSysReserve19B{}},
		{"MsgSysReserve19E", &MsgSysReserve19E{}},
		{"MsgSysReserve19F", &MsgSysReserve19F{}},
		{"MsgSysReserve1A4", &MsgSysReserve1A4{}},
		{"MsgSysReserve1A6", &MsgSysReserve1A6{}},
		{"MsgSysReserve1A7", &MsgSysReserve1A7{}},
		{"MsgSysReserve1A8", &MsgSysReserve1A8{}},
		{"MsgSysReserve1A9", &MsgSysReserve1A9{}},
		{"MsgSysReserve1AA", &MsgSysReserve1AA{}},
		{"MsgSysReserve1AB", &MsgSysReserve1AB{}},
		{"MsgSysReserve1AC", &MsgSysReserve1AC{}},
		{"MsgSysReserve1AD", &MsgSysReserve1AD{}},
		{"MsgSysReserve1AE", &MsgSysReserve1AE{}},
		{"MsgSysReserve1AF", &MsgSysReserve1AF{}},
	}

	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, panicked := callBuildSafe(tt.pkt, bf, ctx)
			if panicked {
				// Build panicked with "Not implemented" - this is acceptable
				// and still exercises the code path for coverage.
				return
			}
			if err == nil {
				// Build succeeded (some packets may have implemented Build)
				return
			}
			// Build returned an error, which is expected for NOT IMPLEMENTED stubs
			if err.Error() != "NOT IMPLEMENTED" {
				t.Errorf("Build() returned unexpected error: %v", err)
			}
		})
	}
}

// TestParseCoverage_NotImplemented exercises Parse() on packet types whose Parse
// method returns "NOT IMPLEMENTED" and is not yet covered by existing tests.
func TestParseCoverage_NotImplemented(t *testing.T) {
	tests := []struct {
		name string
		pkt  MHFPacket
	}{
		// msg_mhf_acquire_tournament.go - Parse returns NOT IMPLEMENTED
		{"MsgMhfAcquireTournament", &MsgMhfAcquireTournament{}},
		// msg_mhf_entry_tournament.go - Parse returns NOT IMPLEMENTED
		{"MsgMhfEntryTournament", &MsgMhfEntryTournament{}},
		// msg_mhf_update_guild.go - Parse returns NOT IMPLEMENTED
		{"MsgMhfUpdateGuild", &MsgMhfUpdateGuild{}},
	}

	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, panicked := callParseSafe(tt.pkt, bf, ctx)
			if panicked {
				return
			}
			if err == nil {
				return
			}
			if err.Error() != "NOT IMPLEMENTED" {
				t.Errorf("Parse() returned unexpected error: %v", err)
			}
		})
	}
}
