package channelserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Objective type constants matching questObjType in questfile.bin.hexpat.
const (
	questObjNone         = uint32(0x00000000)
	questObjHunt         = uint32(0x00000001)
	questObjDeliver      = uint32(0x00000002)
	questObjEsoteric     = uint32(0x00000010)
	questObjCapture      = uint32(0x00000101)
	questObjSlay         = uint32(0x00000201)
	questObjDeliverFlag  = uint32(0x00001002)
	questObjBreakPart    = uint32(0x00004004)
	questObjDamage       = uint32(0x00008004)
	questObjSlayOrDamage = uint32(0x00018004)
	questObjSlayTotal    = uint32(0x00020000)
	questObjSlayAll      = uint32(0x00040000)
)

var questObjTypeMap = map[string]uint32{
	"none":           questObjNone,
	"hunt":           questObjHunt,
	"deliver":        questObjDeliver,
	"esoteric":       questObjEsoteric,
	"capture":        questObjCapture,
	"slay":           questObjSlay,
	"deliver_flag":   questObjDeliverFlag,
	"break_part":     questObjBreakPart,
	"damage":         questObjDamage,
	"slay_or_damage": questObjSlayOrDamage,
	"slay_total":     questObjSlayTotal,
	"slay_all":       questObjSlayAll,
}

// ---- JSON schema types ----

// QuestObjectiveJSON represents a single quest objective.
type QuestObjectiveJSON struct {
	// Type is one of: none, hunt, capture, slay, deliver, deliver_flag,
	// break_part, damage, slay_or_damage, slay_total, slay_all, esoteric.
	Type string `json:"type"`
	// Target is a monster ID for hunt/capture/slay/break_part/damage,
	// or an item ID for deliver/deliver_flag.
	Target uint16 `json:"target"`
	// Count is the quantity required (hunts, item count, etc.).
	Count uint16 `json:"count"`
	// Part is the monster part ID for break_part objectives.
	Part uint16 `json:"part,omitempty"`
}

// QuestRewardItemJSON is one entry in a reward table.
type QuestRewardItemJSON struct {
	Rate     uint16 `json:"rate"`
	Item     uint16 `json:"item"`
	Quantity uint16 `json:"quantity"`
}

// QuestRewardTableJSON is a named reward table with its items.
type QuestRewardTableJSON struct {
	TableID uint8                 `json:"table_id"`
	Items   []QuestRewardItemJSON `json:"items"`
}

