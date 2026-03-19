package channelserver

import (
	"encoding/binary"
	"fmt"
	"math"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseQuestBinary reads a MHF quest binary (ZZ/G10 layout, little-endian)
// and returns a QuestJSON ready for re-compilation with CompileQuestJSON.
//
// The binary layout is described in quest_json.go (CompileQuestJSON).
// Sections guarded by null pointers in the header are skipped; the
// corresponding QuestJSON slices will be nil/empty.
func ParseQuestBinary(data []byte) (*QuestJSON, error) {
	if len(data) < 0x86 {
		return nil, fmt.Errorf("quest binary too short: %d bytes (minimum 0x86)", len(data))
	}

	// ── Helper closures ──────────────────────────────────────────────────
	u8 := func(off int) uint8 {
		return data[off]
	}
	u16 := func(off int) uint16 {
		return binary.LittleEndian.Uint16(data[off:])
	}
	i16 := func(off int) int16 {
		return int16(binary.LittleEndian.Uint16(data[off:]))
	}
	u32 := func(off int) uint32 {
		return binary.LittleEndian.Uint32(data[off:])
	}
	f32 := func(off int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[off:]))
	}

	// check bounds-checks a read of n bytes at off.
	check := func(off, n int, ctx string) error {
		if off < 0 || off+n > len(data) {
			return fmt.Errorf("%s: offset 0x%X len %d out of bounds (file len %d)", ctx, off, n, len(data))
		}
		return nil
	}

	// readSJIS reads a null-terminated Shift-JIS string starting at off.
	readSJIS := func(off int) (string, error) {
		if off < 0 || off >= len(data) {
			return "", fmt.Errorf("string offset 0x%X out of bounds", off)
		}
		end := off
		for end < len(data) && data[end] != 0 {
			end++
		}
		sjis := data[off:end]
		if len(sjis) == 0 {
			return "", nil
		}
		dec := japanese.ShiftJIS.NewDecoder()
		utf8, _, err := transform.Bytes(dec, sjis)
		if err != nil {
			return "", fmt.Errorf("shift-jis decode at 0x%X: %w", off, err)
		}
		return string(utf8), nil
	}

	q := &QuestJSON{}

	// ── Header (0x00–0x43) ───────────────────────────────────────────────
	questTypeFlagsPtr := int(u32(0x00))
	loadedStagesPtr := int(u32(0x04))
	supplyBoxPtr := int(u32(0x08))
	rewardPtr := int(u32(0x0C))
	questAreaPtr := int(u32(0x14))
	largeMonsterPtr := int(u32(0x18))
	areaTransitionsPtr := int(u32(0x1C))
	areaMappingPtr := int(u32(0x20))
	mapInfoPtr := int(u32(0x24))
	gatheringPointsPtr := int(u32(0x28))
	areaFacilitiesPtr := int(u32(0x2C))
	someStringsPtr := int(u32(0x30))
	unk34Ptr := int(u32(0x34)) // stages-end sentinel
	gatheringTablesPtr := int(u32(0x38))

	// ── General Quest Properties (0x44–0x85) ────────────────────────────
	q.MonsterSizeMulti = u16(0x44)
	q.SizeRange = u16(0x46)
	q.StatTable1 = u32(0x48)
	q.MainRankPoints = u32(0x4C)
	// 0x50 unknown u32 — skipped
	q.SubARankPoints = u32(0x54)
	q.SubBRankPoints = u32(0x58)
	// 0x5C questTypeID/unknown — skipped
	// 0x60 padding
	q.StatTable2 = u8(0x61)
	// 0x62–0x72 padding
	// 0x73 questKn1, 0x74 questKn2, 0x76 questKn3 — skipped
	gatheringTablesQty := int(u16(0x78))
	// 0x7A unknown
	area1Zones := int(u8(0x7C))
	// 0x7D–0x7F area2–4Zones (not needed for parsing)

	// ── Main Quest Properties (at questTypeFlagsPtr, 320 bytes) ─────────
	if questTypeFlagsPtr == 0 {
		return nil, fmt.Errorf("questTypeFlagsPtr is null; cannot read main quest properties")
	}
	if err := check(questTypeFlagsPtr, questBodyLenZZ, "mainQuestProperties"); err != nil {
		return nil, err
	}

	mp := questTypeFlagsPtr // shorthand

	q.RankBand = u16(mp + 0x08)
	q.Fee = u32(mp + 0x0C)
	q.RewardMain = u32(mp + 0x10)
	q.RewardSubA = u16(mp + 0x18)
	q.RewardSubB = u16(mp + 0x1C)
	q.HardHRReq = u16(mp + 0x1E)
	questFrames := u32(mp + 0x20)
	q.TimeLimitMinutes = questFrames / (60 * 30)
	q.Map = u32(mp + 0x24)
	questStringsPtr := int(u32(mp + 0x28))
	q.QuestID = u16(mp + 0x2E)

	// +0x30 objectives[3] (8 bytes each)
	objectives, err := parseObjectives(data, mp+0x30)
	if err != nil {
		return nil, err
	}
	q.ObjectiveMain = objectives[0]
	q.ObjectiveSubA = objectives[1]
	q.ObjectiveSubB = objectives[2]

	// +0x4C joinRankMin/Max, postRankMin/Max
	q.JoinRankMin = u16(mp + 0x4C)
	q.JoinRankMax = u16(mp + 0x4E)
	q.PostRankMin = u16(mp + 0x50)
	q.PostRankMax = u16(mp + 0x52)

	// +0x5C forced equipment (6 slots × 4 × u16 = 48 bytes)
	eq, hasEquip := parseForcedEquip(data, mp+0x5C)
	if hasEquip {
		q.ForcedEquipment = eq
	}

	// +0x97 questVariants
	q.QuestVariant1 = u8(mp + 0x97)
	q.QuestVariant2 = u8(mp + 0x98)
	q.QuestVariant3 = u8(mp + 0x99)
	q.QuestVariant4 = u8(mp + 0x9A)

	// ── QuestText strings ────────────────────────────────────────────────
	if questStringsPtr != 0 {
		if err := check(questStringsPtr, 32, "questTextTable"); err != nil {
			return nil, err
		}
		strPtrs := make([]int, 8)
		for i := range strPtrs {
			strPtrs[i] = int(u32(questStringsPtr + i*4))
		}
		texts := make([]string, 8)
		for i, ptr := range strPtrs {
			if ptr == 0 {
				continue
			}
			s, err := readSJIS(ptr)
			if err != nil {
				return nil, fmt.Errorf("string[%d]: %w", i, err)
			}
			texts[i] = s
		}
		q.Title = texts[0]
		q.TextMain = texts[1]
		q.TextSubA = texts[2]
		q.TextSubB = texts[3]
		q.SuccessCond = texts[4]
		q.FailCond = texts[5]
		q.Contractor = texts[6]
		q.Description = texts[7]
	}

	// ── Stages ───────────────────────────────────────────────────────────
	if loadedStagesPtr != 0 && unk34Ptr > loadedStagesPtr {
		off := loadedStagesPtr
		for off+16 <= unk34Ptr {
			if err := check(off, 16, "stage"); err != nil {
				return nil, err
			}
			stageID := u32(off)
			q.Stages = append(q.Stages, QuestStageJSON{StageID: stageID})
			off += 16
		}
	}

	// ── Supply Box ───────────────────────────────────────────────────────
	if supplyBoxPtr != 0 {
		const supplyBoxSize = (24 + 8 + 8) * 4
		if err := check(supplyBoxPtr, supplyBoxSize, "supplyBox"); err != nil {
			return nil, err
		}
		q.SupplyMain = readSupplySlots(data, supplyBoxPtr, 24)
		q.SupplySubA = readSupplySlots(data, supplyBoxPtr+24*4, 8)
		q.SupplySubB = readSupplySlots(data, supplyBoxPtr+24*4+8*4, 8)
	}

	// ── Reward Tables ────────────────────────────────────────────────────
	if rewardPtr != 0 {
		tables, err := parseRewardTables(data, rewardPtr)
		if err != nil {
			return nil, err
		}
		q.Rewards = tables
	}

	// ── Large Monster Spawns ─────────────────────────────────────────────
	if largeMonsterPtr != 0 {
		monsters, err := parseMonsterSpawns(data, largeMonsterPtr, f32)
		if err != nil {
			return nil, err
		}
		q.LargeMonsters = monsters
	}

	// ── Map Sections (questAreaPtr) ──────────────────────────────────────
	// Layout: u32 ptr[] terminated by u32(0), then each mapSection:
	//   u32 loadedStage, u32 unk, u32 spawnTypesPtr, u32 spawnStatsPtr,
	//   u32(0) gap, u16 unk — then spawnTypes and spawnStats data.
	if questAreaPtr != 0 {
		sections, err := parseMapSections(data, questAreaPtr, u32, u16, f32)
		if err != nil {
			return nil, err
		}
		q.MapSections = sections
	}

	// ── Area Mappings (areaMappingPtr) ────────────────────────────────────
	// Read AreaMappings until reaching areaTransitionsPtr (or end of file
	// if areaTransitionsPtr is null). Each entry is 32 bytes.
	if areaMappingPtr != 0 {
		endOff := len(data)
		if areaTransitionsPtr != 0 {
			endOff = areaTransitionsPtr
		}
		mappings, err := parseAreaMappings(data, areaMappingPtr, endOff, f32)
		if err != nil {
			return nil, err
		}
		q.AreaMappings = mappings
	}

	// ── Area Transitions (areaTransitionsPtr) ─────────────────────────────
	// playerAreaChange[area1Zones]: one u32 ptr per zone.
	if areaTransitionsPtr != 0 && area1Zones > 0 {
		transitions, err := parseAreaTransitions(data, areaTransitionsPtr, area1Zones, u32, i16, f32)
		if err != nil {
			return nil, err
		}
		q.AreaTransitions = transitions
	}

	// ── Map Info (mapInfoPtr) ─────────────────────────────────────────────
	if mapInfoPtr != 0 {
		if err := check(mapInfoPtr, 8, "mapInfo"); err != nil {
			return nil, err
		}
		q.MapInfo = &QuestMapInfoJSON{
			MapID:      u32(mapInfoPtr),
			ReturnBCID: u32(mapInfoPtr + 4),
		}
	}

	// ── Gathering Points (gatheringPointsPtr) ─────────────────────────────
	// ptGatheringPoint[area1Zones]: one u32 ptr per zone.
	if gatheringPointsPtr != 0 && area1Zones > 0 {
		gatherPts, err := parseGatheringPoints(data, gatheringPointsPtr, area1Zones, u32, u16, f32)
		if err != nil {
			return nil, err
		}
		q.GatheringPoints = gatherPts
	}

	// ── Area Facilities (areaFacilitiesPtr) ───────────────────────────────
	// ptVar<facPointBlock>[area1Zones]: one u32 ptr per zone.
	if areaFacilitiesPtr != 0 && area1Zones > 0 {
		facilities, err := parseAreaFacilities(data, areaFacilitiesPtr, area1Zones, u32, u16, f32)
		if err != nil {
			return nil, err
		}
		q.AreaFacilities = facilities
	}

	// ── Some Strings (someStringsPtr / unk30) ─────────────────────────────
	// Layout: ptr someStringPtr, ptr questTypePtr (8 bytes at someStringsPtr).
	if someStringsPtr != 0 {
		if err := check(someStringsPtr, 8, "someStrings"); err != nil {
			return nil, err
		}
		someStrP := int(u32(someStringsPtr))
		questTypeP := int(u32(someStringsPtr + 4))
		if someStrP != 0 {
			s, err := readSJIS(someStrP)
			if err != nil {
				return nil, fmt.Errorf("someString: %w", err)
			}
			q.SomeString = s
		}
		if questTypeP != 0 {
			s, err := readSJIS(questTypeP)
			if err != nil {
				return nil, fmt.Errorf("questTypeString: %w", err)
			}
			q.QuestType = s
		}
	}

	// ── Gathering Tables (gatheringTablesPtr) ─────────────────────────────
	// ptVar<gatheringTable>[gatheringTablesQty]: one u32 ptr per table.
	// GatherItem: u16 rate + u16 item, terminated by u16(0xFFFF).
	if gatheringTablesPtr != 0 && gatheringTablesQty > 0 {
		tables, err := parseGatheringTables(data, gatheringTablesPtr, gatheringTablesQty, u32, u16)
		if err != nil {
			return nil, err
		}
		q.GatheringTables = tables
	}

	return q, nil
}

