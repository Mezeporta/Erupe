package channelserver

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// sampleRengokuConfig returns a small but complete RengokuConfig for tests.
func sampleRengokuConfig() RengokuConfig {
	spawnTables := []SpawnTableConfig{
		{Monster1ID: 101, Monster1Variant: 0, Monster2ID: 102, Monster2Variant: 1,
			StatTable: 3, SpawnWeighting: 10},
		{Monster1ID: 103, Monster1Variant: 2, Monster2ID: 104, Monster2Variant: 0,
			SpawnWeighting: 20},
	}
	floors := []FloorConfig{
		{FloorNumber: 1, SpawnTableIndex: 0, PointMulti1: 1.0, PointMulti2: 1.5},
		{FloorNumber: 2, SpawnTableIndex: 1, PointMulti1: 1.2, PointMulti2: 2.0},
		{FloorNumber: 3, SpawnTableIndex: 0, PointMulti1: 1.5, PointMulti2: 2.5, FinalLoop: 1},
	}
	soloFloors := []FloorConfig{
		{FloorNumber: 1, SpawnTableIndex: 0, PointMulti1: 1.0, PointMulti2: 1.5},
		{FloorNumber: 2, SpawnTableIndex: 0, PointMulti1: 1.2, PointMulti2: 2.0},
	}
	return RengokuConfig{
		MultiRoad: RoadConfig{Floors: floors, SpawnTables: spawnTables},
		SoloRoad:  RoadConfig{Floors: soloFloors, SpawnTables: spawnTables[1:]},
	}
}

// TestBuildRengokuBinary_RoundTrip builds a binary from a config and verifies
// that parseRengokuBinary accepts it and reports the expected summary.
func TestBuildRengokuBinary_RoundTrip(t *testing.T) {
	cfg := sampleRengokuConfig()

	bin, err := BuildRengokuBinary(cfg)
	if err != nil {
		t.Fatalf("BuildRengokuBinary: %v", err)
	}

	info, err := parseRengokuBinary(bin)
	if err != nil {
		t.Fatalf("parseRengokuBinary on built binary: %v", err)
	}

	if info.MultiFloors != len(cfg.MultiRoad.Floors) {
		t.Errorf("MultiFloors = %d, want %d", info.MultiFloors, len(cfg.MultiRoad.Floors))
	}
	if info.MultiSpawnTables != len(cfg.MultiRoad.SpawnTables) {
		t.Errorf("MultiSpawnTables = %d, want %d", info.MultiSpawnTables, len(cfg.MultiRoad.SpawnTables))
	}
	if info.SoloFloors != len(cfg.SoloRoad.Floors) {
		t.Errorf("SoloFloors = %d, want %d", info.SoloFloors, len(cfg.SoloRoad.Floors))
	}
	if info.SoloSpawnTables != len(cfg.SoloRoad.SpawnTables) {
		t.Errorf("SoloSpawnTables = %d, want %d", info.SoloSpawnTables, len(cfg.SoloRoad.SpawnTables))
	}
	// Unique monsters: multi has 101,102,103,104; solo has 103,104 → 4 total
	if info.UniqueMonsters != 4 {
		t.Errorf("UniqueMonsters = %d, want 4", info.UniqueMonsters)
	}
}

// TestBuildRengokuBinary_FloatFields verifies that PointMulti1/2 values
// survive the binary encoding intact.
func TestBuildRengokuBinary_FloatFields(t *testing.T) {
	cfg := RengokuConfig{
		MultiRoad: RoadConfig{
			Floors: []FloorConfig{
				{FloorNumber: 1, SpawnTableIndex: 0, PointMulti1: 1.25, PointMulti2: 3.75},
			},
			SpawnTables: []SpawnTableConfig{{Monster1ID: 1}},
		},
		SoloRoad: RoadConfig{
			Floors:      []FloorConfig{{FloorNumber: 1, SpawnTableIndex: 0}},
			SpawnTables: []SpawnTableConfig{{Monster1ID: 2}},
		},
	}

	bin, err := BuildRengokuBinary(cfg)
	if err != nil {
		t.Fatalf("BuildRengokuBinary: %v", err)
	}

	// Re-parse the binary and check that we can read back the float fields.
	// The floor stats for multiDef start at rengokuMinSize (0x44).
	// Layout: floorNumber(4) + spawnTableIndex(4) + unk0(4) + pointMulti1(4) + pointMulti2(4)
	floorBase := rengokuMinSize // 0x44
	pm1Bits := uint32(bin[floorBase+12]) | uint32(bin[floorBase+13])<<8 |
		uint32(bin[floorBase+14])<<16 | uint32(bin[floorBase+15])<<24
	pm2Bits := uint32(bin[floorBase+16]) | uint32(bin[floorBase+17])<<8 |
		uint32(bin[floorBase+18])<<16 | uint32(bin[floorBase+19])<<24

	if got := math.Float32frombits(pm1Bits); got != 1.25 {
		t.Errorf("PointMulti1 = %f, want 1.25", got)
	}
	if got := math.Float32frombits(pm2Bits); got != 3.75 {
		t.Errorf("PointMulti2 = %f, want 3.75", got)
	}
}