// QuestMonsterJSON describes one large monster spawn.
type QuestMonsterJSON struct {
	ID          uint8   `json:"id"`
	SpawnAmount uint32  `json:"spawn_amount"`
	SpawnStage  uint32  `json:"spawn_stage"`
	Orientation uint32  `json:"orientation"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Z           float32 `json:"z"`
}

// QuestSupplyItemJSON is one supply box entry.
type QuestSupplyItemJSON struct {
	Item     uint16 `json:"item"`
	Quantity uint16 `json:"quantity"`
}

// QuestStageJSON is a loaded stage definition.
type QuestStageJSON struct {
	StageID uint32 `json:"stage_id"`
}

// QuestForcedEquipJSON defines forced equipment per slot.
// Each slot is [equipment_id, attach1, attach2, attach3].
// Zero values mean no restriction.
type QuestForcedEquipJSON struct {
	Legs   [4]uint16 `json:"legs,omitempty"`
	Weapon [4]uint16 `json:"weapon,omitempty"`
	Head   [4]uint16 `json:"head,omitempty"`
	Chest  [4]uint16 `json:"chest,omitempty"`
	Arms   [4]uint16 `json:"arms,omitempty"`
	Waist  [4]uint16 `json:"waist,omitempty"`
}

// QuestJSON is the human-readable quest definition.
// Time values: TimeLimitMinutes is converted to frames (×30×60) in the binary.
// Strings: encoded as UTF-8 here, converted to Shift-JIS in the binary.
// All pointer-based sections (gathering, area transitions, facilities) are
// omitted — those fields are set to null pointers so the client uses defaults.
type QuestJSON struct {
	// Quest identification
	QuestID uint16 `json:"quest_id"`

	// Text (UTF-8; converted to Shift-JIS in binary)
	Title       string `json:"title"`
	Description string `json:"description"`
	TextMain    string `json:"text_main"`
	TextSubA    string `json:"text_sub_a"`
	TextSubB    string `json:"text_sub_b"`
	SuccessCond string `json:"success_cond"`
	FailCond    string `json:"fail_cond"`
	Contractor  string `json:"contractor"`

	// General quest properties (generalQuestProperties section, 0x44–0x85)
	MonsterSizeMulti uint16 `json:"monster_size_multi"` // 100 = 100%
	SizeRange        uint16 `json:"size_range"`
	StatTable1       uint32 `json:"stat_table_1,omitempty"`
	StatTable2       uint8  `json:"stat_table_2,omitempty"`
	MainRankPoints   uint32 `json:"main_rank_points"`
	SubARankPoints   uint32 `json:"sub_a_rank_points"`
	SubBRankPoints   uint32 `json:"sub_b_rank_points"`

	// Main quest properties
	Fee              uint32 `json:"fee"`
	RewardMain       uint32 `json:"reward_main"`
	RewardSubA       uint16 `json:"reward_sub_a"`
	RewardSubB       uint16 `json:"reward_sub_b"`
	TimeLimitMinutes uint32 `json:"time_limit_minutes"`
	Map              uint32 `json:"map"`
	RankBand         uint16 `json:"rank_band"`
	HardHRReq        uint16 `json:"hard_hr_req,omitempty"`
	JoinRankMin      uint16 `json:"join_rank_min,omitempty"`
	JoinRankMax      uint16 `json:"join_rank_max,omitempty"`
	PostRankMin      uint16 `json:"post_rank_min,omitempty"`
	PostRankMax      uint16 `json:"post_rank_max,omitempty"`

	// Quest variant flags (see handlers_quest.go makeEventQuest comments)
	QuestVariant1 uint8 `json:"quest_variant1,omitempty"`
	QuestVariant2 uint8 `json:"quest_variant2,omitempty"`
	QuestVariant3 uint8 `json:"quest_variant3,omitempty"`
	QuestVariant4 uint8 `json:"quest_variant4,omitempty"`

	// Objectives
	ObjectiveMain QuestObjectiveJSON `json:"objective_main"`
	ObjectiveSubA QuestObjectiveJSON `json:"objective_sub_a,omitempty"`
	ObjectiveSubB QuestObjectiveJSON `json:"objective_sub_b,omitempty"`

	// Monster spawns
	LargeMonsters []QuestMonsterJSON `json:"large_monsters,omitempty"`

	// Reward tables
	Rewards []QuestRewardTableJSON `json:"rewards,omitempty"`

	// Supply box (main: up to 24, sub_a/sub_b: up to 8 each)
	SupplyMain []QuestSupplyItemJSON `json:"supply_main,omitempty"`
	SupplySubA []QuestSupplyItemJSON `json:"supply_sub_a,omitempty"`
	SupplySubB []QuestSupplyItemJSON `json:"supply_sub_b,omitempty"`

	// Loaded stages
	Stages []QuestStageJSON `json:"stages,omitempty"`

	// Forced equipment (optional)
	ForcedEquipment *QuestForcedEquipJSON `json:"forced_equipment,omitempty"`
}

// toShiftJIS converts a UTF-8 string to a null-terminated Shift-JIS byte slice.
// ASCII-only strings pass through unchanged.
func toShiftJIS(s string) ([]byte, error) {
	enc := japanese.ShiftJIS.NewEncoder()
	out, _, err := transform.Bytes(enc, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("shift-jis encode %q: %w", s, err)
	}
	return append(out, 0x00), nil
}

// writeUint16LE writes a little-endian uint16 to buf.
func writeUint16LE(buf *bytes.Buffer, v uint16) {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}

// writeUint32LE writes a little-endian uint32 to buf.
func writeUint32LE(buf *bytes.Buffer, v uint32) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

// writeFloat32LE writes a little-endian IEEE-754 float32 to buf.
func writeFloat32LE(buf *bytes.Buffer, v float32) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
	buf.Write(b[:])
}

// pad writes n zero bytes to buf.
func pad(buf *bytes.Buffer, n int) {
	buf.Write(make([]byte, n))
}

// objectiveBytes serialises one QuestObjectiveJSON to 8 bytes.
// Layout per hexpat objective.hexpat:
//
//	u32 goalType
//	if hunt/capture/slay/damage/break_part: u8 target, u8 pad
//	else: u16 target
//	if break_part: u16 goalPart
//	else: u16 goalCount
//	if none: trailing padding[4] instead of the above
func objectiveBytes(obj QuestObjectiveJSON) ([]byte, error) {
	goalType, ok := questObjTypeMap[obj.Type]
	if !ok {
		if obj.Type == "" {
			goalType = questObjNone
		} else {
			return nil, fmt.Errorf("unknown objective type %q", obj.Type)
		}
	}

	buf := &bytes.Buffer{}
	writeUint32LE(buf, goalType)

	if goalType == questObjNone {
		pad(buf, 4)
		return buf.Bytes(), nil
	}

	switch goalType {
	case questObjHunt, questObjCapture, questObjSlay, questObjDamage,
		questObjSlayOrDamage, questObjBreakPart:
		buf.WriteByte(uint8(obj.Target))
		buf.WriteByte(0x00)
	default:
		writeUint16LE(buf, obj.Target)
	}

	if goalType == questObjBreakPart {
		writeUint16LE(buf, obj.Part)
	} else {
		writeUint16LE(buf, obj.Count)
	}

	return buf.Bytes(), nil
}

// CompileQuestJSON parses JSON quest data and compiles it to the MHF quest
// binary format (ZZ/G10 version, little-endian, uncompressed).
//
// Binary layout produced:
//
//	0x000–0x043  QuestFileHeader (68 bytes, 17 pointers)
//	0x044–0x085  generalQuestProperties (66 bytes)
//	0x086–0x1C5  mainQuestProperties (320 bytes, questBodyLenZZ)
//	0x1C6+       QuestText pointer table (32 bytes) + strings (Shift-JIS)
//	aligned+     stages, supply box, reward tables, monster spawns
func CompileQuestJSON(data []byte) ([]byte, error) {
	var q QuestJSON
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("parse quest JSON: %w", err)
	}

	// ── Section offsets (computed as we build) ──────────────────────────
	const (
		headerSize    = 68             // 0x44
		genPropSize   = 66             // 0x42
		mainPropSize  = questBodyLenZZ // 320 = 0x140
		questTextSize = 32             // 8 × 4-byte s32p pointers
	)

	questTypeFlagsPtr := uint32(headerSize + genPropSize)            // 0x86
	questStringsTablePtr := questTypeFlagsPtr + uint32(mainPropSize) // 0x1C6

	// ── Build Shift-JIS strings ─────────────────────────────────────────
	// Order matches QuestText struct: title, textMain, textSubA, textSubB,
	// successCond, failCond, contractor, description.
	rawTexts := []string{
		q.Title, q.TextMain, q.TextSubA, q.TextSubB,
		q.SuccessCond, q.FailCond, q.Contractor, q.Description,
	}
	var sjisStrings [][]byte
	for _, s := range rawTexts {
		b, err := toShiftJIS(s)
		if err != nil {
			return nil, err
		}
		sjisStrings = append(sjisStrings, b)
	}

	// Compute absolute pointers for each string (right after the s32p table).
	stringDataStart := questStringsTablePtr + uint32(questTextSize)
	stringPtrs := make([]uint32, len(sjisStrings))
	cursor := stringDataStart
	for i, s := range sjisStrings {
		stringPtrs[i] = cursor
		cursor += uint32(len(s))
	}

	// ── Locate variable sections ─────────────────────────────────────────
	// Offset after all string data, 4-byte aligned.
	align4 := func(n uint32) uint32 { return (n + 3) &^ 3 }
	afterStrings := align4(cursor)

	// Stages: each Stage is u32 stageID + 12 bytes padding = 16 bytes.
	loadedStagesPtr := afterStrings
	stagesSize := uint32(len(q.Stages)) * 16
	afterStages := align4(loadedStagesPtr + stagesSize)
	// unk34 (fixedCoords1Ptr) terminates the stages loop in the hexpat.
	unk34Ptr := afterStages

	// Supply box: main=24×4, subA=8×4, subB=8×4 = 160 bytes total.
	supplyBoxPtr := afterStages
	const supplyBoxSize = (24 + 8 + 8) * 4
	afterSupply := align4(supplyBoxPtr + supplyBoxSize)

	// Reward tables: compute size.
	rewardPtr := afterSupply
	rewardBuf := buildRewardTables(q.Rewards)
	afterRewards := align4(rewardPtr + uint32(len(rewardBuf)))

	// Large monster spawns: each is 60 bytes + 1-byte terminator.
	largeMonsterPtr := afterRewards
	monsterBuf := buildMonsterSpawns(q.LargeMonsters)

	// ── Assemble file ────────────────────────────────────────────────────
	out := &bytes.Buffer{}

	// ── Header (68 bytes) ────────────────────────────────────────────────
	writeUint32LE(out, questTypeFlagsPtr) // 0x00 questTypeFlagsPtr
	writeUint32LE(out, loadedStagesPtr)   // 0x04 loadedStagesPtr
	writeUint32LE(out, supplyBoxPtr)      // 0x08 supplyBoxPtr
	writeUint32LE(out, rewardPtr)         // 0x0C rewardPtr
	writeUint16LE(out, 0)                 // 0x10 subSupplyBoxPtr (unused)
	out.WriteByte(0)                      // 0x12 hidden
	out.WriteByte(0)                      // 0x13 subSupplyBoxLen
	writeUint32LE(out, 0)                 // 0x14 questAreaPtr (null)
	writeUint32LE(out, largeMonsterPtr)   // 0x18 largeMonsterPtr
	writeUint32LE(out, 0)                 // 0x1C areaTransitionsPtr (null)
	writeUint32LE(out, 0)                 // 0x20 areaMappingPtr (null)
	writeUint32LE(out, 0)                 // 0x24 mapInfoPtr (null)
	writeUint32LE(out, 0)                 // 0x28 gatheringPointsPtr (null)
	writeUint32LE(out, 0)                 // 0x2C areaFacilitiesPtr (null)
	writeUint32LE(out, 0)                 // 0x30 someStringsPtr (null)
	writeUint32LE(out, unk34Ptr)          // 0x34 fixedCoords1Ptr (stages end)
	writeUint32LE(out, 0)                 // 0x38 gatheringTablesPtr (null)
	writeUint32LE(out, 0)                 // 0x3C fixedCoords2Ptr (null)
	writeUint32LE(out, 0)                 // 0x40 fixedInfoPtr (null)

	if out.Len() != headerSize {
		return nil, fmt.Errorf("header size mismatch: got %d want %d", out.Len(), headerSize)
	}

	// ── General Quest Properties (66 bytes, 0x44–0x85) ──────────────────
	writeUint16LE(out, q.MonsterSizeMulti) // 0x44 monsterSizeMulti
	writeUint16LE(out, q.SizeRange)        // 0x46 sizeRange
	writeUint32LE(out, q.StatTable1)       // 0x48 statTable1
	writeUint32LE(out, q.MainRankPoints)   // 0x4C mainRankPoints
	writeUint32LE(out, 0)                  // 0x50 unknown
	writeUint32LE(out, q.SubARankPoints)   // 0x54 subARankPoints
	writeUint32LE(out, q.SubBRankPoints)   // 0x58 subBRankPoints
	writeUint32LE(out, 0)                  // 0x5C questTypeID / unknown
	out.WriteByte(0)                       // 0x60 padding
	out.WriteByte(q.StatTable2)            // 0x61 statTable2
	pad(out, 0x11)                         // 0x62–0x72 padding
	out.WriteByte(0)                       // 0x73 questKn1
	writeUint16LE(out, 0)                  // 0x74 questKn2
	writeUint16LE(out, 0)                  // 0x76 questKn3
	writeUint16LE(out, 0)                  // 0x78 gatheringTablesQty
	writeUint16LE(out, 0)                  // 0x7A unknown
	out.WriteByte(0)                       // 0x7C area1Zones
	out.WriteByte(0)                       // 0x7D area2Zones
	out.WriteByte(0)                       // 0x7E area3Zones
	out.WriteByte(0)                       // 0x7F area4Zones
	writeUint16LE(out, 0)                  // 0x80 unknown
	writeUint16LE(out, 0)                  // 0x82 unknown
	writeUint16LE(out, 0)                  // 0x84 unknown

	if out.Len() != headerSize+genPropSize {
		return nil, fmt.Errorf("genProp size mismatch: got %d want %d", out.Len(), headerSize+genPropSize)
	}

	// ── Main Quest Properties (320 bytes, 0x86–0x1C5) ───────────────────
	// Matches mainQuestProperties struct in questfile.bin.hexpat.
	mainStart := out.Len()
	out.WriteByte(0)                             // +0x00 unknown
	out.WriteByte(0)                             // +0x01 musicMode
	out.WriteByte(0)                             // +0x02 localeFlags
	out.WriteByte(0)                             // +0x03 unknown
	out.WriteByte(0)                             // +0x04 rankingID
	out.WriteByte(0)                             // +0x05 unknown
	writeUint16LE(out, 0)                        // +0x06 unknown
	writeUint16LE(out, q.RankBand)               // +0x08 rankBand
	writeUint16LE(out, 0)                        // +0x0A questTypeID
	writeUint32LE(out, q.Fee)                    // +0x0C questFee
	writeUint32LE(out, q.RewardMain)             // +0x10 rewardMain
	writeUint32LE(out, 0)                        // +0x14 cartsOrReduction
	writeUint16LE(out, q.RewardSubA)             // +0x18 rewardA
	writeUint16LE(out, 0)                        // +0x1A padding
	writeUint16LE(out, q.RewardSubB)             // +0x1C rewardB
	writeUint16LE(out, q.HardHRReq)              // +0x1E hardHRReq
	writeUint32LE(out, q.TimeLimitMinutes*60*30) // +0x20 questTime (frames at 30Hz)
	writeUint32LE(out, q.Map)                    // +0x24 questMap
	writeUint32LE(out, questStringsTablePtr)     // +0x28 questStringsPtr
	writeUint16LE(out, 0)                        // +0x2C unknown
	writeUint16LE(out, q.QuestID)                // +0x2E questID

	// +0x30 objectives[3] (8 bytes each)
	for _, obj := range []QuestObjectiveJSON{q.ObjectiveMain, q.ObjectiveSubA, q.ObjectiveSubB} {
		b, err := objectiveBytes(obj)
		if err != nil {
			return nil, err
		}
		out.Write(b)
	}

	// +0x48 post-objectives fields
	out.WriteByte(0)                  // +0x48 unknown
	out.WriteByte(0)                  // +0x49 unknown
	writeUint16LE(out, 0)             // +0x4A padding
	writeUint16LE(out, q.JoinRankMin) // +0x4C joinRankMin
	writeUint16LE(out, q.JoinRankMax) // +0x4E joinRankMax
	writeUint16LE(out, q.PostRankMin) // +0x50 postRankMin
	writeUint16LE(out, q.PostRankMax) // +0x52 postRankMax
	pad(out, 8)                       // +0x54 padding[8]

	// +0x5C forced equipment (6 slots × 4 u16 = 48 bytes)
	eq := q.ForcedEquipment
	if eq == nil {
		eq = &QuestForcedEquipJSON{}
	}
	for _, slot := range [][4]uint16{eq.Legs, eq.Weapon, eq.Head, eq.Chest, eq.Arms, eq.Waist} {
		for _, v := range slot {
			writeUint16LE(out, v)
		}
	}

	// +0x8C unknown u32
	writeUint32LE(out, 0)

	// +0x90 monster variants[3] + mapVariant
	out.WriteByte(0) // monsterVariants[0]
	out.WriteByte(0) // monsterVariants[1]
	out.WriteByte(0) // monsterVariants[2]
	out.WriteByte(0) // mapVariant

	// +0x94 requiredItemType (ItemID = u16), requiredItemCount
	writeUint16LE(out, 0)
	out.WriteByte(0) // requiredItemCount

	// +0x97 questVariants
	out.WriteByte(q.QuestVariant1)
	out.WriteByte(q.QuestVariant2)
	out.WriteByte(q.QuestVariant3)
	out.WriteByte(q.QuestVariant4)

	// +0x9B padding[5]
	pad(out, 5)

	// +0xA0 allowedEquipBitmask, points
	writeUint32LE(out, 0) // allowedEquipBitmask
	writeUint32LE(out, 0) // mainPoints
	writeUint32LE(out, 0) // subAPoints
	writeUint32LE(out, 0) // subBPoints

	// +0xB0 rewardItems[3] (ItemID = u16, 3 items = 6 bytes)
	pad(out, 6)

	// +0xB6 interception section (non-SlayAll path: padding[3] + MonsterID[1] = 4 bytes)
	pad(out, 4)

	// +0xBA padding[0xA] = 10 bytes
	pad(out, 10)

	// +0xC4 questClearsAllowed
	writeUint32LE(out, 0)

	// +0xC8 = 200 bytes so far for documented fields. ZZ body = 320 bytes.
	// Zero-pad the remaining unknown ZZ-specific fields.
	writtenInMain := out.Len() - mainStart
	if writtenInMain < mainPropSize {
		pad(out, mainPropSize-writtenInMain)
	} else if writtenInMain > mainPropSize {
		return nil, fmt.Errorf("mainQuestProperties overflowed: wrote %d, max %d", writtenInMain, mainPropSize)
	}

	if out.Len() != int(questTypeFlagsPtr)+mainPropSize {
		return nil, fmt.Errorf("main prop end mismatch: at %d, want %d", out.Len(), int(questTypeFlagsPtr)+mainPropSize)
	}

	// ── QuestText pointer table (32 bytes) ───────────────────────────────
	for _, ptr := range stringPtrs {
		writeUint32LE(out, ptr)
	}

	// ── String data ──────────────────────────────────────────────────────
	for _, s := range sjisStrings {
		out.Write(s)
	}

	// Pad to afterStrings alignment.
	for uint32(out.Len()) < afterStrings {
		out.WriteByte(0)
	}

	// ── Stages ───────────────────────────────────────────────────────────
	// Each Stage: u32 stageID + 12 bytes padding = 16 bytes.
	for _, st := range q.Stages {
		writeUint32LE(out, st.StageID)
		pad(out, 12)
	}
	for uint32(out.Len()) < afterStages {
		out.WriteByte(0)
	}

	// ── Supply Box ───────────────────────────────────────────────────────
	// Three sections: main (24 slots), subA (8 slots), subB (8 slots).
	type slot struct {
		items []QuestSupplyItemJSON
		max   int
	}
	for _, section := range []slot{
		{q.SupplyMain, 24},
		{q.SupplySubA, 8},
		{q.SupplySubB, 8},
	} {
		written := 0
		for _, item := range section.items {
			if written >= section.max {
				break
			}
			writeUint16LE(out, item.Item)
			writeUint16LE(out, item.Quantity)
			written++
		}
		// Pad remaining slots with zeros.
		for written < section.max {
			writeUint32LE(out, 0)
			written++
		}
	}

	// ── Reward Tables ────────────────────────────────────────────────────
	// Written immediately after the supply box (at rewardPtr), then padded
	// to 4-byte alignment before the monster spawn list.
	out.Write(rewardBuf)
	for uint32(out.Len()) < largeMonsterPtr {
		out.WriteByte(0)
	}

	// ── Large Monster Spawns ─────────────────────────────────────────────
	out.Write(monsterBuf)

	return out.Bytes(), nil
}

// buildRewardTables serialises the reward table array and all reward item lists.
// Layout per hexpat:
//
//	RewardTable[] { u8 tableId, u8 pad, u16 pad, u32 tableOffset } terminated by int16(-1)
//	RewardItem[]  { u16 rate, u16 item, u16 quantity }             terminated by int16(-1)
func buildRewardTables(tables []QuestRewardTableJSON) []byte {
	if len(tables) == 0 {
		// Empty: just the terminator.
		b := [2]byte{0xFF, 0xFF}
		return b[:]
	}

	headers := &bytes.Buffer{}
	itemData := &bytes.Buffer{}

	// Header array size = len(tables) × 8 bytes + 2-byte terminator.
	headerArraySize := uint32(len(tables)*8 + 2)

	for _, t := range tables {
		// tableOffset is relative to the start of rewardPtr in the file.
		// We compute it as headerArraySize + offset into itemData.
		tableOffset := headerArraySize + uint32(itemData.Len())

		headers.WriteByte(t.TableID)
		headers.WriteByte(0)      // padding
		writeUint16LE(headers, 0) // padding
		writeUint32LE(headers, tableOffset)

		for _, item := range t.Items {
			writeUint16LE(itemData, item.Rate)
			writeUint16LE(itemData, item.Item)
			writeUint16LE(itemData, item.Quantity)
		}
		// Terminate this table's item list with -1.
		writeUint16LE(itemData, 0xFFFF)
	}
	// Terminate the table header array.
	writeUint16LE(headers, 0xFFFF)

	return append(headers.Bytes(), itemData.Bytes()...)
}

// buildMonsterSpawns serialises the large monster spawn list.
// Each entry is 60 bytes; terminated with a 0xFF byte.
func buildMonsterSpawns(monsters []QuestMonsterJSON) []byte {
	buf := &bytes.Buffer{}
	for _, m := range monsters {
		buf.WriteByte(m.ID)
		pad(buf, 3)                       // +0x01 padding[3]
		writeUint32LE(buf, m.SpawnAmount) // +0x04
		writeUint32LE(buf, m.SpawnStage)  // +0x08
		pad(buf, 16)                      // +0x0C padding[0x10]
		writeUint32LE(buf, m.Orientation) // +0x1C
		writeFloat32LE(buf, m.X)          // +0x20
		writeFloat32LE(buf, m.Y)          // +0x24
		writeFloat32LE(buf, m.Z)          // +0x28
		pad(buf, 16)                      // +0x2C padding[0x10]
	}
	buf.WriteByte(0xFF) // terminator
	return buf.Bytes()
}