// ── Section parsers ──────────────────────────────────────────────────────────

// parseObjectives reads the three 8-byte objective entries at off.
func parseObjectives(data []byte, off int) ([3]QuestObjectiveJSON, error) {
	var objs [3]QuestObjectiveJSON
	for i := range objs {
		base := off + i*8
		if base+8 > len(data) {
			return objs, fmt.Errorf("objective[%d] at 0x%X out of bounds", i, base)
		}
		goalType := binary.LittleEndian.Uint32(data[base:])
		typeName, ok := objTypeToString(goalType)
		if !ok {
			typeName = "none"
		}
		obj := QuestObjectiveJSON{Type: typeName}

		if goalType != questObjNone {
			switch goalType {
			case questObjHunt, questObjCapture, questObjSlay, questObjDamage,
				questObjSlayOrDamage, questObjBreakPart:
				obj.Target = uint16(data[base+4])
				// data[base+5] is padding
			default:
				obj.Target = binary.LittleEndian.Uint16(data[base+4:])
			}

			secondary := binary.LittleEndian.Uint16(data[base+6:])
			if goalType == questObjBreakPart {
				obj.Part = secondary
			} else {
				obj.Count = secondary
			}
		}
		objs[i] = obj
	}
	return objs, nil
}

// parseForcedEquip reads 6 slots × 4 uint16 at off.
// Returns nil, false if all values are zero (no forced equipment).
func parseForcedEquip(data []byte, off int) (*QuestForcedEquipJSON, bool) {
	eq := &QuestForcedEquipJSON{}
	slots := []*[4]uint16{&eq.Legs, &eq.Weapon, &eq.Head, &eq.Chest, &eq.Arms, &eq.Waist}
	anyNonZero := false
	for _, slot := range slots {
		for j := range slot {
			v := binary.LittleEndian.Uint16(data[off:])
			slot[j] = v
			if v != 0 {
				anyNonZero = true
			}
			off += 2
		}
	}
	if !anyNonZero {
		return nil, false
	}
	return eq, true
}

