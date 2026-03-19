package channelserver

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
)

// minimalQuestJSON is a small but complete quest used across many test cases.
var minimalQuestJSON = `{
	"quest_id": 1,
	"title": "Test Quest",
	"description": "A test quest.",
	"text_main": "Hunt the Rathalos.",
	"text_sub_a": "",
	"text_sub_b": "",
	"success_cond": "Slay the Rathalos.",
	"fail_cond": "Time runs out or all hunters faint.",
	"contractor": "Guild Master",
	"monster_size_multi": 100,
	"stat_table_1": 0,
	"main_rank_points": 120,
	"sub_a_rank_points": 60,
	"sub_b_rank_points": 0,
	"fee": 500,
	"reward_main": 5000,
	"reward_sub_a": 1000,
	"reward_sub_b": 0,
	"time_limit_minutes": 50,
	"map": 2,
	"rank_band": 0,
	"objective_main": {"type": "hunt", "target": 11, "count": 1},
	"objective_sub_a": {"type": "deliver", "target": 149, "count": 3},
	"objective_sub_b": {"type": "none"},
	"large_monsters": [
		{"id": 11, "spawn_amount": 1, "spawn_stage": 5, "orientation": 180, "x": 1500.0, "y": 0.0, "z": -2000.0}
	],
	"rewards": [
		{
			"table_id": 1,
			"items": [
				{"rate": 50, "item": 149, "quantity": 1},
				{"rate": 30, "item": 153, "quantity": 1}
			]
		}
	],
	"supply_main": [
		{"item": 1, "quantity": 5}
	],
	"stages": [
		{"stage_id": 2}
	]
}`

// ── Compiler tests (existing) ────────────────────────────────────────────────

func TestCompileQuestJSON_MinimalQuest(t *testing.T) {
	data, err := CompileQuestJSON([]byte(minimalQuestJSON))
	if err != nil {
		t.Fatalf("CompileQuestJSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty output")
	}

	// Header check: first pointer (questTypeFlagsPtr) must equal headerSize+genPropSize = 0x86
	questTypeFlagsPtr := binary.LittleEndian.Uint32(data[0:4])
	const expectedBodyStart = uint32(68 + 66) // 0x86
	if questTypeFlagsPtr != expectedBodyStart {
		t.Errorf("questTypeFlagsPtr = 0x%X, want 0x%X", questTypeFlagsPtr, expectedBodyStart)
	}

	// QuestStringsPtr (mainQuestProperties+40) must point past the body.
	questStringsPtr := binary.LittleEndian.Uint32(data[questTypeFlagsPtr+40 : questTypeFlagsPtr+44])
	if questStringsPtr < questTypeFlagsPtr+questBodyLenZZ {
		t.Errorf("questStringsPtr 0x%X is inside main body (ends at 0x%X)", questStringsPtr, questTypeFlagsPtr+questBodyLenZZ)
	}

	// QuestStringsPtr must be within the file.
	if int(questStringsPtr) >= len(data) {
		t.Errorf("questStringsPtr 0x%X out of range (file len %d)", questStringsPtr, len(data))
	}

	// The quest text pointer table: 8 string pointers, all within the file.
	for i := 0; i < 8; i++ {
		off := int(questStringsPtr) + i*4
		if off+4 > len(data) {
			t.Fatalf("string pointer %d out of bounds", i)
		}
		strPtr := binary.LittleEndian.Uint32(data[off : off+4])
		if int(strPtr) >= len(data) {
			t.Errorf("string pointer %d = 0x%X out of file range (%d bytes)", i, strPtr, len(data))
		}
	}

	// QuestID at mainQuestProperties+0x2E.
	questID := binary.LittleEndian.Uint16(data[questTypeFlagsPtr+0x2E : questTypeFlagsPtr+0x30])
	if questID != 1 {
		t.Errorf("questID = %d, want 1", questID)
	}

	// QuestTime at mainQuestProperties+0x20: 50 minutes × 60s × 30Hz = 90000 frames.
	questTime := binary.LittleEndian.Uint32(data[questTypeFlagsPtr+0x20 : questTypeFlagsPtr+0x24])
	if questTime != 90000 {
		t.Errorf("questTime = %d frames, want 90000 (50min)", questTime)
	}
}

func TestCompileQuestJSON_BadObjectiveType(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.ObjectiveMain.Type = "invalid_type"
	b, _ := json.Marshal(q)

	_, err := CompileQuestJSON(b)
	if err == nil {
		t.Fatal("expected error for invalid objective type, got nil")
	}
}

func TestCompileQuestJSON_AllObjectiveTypes(t *testing.T) {
	types := []string{
		"none", "hunt", "capture", "slay", "deliver", "deliver_flag",
		"break_part", "damage", "slay_or_damage", "slay_total", "slay_all", "esoteric",
	}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			var q QuestJSON
			_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
			q.ObjectiveMain.Type = typ
			b, _ := json.Marshal(q)
			if _, err := CompileQuestJSON(b); err != nil {
				t.Fatalf("CompileQuestJSON with type %q: %v", typ, err)
			}
		})
	}
}

func TestCompileQuestJSON_EmptyRewards(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Rewards = nil
	b, _ := json.Marshal(q)
	if _, err := CompileQuestJSON(b); err != nil {
		t.Fatalf("unexpected error with no rewards: %v", err)
	}
}

func TestCompileQuestJSON_MultipleRewardTables(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Rewards = []QuestRewardTableJSON{
		{TableID: 1, Items: []QuestRewardItemJSON{{Rate: 50, Item: 149, Quantity: 1}}},
		{TableID: 2, Items: []QuestRewardItemJSON{{Rate: 100, Item: 153, Quantity: 2}}},
	}
	b, _ := json.Marshal(q)
	data, err := CompileQuestJSON(b)
	if err != nil {
		t.Fatalf("CompileQuestJSON: %v", err)
	}

	// Verify reward pointer points into the file.
	rewardPtr := binary.LittleEndian.Uint32(data[0x0C:0x10])
	if int(rewardPtr) >= len(data) {
		t.Errorf("rewardPtr 0x%X out of file range (%d)", rewardPtr, len(data))
	}
}

// ── Parser tests ─────────────────────────────────────────────────────────────