// TestBuildRengokuBinary_ValidationErrors verifies that out-of-range
// spawn_table_index values are caught before the binary is built.
func TestBuildRengokuBinary_ValidationErrors(t *testing.T) {
	cases := []struct {
		name    string
		cfg     RengokuConfig
		wantErr string
	}{
		{
			name: "multi_index_out_of_range",
			cfg: RengokuConfig{
				MultiRoad: RoadConfig{
					Floors:      []FloorConfig{{FloorNumber: 1, SpawnTableIndex: 5}},
					SpawnTables: []SpawnTableConfig{{Monster1ID: 1}},
				},
				SoloRoad: RoadConfig{
					Floors:      []FloorConfig{{FloorNumber: 1, SpawnTableIndex: 0}},
					SpawnTables: []SpawnTableConfig{{Monster1ID: 2}},
				},
			},
			wantErr: "multi_road",
		},
		{
			name: "solo_index_out_of_range",
			cfg: RengokuConfig{
				MultiRoad: RoadConfig{
					Floors:      []FloorConfig{{FloorNumber: 1, SpawnTableIndex: 0}},
					SpawnTables: []SpawnTableConfig{{Monster1ID: 1}},
				},
				SoloRoad: RoadConfig{
					Floors:      []FloorConfig{{FloorNumber: 1, SpawnTableIndex: 99}},
					SpawnTables: []SpawnTableConfig{{Monster1ID: 2}},
				},
			},
			wantErr: "solo_road",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := BuildRengokuBinary(tc.cfg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestLoadRengokuBinary_BinPreferredOverJSON writes both a JSON file and a
// .bin file and verifies that the .bin source is used (consistent with the
// quest and scenario loaders).
func TestLoadRengokuBinary_BinPreferredOverJSON(t *testing.T) {
	dir := t.TempDir()
	logger, _ := zap.NewDevelopment()

	// Write a valid rengoku_data.json (would produce a much larger binary).
	cfg := sampleRengokuConfig()
	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rengoku_data.json"), jsonBytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Write a minimal valid-magic .bin — should be preferred over JSON.
	binData := make([]byte, 16) // 16-byte ECD header, zero payload
	binData[0], binData[1], binData[2], binData[3] = 0x65, 0x63, 0x64, 0x1A
	if err := os.WriteFile(filepath.Join(dir, "rengoku_data.bin"), binData, 0644); err != nil {
		t.Fatal(err)
	}

	result := loadRengokuBinary(dir, logger)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The JSON-built binary would be much larger; 16 bytes confirms .bin was used.
	if len(result) != 16 {
		t.Errorf("result is %d bytes — looks like JSON was used instead of .bin", len(result))
	}
}

// TestLoadRengokuBinary_JSONFallbackWhenNoBin verifies that when no .bin file
// is present, loadRengokuBinary falls back to rengoku_data.json.
func TestLoadRengokuBinary_JSONFallbackWhenNoBin(t *testing.T) {
	dir := t.TempDir()
	logger, _ := zap.NewDevelopment()

	cfg := sampleRengokuConfig()
	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rengoku_data.json"), jsonBytes, 0644); err != nil {
		t.Fatal(err)
	}

	result := loadRengokuBinary(dir, logger)
	if result == nil {
		t.Fatal("expected fallback to JSON, got nil")
	}
	// JSON-built result is much larger than 16 bytes.
	if len(result) <= 16 {
		t.Errorf("result is %d bytes — JSON fallback likely did not run", len(result))
	}
}