// readSupplySlots reads n supply item slots (each 4 bytes: u16 item + u16 qty)
// starting at off and returns only non-empty entries (item != 0).
func readSupplySlots(data []byte, off, n int) []QuestSupplyItemJSON {
	var out []QuestSupplyItemJSON
	for i := 0; i < n; i++ {
		base := off + i*4
		item := binary.LittleEndian.Uint16(data[base:])
		qty := binary.LittleEndian.Uint16(data[base+2:])
		if item == 0 {
			continue
		}
		out = append(out, QuestSupplyItemJSON{Item: item, Quantity: qty})
	}
	return out
}

// parseRewardTables reads the reward table array starting at baseOff.
// Header array: {u8 tableId, u8 pad, u16 pad, u32 tableOffset} per entry,
// terminated by int16(-1). tableOffset is relative to baseOff.
// Each item list: {u16 rate, u16 item, u16 quantity} terminated by int16(-1).
func parseRewardTables(data []byte, baseOff int) ([]QuestRewardTableJSON, error) {
	var tables []QuestRewardTableJSON
	off := baseOff
	for {
		if off+2 > len(data) {
			return nil, fmt.Errorf("reward table header truncated at 0x%X", off)
		}
		if binary.LittleEndian.Uint16(data[off:]) == 0xFFFF {
			break
		}
		if off+8 > len(data) {
			return nil, fmt.Errorf("reward table header entry truncated at 0x%X", off)
		}
		tableID := data[off]
		tableOff := int(binary.LittleEndian.Uint32(data[off+4:])) + baseOff
		off += 8

		items, err := parseRewardItems(data, tableOff)
		if err != nil {
			return nil, fmt.Errorf("reward table %d items: %w", tableID, err)
		}
		tables = append(tables, QuestRewardTableJSON{TableID: tableID, Items: items})
	}
	return tables, nil
}