func TestParseQuestBinary_TooShort(t *testing.T) {
	_, err := ParseQuestBinary([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for undersized input, got nil")
	}
}

func TestParseQuestBinary_NullQuestTypeFlagsPtr(t *testing.T) {
	// Build a buffer that is long enough but has a null questTypeFlagsPtr.
	buf := make([]byte, 0x200)
	// questTypeFlagsPtr at 0x00 = 0 (null)
	binary.LittleEndian.PutUint32(buf[0x00:], 0)
	_, err := ParseQuestBinary(buf)
	if err == nil {
		t.Fatal("expected error for null questTypeFlagsPtr, got nil")
	}
}

func TestParseQuestBinary_MinimalQuest(t *testing.T) {
	data, err := CompileQuestJSON([]byte(minimalQuestJSON))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	q, err := ParseQuestBinary(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Identification
	if q.QuestID != 1 {
		t.Errorf("QuestID = %d, want 1", q.QuestID)
	}

	// Text strings
	if q.Title != "Test Quest" {
		t.Errorf("Title = %q, want %q", q.Title, "Test Quest")
	}
	if q.Description != "A test quest." {
		t.Errorf("Description = %q, want %q", q.Description, "A test quest.")
	}
	if q.TextMain != "Hunt the Rathalos." {
		t.Errorf("TextMain = %q, want %q", q.TextMain, "Hunt the Rathalos.")
	}
	if q.SuccessCond != "Slay the Rathalos." {
		t.Errorf("SuccessCond = %q, want %q", q.SuccessCond, "Slay the Rathalos.")
	}
	if q.FailCond != "Time runs out or all hunters faint." {
		t.Errorf("FailCond = %q, want %q", q.FailCond, "Time runs out or all hunters faint.")
	}
	if q.Contractor != "Guild Master" {
		t.Errorf("Contractor = %q, want %q", q.Contractor, "Guild Master")
	}

	// Numeric fields
	if q.MonsterSizeMulti != 100 {
		t.Errorf("MonsterSizeMulti = %d, want 100", q.MonsterSizeMulti)
	}
	if q.MainRankPoints != 120 {
		t.Errorf("MainRankPoints = %d, want 120", q.MainRankPoints)
	}
	if q.SubARankPoints != 60 {
		t.Errorf("SubARankPoints = %d, want 60", q.SubARankPoints)
	}
	if q.SubBRankPoints != 0 {
		t.Errorf("SubBRankPoints = %d, want 0", q.SubBRankPoints)
	}
	if q.Fee != 500 {
		t.Errorf("Fee = %d, want 500", q.Fee)
	}
	if q.RewardMain != 5000 {
		t.Errorf("RewardMain = %d, want 5000", q.RewardMain)
	}
	if q.RewardSubA != 1000 {
		t.Errorf("RewardSubA = %d, want 1000", q.RewardSubA)
	}
	if q.TimeLimitMinutes != 50 {
		t.Errorf("TimeLimitMinutes = %d, want 50", q.TimeLimitMinutes)
	}
	if q.Map != 2 {
		t.Errorf("Map = %d, want 2", q.Map)
	}

	// Objectives
	if q.ObjectiveMain.Type != "hunt" {
		t.Errorf("ObjectiveMain.Type = %q, want hunt", q.ObjectiveMain.Type)
	}
	if q.ObjectiveMain.Target != 11 {
		t.Errorf("ObjectiveMain.Target = %d, want 11", q.ObjectiveMain.Target)
	}
	if q.ObjectiveMain.Count != 1 {
		t.Errorf("ObjectiveMain.Count = %d, want 1", q.ObjectiveMain.Count)
	}
	if q.ObjectiveSubA.Type != "deliver" {
		t.Errorf("ObjectiveSubA.Type = %q, want deliver", q.ObjectiveSubA.Type)
	}
	if q.ObjectiveSubA.Target != 149 {
		t.Errorf("ObjectiveSubA.Target = %d, want 149", q.ObjectiveSubA.Target)
	}
	if q.ObjectiveSubA.Count != 3 {
		t.Errorf("ObjectiveSubA.Count = %d, want 3", q.ObjectiveSubA.Count)
	}
	if q.ObjectiveSubB.Type != "none" {
		t.Errorf("ObjectiveSubB.Type = %q, want none", q.ObjectiveSubB.Type)
	}

	// Stages
	if len(q.Stages) != 1 {
		t.Fatalf("Stages len = %d, want 1", len(q.Stages))
	}
	if q.Stages[0].StageID != 2 {
		t.Errorf("Stages[0].StageID = %d, want 2", q.Stages[0].StageID)
	}

	// Supply box
	if len(q.SupplyMain) != 1 {
		t.Fatalf("SupplyMain len = %d, want 1", len(q.SupplyMain))
	}
	if q.SupplyMain[0].Item != 1 || q.SupplyMain[0].Quantity != 5 {
		t.Errorf("SupplyMain[0] = {%d, %d}, want {1, 5}", q.SupplyMain[0].Item, q.SupplyMain[0].Quantity)
	}
	if len(q.SupplySubA) != 0 {
		t.Errorf("SupplySubA len = %d, want 0", len(q.SupplySubA))
	}

	// Rewards
	if len(q.Rewards) != 1 {
		t.Fatalf("Rewards len = %d, want 1", len(q.Rewards))
	}
	rt := q.Rewards[0]
	if rt.TableID != 1 {
		t.Errorf("Rewards[0].TableID = %d, want 1", rt.TableID)
	}
	if len(rt.Items) != 2 {
		t.Fatalf("Rewards[0].Items len = %d, want 2", len(rt.Items))
	}
	if rt.Items[0].Rate != 50 || rt.Items[0].Item != 149 || rt.Items[0].Quantity != 1 {
		t.Errorf("Rewards[0].Items[0] = %+v, want {50 149 1}", rt.Items[0])
	}
	if rt.Items[1].Rate != 30 || rt.Items[1].Item != 153 || rt.Items[1].Quantity != 1 {
		t.Errorf("Rewards[0].Items[1] = %+v, want {30 153 1}", rt.Items[1])
	}

	// Large monsters
	if len(q.LargeMonsters) != 1 {
		t.Fatalf("LargeMonsters len = %d, want 1", len(q.LargeMonsters))
	}
	m := q.LargeMonsters[0]
	if m.ID != 11 {
		t.Errorf("LargeMonsters[0].ID = %d, want 11", m.ID)
	}
	if m.SpawnAmount != 1 {
		t.Errorf("LargeMonsters[0].SpawnAmount = %d, want 1", m.SpawnAmount)
	}
	if m.SpawnStage != 5 {
		t.Errorf("LargeMonsters[0].SpawnStage = %d, want 5", m.SpawnStage)
	}
	if m.Orientation != 180 {
		t.Errorf("LargeMonsters[0].Orientation = %d, want 180", m.Orientation)
	}
	if m.X != 1500.0 {
		t.Errorf("LargeMonsters[0].X = %v, want 1500.0", m.X)
	}
	if m.Y != 0.0 {
		t.Errorf("LargeMonsters[0].Y = %v, want 0.0", m.Y)
	}
	if m.Z != -2000.0 {
		t.Errorf("LargeMonsters[0].Z = %v, want -2000.0", m.Z)
	}
}

// ── Round-trip tests ─────────────────────────────────────────────────────────

// roundTrip compiles JSON → binary, parses back to QuestJSON, re-serializes
// to JSON, compiles again, and asserts the two binaries are byte-for-byte equal.
func roundTrip(t *testing.T, label, jsonSrc string) {
	t.Helper()

	bin1, err := CompileQuestJSON([]byte(jsonSrc))
	if err != nil {
		t.Fatalf("%s: compile(1): %v", label, err)
	}

	q, err := ParseQuestBinary(bin1)
	if err != nil {
		t.Fatalf("%s: parse: %v", label, err)
	}

	jsonOut, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("%s: marshal: %v", label, err)
	}

	bin2, err := CompileQuestJSON(jsonOut)
	if err != nil {
		t.Fatalf("%s: compile(2): %v", label, err)
	}

	if !bytes.Equal(bin1, bin2) {
		t.Errorf("%s: round-trip binary mismatch (bin1 len=%d, bin2 len=%d)", label, len(bin1), len(bin2))
		// Find first differing byte to aid debugging.
		limit := len(bin1)
		if len(bin2) < limit {
			limit = len(bin2)
		}
		for i := 0; i < limit; i++ {
			if bin1[i] != bin2[i] {
				t.Errorf("  first diff at offset 0x%X: bin1=0x%02X bin2=0x%02X", i, bin1[i], bin2[i])
				break
			}
		}
	}
}

func TestRoundTrip_MinimalQuest(t *testing.T) {
	roundTrip(t, "minimal", minimalQuestJSON)
}

func TestRoundTrip_NoRewards(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Rewards = nil
	b, _ := json.Marshal(q)
	roundTrip(t, "no rewards", string(b))
}

func TestRoundTrip_NoMonsters(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.LargeMonsters = nil
	b, _ := json.Marshal(q)
	roundTrip(t, "no monsters", string(b))
}

func TestRoundTrip_NoStages(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Stages = nil
	b, _ := json.Marshal(q)
	roundTrip(t, "no stages", string(b))
}

func TestRoundTrip_MultipleStages(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Stages = []QuestStageJSON{{StageID: 2}, {StageID: 5}, {StageID: 11}}
	b, _ := json.Marshal(q)
	roundTrip(t, "multiple stages", string(b))
}

func TestRoundTrip_MultipleMonsters(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.LargeMonsters = []QuestMonsterJSON{
		{ID: 11, SpawnAmount: 1, SpawnStage: 5, Orientation: 180, X: 1500.0, Y: 0.0, Z: -2000.0},
		{ID: 37, SpawnAmount: 2, SpawnStage: 3, Orientation: 90, X: 0.0, Y: 50.0, Z: 300.0},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "multiple monsters", string(b))
}

func TestRoundTrip_MultipleRewardTables(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.Rewards = []QuestRewardTableJSON{
		{TableID: 1, Items: []QuestRewardItemJSON{
			{Rate: 50, Item: 149, Quantity: 1},
			{Rate: 50, Item: 153, Quantity: 2},
		}},
		{TableID: 2, Items: []QuestRewardItemJSON{
			{Rate: 100, Item: 200, Quantity: 3},
		}},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "multiple reward tables", string(b))
}

func TestRoundTrip_FullSupplyBox(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	// Fill supply box to capacity: 24 main + 8 subA + 8 subB.
	q.SupplyMain = make([]QuestSupplyItemJSON, 24)
	for i := range q.SupplyMain {
		q.SupplyMain[i] = QuestSupplyItemJSON{Item: uint16(i + 1), Quantity: uint16(i + 1)}
	}
	q.SupplySubA = []QuestSupplyItemJSON{{Item: 10, Quantity: 2}, {Item: 20, Quantity: 1}}
	q.SupplySubB = []QuestSupplyItemJSON{{Item: 30, Quantity: 5}}
	b, _ := json.Marshal(q)
	roundTrip(t, "full supply box", string(b))
}

func TestRoundTrip_BreakPartObjective(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.ObjectiveMain = QuestObjectiveJSON{Type: "break_part", Target: 11, Part: 3}
	b, _ := json.Marshal(q)
	roundTrip(t, "break_part objective", string(b))
}

func TestRoundTrip_AllObjectiveTypes(t *testing.T) {
	types := []string{
		"none", "hunt", "capture", "slay", "deliver", "deliver_flag",
		"break_part", "damage", "slay_or_damage", "slay_total", "slay_all", "esoteric",
	}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			var q QuestJSON
			_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
			q.ObjectiveMain = QuestObjectiveJSON{Type: typ, Target: 11, Count: 1}
			b, _ := json.Marshal(q)
			roundTrip(t, typ, string(b))
		})
	}
}

