package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// TestMsgMhfSavedataParse tests parsing MsgMhfSavedata
func TestMsgMhfSavedataParse(t *testing.T) {
	pkt := FromOpcode(network.MSG_MHF_SAVEDATA)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_MHF_SAVEDATA) returned nil")
	}
	if pkt.Opcode() != network.MSG_MHF_SAVEDATA {
		t.Errorf("Opcode() = %s, want MSG_MHF_SAVEDATA", pkt.Opcode())
	}
}

// TestMsgMhfLoaddataParse tests parsing MsgMhfLoaddata
func TestMsgMhfLoaddataParse(t *testing.T) {
	pkt := FromOpcode(network.MSG_MHF_LOADDATA)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_MHF_LOADDATA) returned nil")
	}
	if pkt.Opcode() != network.MSG_MHF_LOADDATA {
		t.Errorf("Opcode() = %s, want MSG_MHF_LOADDATA", pkt.Opcode())
	}
}

// TestMsgMhfListMemberOpcode tests MsgMhfListMember Opcode
func TestMsgMhfListMemberOpcode(t *testing.T) {
	pkt := &MsgMhfListMember{}
	if pkt.Opcode() != network.MSG_MHF_LIST_MEMBER {
		t.Errorf("Opcode() = %s, want MSG_MHF_LIST_MEMBER", pkt.Opcode())
	}
}