// parseRewardItems reads a null-terminated reward item list at off.
func parseRewardItems(data []byte, off int) ([]QuestRewardItemJSON, error) {
	var items []QuestRewardItemJSON
	for {
		if off+2 > len(data) {
			return nil, fmt.Errorf("reward item list truncated at 0x%X", off)
		}
		if binary.LittleEndian.Uint16(data[off:]) == 0xFFFF {
			break
		}
		if off+6 > len(data) {
			return nil, fmt.Errorf("reward item entry truncated at 0x%X", off)
		}
		rate := binary.LittleEndian.Uint16(data[off:])
		item := binary.LittleEndian.Uint16(data[off+2:])
		qty := binary.LittleEndian.Uint16(data[off+4:])
		items = append(items, QuestRewardItemJSON{Rate: rate, Item: item, Quantity: qty})
		off += 6
	}
	return items, nil
}

// parseMonsterSpawns reads large monster spawn entries at baseOff.
// Each entry is 60 bytes; the list is terminated by a 0xFF byte.
func parseMonsterSpawns(data []byte, baseOff int, f32fn func(int) float32) ([]QuestMonsterJSON, error) {
	var monsters []QuestMonsterJSON
	off := baseOff
	const entrySize = 60
	for {
		if off >= len(data) {
			return nil, fmt.Errorf("monster spawn list unterminated at end of file")
		}
		if data[off] == 0xFF {
			break
		}
		if off+entrySize > len(data) {
			return nil, fmt.Errorf("monster spawn entry at 0x%X truncated", off)
		}
		m := QuestMonsterJSON{
			ID:          data[off],
			SpawnAmount: binary.LittleEndian.Uint32(data[off+4:]),
			SpawnStage:  binary.LittleEndian.Uint32(data[off+8:]),
			// +0x0C padding[16]
			Orientation: binary.LittleEndian.Uint32(data[off+0x1C:]),
			X:           f32fn(off + 0x20),
			Y:           f32fn(off + 0x24),
			Z:           f32fn(off + 0x28),
			// +0x2C padding[16]
		}
		monsters = append(monsters, m)
		off += entrySize
	}
	return monsters, nil
}