func TestRoundTrip_RankFields(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.RankBand = 7
	q.HardHRReq = 300
	q.JoinRankMin = 100
	q.JoinRankMax = 999
	q.PostRankMin = 50
	q.PostRankMax = 500
	b, _ := json.Marshal(q)
	roundTrip(t, "rank fields", string(b))
}

func TestRoundTrip_QuestVariants(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.QuestVariant1 = 1
	q.QuestVariant2 = 2
	q.QuestVariant3 = 4
	q.QuestVariant4 = 8
	b, _ := json.Marshal(q)
	roundTrip(t, "quest variants", string(b))
}

func TestRoundTrip_EmptyQuest(t *testing.T) {
	q := QuestJSON{
		QuestID:          999,
		TimeLimitMinutes: 30,
		MonsterSizeMulti: 100,
		ObjectiveMain:    QuestObjectiveJSON{Type: "slay_all"},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "empty quest", string(b))
}

// ── New section round-trip tests ─────────────────────────────────────────────

func TestRoundTrip_MapSections(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.MapSections = []QuestMapSectionJSON{
		{
			LoadedStage:   5,
			SpawnMonsters: []uint8{0x0F, 0x33}, // Khezu, Blangonga
			MinionSpawns: []QuestMinionSpawnJSON{
				{Monster: 0x0F, SpawnToggle: 1, SpawnAmount: 3, X: 100.0, Y: 0.0, Z: -200.0},
				{Monster: 0x33, SpawnToggle: 1, SpawnAmount: 2, X: 250.0, Y: 5.0, Z: 300.0},
			},
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "map sections", string(b))
}

func TestRoundTrip_MapSectionsMultiple(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.MapSections = []QuestMapSectionJSON{
		{
			LoadedStage:   2,
			SpawnMonsters: []uint8{0x06},
			MinionSpawns: []QuestMinionSpawnJSON{
				{Monster: 0x06, SpawnToggle: 1, SpawnAmount: 4, X: 50.0, Y: 0.0, Z: 50.0},
			},
		},
		{
			LoadedStage:   3,
			SpawnMonsters: nil,
			MinionSpawns:  nil,
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "map sections multiple", string(b))
}

func TestRoundTrip_AreaTransitions(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.AreaTransitions = []QuestAreaTransitionsJSON{
		{
			Transitions: []QuestAreaTransitionJSON{
				{
					TargetStageID1: 3,
					StageVariant:   0,
					CurrentX:       100.0,
					CurrentY:       0.0,
					CurrentZ:       50.0,
					TransitionBox:  [5]float32{10.0, 5.0, 10.0, 0.0, 0.0},
					TargetX:        -100.0,
					TargetY:        0.0,
					TargetZ:        -50.0,
					TargetRotation: [2]int16{90, 0},
				},
			},
		},
		{
			// Zone 2: no transitions (null pointer).
			Transitions: nil,
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "area transitions", string(b))
}

func TestRoundTrip_AreaMappings(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	// AreaMappings without AreaTransitions: the parser reads until areaTransitionsPtr,
	// which will be null, so it reads until end of file's mapping section. To make
	// this round-trip cleanly, add both together.
	q.AreaTransitions = []QuestAreaTransitionsJSON{{}, {}}
	q.AreaMappings = []QuestAreaMappingJSON{
		{AreaX: 100.0, AreaZ: 200.0, BaseX: 10.0, BaseZ: 20.0, KnPos: 5.0},
		{AreaX: 300.0, AreaZ: 400.0, BaseX: 30.0, BaseZ: 40.0, KnPos: 7.5},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "area mappings", string(b))
}

func TestRoundTrip_MapInfo(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.MapInfo = &QuestMapInfoJSON{
		MapID:      2,
		ReturnBCID: 1,
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "map info", string(b))
}

func TestRoundTrip_GatheringPoints(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.AreaTransitions = []QuestAreaTransitionsJSON{{}, {}}
	q.GatheringPoints = []QuestAreaGatheringJSON{
		{
			Points: []QuestGatheringPointJSON{
				{X: 50.0, Y: 0.0, Z: 100.0, Range: 3.0, GatheringID: 5, MaxCount: 3, MinCount: 1},
				{X: 150.0, Y: 0.0, Z: 200.0, Range: 3.0, GatheringID: 6, MaxCount: 2, MinCount: 1},
			},
		},
		{
			// Zone 2: no gathering points (null pointer).
			Points: nil,
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "gathering points", string(b))
}

func TestRoundTrip_AreaFacilities(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.AreaTransitions = []QuestAreaTransitionsJSON{{}, {}}
	q.AreaFacilities = []QuestAreaFacilitiesJSON{
		{
			Points: []QuestFacilityPointJSON{
				{Type: 1, X: 10.0, Y: 0.0, Z: -5.0, Range: 2.0, ID: 1},  // cooking
				{Type: 7, X: 20.0, Y: 0.0, Z: -10.0, Range: 3.0, ID: 2}, // red box
			},
		},
		{
			// Zone 2: no facilities (null pointer).
			Points: nil,
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "area facilities", string(b))
}

func TestRoundTrip_SomeStrings(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.SomeString = "extra info"
	q.QuestType = "standard"
	b, _ := json.Marshal(q)
	roundTrip(t, "some strings", string(b))
}

func TestRoundTrip_SomeStringOnly(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.SomeString = "only this string"
	b, _ := json.Marshal(q)
	roundTrip(t, "some string only", string(b))
}

func TestRoundTrip_GatheringTables(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.GatheringTables = []QuestGatheringTableJSON{
		{
			Items: []QuestGatherItemJSON{
				{Rate: 50, Item: 100},
				{Rate: 30, Item: 101},
				{Rate: 20, Item: 102},
			},
		},
		{
			Items: []QuestGatherItemJSON{
				{Rate: 100, Item: 200},
			},
		},
	}
	b, _ := json.Marshal(q)
	roundTrip(t, "gathering tables", string(b))
}

func TestRoundTrip_AllSections(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)

	q.MapSections = []QuestMapSectionJSON{
		{
			LoadedStage:   5,
			SpawnMonsters: []uint8{0x0F},
			MinionSpawns: []QuestMinionSpawnJSON{
				{Monster: 0x0F, SpawnToggle: 1, SpawnAmount: 2, X: 100.0, Y: 0.0, Z: -100.0},
			},
		},
	}
	q.AreaTransitions = []QuestAreaTransitionsJSON{
		{
			Transitions: []QuestAreaTransitionJSON{
				{
					TargetStageID1: 2,
					StageVariant:   0,
					CurrentX:       50.0,
					CurrentY:       0.0,
					CurrentZ:       25.0,
					TransitionBox:  [5]float32{5.0, 5.0, 5.0, 0.0, 0.0},
					TargetX:        -50.0,
					TargetY:        0.0,
					TargetZ:        -25.0,
					TargetRotation: [2]int16{180, 0},
				},
			},
		},
		{Transitions: nil},
	}
	q.AreaMappings = []QuestAreaMappingJSON{
		{AreaX: 100.0, AreaZ: 200.0, BaseX: 10.0, BaseZ: 20.0, KnPos: 1.0},
	}
	q.MapInfo = &QuestMapInfoJSON{MapID: 2, ReturnBCID: 0}
	q.GatheringPoints = []QuestAreaGatheringJSON{
		{
			Points: []QuestGatheringPointJSON{
				{X: 75.0, Y: 0.0, Z: 150.0, Range: 2.5, GatheringID: 3, MaxCount: 3, MinCount: 1},
			},
		},
		{Points: nil},
	}
	q.AreaFacilities = []QuestAreaFacilitiesJSON{
		{
			Points: []QuestFacilityPointJSON{
				{Type: 3, X: 5.0, Y: 0.0, Z: -5.0, Range: 2.0, ID: 10},
			},
		},
		{Points: nil},
	}
	q.SomeString = "test string"
	q.QuestType = "hunt"
	q.GatheringTables = []QuestGatheringTableJSON{
		{
			Items: []QuestGatherItemJSON{
				{Rate: 60, Item: 300},
				{Rate: 40, Item: 301},
			},
		},
	}

	b, _ := json.Marshal(q)
	roundTrip(t, "all sections", string(b))
}

// ── Golden file test ─────────────────────────────────────────────────────────
//
// This test manually constructs expected binary bytes at specific offsets and
// verifies the compiler produces them exactly for minimalQuestJSON.
// Hard-coded values are derived from the documented binary layout.
//
// Layout constants for minimalQuestJSON:
//
//	headerSize      = 68   (0x44)
//	genPropSize     = 66   (0x42)
//	mainPropOffset  = 0x86 (= headerSize + genPropSize)
//	questStringsPtr = 0x1C6 (= mainPropOffset + 320)
func TestGolden_MinimalQuestBinaryLayout(t *testing.T) {
	data, err := CompileQuestJSON([]byte(minimalQuestJSON))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	const (
		mainPropOffset  = 0x86
		questStringsPtr = uint32(mainPropOffset + questBodyLenZZ) // 0x1C6
	)

	// ── Header (0x00–0x43) ───────────────────────────────────────────────
	assertU32(t, data, 0x00, mainPropOffset, "questTypeFlagsPtr")
	assertU16(t, data, 0x10, 0, "subSupplyBoxPtr (unused)")
	assertByte(t, data, 0x12, 0, "hidden")
	assertByte(t, data, 0x13, 0, "subSupplyBoxLen")
	assertU32(t, data, 0x14, 0, "questAreaPtr (null)")
	assertU32(t, data, 0x1C, 0, "areaTransitionsPtr (null)")
	assertU32(t, data, 0x20, 0, "areaMappingPtr (null)")
	assertU32(t, data, 0x24, 0, "mapInfoPtr (null)")
	assertU32(t, data, 0x28, 0, "gatheringPointsPtr (null)")
	assertU32(t, data, 0x2C, 0, "areaFacilitiesPtr (null)")
	assertU32(t, data, 0x30, 0, "someStringsPtr (null)")
	assertU32(t, data, 0x38, 0, "gatheringTablesPtr (null)")
	assertU32(t, data, 0x3C, 0, "fixedCoords2Ptr (null)")
	assertU32(t, data, 0x40, 0, "fixedInfoPtr (null)")

	loadedStagesPtr := binary.LittleEndian.Uint32(data[0x04:])
	unk34Ptr := binary.LittleEndian.Uint32(data[0x34:])
	if unk34Ptr != loadedStagesPtr+16 {
		t.Errorf("unk34Ptr 0x%X != loadedStagesPtr+16 (0x%X); expected exactly 1 stage × 16 bytes",
			unk34Ptr, loadedStagesPtr+16)
	}

	// ── General Quest Properties (0x44–0x85) ────────────────────────────
	assertU16(t, data, 0x44, 100, "monsterSizeMulti")
	assertU16(t, data, 0x46, 0, "sizeRange")
	assertU32(t, data, 0x48, 0, "statTable1")
	assertU32(t, data, 0x4C, 120, "mainRankPoints")
	assertU32(t, data, 0x50, 0, "unknown@0x50")
	assertU32(t, data, 0x54, 60, "subARankPoints")
	assertU32(t, data, 0x58, 0, "subBRankPoints")
	assertU32(t, data, 0x5C, 0, "questTypeID@0x5C")
	assertByte(t, data, 0x60, 0, "padding@0x60")
	assertByte(t, data, 0x61, 0, "statTable2")
	// 0x62–0x72: padding (17 bytes of zeros)
	for i := 0x62; i <= 0x72; i++ {
		assertByte(t, data, i, 0, "padding")
	}
	assertByte(t, data, 0x73, 0, "questKn1")
	assertU16(t, data, 0x74, 0, "questKn2")
	assertU16(t, data, 0x76, 0, "questKn3")
	assertU16(t, data, 0x78, 0, "gatheringTablesQty")
	assertByte(t, data, 0x7C, 0, "area1Zones")
	assertByte(t, data, 0x7D, 0, "area2Zones")
	assertByte(t, data, 0x7E, 0, "area3Zones")
	assertByte(t, data, 0x7F, 0, "area4Zones")

	// ── Main Quest Properties (0x86–0x1C5) ──────────────────────────────
	mp := mainPropOffset
	assertByte(t, data, mp+0x00, 0, "mp.unknown@+0x00")
	assertByte(t, data, mp+0x01, 0, "mp.musicMode")
	assertByte(t, data, mp+0x02, 0, "mp.localeFlags")
	assertByte(t, data, mp+0x08, 0, "mp.rankBand lo") // rankBand = 0
	assertByte(t, data, mp+0x09, 0, "mp.rankBand hi")
	// questFee = 500 → LE bytes: 0xF4 0x01 0x00 0x00
	assertU32(t, data, mp+0x0C, 500, "mp.questFee")
	// rewardMain = 5000 → LE: 0x88 0x13 0x00 0x00
	assertU32(t, data, mp+0x10, 5000, "mp.rewardMain")
	assertU32(t, data, mp+0x14, 0, "mp.cartsOrReduction")
	// rewardA = 1000 → LE: 0xE8 0x03
	assertU16(t, data, mp+0x18, 1000, "mp.rewardA")
	assertU16(t, data, mp+0x1A, 0, "mp.padding@+0x1A")
	assertU16(t, data, mp+0x1C, 0, "mp.rewardB")
	assertU16(t, data, mp+0x1E, 0, "mp.hardHRReq")
	// questTime = 50 × 60 × 30 = 90000 → LE: 0x10 0x5F 0x01 0x00
	assertU32(t, data, mp+0x20, 90000, "mp.questTime")
	assertU32(t, data, mp+0x24, 2, "mp.questMap")
	assertU32(t, data, mp+0x28, uint32(questStringsPtr), "mp.questStringsPtr")
	assertU16(t, data, mp+0x2C, 0, "mp.unknown@+0x2C")
	assertU16(t, data, mp+0x2E, 1, "mp.questID")

	// Objective[0]: hunt, target=11, count=1
	assertU32(t, data, mp+0x30, questObjHunt, "obj[0].goalType")
	assertByte(t, data, mp+0x34, 11, "obj[0].target")
	assertByte(t, data, mp+0x35, 0, "obj[0].pad")
	assertU16(t, data, mp+0x36, 1, "obj[0].count")

	// Objective[1]: deliver, target=149, count=3
	assertU32(t, data, mp+0x38, questObjDeliver, "obj[1].goalType")
	assertU16(t, data, mp+0x3C, 149, "obj[1].target")
	assertU16(t, data, mp+0x3E, 3, "obj[1].count")

	// Objective[2]: none
	assertU32(t, data, mp+0x40, questObjNone, "obj[2].goalType")
	assertU32(t, data, mp+0x44, 0, "obj[2].trailing pad")

	assertU16(t, data, mp+0x4C, 0, "mp.joinRankMin")
	assertU16(t, data, mp+0x4E, 0, "mp.joinRankMax")
	assertU16(t, data, mp+0x50, 0, "mp.postRankMin")
	assertU16(t, data, mp+0x52, 0, "mp.postRankMax")

	// forced equip: 6 slots × 4 × 2 = 48 bytes, all zero
	for i := 0; i < 48; i++ {
		assertByte(t, data, mp+0x5C+i, 0, "forced equip zero")
	}

	assertByte(t, data, mp+0x97, 0, "mp.questVariant1")
	assertByte(t, data, mp+0x98, 0, "mp.questVariant2")
	assertByte(t, data, mp+0x99, 0, "mp.questVariant3")
	assertByte(t, data, mp+0x9A, 0, "mp.questVariant4")

	// ── QuestText pointer table (0x1C6–0x1E5) ───────────────────────────
	for i := 0; i < 8; i++ {
		off := int(questStringsPtr) + i*4
		strPtr := int(binary.LittleEndian.Uint32(data[off:]))
		if strPtr < 0 || strPtr >= len(data) {
			t.Errorf("string[%d] ptr 0x%X out of bounds (len=%d)", i, strPtr, len(data))
		}
	}

	// Title pointer → "Test Quest"
	titlePtr := int(binary.LittleEndian.Uint32(data[int(questStringsPtr):]))
	end := titlePtr
	for end < len(data) && data[end] != 0 {
		end++
	}
	if string(data[titlePtr:end]) != "Test Quest" {
		t.Errorf("title bytes = %q, want %q", data[titlePtr:end], "Test Quest")
	}

	// ── Stage entry (1 stage: stageID=2) ────────────────────────────────
	assertU32(t, data, int(loadedStagesPtr), 2, "stage[0].stageID")
	for i := 1; i < 16; i++ {
		assertByte(t, data, int(loadedStagesPtr)+i, 0, "stage padding")
	}

	// ── Supply box: main[0] = {item:1, qty:5} ───────────────────────────
	supplyBoxPtr := int(binary.LittleEndian.Uint32(data[0x08:]))
	assertU16(t, data, supplyBoxPtr, 1, "supply_main[0].item")
	assertU16(t, data, supplyBoxPtr+2, 5, "supply_main[0].quantity")
	for i := 1; i < 24; i++ {
		assertU32(t, data, supplyBoxPtr+i*4, 0, "supply_main slot empty")
	}
	subABase := supplyBoxPtr + 24*4
	for i := 0; i < 8; i++ {
		assertU32(t, data, subABase+i*4, 0, "supply_subA slot empty")
	}
	subBBase := subABase + 8*4
	for i := 0; i < 8; i++ {
		assertU32(t, data, subBBase+i*4, 0, "supply_subB slot empty")
	}

	// ── Reward table ────────────────────────────────────────────────────
	rewardPtr := int(binary.LittleEndian.Uint32(data[0x0C:]))
	assertByte(t, data, rewardPtr, 1, "reward header[0].tableID")
	assertByte(t, data, rewardPtr+1, 0, "reward header[0].pad1")
	assertU16(t, data, rewardPtr+2, 0, "reward header[0].pad2")
	// headerArraySize = 1×8 + 2 = 10
	assertU32(t, data, rewardPtr+4, 10, "reward header[0].tableOffset")
	assertU16(t, data, rewardPtr+8, 0xFFFF, "reward header terminator")
	itemsBase := rewardPtr + 10
	assertU16(t, data, itemsBase, 50, "reward[0].items[0].rate")
	assertU16(t, data, itemsBase+2, 149, "reward[0].items[0].item")
	assertU16(t, data, itemsBase+4, 1, "reward[0].items[0].quantity")
	assertU16(t, data, itemsBase+6, 30, "reward[0].items[1].rate")
	assertU16(t, data, itemsBase+8, 153, "reward[0].items[1].item")
	assertU16(t, data, itemsBase+10, 1, "reward[0].items[1].quantity")
	assertU16(t, data, itemsBase+12, 0xFFFF, "reward item terminator")

	// ── Large monster spawn ──────────────────────────────────────────────
	largeMonsterPtr := int(binary.LittleEndian.Uint32(data[0x18:]))
	assertByte(t, data, largeMonsterPtr, 11, "monster[0].id")
	assertByte(t, data, largeMonsterPtr+1, 0, "monster[0].pad1")
	assertByte(t, data, largeMonsterPtr+2, 0, "monster[0].pad2")
	assertByte(t, data, largeMonsterPtr+3, 0, "monster[0].pad3")
	assertU32(t, data, largeMonsterPtr+4, 1, "monster[0].spawnAmount")
	assertU32(t, data, largeMonsterPtr+8, 5, "monster[0].spawnStage")
	for i := 0; i < 16; i++ {
		assertByte(t, data, largeMonsterPtr+0x0C+i, 0, "monster[0].pad16")
	}
	assertU32(t, data, largeMonsterPtr+0x1C, 180, "monster[0].orientation")
	assertF32(t, data, largeMonsterPtr+0x20, 1500.0, "monster[0].x")
	assertF32(t, data, largeMonsterPtr+0x24, 0.0, "monster[0].y")
	assertF32(t, data, largeMonsterPtr+0x28, -2000.0, "monster[0].z")
	for i := 0; i < 16; i++ {
		assertByte(t, data, largeMonsterPtr+0x2C+i, 0, "monster[0].trailing_pad")
	}
	assertByte(t, data, largeMonsterPtr+60, 0xFF, "monster list terminator")

	// ── Total file size ──────────────────────────────────────────────────
	minExpectedLen := largeMonsterPtr + 61
	if len(data) < minExpectedLen {
		t.Errorf("file too short: len=%d, need at least %d", len(data), minExpectedLen)
	}
}

// ── Golden test: generalQuestProperties with populated sections ───────────────

func TestGolden_GeneralQuestPropertiesCounts(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.AreaTransitions = []QuestAreaTransitionsJSON{{}, {}, {}}
	q.GatheringTables = []QuestGatheringTableJSON{
		{Items: []QuestGatherItemJSON{{Rate: 100, Item: 1}}},
		{Items: []QuestGatherItemJSON{{Rate: 100, Item: 2}}},
	}

	b, _ := json.Marshal(q)
	data, err := CompileQuestJSON(b)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	// area1Zones at 0x7C should be 3.
	assertByte(t, data, 0x7C, 3, "area1Zones")
	// gatheringTablesQty at 0x78 should be 2.
	assertU16(t, data, 0x78, 2, "gatheringTablesQty")
}

// ── Golden test: map sections binary layout ───────────────────────────────────

func TestGolden_MapSectionsBinaryLayout(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.MapSections = []QuestMapSectionJSON{
		{
			LoadedStage:   7,
			SpawnMonsters: []uint8{0x0B}, // Rathalos
			MinionSpawns: []QuestMinionSpawnJSON{
				{Monster: 0x0B, SpawnToggle: 1, SpawnAmount: 2, X: 500.0, Y: 10.0, Z: -300.0},
			},
		},
	}

	data, err := CompileQuestJSON(func() []byte { b, _ := json.Marshal(q); return b }())
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	// questAreaPtr must be non-null.
	questAreaPtr := int(binary.LittleEndian.Uint32(data[0x14:]))
	if questAreaPtr == 0 {
		t.Fatal("questAreaPtr is null, expected non-null")
	}

	// First entry in pointer array must be non-null (points to mapSection).
	sectionPtr := int(binary.LittleEndian.Uint32(data[questAreaPtr:]))
	if sectionPtr == 0 {
		t.Fatal("mapSection[0] ptr is null")
	}

	// Terminator after the pointer.
	terminatorOff := questAreaPtr + 4
	if terminatorOff+4 > len(data) {
		t.Fatalf("terminator out of bounds")
	}
	termVal := binary.LittleEndian.Uint32(data[terminatorOff:])
	if termVal != 0 {
		t.Errorf("pointer array terminator = 0x%08X, want 0", termVal)
	}

	// mapSection at sectionPtr: loadedStage = 7.
	if sectionPtr+16 > len(data) {
		t.Fatalf("mapSection out of bounds")
	}
	loadedStage := binary.LittleEndian.Uint32(data[sectionPtr:])
	if loadedStage != 7 {
		t.Errorf("mapSection.loadedStage = %d, want 7", loadedStage)
	}

	// spawnTypes and spawnStats ptrs must be non-null.
	spawnTypesPtr := int(binary.LittleEndian.Uint32(data[sectionPtr+8:]))
	spawnStatsPtr := int(binary.LittleEndian.Uint32(data[sectionPtr+12:]))
	if spawnTypesPtr == 0 {
		t.Fatal("spawnTypesPtr is null")
	}
	if spawnStatsPtr == 0 {
		t.Fatal("spawnStatsPtr is null")
	}

	// spawnTypes: first entry = Rathalos (0x0B) + pad[3], then 0xFFFF terminator.
	if spawnTypesPtr+6 > len(data) {
		t.Fatalf("spawnTypes data out of bounds")
	}
	if data[spawnTypesPtr] != 0x0B {
		t.Errorf("spawnTypes[0].monster = 0x%02X, want 0x0B", data[spawnTypesPtr])
	}
	termU16 := binary.LittleEndian.Uint16(data[spawnTypesPtr+4:])
	if termU16 != 0xFFFF {
		t.Errorf("spawnTypes terminator = 0x%04X, want 0xFFFF", termU16)
	}

	// spawnStats: first entry monster = Rathalos (0x0B).
	if data[spawnStatsPtr] != 0x0B {
		t.Errorf("spawnStats[0].monster = 0x%02X, want 0x0B", data[spawnStatsPtr])
	}
	// spawnToggle at +2 = 1.
	spawnToggle := binary.LittleEndian.Uint16(data[spawnStatsPtr+2:])
	if spawnToggle != 1 {
		t.Errorf("spawnStats[0].spawnToggle = %d, want 1", spawnToggle)
	}
	// spawnAmount at +4 = 2.
	spawnAmount := binary.LittleEndian.Uint32(data[spawnStatsPtr+4:])
	if spawnAmount != 2 {
		t.Errorf("spawnStats[0].spawnAmount = %d, want 2", spawnAmount)
	}
	// xPos at +0x20 = 500.0.
	xBits := binary.LittleEndian.Uint32(data[spawnStatsPtr+0x20:])
	xPos := math.Float32frombits(xBits)
	if xPos != 500.0 {
		t.Errorf("spawnStats[0].x = %v, want 500.0", xPos)
	}
}

// ── Golden test: gathering tables binary layout ───────────────────────────────

func TestGolden_GatheringTablesBinaryLayout(t *testing.T) {
	var q QuestJSON
	_ = json.Unmarshal([]byte(minimalQuestJSON), &q)
	q.GatheringTables = []QuestGatheringTableJSON{
		{Items: []QuestGatherItemJSON{{Rate: 75, Item: 500}, {Rate: 25, Item: 501}}},
	}

	b, _ := json.Marshal(q)
	data, err := CompileQuestJSON(b)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	// gatheringTablesPtr must be non-null.
	gatherTablesPtr := int(binary.LittleEndian.Uint32(data[0x38:]))
	if gatherTablesPtr == 0 {
		t.Fatal("gatheringTablesPtr is null")
	}

	// gatheringTablesQty at 0x78 must be 1.
	assertU16(t, data, 0x78, 1, "gatheringTablesQty")

	// Table 0: pointer to item data.
	tblPtr := int(binary.LittleEndian.Uint32(data[gatherTablesPtr:]))
	if tblPtr == 0 {
		t.Fatal("gathering table[0] ptr is null")
	}

	// Item 0: rate=75, item=500.
	if tblPtr+4 > len(data) {
		t.Fatalf("gathering table items out of bounds")
	}
	rate0 := binary.LittleEndian.Uint16(data[tblPtr:])
	item0 := binary.LittleEndian.Uint16(data[tblPtr+2:])
	if rate0 != 75 {
		t.Errorf("table[0].items[0].rate = %d, want 75", rate0)
	}
	if item0 != 500 {
		t.Errorf("table[0].items[0].item = %d, want 500", item0)
	}

	// Item 1: rate=25, item=501.
	rate1 := binary.LittleEndian.Uint16(data[tblPtr+4:])
	item1 := binary.LittleEndian.Uint16(data[tblPtr+6:])
	if rate1 != 25 {
		t.Errorf("table[0].items[1].rate = %d, want 25", rate1)
	}
	if item1 != 501 {
		t.Errorf("table[0].items[1].item = %d, want 501", item1)
	}

	// Terminator: 0xFFFF.
	term := binary.LittleEndian.Uint16(data[tblPtr+8:])
	if term != 0xFFFF {
		t.Errorf("gathering table terminator = 0x%04X, want 0xFFFF", term)
	}
}

// ── Objective encoding golden tests ─────────────────────────────────────────

func TestGolden_ObjectiveEncoding(t *testing.T) {
	cases := []struct {
		name    string
		obj     QuestObjectiveJSON
		wantRaw [8]byte // goalType(4) + payload(4)
	}{
		{
			name: "none",
			obj:  QuestObjectiveJSON{Type: "none"},
			// goalType=0x00000000, trailing zeros
			wantRaw: [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "hunt target=11 count=1",
			obj:  QuestObjectiveJSON{Type: "hunt", Target: 11, Count: 1},
			// goalType=0x00000001, u8(11)=0x0B, u8(0), u16(1)=0x01 0x00
			wantRaw: [8]byte{0x01, 0x00, 0x00, 0x00, 0x0B, 0x00, 0x01, 0x00},
		},
		{
			name: "capture target=11 count=1",
			obj:  QuestObjectiveJSON{Type: "capture", Target: 11, Count: 1},
			// goalType=0x00000101
			wantRaw: [8]byte{0x01, 0x01, 0x00, 0x00, 0x0B, 0x00, 0x01, 0x00},
		},
		{
			name: "slay target=37 count=3",
			obj:  QuestObjectiveJSON{Type: "slay", Target: 37, Count: 3},
			// goalType=0x00000201, u8(37)=0x25, u8(0), u16(3)=0x03 0x00
			wantRaw: [8]byte{0x01, 0x02, 0x00, 0x00, 0x25, 0x00, 0x03, 0x00},
		},
		{
			name: "deliver target=149 count=3",
			obj:  QuestObjectiveJSON{Type: "deliver", Target: 149, Count: 3},
			// goalType=0x00000002, u16(149)=0x95 0x00, u16(3)=0x03 0x00
			wantRaw: [8]byte{0x02, 0x00, 0x00, 0x00, 0x95, 0x00, 0x03, 0x00},
		},
		{
			name: "break_part target=11 part=3",
			obj:  QuestObjectiveJSON{Type: "break_part", Target: 11, Part: 3},
			// goalType=0x00004004, u8(11)=0x0B, u8(0), u16(part=3)=0x03 0x00
			wantRaw: [8]byte{0x04, 0x40, 0x00, 0x00, 0x0B, 0x00, 0x03, 0x00},
		},
		{
			name: "slay_all",
			obj:  QuestObjectiveJSON{Type: "slay_all"},
			// goalType=0x00040000 — slay_all uses default (deliver) path: u16(target), u16(count)
			wantRaw: [8]byte{0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := objectiveBytes(tc.obj)
			if err != nil {
				t.Fatalf("objectiveBytes: %v", err)
			}
			if len(got) != 8 {
				t.Fatalf("len(got) = %d, want 8", len(got))
			}
			if [8]byte(got) != tc.wantRaw {
				t.Errorf("bytes = %v, want %v", got, tc.wantRaw[:])
			}
		})
	}
}

// ── Helper assertions ────────────────────────────────────────────────────────

func assertByte(t *testing.T, data []byte, off int, want byte, label string) {
	t.Helper()
	if off >= len(data) {
		t.Errorf("%s @ 0x%X: out of bounds (len=%d)", label, off, len(data))
		return
	}
	if data[off] != want {
		t.Errorf("%s @ 0x%X: got 0x%02X, want 0x%02X", label, off, data[off], want)
	}
}

func assertU16(t *testing.T, data []byte, off int, want uint16, label string) {
	t.Helper()
	if off+2 > len(data) {
		t.Errorf("%s @ 0x%X: out of bounds (len=%d)", label, off, len(data))
		return
	}
	got := binary.LittleEndian.Uint16(data[off:])
	if got != want {
		t.Errorf("%s @ 0x%X: got %d (0x%04X), want %d (0x%04X)", label, off, got, got, want, want)
	}
}

func assertU32(t *testing.T, data []byte, off int, want uint32, label string) {
	t.Helper()
	if off+4 > len(data) {
		t.Errorf("%s @ 0x%X: out of bounds (len=%d)", label, off, len(data))
		return
	}
	got := binary.LittleEndian.Uint32(data[off:])
	if got != want {
		t.Errorf("%s @ 0x%X: got %d (0x%08X), want %d (0x%08X)", label, off, got, got, want, want)
	}
}

func assertF32(t *testing.T, data []byte, off int, want float32, label string) {
	t.Helper()
	if off+4 > len(data) {
		t.Errorf("%s @ 0x%X: out of bounds (len=%d)", label, off, len(data))
		return
	}
	got := math.Float32frombits(binary.LittleEndian.Uint32(data[off:]))
	if got != want {
		t.Errorf("%s @ 0x%X: got %v, want %v", label, off, got, want)
	}
}
