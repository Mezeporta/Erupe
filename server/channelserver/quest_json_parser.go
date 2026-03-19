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
	u32 := func(off int) uint32 {
		return binary.LittleEndian.Uint32(data[off:])
	}
	f32 := func(off int) float32 {
		return math.Float32frombits(binary.LittleEndian.Uint32(data[off:]))
	}

	// bounds checks a read of n bytes at off.
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
	// 0x10 subSupplyBoxPtr (u16), 0x12 hidden, 0x13 subSupplyBoxLen — not in QuestJSON
	// 0x14 questAreaPtr — null, not parsed
	largeMonsterPtr := int(u32(0x18))
	// 0x1C areaTransitionsPtr — null, not parsed
	// 0x20 areaMappingPtr — null, not parsed
	// 0x24 mapInfoPtr — null, not parsed
	// 0x28 gatheringPointsPtr — null, not parsed
	// 0x2C areaFacilitiesPtr — null, not parsed
	// 0x30 someStringsPtr — null, not parsed
	unk34Ptr := int(u32(0x34)) // stages-end sentinel
	// 0x38 gatheringTablesPtr — null, not parsed
	// 0x3C fixedCoords2Ptr — null, not parsed
	// 0x40 fixedInfoPtr — null, not parsed

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
	// 0x62–0x85 padding, questKn1/2/3, gatheringTablesQty, zone counts, unknowns — skipped

	// ── Main Quest Properties (at questTypeFlagsPtr, 320 bytes) ─────────
	if questTypeFlagsPtr == 0 {
		return nil, fmt.Errorf("questTypeFlagsPtr is null; cannot read main quest properties")
	}
	if err := check(questTypeFlagsPtr, questBodyLenZZ, "mainQuestProperties"); err != nil {
		return nil, err
	}

	mp := questTypeFlagsPtr // shorthand

	// +0x08 rankBand
	q.RankBand = u16(mp + 0x08)
	// +0x0C questFee
	q.Fee = u32(mp + 0x0C)
	// +0x10 rewardMain
	q.RewardMain = u32(mp + 0x10)
	// +0x18 rewardA
	q.RewardSubA = u16(mp + 0x18)
	// +0x1C rewardB
	q.RewardSubB = u16(mp + 0x1C)
	// +0x1E hardHRReq
	q.HardHRReq = u16(mp + 0x1E)
	// +0x20 questTime (frames at 30 Hz → minutes)
	questFrames := u32(mp + 0x20)
	q.TimeLimitMinutes = questFrames / (60 * 30)
	// +0x24 questMap
	q.Map = u32(mp + 0x24)
	// +0x28 questStringsPtr (absolute file offset)
	questStringsPtr := int(u32(mp + 0x28))
	// +0x2E questID
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
		// 8 pointers × 4 bytes: title, textMain, textSubA, textSubB,
		// successCond, failCond, contractor, description.
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
	// Guarded by loadedStagesPtr; terminated when we reach unk34Ptr.
	// Each stage: u32 stageID + 12 bytes padding = 16 bytes.
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
	// Guarded by supplyBoxPtr. Layout: main(24) + subA(8) + subB(8) × 4 bytes each.
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
	// Guarded by rewardPtr. Header array terminated by int16(-1); item lists
	// each terminated by int16(-1).
	if rewardPtr != 0 {
		tables, err := parseRewardTables(data, rewardPtr)
		if err != nil {
			return nil, err
		}
		q.Rewards = tables
	}

	// ── Large Monster Spawns ─────────────────────────────────────────────
	// Guarded by largeMonsterPtr. Each entry is 60 bytes; terminated by 0xFF.
	if largeMonsterPtr != 0 {
		monsters, err := parseMonsterSpawns(data, largeMonsterPtr, f32)
		if err != nil {
			return nil, err
		}
		q.LargeMonsters = monsters
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
		// Check for terminator (0xFFFF).
		if binary.LittleEndian.Uint16(data[off:]) == 0xFFFF {
			break
		}
		if off+8 > len(data) {
			return nil, fmt.Errorf("reward table header entry truncated at 0x%X", off)
		}
		tableID := data[off]
		tableOff := int(binary.LittleEndian.Uint32(data[off+4:])) + baseOff
		off += 8

		// Read items at tableOff.
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