// parseMapSections reads the MapZones structure at baseOff.
// Layout: u32 ptr[] terminated by u32(0); each ptr points to a mapSection:
//
//	u32 loadedStage, u32 unk, u32 spawnTypesPtr, u32 spawnStatsPtr.
//
// After the 16-byte mapSection: u32(0) gap + u16 unk (2 bytes).
// spawnTypes: varPaddT<MonsterID,3> = u8+pad[3] per entry, terminated by 0xFFFF.
// spawnStats: MinionSpawn (60 bytes) per entry, terminated by 0xFFFF in first 2 bytes.
func parseMapSections(data []byte, baseOff int,
	u32fn func(int) uint32,
	u16fn func(int) uint16,
	f32fn func(int) float32,
) ([]QuestMapSectionJSON, error) {
	var sections []QuestMapSectionJSON

	// Read pointer array (terminated by u32(0)).
	off := baseOff
	for {
		if off+4 > len(data) {
			return nil, fmt.Errorf("mapSection pointer array truncated at 0x%X", off)
		}
		ptr := int(u32fn(off))
		off += 4
		if ptr == 0 {
			break
		}

		// Read mapSection at ptr.
		if ptr+16 > len(data) {
			return nil, fmt.Errorf("mapSection at 0x%X truncated", ptr)
		}
		loadedStage := u32fn(ptr)
		// ptr+4 is unk u32 — skip
		spawnTypesPtr := int(u32fn(ptr + 8))
		spawnStatsPtr := int(u32fn(ptr + 12))

		ms := QuestMapSectionJSON{LoadedStage: loadedStage}

		// Read spawnTypes: varPaddT<MonsterID,3> terminated by 0xFFFF.
		if spawnTypesPtr != 0 {
			stOff := spawnTypesPtr
			for {
				if stOff+2 > len(data) {
					return nil, fmt.Errorf("spawnTypes at 0x%X truncated", stOff)
				}
				if u16fn(stOff) == 0xFFFF {
					break
				}
				if stOff+4 > len(data) {
					return nil, fmt.Errorf("spawnType entry at 0x%X truncated", stOff)
				}
				monID := data[stOff]
				ms.SpawnMonsters = append(ms.SpawnMonsters, monID)
				stOff += 4 // u8 + pad[3]
			}
		}

		// Read spawnStats: MinionSpawn terminated by 0xFFFF in first 2 bytes.
		if spawnStatsPtr != 0 {
			const minionSize = 60
			ssOff := spawnStatsPtr
			for {
				if ssOff+2 > len(data) {
					return nil, fmt.Errorf("spawnStats at 0x%X truncated", ssOff)
				}
				// Terminator: first 2 bytes == 0xFFFF.
				if u16fn(ssOff) == 0xFFFF {
					break
				}
				if ssOff+minionSize > len(data) {
					return nil, fmt.Errorf("minionSpawn at 0x%X truncated", ssOff)
				}
				spawn := QuestMinionSpawnJSON{
					Monster: data[ssOff],
					// ssOff+1 padding
					SpawnToggle: u16fn(ssOff + 2),
					SpawnAmount: u32fn(ssOff + 4),
					// +8 unk u32, +0xC pad[16], +0x1C unk u32
					X: f32fn(ssOff + 0x20),
					Y: f32fn(ssOff + 0x24),
					Z: f32fn(ssOff + 0x28),
				}
				ms.MinionSpawns = append(ms.MinionSpawns, spawn)
				ssOff += minionSize
			}
		}

		sections = append(sections, ms)
	}

	return sections, nil
}

