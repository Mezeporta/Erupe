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

// QuestMinionSpawnJSON is one minion spawn entry within a map section.
type QuestMinionSpawnJSON struct {
	Monster     uint8   `json:"monster"`
	SpawnToggle uint16  `json:"spawn_toggle"`
	SpawnAmount uint32  `json:"spawn_amount"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Z           float32 `json:"z"`
}

// QuestMapSectionJSON defines one map section with its minion spawns.
// Each section corresponds to a loaded stage area.
type QuestMapSectionJSON struct {
	LoadedStage   uint32                 `json:"loaded_stage"`
	SpawnMonsters []uint8                `json:"spawn_monsters,omitempty"` // monster IDs for spawn type list
	MinionSpawns  []QuestMinionSpawnJSON `json:"minion_spawns,omitempty"`
}

// QuestAreaTransitionJSON is one zone transition (floatSet).
type QuestAreaTransitionJSON struct {
	TargetStageID1 int16      `json:"target_stage_id"`
	StageVariant   int16      `json:"stage_variant"`
	CurrentX       float32    `json:"current_x"`
	CurrentY       float32    `json:"current_y"`
	CurrentZ       float32    `json:"current_z"`
	TransitionBox  [5]float32 `json:"transition_box"`
	TargetX        float32    `json:"target_x"`
	TargetY        float32    `json:"target_y"`
	TargetZ        float32    `json:"target_z"`
	TargetRotation [2]int16   `json:"target_rotation"`
}

// QuestAreaTransitionsJSON holds the transitions for one area zone entry.
// The pointer may be null (empty transitions list) for zones without transitions.
type QuestAreaTransitionsJSON struct {
	Transitions []QuestAreaTransitionJSON `json:"transitions,omitempty"`
}

// QuestAreaMappingJSON defines coordinate mappings between area and base map.
// Layout: 32 bytes per entry (Area_xPos, Area_zPos, pad8, Base_xPos, Base_zPos, kn_Pos, pad4).
type QuestAreaMappingJSON struct {
	AreaX float32 `json:"area_x"`
	AreaZ float32 `json:"area_z"`
	BaseX float32 `json:"base_x"`
	BaseZ float32 `json:"base_z"`
	KnPos float32 `json:"kn_pos"`
}

// QuestMapInfoJSON contains the map ID and return base camp ID.
type QuestMapInfoJSON struct {
	MapID      uint32 `json:"map_id"`
	ReturnBCID uint32 `json:"return_bc_id"`
}

// QuestGatheringPointJSON is one gathering point (24 bytes).
type QuestGatheringPointJSON struct {
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Z           float32 `json:"z"`
	Range       float32 `json:"range"`
	GatheringID uint16  `json:"gathering_id"`
	MaxCount    uint16  `json:"max_count"`
	MinCount    uint16  `json:"min_count"`
}

// QuestAreaGatheringJSON holds up to 4 gathering points for one area zone entry.
// A nil/empty list means the pointer is null for this zone.
type QuestAreaGatheringJSON struct {
	Points []QuestGatheringPointJSON `json:"points,omitempty"`
}

// QuestFacilityPointJSON is one facility point (24 bytes, facPoint in hexpat).
type QuestFacilityPointJSON struct {
	Type  uint16  `json:"type"` // SpecAc: 1=cooking, 2=fishing, 3=bluebox, etc.
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Z     float32 `json:"z"`
	Range float32 `json:"range"`
	ID    uint16  `json:"id"`
}

// QuestAreaFacilitiesJSON holds the facilities block for one area zone entry.
// A nil/empty list means the pointer is null for this zone.
type QuestAreaFacilitiesJSON struct {
	Points []QuestFacilityPointJSON `json:"points,omitempty"`
}

// QuestGatherItemJSON is one entry in a gathering table.
type QuestGatherItemJSON struct {
	Rate uint16 `json:"rate"`
	Item uint16 `json:"item"`
}

// QuestGatheringTableJSON is one gathering loot table.
type QuestGatheringTableJSON struct {
	Items []QuestGatherItemJSON `json:"items,omitempty"`
}

// QuestJSON is the human-readable quest definition.
// Time values: TimeLimitMinutes is converted to frames (×30×60) in the binary.
// Strings: encoded as UTF-8 here, converted to Shift-JIS in the binary.
type QuestJSON struct {
	// Quest identification
	QuestID uint16 `json:"quest_id"`

	// Text (UTF-8; converted to Shift-JIS in binary).
	//
	// Each field accepts either a plain JSON string (single-language, treated
	// as the value for every language) or a language-keyed object:
	//
	//	"title": "リオレウス"
	//	"title": { "jp": "リオレウス", "en": "Rathalos", "fr": "Rathalos" }
	//
	// CompileQuestJSON resolves these based on the compiling session's
	// language preference (see #188 phase B).
	Title       LocalizedString `json:"title"`
	Description LocalizedString `json:"description"`
	TextMain    LocalizedString `json:"text_main"`
	TextSubA    LocalizedString `json:"text_sub_a"`
	TextSubB    LocalizedString `json:"text_sub_b"`
	SuccessCond LocalizedString `json:"success_cond"`
	FailCond    LocalizedString `json:"fail_cond"`
	Contractor  LocalizedString `json:"contractor"`

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

	// Map sections with minion spawns (questAreaPtr)
	MapSections []QuestMapSectionJSON `json:"map_sections,omitempty"`

	// Area transitions per zone (areaTransitionsPtr); one entry per zone.
	// Length determines area1Zones in generalQuestProperties.
	AreaTransitions []QuestAreaTransitionsJSON `json:"area_transitions,omitempty"`

	// Area coordinate mappings (areaMappingPtr)
	AreaMappings []QuestAreaMappingJSON `json:"area_mappings,omitempty"`

	// Map info: map ID + return base camp ID (mapInfoPtr)
	MapInfo *QuestMapInfoJSON `json:"map_info,omitempty"`

	// Per-zone gathering points (gatheringPointsPtr); one entry per zone.
	GatheringPoints []QuestAreaGatheringJSON `json:"gathering_points,omitempty"`

	// Per-zone area facilities (areaFacilitiesPtr); one entry per zone.
	AreaFacilities []QuestAreaFacilitiesJSON `json:"area_facilities,omitempty"`

	// Additional metadata strings (someStringsPtr / unk30). Optional.
	SomeString string `json:"some_string,omitempty"`
	QuestType  string `json:"quest_type_string,omitempty"`

	// Gathering loot tables (gatheringTablesPtr)
	GatheringTables []QuestGatheringTableJSON `json:"gathering_tables,omitempty"`
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

// writeInt16LE writes a little-endian int16 to buf.
func writeInt16LE(buf *bytes.Buffer, v int16) {
	writeUint16LE(buf, uint16(v))
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

// questBuilder is a small helper for building a quest binary with pointer patching.
// All pointers are absolute offsets from the start of the buffer (file start).
type questBuilder struct {
	out *bytes.Buffer
}

// reserve writes a u32(0) placeholder and returns its offset in the buffer.
func (b *questBuilder) reserve() int {
	off := b.out.Len()
	writeUint32LE(b.out, 0)
	return off
}

// patch writes the current buffer length as a u32 at the previously reserved offset.
func (b *questBuilder) patch(reservedOff int) {
	binary.LittleEndian.PutUint32(b.out.Bytes()[reservedOff:], uint32(b.out.Len()))
}

// patchValue writes a specific uint32 value at a previously reserved offset.
func (b *questBuilder) patchValue(reservedOff int, v uint32) {
	binary.LittleEndian.PutUint32(b.out.Bytes()[reservedOff:], v)
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
//	aligned+     stages, supply box, reward tables, monster spawns,
//	             map sections, area mappings, area transitions,
//	             map info, gathering points, area facilities,
//	             some strings, gathering tables
func CompileQuestJSON(data []byte, lang string) ([]byte, error) {
	var q QuestJSON
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("parse quest JSON: %w", err)
	}

	// ── Compute counts before writing generalQuestProperties ─────────────
	numZones := len(q.AreaTransitions)
	numGatheringTables := len(q.GatheringTables)

	// Validate zone-length consistency.
	if len(q.GatheringPoints) != 0 && len(q.GatheringPoints) != numZones {
		return nil, fmt.Errorf("GatheringPoints len (%d) must equal AreaTransitions len (%d) or be 0",
			len(q.GatheringPoints), numZones)
	}
	if len(q.AreaFacilities) != 0 && len(q.AreaFacilities) != numZones {
		return nil, fmt.Errorf("AreaFacilities len (%d) must equal AreaTransitions len (%d) or be 0",
			len(q.AreaFacilities), numZones)
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
	// successCond, failCond, contractor, description. Each LocalizedString
	// is resolved against the requesting session's language — plain-string
	// JSON fields resolve to their literal value for every language.
	rawTexts := []string{
		q.Title.Resolve(lang), q.TextMain.Resolve(lang),
		q.TextSubA.Resolve(lang), q.TextSubB.Resolve(lang),
		q.SuccessCond.Resolve(lang), q.FailCond.Resolve(lang),
		q.Contractor.Resolve(lang), q.Description.Resolve(lang),
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
	afterMonsters := align4(largeMonsterPtr + uint32(len(monsterBuf)))

	// ── Assemble file ────────────────────────────────────────────────────
	qb := &questBuilder{out: &bytes.Buffer{}}

	// ── Header placeholders (68 bytes) ────────────────────────────────────
	// We'll write the header now with known values; variable section pointers
	// that depend on the preceding variable sections are also known at this
	// point because we computed them above. The new sections (area, gathering,
	// etc.) will be appended after the monster spawns and patched in.
	hdrQuestAreaOff := 0x14    // questAreaPtr placeholder
	hdrAreaTransOff := 0x1C    // areaTransitionsPtr placeholder
	hdrAreaMappingOff := 0x20  // areaMappingPtr placeholder
	hdrMapInfoOff := 0x24      // mapInfoPtr placeholder
	hdrGatherPtsOff := 0x28    // gatheringPointsPtr placeholder
	hdrFacilitiesOff := 0x2C   // areaFacilitiesPtr placeholder
	hdrSomeStringsOff := 0x30  // someStringsPtr placeholder
	hdrGatherTablesOff := 0x38 // gatheringTablesPtr placeholder

	writeUint32LE(qb.out, questTypeFlagsPtr) // 0x00 questTypeFlagsPtr
	writeUint32LE(qb.out, loadedStagesPtr)   // 0x04 loadedStagesPtr
	writeUint32LE(qb.out, supplyBoxPtr)      // 0x08 supplyBoxPtr
	writeUint32LE(qb.out, rewardPtr)         // 0x0C rewardPtr
	writeUint16LE(qb.out, 0)                 // 0x10 subSupplyBoxPtr (unused)
	qb.out.WriteByte(0)                      // 0x12 hidden
	qb.out.WriteByte(0)                      // 0x13 subSupplyBoxLen
	writeUint32LE(qb.out, 0)                 // 0x14 questAreaPtr (patched later)
	writeUint32LE(qb.out, largeMonsterPtr)   // 0x18 largeMonsterPtr
	writeUint32LE(qb.out, 0)                 // 0x1C areaTransitionsPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x20 areaMappingPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x24 mapInfoPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x28 gatheringPointsPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x2C areaFacilitiesPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x30 someStringsPtr (patched later)
	writeUint32LE(qb.out, unk34Ptr)          // 0x34 fixedCoords1Ptr (stages end)
	writeUint32LE(qb.out, 0)                 // 0x38 gatheringTablesPtr (patched later)
	writeUint32LE(qb.out, 0)                 // 0x3C fixedCoords2Ptr (null)
	writeUint32LE(qb.out, 0)                 // 0x40 fixedInfoPtr (null)

	if qb.out.Len() != headerSize {
		return nil, fmt.Errorf("header size mismatch: got %d want %d", qb.out.Len(), headerSize)
	}

	// ── General Quest Properties (66 bytes, 0x44–0x85) ──────────────────
	writeUint16LE(qb.out, q.MonsterSizeMulti)         // 0x44 monsterSizeMulti
	writeUint16LE(qb.out, q.SizeRange)                // 0x46 sizeRange
	writeUint32LE(qb.out, q.StatTable1)               // 0x48 statTable1
	writeUint32LE(qb.out, q.MainRankPoints)           // 0x4C mainRankPoints
	writeUint32LE(qb.out, 0)                          // 0x50 unknown
	writeUint32LE(qb.out, q.SubARankPoints)           // 0x54 subARankPoints
	writeUint32LE(qb.out, q.SubBRankPoints)           // 0x58 subBRankPoints
	writeUint32LE(qb.out, 0)                          // 0x5C questTypeID / unknown
	qb.out.WriteByte(0)                               // 0x60 padding
	qb.out.WriteByte(q.StatTable2)                    // 0x61 statTable2
	pad(qb.out, 0x11)                                 // 0x62–0x72 padding
	qb.out.WriteByte(0)                               // 0x73 questKn1
	writeUint16LE(qb.out, 0)                          // 0x74 questKn2
	writeUint16LE(qb.out, 0)                          // 0x76 questKn3
	writeUint16LE(qb.out, uint16(numGatheringTables)) // 0x78 gatheringTablesQty
	writeUint16LE(qb.out, 0)                          // 0x7A unknown
	qb.out.WriteByte(uint8(numZones))                 // 0x7C area1Zones
	qb.out.WriteByte(0)                               // 0x7D area2Zones
	qb.out.WriteByte(0)                               // 0x7E area3Zones
	qb.out.WriteByte(0)                               // 0x7F area4Zones
	writeUint16LE(qb.out, 0)                          // 0x80 unknown
	writeUint16LE(qb.out, 0)                          // 0x82 unknown
	writeUint16LE(qb.out, 0)                          // 0x84 unknown

	if qb.out.Len() != headerSize+genPropSize {
		return nil, fmt.Errorf("genProp size mismatch: got %d want %d", qb.out.Len(), headerSize+genPropSize)
	}

	// ── Main Quest Properties (320 bytes, 0x86–0x1C5) ───────────────────
	mainStart := qb.out.Len()
	qb.out.WriteByte(0)                             // +0x00 unknown
	qb.out.WriteByte(0)                             // +0x01 musicMode
	qb.out.WriteByte(0)                             // +0x02 localeFlags
	qb.out.WriteByte(0)                             // +0x03 unknown
	qb.out.WriteByte(0)                             // +0x04 rankingID
	qb.out.WriteByte(0)                             // +0x05 unknown
	writeUint16LE(qb.out, 0)                        // +0x06 unknown
	writeUint16LE(qb.out, q.RankBand)               // +0x08 rankBand
	writeUint16LE(qb.out, 0)                        // +0x0A questTypeID
	writeUint32LE(qb.out, q.Fee)                    // +0x0C questFee
	writeUint32LE(qb.out, q.RewardMain)             // +0x10 rewardMain
	writeUint32LE(qb.out, 0)                        // +0x14 cartsOrReduction
	writeUint16LE(qb.out, q.RewardSubA)             // +0x18 rewardA
	writeUint16LE(qb.out, 0)                        // +0x1A padding
	writeUint16LE(qb.out, q.RewardSubB)             // +0x1C rewardB
	writeUint16LE(qb.out, q.HardHRReq)              // +0x1E hardHRReq
	writeUint32LE(qb.out, q.TimeLimitMinutes*60*30) // +0x20 questTime (frames at 30Hz)
	writeUint32LE(qb.out, q.Map)                    // +0x24 questMap
	writeUint32LE(qb.out, questStringsTablePtr)     // +0x28 questStringsPtr
	writeUint16LE(qb.out, 0)                        // +0x2C unknown
	writeUint16LE(qb.out, q.QuestID)                // +0x2E questID

	// +0x30 objectives[3] (8 bytes each)
	for _, obj := range []QuestObjectiveJSON{q.ObjectiveMain, q.ObjectiveSubA, q.ObjectiveSubB} {
		b, err := objectiveBytes(obj)
		if err != nil {
			return nil, err
		}
		qb.out.Write(b)
	}

	// +0x48 post-objectives fields
	qb.out.WriteByte(0)                  // +0x48 unknown
	qb.out.WriteByte(0)                  // +0x49 unknown
	writeUint16LE(qb.out, 0)             // +0x4A padding
	writeUint16LE(qb.out, q.JoinRankMin) // +0x4C joinRankMin
	writeUint16LE(qb.out, q.JoinRankMax) // +0x4E joinRankMax
	writeUint16LE(qb.out, q.PostRankMin) // +0x50 postRankMin
	writeUint16LE(qb.out, q.PostRankMax) // +0x52 postRankMax
	pad(qb.out, 8)                       // +0x54 padding[8]

	// +0x5C forced equipment (6 slots × 4 u16 = 48 bytes)
	eq := q.ForcedEquipment
	if eq == nil {
		eq = &QuestForcedEquipJSON{}
	}
	for _, slot := range [][4]uint16{eq.Legs, eq.Weapon, eq.Head, eq.Chest, eq.Arms, eq.Waist} {
		for _, v := range slot {
			writeUint16LE(qb.out, v)
		}
	}

	// +0x8C unknown u32
	writeUint32LE(qb.out, 0)

	// +0x90 monster variants[3] + mapVariant
	qb.out.WriteByte(0) // monsterVariants[0]
	qb.out.WriteByte(0) // monsterVariants[1]
	qb.out.WriteByte(0) // monsterVariants[2]
	qb.out.WriteByte(0) // mapVariant

	// +0x94 requiredItemType (ItemID = u16), requiredItemCount
	writeUint16LE(qb.out, 0)
	qb.out.WriteByte(0) // requiredItemCount

	// +0x97 questVariants
	qb.out.WriteByte(q.QuestVariant1)
	qb.out.WriteByte(q.QuestVariant2)
	qb.out.WriteByte(q.QuestVariant3)
	qb.out.WriteByte(q.QuestVariant4)

	// +0x9B padding[5]
	pad(qb.out, 5)

	// +0xA0 allowedEquipBitmask, points
	writeUint32LE(qb.out, 0) // allowedEquipBitmask
	writeUint32LE(qb.out, 0) // mainPoints
	writeUint32LE(qb.out, 0) // subAPoints
	writeUint32LE(qb.out, 0) // subBPoints

	// +0xB0 rewardItems[3] (ItemID = u16, 3 items = 6 bytes)
	pad(qb.out, 6)

	// +0xB6 interception section (non-SlayAll path: padding[3] + MonsterID[1] = 4 bytes)
	pad(qb.out, 4)

	// +0xBA padding[0xA] = 10 bytes
	pad(qb.out, 10)

	// +0xC4 questClearsAllowed
	writeUint32LE(qb.out, 0)

	// +0xC8 = 200 bytes so far for documented fields. ZZ body = 320 bytes.
	// Zero-pad the remaining unknown ZZ-specific fields.
	writtenInMain := qb.out.Len() - mainStart
	if writtenInMain < mainPropSize {
		pad(qb.out, mainPropSize-writtenInMain)
	} else if writtenInMain > mainPropSize {
		return nil, fmt.Errorf("mainQuestProperties overflowed: wrote %d, max %d", writtenInMain, mainPropSize)
	}

	if qb.out.Len() != int(questTypeFlagsPtr)+mainPropSize {
		return nil, fmt.Errorf("main prop end mismatch: at %d, want %d", qb.out.Len(), int(questTypeFlagsPtr)+mainPropSize)
	}

	// ── QuestText pointer table (32 bytes) ───────────────────────────────
	for _, ptr := range stringPtrs {
		writeUint32LE(qb.out, ptr)
	}

	// ── String data ──────────────────────────────────────────────────────
	for _, s := range sjisStrings {
		qb.out.Write(s)
	}

	// Pad to afterStrings alignment.
	for uint32(qb.out.Len()) < afterStrings {
		qb.out.WriteByte(0)
	}

	// ── Stages ───────────────────────────────────────────────────────────
	for _, st := range q.Stages {
		writeUint32LE(qb.out, st.StageID)
		pad(qb.out, 12)
	}
	for uint32(qb.out.Len()) < afterStages {
		qb.out.WriteByte(0)
	}

	// ── Supply Box ───────────────────────────────────────────────────────
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
			writeUint16LE(qb.out, item.Item)
			writeUint16LE(qb.out, item.Quantity)
			written++
		}
		for written < section.max {
			writeUint32LE(qb.out, 0)
			written++
		}
	}

	// ── Reward Tables ────────────────────────────────────────────────────
	qb.out.Write(rewardBuf)
	for uint32(qb.out.Len()) < largeMonsterPtr {
		qb.out.WriteByte(0)
	}

	// ── Large Monster Spawns ─────────────────────────────────────────────
	qb.out.Write(monsterBuf)
	for uint32(qb.out.Len()) < afterMonsters {
		qb.out.WriteByte(0)
	}

	// ── Variable sections: map sections, area mappings, transitions, etc. ──
	// All written at afterMonsters and beyond, pointers patched into header.

	// ── Map Sections (questAreaPtr) ──────────────────────────────────────
	// Layout:
	//   u32 ptr[0], u32 ptr[1], ..., u32(0) terminator
	//   For each section:
	//     mapSection: u32 loadedStage, u32 unk, u32 spawnTypesPtr, u32 spawnStatsPtr
	//     u32(0) gap, u16 unk (= 6 bytes after mapSection)
	//     spawnTypes data: (MonsterID u8 + pad[3]) per entry, terminated by 0xFFFF
	//     spawnStats data: MinionSpawn (60 bytes) per entry, terminated by 0xFFFF
	if len(q.MapSections) > 0 {
		questAreaOff := qb.out.Len()
		qb.patchValue(hdrQuestAreaOff, uint32(questAreaOff))

		// Write pointer array (one u32 per section + terminator).
		sectionPtrOffs := make([]int, len(q.MapSections))
		for i := range q.MapSections {
			sectionPtrOffs[i] = qb.reserve()
		}
		writeUint32LE(qb.out, 0) // terminator

		// Write each mapSection block.
		type sectionPtrs struct {
			spawnTypesOff int
			spawnStatsOff int
		}
		internalPtrs := make([]sectionPtrs, len(q.MapSections))

		for i, ms := range q.MapSections {
			// Patch the pointer-array entry to point here.
			qb.patch(sectionPtrOffs[i])

			// mapSection: loadedStage, unk, spawnTypesPtr, spawnStatsPtr
			writeUint32LE(qb.out, ms.LoadedStage)
			writeUint32LE(qb.out, 0) // unk
			internalPtrs[i].spawnTypesOff = qb.reserve()
			internalPtrs[i].spawnStatsOff = qb.reserve()

			// u32(0) gap + u16 unk immediately after the 16-byte mapSection.
			writeUint32LE(qb.out, 0)
			writeUint16LE(qb.out, 0)
		}

		// Write spawn data for each section.
		for i, ms := range q.MapSections {
			// spawnTypes: varPaddT<MonsterID,3> = u8 monster + pad[3] per entry.
			// Terminated by first 2 bytes == 0xFFFF.
			qb.patch(internalPtrs[i].spawnTypesOff)
			for _, monID := range ms.SpawnMonsters {
				qb.out.WriteByte(monID)
				pad(qb.out, 3)
			}
			writeUint16LE(qb.out, 0xFFFF) // terminator

			// Align to 4 bytes before spawnStats.
			for qb.out.Len()%4 != 0 {
				qb.out.WriteByte(0)
			}

			// spawnStats: MinionSpawn per entry (60 bytes), terminated by 0xFFFF.
			qb.patch(internalPtrs[i].spawnStatsOff)
			for _, ms2 := range ms.MinionSpawns {
				qb.out.WriteByte(ms2.Monster)
				qb.out.WriteByte(0)                    // padding[1]
				writeUint16LE(qb.out, ms2.SpawnToggle) // spawnToggle
				writeUint32LE(qb.out, ms2.SpawnAmount) // spawnAmount
				writeUint32LE(qb.out, 0)               // unk u32
				pad(qb.out, 0x10)                      // padding[16]
				writeUint32LE(qb.out, 0)               // unk u32
				writeFloat32LE(qb.out, ms2.X)
				writeFloat32LE(qb.out, ms2.Y)
				writeFloat32LE(qb.out, ms2.Z)
				pad(qb.out, 0x10) // padding[16]
			}
			writeUint16LE(qb.out, 0xFFFF) // terminator

			// Align for next section.
			for qb.out.Len()%4 != 0 {
				qb.out.WriteByte(0)
			}
		}
	}

	// ── Area Mappings (areaMappingPtr) ────────────────────────────────────
	// Written BEFORE area transitions so the parser can use
	// "read until areaTransitionsPtr" to know the count.
	// Layout: AreaMappings[n] × 32 bytes each, back-to-back.
	//   float area_xPos, float area_zPos, pad[8],
	//   float base_xPos, float base_zPos, float kn_Pos, pad[4]
	if len(q.AreaMappings) > 0 {
		areaMappingOff := qb.out.Len()
		qb.patchValue(hdrAreaMappingOff, uint32(areaMappingOff))

		for _, am := range q.AreaMappings {
			writeFloat32LE(qb.out, am.AreaX)
			writeFloat32LE(qb.out, am.AreaZ)
			pad(qb.out, 8)
			writeFloat32LE(qb.out, am.BaseX)
			writeFloat32LE(qb.out, am.BaseZ)
			writeFloat32LE(qb.out, am.KnPos)
			pad(qb.out, 4)
		}
	}

	// ── Area Transitions (areaTransitionsPtr) ─────────────────────────────
	// Layout: playerAreaChange[area1Zones] = u32 ptr per zone.
	// Then floatSet arrays for each zone with transitions.
	if numZones > 0 {
		areaTransOff := qb.out.Len()
		qb.patchValue(hdrAreaTransOff, uint32(areaTransOff))

		// Write pointer array.
		zonePtrOffs := make([]int, numZones)
		for i := range q.AreaTransitions {
			zonePtrOffs[i] = qb.reserve()
		}

		// Write floatSet arrays for non-empty zones.
		for i, zone := range q.AreaTransitions {
			if len(zone.Transitions) == 0 {
				// Null pointer — leave as 0.
				continue
			}
			qb.patch(zonePtrOffs[i])
			for _, tr := range zone.Transitions {
				writeInt16LE(qb.out, tr.TargetStageID1)
				writeInt16LE(qb.out, tr.StageVariant)
				writeFloat32LE(qb.out, tr.CurrentX)
				writeFloat32LE(qb.out, tr.CurrentY)
				writeFloat32LE(qb.out, tr.CurrentZ)
				for _, f := range tr.TransitionBox {
					writeFloat32LE(qb.out, f)
				}
				writeFloat32LE(qb.out, tr.TargetX)
				writeFloat32LE(qb.out, tr.TargetY)
				writeFloat32LE(qb.out, tr.TargetZ)
				for _, r := range tr.TargetRotation {
					writeInt16LE(qb.out, r)
				}
			}
			// Terminate with s16(-1).
			writeInt16LE(qb.out, -1)
			// Align.
			for qb.out.Len()%4 != 0 {
				qb.out.WriteByte(0)
			}
		}
	}

	// ── Map Info (mapInfoPtr) ─────────────────────────────────────────────
	if q.MapInfo != nil {
		mapInfoOff := qb.out.Len()
		qb.patchValue(hdrMapInfoOff, uint32(mapInfoOff))
		writeUint32LE(qb.out, q.MapInfo.MapID)
		writeUint32LE(qb.out, q.MapInfo.ReturnBCID)
	}

	// ── Gathering Points (gatheringPointsPtr) ─────────────────────────────
	// Layout: ptGatheringPoint[area1Zones] = u32 ptr per zone.
	// Each non-null ptr points to gatheringPoint[4] terminated by xPos=-1.0.
	if numZones > 0 && len(q.GatheringPoints) > 0 {
		gatherPtsOff := qb.out.Len()
		qb.patchValue(hdrGatherPtsOff, uint32(gatherPtsOff))

		// Write pointer array.
		gpPtrOffs := make([]int, numZones)
		for i := range q.GatheringPoints {
			gpPtrOffs[i] = qb.reserve()
		}

		// Write gathering point arrays for non-empty zones.
		for i, zone := range q.GatheringPoints {
			if len(zone.Points) == 0 {
				continue
			}
			qb.patch(gpPtrOffs[i])
			for _, gp := range zone.Points {
				writeFloat32LE(qb.out, gp.X)
				writeFloat32LE(qb.out, gp.Y)
				writeFloat32LE(qb.out, gp.Z)
				writeFloat32LE(qb.out, gp.Range)
				writeUint16LE(qb.out, gp.GatheringID)
				writeUint16LE(qb.out, gp.MaxCount)
				pad(qb.out, 2)
				writeUint16LE(qb.out, gp.MinCount)
			}
			// Terminator: xPos == -1.0 (0xBF800000).
			writeFloat32LE(qb.out, float32(math.Float32frombits(0xBF800000)))
			// Pad terminator entry to 24 bytes total (only wrote 4).
			pad(qb.out, 20)
		}
	}

	// ── Area Facilities (areaFacilitiesPtr) ───────────────────────────────
	// Layout: ptVar<facPointBlock>[area1Zones] = u32 ptr per zone.
	// Each non-null ptr points to a facPointBlock.
	// facPoint: pad[2] + SpecAc(u16) + xPos + yPos + zPos + range + id(u16) + pad[2] = 24 bytes
	// facPointBlock: facPoints[] terminated by (xPos-at-$+4 == 0xBF800000) + pad[0xC] + float + float
	// Terminator layout: write pad[2]+type[2] then float32(-1.0) to trigger termination,
	// then block footer: pad[0xC] + float(0) + float(0).
	if numZones > 0 && len(q.AreaFacilities) > 0 {
		facOff := qb.out.Len()
		qb.patchValue(hdrFacilitiesOff, uint32(facOff))

		facPtrOffs := make([]int, numZones)
		for i := range q.AreaFacilities {
			facPtrOffs[i] = qb.reserve()
		}

		for i, zone := range q.AreaFacilities {
			if len(zone.Points) == 0 {
				continue
			}
			qb.patch(facPtrOffs[i])

			for _, fp := range zone.Points {
				pad(qb.out, 2)                 // pad[2]
				writeUint16LE(qb.out, fp.Type) // SpecAc type
				writeFloat32LE(qb.out, fp.X)
				writeFloat32LE(qb.out, fp.Y)
				writeFloat32LE(qb.out, fp.Z)
				writeFloat32LE(qb.out, fp.Range)
				writeUint16LE(qb.out, fp.ID)
				pad(qb.out, 2) // pad[2]
			}

			// Terminator: the while condition checks read_unsigned($+4,4).
			// Write 4 bytes header (pad[2]+type[2]) then float32(-1.0).
			pad(qb.out, 2)
			writeUint16LE(qb.out, 0)
			writeFloat32LE(qb.out, float32(math.Float32frombits(0xBF800000)))

			// Block footer: padding[0xC] + float(0) + float(0) = 20 bytes.
			pad(qb.out, 0xC)
			writeFloat32LE(qb.out, 0)
			writeFloat32LE(qb.out, 0)
		}
	}

	// ── Some Strings (someStringsPtr / unk30) ─────────────────────────────
	// Layout at unk30: ptr someStringPtr, ptr questTypePtr (8 bytes),
	// then the string data.
	hasSomeStrings := q.SomeString != "" || q.QuestType != ""
	if hasSomeStrings {
		someStringsOff := qb.out.Len()
		qb.patchValue(hdrSomeStringsOff, uint32(someStringsOff))

		// Two pointer slots.
		someStrPtrOff := qb.reserve()
		questTypePtrOff := qb.reserve()

		if q.SomeString != "" {
			qb.patch(someStrPtrOff)
			b, err := toShiftJIS(q.SomeString)
			if err != nil {
				return nil, err
			}
			qb.out.Write(b)
		}

		if q.QuestType != "" {
			qb.patch(questTypePtrOff)
			b, err := toShiftJIS(q.QuestType)
			if err != nil {
				return nil, err
			}
			qb.out.Write(b)
		}
	}

	// ── Gathering Tables (gatheringTablesPtr) ─────────────────────────────
	// Layout: ptVar<gatheringTable>[gatheringTablesQty] = u32 ptr per table.
	// Each ptr points to GatherItem[] terminated by u16(0xFFFF).
	// GatherItem: u16 rate + u16 item = 4 bytes.
	if numGatheringTables > 0 {
		gatherTablesOff := qb.out.Len()
		qb.patchValue(hdrGatherTablesOff, uint32(gatherTablesOff))

		tblPtrOffs := make([]int, numGatheringTables)
		for i := range q.GatheringTables {
			tblPtrOffs[i] = qb.reserve()
		}

		for i, tbl := range q.GatheringTables {
			qb.patch(tblPtrOffs[i])
			for _, item := range tbl.Items {
				writeUint16LE(qb.out, item.Rate)
				writeUint16LE(qb.out, item.Item)
			}
			writeUint16LE(qb.out, 0xFFFF) // terminator
		}
	}

	return qb.out.Bytes(), nil
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