// TestMsgMhfOprMemberOpcode tests MsgMhfOprMember Opcode
func TestMsgMhfOprMemberOpcode(t *testing.T) {
	pkt := &MsgMhfOprMember{}
	if pkt.Opcode() != network.MSG_MHF_OPR_MEMBER {
		t.Errorf("Opcode() = %s, want MSG_MHF_OPR_MEMBER", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateDistItemOpcode tests MsgMhfEnumerateDistItem Opcode
func TestMsgMhfEnumerateDistItemOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateDistItem{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_DIST_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_DIST_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfApplyDistItemOpcode tests MsgMhfApplyDistItem Opcode
func TestMsgMhfApplyDistItemOpcode(t *testing.T) {
	pkt := &MsgMhfApplyDistItem{}
	if pkt.Opcode() != network.MSG_MHF_APPLY_DIST_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_APPLY_DIST_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfAcquireDistItemOpcode tests MsgMhfAcquireDistItem Opcode
func TestMsgMhfAcquireDistItemOpcode(t *testing.T) {
	pkt := &MsgMhfAcquireDistItem{}
	if pkt.Opcode() != network.MSG_MHF_ACQUIRE_DIST_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ACQUIRE_DIST_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfGetDistDescriptionOpcode tests MsgMhfGetDistDescription Opcode
func TestMsgMhfGetDistDescriptionOpcode(t *testing.T) {
	pkt := &MsgMhfGetDistDescription{}
	if pkt.Opcode() != network.MSG_MHF_GET_DIST_DESCRIPTION {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_DIST_DESCRIPTION", pkt.Opcode())
	}
}

// TestMsgMhfSendMailOpcode tests MsgMhfSendMail Opcode
func TestMsgMhfSendMailOpcode(t *testing.T) {
	pkt := &MsgMhfSendMail{}
	if pkt.Opcode() != network.MSG_MHF_SEND_MAIL {
		t.Errorf("Opcode() = %s, want MSG_MHF_SEND_MAIL", pkt.Opcode())
	}
}

// TestMsgMhfReadMailOpcode tests MsgMhfReadMail Opcode
func TestMsgMhfReadMailOpcode(t *testing.T) {
	pkt := &MsgMhfReadMail{}
	if pkt.Opcode() != network.MSG_MHF_READ_MAIL {
		t.Errorf("Opcode() = %s, want MSG_MHF_READ_MAIL", pkt.Opcode())
	}
}

// TestMsgMhfListMailOpcode tests MsgMhfListMail Opcode
func TestMsgMhfListMailOpcode(t *testing.T) {
	pkt := &MsgMhfListMail{}
	if pkt.Opcode() != network.MSG_MHF_LIST_MAIL {
		t.Errorf("Opcode() = %s, want MSG_MHF_LIST_MAIL", pkt.Opcode())
	}
}

// TestMsgMhfOprtMailOpcode tests MsgMhfOprtMail Opcode
func TestMsgMhfOprtMailOpcode(t *testing.T) {
	pkt := &MsgMhfOprtMail{}
	if pkt.Opcode() != network.MSG_MHF_OPRT_MAIL {
		t.Errorf("Opcode() = %s, want MSG_MHF_OPRT_MAIL", pkt.Opcode())
	}
}

// TestMsgMhfLoadFavoriteQuestOpcode tests MsgMhfLoadFavoriteQuest Opcode
func TestMsgMhfLoadFavoriteQuestOpcode(t *testing.T) {
	pkt := &MsgMhfLoadFavoriteQuest{}
	if pkt.Opcode() != network.MSG_MHF_LOAD_FAVORITE_QUEST {
		t.Errorf("Opcode() = %s, want MSG_MHF_LOAD_FAVORITE_QUEST", pkt.Opcode())
	}
}

// TestMsgMhfSaveFavoriteQuestOpcode tests MsgMhfSaveFavoriteQuest Opcode
func TestMsgMhfSaveFavoriteQuestOpcode(t *testing.T) {
	pkt := &MsgMhfSaveFavoriteQuest{}
	if pkt.Opcode() != network.MSG_MHF_SAVE_FAVORITE_QUEST {
		t.Errorf("Opcode() = %s, want MSG_MHF_SAVE_FAVORITE_QUEST", pkt.Opcode())
	}
}

// TestMsgMhfRegisterEventOpcode tests MsgMhfRegisterEvent Opcode
func TestMsgMhfRegisterEventOpcode(t *testing.T) {
	pkt := &MsgMhfRegisterEvent{}
	if pkt.Opcode() != network.MSG_MHF_REGISTER_EVENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_REGISTER_EVENT", pkt.Opcode())
	}
}

// TestMsgMhfReleaseEventOpcode tests MsgMhfReleaseEvent Opcode
func TestMsgMhfReleaseEventOpcode(t *testing.T) {
	pkt := &MsgMhfReleaseEvent{}
	if pkt.Opcode() != network.MSG_MHF_RELEASE_EVENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_RELEASE_EVENT", pkt.Opcode())
	}
}

// TestMsgMhfTransitMessageOpcode tests MsgMhfTransitMessage Opcode
func TestMsgMhfTransitMessageOpcode(t *testing.T) {
	pkt := &MsgMhfTransitMessage{}
	if pkt.Opcode() != network.MSG_MHF_TRANSIT_MESSAGE {
		t.Errorf("Opcode() = %s, want MSG_MHF_TRANSIT_MESSAGE", pkt.Opcode())
	}
}

// TestMsgMhfPresentBoxOpcode tests MsgMhfPresentBox Opcode
func TestMsgMhfPresentBoxOpcode(t *testing.T) {
	pkt := &MsgMhfPresentBox{}
	if pkt.Opcode() != network.MSG_MHF_PRESENT_BOX {
		t.Errorf("Opcode() = %s, want MSG_MHF_PRESENT_BOX", pkt.Opcode())
	}
}

// TestMsgMhfServerCommandOpcode tests MsgMhfServerCommand Opcode
func TestMsgMhfServerCommandOpcode(t *testing.T) {
	pkt := &MsgMhfServerCommand{}
	if pkt.Opcode() != network.MSG_MHF_SERVER_COMMAND {
		t.Errorf("Opcode() = %s, want MSG_MHF_SERVER_COMMAND", pkt.Opcode())
	}
}

// TestMsgMhfShutClientOpcode tests MsgMhfShutClient Opcode
func TestMsgMhfShutClientOpcode(t *testing.T) {
	pkt := &MsgMhfShutClient{}
	if pkt.Opcode() != network.MSG_MHF_SHUT_CLIENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_SHUT_CLIENT", pkt.Opcode())
	}
}

// TestMsgMhfAnnounceOpcode tests MsgMhfAnnounce Opcode
func TestMsgMhfAnnounceOpcode(t *testing.T) {
	pkt := &MsgMhfAnnounce{}
	if pkt.Opcode() != network.MSG_MHF_ANNOUNCE {
		t.Errorf("Opcode() = %s, want MSG_MHF_ANNOUNCE", pkt.Opcode())
	}
}

// TestMsgMhfSetLoginwindowOpcode tests MsgMhfSetLoginwindow Opcode
func TestMsgMhfSetLoginwindowOpcode(t *testing.T) {
	pkt := &MsgMhfSetLoginwindow{}
	if pkt.Opcode() != network.MSG_MHF_SET_LOGINWINDOW {
		t.Errorf("Opcode() = %s, want MSG_MHF_SET_LOGINWINDOW", pkt.Opcode())
	}
}

// TestMsgMhfGetCaUniqueIDOpcode tests MsgMhfGetCaUniqueID Opcode
func TestMsgMhfGetCaUniqueIDOpcode(t *testing.T) {
	pkt := &MsgMhfGetCaUniqueID{}
	if pkt.Opcode() != network.MSG_MHF_GET_CA_UNIQUE_ID {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_CA_UNIQUE_ID", pkt.Opcode())
	}
}

// TestMsgMhfSetCaAchievementOpcode tests MsgMhfSetCaAchievement Opcode
func TestMsgMhfSetCaAchievementOpcode(t *testing.T) {
	pkt := &MsgMhfSetCaAchievement{}
	if pkt.Opcode() != network.MSG_MHF_SET_CA_ACHIEVEMENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_SET_CA_ACHIEVEMENT", pkt.Opcode())
	}
}

// TestMsgMhfCaravanMyScoreOpcode tests MsgMhfCaravanMyScore Opcode
func TestMsgMhfCaravanMyScoreOpcode(t *testing.T) {
	pkt := &MsgMhfCaravanMyScore{}
	if pkt.Opcode() != network.MSG_MHF_CARAVAN_MY_SCORE {
		t.Errorf("Opcode() = %s, want MSG_MHF_CARAVAN_MY_SCORE", pkt.Opcode())
	}
}

// TestMsgMhfCaravanRankingOpcode tests MsgMhfCaravanRanking Opcode
func TestMsgMhfCaravanRankingOpcode(t *testing.T) {
	pkt := &MsgMhfCaravanRanking{}
	if pkt.Opcode() != network.MSG_MHF_CARAVAN_RANKING {
		t.Errorf("Opcode() = %s, want MSG_MHF_CARAVAN_RANKING", pkt.Opcode())
	}
}

// TestMsgMhfCaravanMyRankOpcode tests MsgMhfCaravanMyRank Opcode
func TestMsgMhfCaravanMyRankOpcode(t *testing.T) {
	pkt := &MsgMhfCaravanMyRank{}
	if pkt.Opcode() != network.MSG_MHF_CARAVAN_MY_RANK {
		t.Errorf("Opcode() = %s, want MSG_MHF_CARAVAN_MY_RANK", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateQuestOpcode tests MsgMhfEnumerateQuest Opcode
func TestMsgMhfEnumerateQuestOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateQuest{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_QUEST {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_QUEST", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateEventOpcode tests MsgMhfEnumerateEvent Opcode
func TestMsgMhfEnumerateEventOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateEvent{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_EVENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_EVENT", pkt.Opcode())
	}
}

// TestMsgMhfEnumeratePriceOpcode tests MsgMhfEnumeratePrice Opcode
func TestMsgMhfEnumeratePriceOpcode(t *testing.T) {
	pkt := &MsgMhfEnumeratePrice{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_PRICE {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_PRICE", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateRankingOpcode tests MsgMhfEnumerateRanking Opcode
func TestMsgMhfEnumerateRankingOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateRanking{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_RANKING {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_RANKING", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateOrderOpcode tests MsgMhfEnumerateOrder Opcode
func TestMsgMhfEnumerateOrderOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateOrder{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_ORDER {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_ORDER", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateShopOpcode tests MsgMhfEnumerateShop Opcode
func TestMsgMhfEnumerateShopOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateShop{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_SHOP {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_SHOP", pkt.Opcode())
	}
}

// TestMsgMhfGetExtraInfoOpcode tests MsgMhfGetExtraInfo Opcode
func TestMsgMhfGetExtraInfoOpcode(t *testing.T) {
	pkt := &MsgMhfGetExtraInfo{}
	if pkt.Opcode() != network.MSG_MHF_GET_EXTRA_INFO {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_EXTRA_INFO", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateItemOpcode tests MsgMhfEnumerateItem Opcode
func TestMsgMhfEnumerateItemOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateItem{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfAcquireItemOpcode tests MsgMhfAcquireItem Opcode
func TestMsgMhfAcquireItemOpcode(t *testing.T) {
	pkt := &MsgMhfAcquireItem{}
	if pkt.Opcode() != network.MSG_MHF_ACQUIRE_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ACQUIRE_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfTransferItemOpcode tests MsgMhfTransferItem Opcode
func TestMsgMhfTransferItemOpcode(t *testing.T) {
	pkt := &MsgMhfTransferItem{}
	if pkt.Opcode() != network.MSG_MHF_TRANSFER_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_TRANSFER_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfEntryRookieGuildOpcode tests MsgMhfEntryRookieGuild Opcode
func TestMsgMhfEntryRookieGuildOpcode(t *testing.T) {
	pkt := &MsgMhfEntryRookieGuild{}
	if pkt.Opcode() != network.MSG_MHF_ENTRY_ROOKIE_GUILD {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENTRY_ROOKIE_GUILD", pkt.Opcode())
	}
}

// TestMsgCaExchangeItemOpcode tests MsgCaExchangeItem Opcode
func TestMsgCaExchangeItemOpcode(t *testing.T) {
	pkt := &MsgCaExchangeItem{}
	if pkt.Opcode() != network.MSG_CA_EXCHANGE_ITEM {
		t.Errorf("Opcode() = %s, want MSG_CA_EXCHANGE_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateCampaignOpcode tests MsgMhfEnumerateCampaign Opcode
func TestMsgMhfEnumerateCampaignOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateCampaign{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_CAMPAIGN {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_CAMPAIGN", pkt.Opcode())
	}
}

// TestMsgMhfStateCampaignOpcode tests MsgMhfStateCampaign Opcode
func TestMsgMhfStateCampaignOpcode(t *testing.T) {
	pkt := &MsgMhfStateCampaign{}
	if pkt.Opcode() != network.MSG_MHF_STATE_CAMPAIGN {
		t.Errorf("Opcode() = %s, want MSG_MHF_STATE_CAMPAIGN", pkt.Opcode())
	}
}

// TestMsgMhfApplyCampaignOpcode tests MsgMhfApplyCampaign Opcode
func TestMsgMhfApplyCampaignOpcode(t *testing.T) {
	pkt := &MsgMhfApplyCampaign{}
	if pkt.Opcode() != network.MSG_MHF_APPLY_CAMPAIGN {
		t.Errorf("Opcode() = %s, want MSG_MHF_APPLY_CAMPAIGN", pkt.Opcode())
	}
}

// TestMsgMhfCreateJointOpcode tests MsgMhfCreateJoint Opcode
func TestMsgMhfCreateJointOpcode(t *testing.T) {
	pkt := &MsgMhfCreateJoint{}
	if pkt.Opcode() != network.MSG_MHF_CREATE_JOINT {
		t.Errorf("Opcode() = %s, want MSG_MHF_CREATE_JOINT", pkt.Opcode())
	}
}

// TestMsgMhfOperateJointOpcode tests MsgMhfOperateJoint Opcode
func TestMsgMhfOperateJointOpcode(t *testing.T) {
	pkt := &MsgMhfOperateJoint{}
	if pkt.Opcode() != network.MSG_MHF_OPERATE_JOINT {
		t.Errorf("Opcode() = %s, want MSG_MHF_OPERATE_JOINT", pkt.Opcode())
	}
}

// TestMsgMhfInfoJointOpcode tests MsgMhfInfoJoint Opcode
func TestMsgMhfInfoJointOpcode(t *testing.T) {
	pkt := &MsgMhfInfoJoint{}
	if pkt.Opcode() != network.MSG_MHF_INFO_JOINT {
		t.Errorf("Opcode() = %s, want MSG_MHF_INFO_JOINT", pkt.Opcode())
	}
}

// TestMsgMhfGetCogInfoOpcode tests MsgMhfGetCogInfo Opcode
func TestMsgMhfGetCogInfoOpcode(t *testing.T) {
	pkt := &MsgMhfGetCogInfo{}
	if pkt.Opcode() != network.MSG_MHF_GET_COG_INFO {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_COG_INFO", pkt.Opcode())
	}
}

// TestMsgMhfCheckMonthlyItemOpcode tests MsgMhfCheckMonthlyItem Opcode
func TestMsgMhfCheckMonthlyItemOpcode(t *testing.T) {
	pkt := &MsgMhfCheckMonthlyItem{}
	if pkt.Opcode() != network.MSG_MHF_CHECK_MONTHLY_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_CHECK_MONTHLY_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfAcquireMonthlyItemOpcode tests MsgMhfAcquireMonthlyItem Opcode
func TestMsgMhfAcquireMonthlyItemOpcode(t *testing.T) {
	pkt := &MsgMhfAcquireMonthlyItem{}
	if pkt.Opcode() != network.MSG_MHF_ACQUIRE_MONTHLY_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ACQUIRE_MONTHLY_ITEM", pkt.Opcode())
	}
}

// TestMsgMhfCheckWeeklyStampOpcode tests MsgMhfCheckWeeklyStamp Opcode
func TestMsgMhfCheckWeeklyStampOpcode(t *testing.T) {
	pkt := &MsgMhfCheckWeeklyStamp{}
	if pkt.Opcode() != network.MSG_MHF_CHECK_WEEKLY_STAMP {
		t.Errorf("Opcode() = %s, want MSG_MHF_CHECK_WEEKLY_STAMP", pkt.Opcode())
	}
}

// TestMsgMhfExchangeWeeklyStampOpcode tests MsgMhfExchangeWeeklyStamp Opcode
func TestMsgMhfExchangeWeeklyStampOpcode(t *testing.T) {
	pkt := &MsgMhfExchangeWeeklyStamp{}
	if pkt.Opcode() != network.MSG_MHF_EXCHANGE_WEEKLY_STAMP {
		t.Errorf("Opcode() = %s, want MSG_MHF_EXCHANGE_WEEKLY_STAMP", pkt.Opcode())
	}
}

// TestMsgMhfCreateMercenaryOpcode tests MsgMhfCreateMercenary Opcode
func TestMsgMhfCreateMercenaryOpcode(t *testing.T) {
	pkt := &MsgMhfCreateMercenary{}
	if pkt.Opcode() != network.MSG_MHF_CREATE_MERCENARY {
		t.Errorf("Opcode() = %s, want MSG_MHF_CREATE_MERCENARY", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateMercenaryLogOpcode tests MsgMhfEnumerateMercenaryLog Opcode
func TestMsgMhfEnumerateMercenaryLogOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateMercenaryLog{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_MERCENARY_LOG {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_MERCENARY_LOG", pkt.Opcode())
	}
}

// TestMsgMhfEnumerateGuacotOpcode tests MsgMhfEnumerateGuacot Opcode
func TestMsgMhfEnumerateGuacotOpcode(t *testing.T) {
	pkt := &MsgMhfEnumerateGuacot{}
	if pkt.Opcode() != network.MSG_MHF_ENUMERATE_GUACOT {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENUMERATE_GUACOT", pkt.Opcode())
	}
}

// TestMsgMhfUpdateGuacotOpcode tests MsgMhfUpdateGuacot Opcode
func TestMsgMhfUpdateGuacotOpcode(t *testing.T) {
	pkt := &MsgMhfUpdateGuacot{}
	if pkt.Opcode() != network.MSG_MHF_UPDATE_GUACOT {
		t.Errorf("Opcode() = %s, want MSG_MHF_UPDATE_GUACOT", pkt.Opcode())
	}
}

// TestMsgMhfEnterTournamentQuestOpcode tests MsgMhfEnterTournamentQuest Opcode
func TestMsgMhfEnterTournamentQuestOpcode(t *testing.T) {
	pkt := &MsgMhfEnterTournamentQuest{}
	if pkt.Opcode() != network.MSG_MHF_ENTER_TOURNAMENT_QUEST {
		t.Errorf("Opcode() = %s, want MSG_MHF_ENTER_TOURNAMENT_QUEST", pkt.Opcode())
	}
}

// TestMsgMhfResetAchievementOpcode tests MsgMhfResetAchievement Opcode
func TestMsgMhfResetAchievementOpcode(t *testing.T) {
	pkt := &MsgMhfResetAchievement{}
	if pkt.Opcode() != network.MSG_MHF_RESET_ACHIEVEMENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_RESET_ACHIEVEMENT", pkt.Opcode())
	}
}

// TestMsgMhfPaymentAchievementOpcode tests MsgMhfPaymentAchievement Opcode
func TestMsgMhfPaymentAchievementOpcode(t *testing.T) {
	pkt := &MsgMhfPaymentAchievement{}
	if pkt.Opcode() != network.MSG_MHF_PAYMENT_ACHIEVEMENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_PAYMENT_ACHIEVEMENT", pkt.Opcode())
	}
}

// TestMsgMhfDisplayedAchievementOpcode tests MsgMhfDisplayedAchievement Opcode
func TestMsgMhfDisplayedAchievementOpcode(t *testing.T) {
	pkt := &MsgMhfDisplayedAchievement{}
	if pkt.Opcode() != network.MSG_MHF_DISPLAYED_ACHIEVEMENT {
		t.Errorf("Opcode() = %s, want MSG_MHF_DISPLAYED_ACHIEVEMENT", pkt.Opcode())
	}
}

// TestMsgMhfGetBbsSnsStatusOpcode tests MsgMhfGetBbsSnsStatus Opcode
func TestMsgMhfGetBbsSnsStatusOpcode(t *testing.T) {
	pkt := &MsgMhfGetBbsSnsStatus{}
	if pkt.Opcode() != network.MSG_MHF_GET_BBS_SNS_STATUS {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_BBS_SNS_STATUS", pkt.Opcode())
	}
}

// TestMsgMhfApplyBbsArticleOpcode tests MsgMhfApplyBbsArticle Opcode
func TestMsgMhfApplyBbsArticleOpcode(t *testing.T) {
	pkt := &MsgMhfApplyBbsArticle{}
	if pkt.Opcode() != network.MSG_MHF_APPLY_BBS_ARTICLE {
		t.Errorf("Opcode() = %s, want MSG_MHF_APPLY_BBS_ARTICLE", pkt.Opcode())
	}
}

// TestMsgMhfGetEtcPointsOpcode tests MsgMhfGetEtcPoints Opcode
func TestMsgMhfGetEtcPointsOpcode(t *testing.T) {
	pkt := &MsgMhfGetEtcPoints{}
	if pkt.Opcode() != network.MSG_MHF_GET_ETC_POINTS {
		t.Errorf("Opcode() = %s, want MSG_MHF_GET_ETC_POINTS", pkt.Opcode())
	}
}

// TestMsgMhfUpdateEtcPointOpcode tests MsgMhfUpdateEtcPoint Opcode
func TestMsgMhfUpdateEtcPointOpcode(t *testing.T) {
	pkt := &MsgMhfUpdateEtcPoint{}
	if pkt.Opcode() != network.MSG_MHF_UPDATE_ETC_POINT {
		t.Errorf("Opcode() = %s, want MSG_MHF_UPDATE_ETC_POINT", pkt.Opcode())
	}
}

// TestAchievementPacketParse tests simple achievement packet parsing
func TestAchievementPacketParse(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(5)    // AchievementID
	bf.WriteUint16(100) // Unk1
	bf.WriteUint16(200) // Unk2
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfAddAchievement{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AchievementID != 5 {
		t.Errorf("AchievementID = %d, want 5", pkt.AchievementID)
	}
	if pkt.Unk1 != 100 {
		t.Errorf("Unk1 = %d, want 100", pkt.Unk1)
	}
	if pkt.Unk2 != 200 {
		t.Errorf("Unk2 = %d, want 200", pkt.Unk2)
	}
}