// parseAreaMappings reads AreaMappings entries at baseOff until endOff.
// Each entry is 32 bytes: float areaX, float areaZ, pad[8],
// float baseX, float baseZ, float knPos, pad[4].
func parseAreaMappings(data []byte, baseOff, endOff int, f32fn func(int) float32) ([]QuestAreaMappingJSON, error) {
	var mappings []QuestAreaMappingJSON
	const entrySize = 32
	off := baseOff
	for off+entrySize <= endOff {
		if off+entrySize > len(data) {
			return nil, fmt.Errorf("areaMapping at 0x%X truncated", off)
		}
		am := QuestAreaMappingJSON{
			AreaX: f32fn(off),
			AreaZ: f32fn(off + 4),
			// off+8: pad[8]
			BaseX: f32fn(off + 16),
			BaseZ: f32fn(off + 20),
			KnPos: f32fn(off + 24),
			// off+28: pad[4]
		}
		mappings = append(mappings, am)
		off += entrySize
	}
	return mappings, nil
}

// parseAreaTransitions reads playerAreaChange[numZones] at baseOff.
// Each entry is a u32 pointer to a floatSet array terminated by s16(-1).
// floatSet: s16 targetStageId + s16 stageVariant + float[3] current + float[5] box +
// float[3] target + s16[2] rotation = 52 bytes.
func parseAreaTransitions(data []byte, baseOff, numZones int,
	u32fn func(int) uint32,
	i16fn func(int) int16,
	f32fn func(int) float32,
) ([]QuestAreaTransitionsJSON, error) {
	result := make([]QuestAreaTransitionsJSON, numZones)

	if baseOff+numZones*4 > len(data) {
		return nil, fmt.Errorf("areaTransitions pointer array at 0x%X truncated", baseOff)
	}

	for i := 0; i < numZones; i++ {
		ptr := int(u32fn(baseOff + i*4))
		if ptr == 0 {
			// Null pointer — no transitions for this zone.
			continue
		}

		// Read floatSet entries until targetStageId1 == -1.
		var transitions []QuestAreaTransitionJSON
		off := ptr
		for {
			if off+2 > len(data) {
				return nil, fmt.Errorf("floatSet at 0x%X truncated", off)
			}
			targetStageID := i16fn(off)
			if targetStageID == -1 {
				break
			}
			// Each floatSet is 52 bytes:
			//   s16 targetStageId1 + s16 stageVariant = 4
			//   float[3] current = 12
			//   float[5] transitionBox = 20
			//   float[3] target = 12
			//   s16[2] rotation = 4
			// Total = 52
			const floatSetSize = 52
			if off+floatSetSize > len(data) {
				return nil, fmt.Errorf("floatSet at 0x%X truncated (need %d bytes)", off, floatSetSize)
			}
			tr := QuestAreaTransitionJSON{
				TargetStageID1: targetStageID,
				StageVariant:   i16fn(off + 2),
				CurrentX:       f32fn(off + 4),
				CurrentY:       f32fn(off + 8),
				CurrentZ:       f32fn(off + 12),
				TargetX:        f32fn(off + 36),
				TargetY:        f32fn(off + 40),
				TargetZ:        f32fn(off + 44),
			}
			for j := 0; j < 5; j++ {
				tr.TransitionBox[j] = f32fn(off + 16 + j*4)
			}
			tr.TargetRotation[0] = i16fn(off + 48)
			tr.TargetRotation[1] = i16fn(off + 50)
			transitions = append(transitions, tr)
			off += floatSetSize
		}
		result[i] = QuestAreaTransitionsJSON{Transitions: transitions}
	}

	return result, nil
}

