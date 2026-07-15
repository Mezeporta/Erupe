package channelserver

// Raviente quest type codes
const (
	QuestTypeSpecialTool      = uint8(9)
	QuestTypeRegularRaviente  = uint8(16)
	QuestTypeViolentRaviente  = uint8(22)
	QuestTypeBerserkRaviente  = uint8(40)
	QuestTypeExtremeRaviente  = uint8(50)
	QuestTypeSmallBerserkRavi = uint8(51)
)

// Diva Defense (interception) quest type codes. Verified exhaustively against
// every ripped 58xxx quest_id row in EventQuests.json: quest_type is always
// 46, 47, or 48 for those rows, and never appears on any other quest_id.
const (
	QuestTypeDivaDefenseA = uint8(46)
	QuestTypeDivaDefenseB = uint8(47)
	QuestTypeDivaDefenseC = uint8(48)
)

func isDivaDefenseQuestType(questType uint8) bool {
	return questType == QuestTypeDivaDefenseA ||
		questType == QuestTypeDivaDefenseB ||
		questType == QuestTypeDivaDefenseC
}

// Event quest binary frame offsets
const (
	questFrameTimeFlagOffset = 25
	questFrameVariant3Offset = 175
)

// Quest body lengths per game version
const (
	questBodyLenS6   = 160
	questBodyLenF5   = 168
	questBodyLenG101 = 192
	questBodyLenZ1   = 224
	questBodyLenZZ   = 320
)

// BackportQuest constants
const (
	questRewardTableBase    = uint32(96)
	questStringPointerOff   = 40
	questStringTablePadding = 32
	questStringCount        = 8
)

// BackportQuest fill lengths per version
const (
	questBackportFillS6   = uint32(44)
	questBackportFillF5   = uint32(52)
	questBackportFillG101 = uint32(76)
	questBackportFillZZ   = uint32(108)
)

// Tune value count limits per game version
const (
	tuneLimitG1   = 256
	tuneLimitG3   = 283
	tuneLimitGG   = 315
	tuneLimitG61  = 332
	tuneLimitG7   = 339
	tuneLimitG81  = 396
	tuneLimitG91  = 694
	tuneLimitG101 = 704
	tuneLimitZ2   = 750
	tuneLimitZZ   = 770
)

// Event quest data size bounds
const (
	questDataMaxLen = 896
	questDataMinLen = 352
)