// parseGatheringPoints reads ptGatheringPoint[numZones] at baseOff.
// Each entry is a u32 pointer to gatheringPoint[4] terminated by xPos==-1.0.
// gatheringPoint: float xPos, yPos, zPos, range, u16 gatheringID, u16 maxCount, pad[2], u16 minCount = 24 bytes.
func parseGatheringPoints(data []byte, baseOff, numZones int,
	u32fn func(int) uint32,
	u16fn func(int) uint16,
	f32fn func(int) float32,
) ([]QuestAreaGatheringJSON, error) {
	result := make([]QuestAreaGatheringJSON, numZones)

	if baseOff+numZones*4 > len(data) {
		return nil, fmt.Errorf("gatheringPoints pointer array at 0x%X truncated", baseOff)
	}

	const sentinel = uint32(0xBF800000) // float32(-1.0)
	const pointSize = 24

	for i := 0; i < numZones; i++ {
		ptr := int(u32fn(baseOff + i*4))
		if ptr == 0 {
			continue
		}

		var points []QuestGatheringPointJSON
		off := ptr
		for {
			if off+4 > len(data) {
				return nil, fmt.Errorf("gatheringPoint at 0x%X truncated", off)
			}
			// Terminator: xPos bit pattern == 0xBF800000 (-1.0f).
			if binary.LittleEndian.Uint32(data[off:]) == sentinel {
				break
			}
			if off+pointSize > len(data) {
				return nil, fmt.Errorf("gatheringPoint entry at 0x%X truncated", off)
			}
			gp := QuestGatheringPointJSON{
				X:           f32fn(off),
				Y:           f32fn(off + 4),
				Z:           f32fn(off + 8),
				Range:       f32fn(off + 12),
				GatheringID: u16fn(off + 16),
				MaxCount:    u16fn(off + 18),
				// off+20 pad[2]
				MinCount: u16fn(off + 22),
			}
			points = append(points, gp)
			off += pointSize
		}
		result[i] = QuestAreaGatheringJSON{Points: points}
	}

	return result, nil
}

// parseAreaFacilities reads ptVar<facPointBlock>[numZones] at baseOff.
// Each entry is a u32 pointer to a facPointBlock.
// facPoint: pad[2] + SpecAc(u16) + xPos + yPos + zPos + range + id(u16) + pad[2] = 24 bytes.
// Termination: the loop condition checks read_unsigned($+4,4) != 0xBF800000.
// So a facPoint whose xPos (at offset +4 from start of that potential entry) == -1.0 terminates.
// After all facPoints: padding[0xC] + float + float = 20 bytes (block footer, not parsed into JSON).
func parseAreaFacilities(data []byte, baseOff, numZones int,
	u32fn func(int) uint32,
	u16fn func(int) uint16,
	f32fn func(int) float32,
) ([]QuestAreaFacilitiesJSON, error) {
	result := make([]QuestAreaFacilitiesJSON, numZones)

	if baseOff+numZones*4 > len(data) {
		return nil, fmt.Errorf("areaFacilities pointer array at 0x%X truncated", baseOff)
	}

	const sentinel = uint32(0xBF800000)
	const pointSize = 24

	for i := 0; i < numZones; i++ {
		ptr := int(u32fn(baseOff + i*4))
		if ptr == 0 {
			continue
		}

		var points []QuestFacilityPointJSON
		off := ptr
		for off+8 <= len(data) {
			// Check: read_unsigned($+4, 4) == sentinel means terminate.
			// $+4 is the xPos field of the potential next facPoint.
			if binary.LittleEndian.Uint32(data[off+4:]) == sentinel {
				break
			}
			if off+pointSize > len(data) {
				return nil, fmt.Errorf("facPoint at 0x%X truncated", off)
			}
			fp := QuestFacilityPointJSON{
				// off+0: pad[2]
				Type:  u16fn(off + 2),
				X:     f32fn(off + 4),
				Y:     f32fn(off + 8),
				Z:     f32fn(off + 12),
				Range: f32fn(off + 16),
				ID:    u16fn(off + 20),
				// off+22: pad[2]
			}
			points = append(points, fp)
			off += pointSize
		}
		result[i] = QuestAreaFacilitiesJSON{Points: points}
	}

	return result, nil
}

// parseGatheringTables reads ptVar<gatheringTable>[count] at baseOff.
// Each entry is a u32 pointer to GatherItem[] terminated by u16(0xFFFF).
// GatherItem: u16 rate + u16 item = 4 bytes.
func parseGatheringTables(data []byte, baseOff, count int,
	u32fn func(int) uint32,
	u16fn func(int) uint16,
) ([]QuestGatheringTableJSON, error) {
	result := make([]QuestGatheringTableJSON, count)

	if baseOff+count*4 > len(data) {
		return nil, fmt.Errorf("gatheringTables pointer array at 0x%X truncated", baseOff)
	}

	for i := 0; i < count; i++ {
		ptr := int(u32fn(baseOff + i*4))
		if ptr == 0 {
			continue
		}

		var items []QuestGatherItemJSON
		off := ptr
		for {
			if off+2 > len(data) {
				return nil, fmt.Errorf("gatheringTable at 0x%X truncated", off)
			}
			if u16fn(off) == 0xFFFF {
				break
			}
			if off+4 > len(data) {
				return nil, fmt.Errorf("gatherItem at 0x%X truncated", off)
			}
			items = append(items, QuestGatherItemJSON{
				Rate: u16fn(off),
				Item: u16fn(off + 2),
			})
			off += 4
		}
		result[i] = QuestGatheringTableJSON{Items: items}
	}

	return result, nil
}

// objTypeToString maps a uint32 goal type to its JSON string name.
// Returns "", false for unknown types.
func objTypeToString(t uint32) (string, bool) {
	for name, v := range questObjTypeMap {
		if v == t {
			return name, true
		}
	}
	return "", false
}
