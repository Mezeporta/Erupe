package channelserver

/*
	JSON-based rengoku_data.bin builder.

	Operators can place rengoku_data.json in the bin/ directory instead of
	(or alongside) rengoku_data.bin. When the JSON file is found it takes
	precedence: it is parsed, validated, assembled into the raw binary layout,
	and ECD-encrypted before being cached. The .bin file is used as a fallback.

	Binary layout produced by BuildRengokuBinary:
	  0x00–0x13  header  (20 bytes: magic + version + zeros)
	  0x14–0x2B  multiDef RoadMode  (24 bytes)
	  0x2C–0x43  soloDef  RoadMode  (24 bytes)
	  -- multi road data --
	  floorStats[]          (floorStatsCount × 24 bytes)
	  spawnTablePtrs[]      (spawnTablePtrCount × 4 bytes)
	  spawnCountPtrs[]      (spawnTablePtrCount × 4 bytes, zeroed)
	  spawnTables[]         (spawnTablePtrCount × 32 bytes)
	  -- solo road data --  (same sub-layout)
*/

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"erupe-ce/common/decryption"

	"go.uber.org/zap"
)

// ─── JSON schema ────────────────────────────────────────────────────────────

// RengokuConfig is the top-level JSON structure for rengoku_data.json.
type RengokuConfig struct {
	MultiRoad RoadConfig `json:"multi_road"`
	SoloRoad  RoadConfig `json:"solo_road"`
}

// RoadConfig describes one road mode (multi or solo) with its floors and
// spawn tables. Floors reference spawn tables by zero-based index.
type RoadConfig struct {
	Floors      []FloorConfig      `json:"floors"`
	SpawnTables []SpawnTableConfig `json:"spawn_tables"`
}

// FloorConfig describes one floor within a road mode.
//
//   - SpawnTableIndex: zero-based index into this road's SpawnTables slice,
//     selecting which monster configuration is active on this floor.
//   - PointMulti1/2: point multipliers applied to rewards on this floor.
//   - FinalLoop: non-zero on the last floor of a loop cycle.
type FloorConfig struct {
	FloorNumber     uint32  `json:"floor_number"`
	SpawnTableIndex uint32  `json:"spawn_table_index"`
	Unk0            uint32  `json:"unk0,omitempty"`
	PointMulti1     float32 `json:"point_multi_1"`
	PointMulti2     float32 `json:"point_multi_2"`
	FinalLoop       uint32  `json:"final_loop,omitempty"`
}

// SpawnTableConfig describes the two monsters that appear together on a floor.
type SpawnTableConfig struct {
	Monster1ID      uint32 `json:"monster1_id"`
	Monster1Variant uint32 `json:"monster1_variant,omitempty"`
	Monster2ID      uint32 `json:"monster2_id"`
	Monster2Variant uint32 `json:"monster2_variant,omitempty"`
	StatTable       uint32 `json:"stat_table,omitempty"`
	MapZoneOverride uint32 `json:"map_zone_override,omitempty"`
	SpawnWeighting  uint32 `json:"spawn_weighting,omitempty"`
	AdditionalFlag  uint32 `json:"additional_flag,omitempty"`
}

// ─── Builder ─────────────────────────────────────────────────────────────────

// BuildRengokuBinary assembles a raw (unencrypted, uncompressed) rengoku
// binary from a RengokuConfig. The result can be passed to EncodeECD and
// served directly to clients.
func BuildRengokuBinary(cfg RengokuConfig) ([]byte, error) {
	if err := validateRengokuConfig(cfg); err != nil {
		return nil, err
	}

	// ── Offset plan ──────────────────────────────────────────────────────────
	// Fixed regions: header (0x14) + two RoadModes (2×24) = 0x44
	const dataStart = uint32(rengokuMinSize) // 0x44

	// Multi road sections
	mFloorOff := dataStart
	mFloorSz := uint32(len(cfg.MultiRoad.Floors)) * floorStatsByteSize
	mPtrsOff := mFloorOff + mFloorSz
	mPtrsSz := uint32(len(cfg.MultiRoad.SpawnTables)) * spawnPtrEntrySize
	mCntOff := mPtrsOff + mPtrsSz
	mCntSz := uint32(len(cfg.MultiRoad.SpawnTables)) * spawnPtrEntrySize
	mTablesOff := mCntOff + mCntSz
	mTablesSz := uint32(len(cfg.MultiRoad.SpawnTables)) * spawnTableByteSize

	// Solo road sections (appended directly after multi)
	sFloorOff := mTablesOff + mTablesSz
	sFloorSz := uint32(len(cfg.SoloRoad.Floors)) * floorStatsByteSize
	sPtrsOff := sFloorOff + sFloorSz
	sPtrsSz := uint32(len(cfg.SoloRoad.SpawnTables)) * spawnPtrEntrySize
	sCntOff := sPtrsOff + sPtrsSz
	sCntSz := uint32(len(cfg.SoloRoad.SpawnTables)) * spawnPtrEntrySize
	sTablesOff := sCntOff + sCntSz
	sTablesSz := uint32(len(cfg.SoloRoad.SpawnTables)) * spawnTableByteSize

	totalSize := sTablesOff + sTablesSz
	buf := make([]byte, totalSize)

	// ── Header ───────────────────────────────────────────────────────────────
	buf[0], buf[1], buf[2], buf[3] = 'r', 'e', 'f', 0x1A
	buf[4] = 1 // version

	le := binary.LittleEndian

	// ── RoadMode structs ─────────────────────────────────────────────────────
	writeRoadMode(buf, 0x14, le, RoadModeFields{
		FloorCount:   uint32(len(cfg.MultiRoad.Floors)),
		SpawnCount:   uint32(len(cfg.MultiRoad.SpawnTables)),
		TablePtrCnt:  uint32(len(cfg.MultiRoad.SpawnTables)),
		FloorPtr:     mFloorOff,
		TablePtrsPtr: mPtrsOff,
		CountPtrsPtr: mCntOff,
	})
	writeRoadMode(buf, 0x2C, le, RoadModeFields{
		FloorCount:   uint32(len(cfg.SoloRoad.Floors)),
		SpawnCount:   uint32(len(cfg.SoloRoad.SpawnTables)),
		TablePtrCnt:  uint32(len(cfg.SoloRoad.SpawnTables)),
		FloorPtr:     sFloorOff,
		TablePtrsPtr: sPtrsOff,
		CountPtrsPtr: sCntOff,
	})

	// ── Data sections ────────────────────────────────────────────────────────
	writeFloors(buf, cfg.MultiRoad.Floors, mFloorOff, le)
	writeSpawnSection(buf, cfg.MultiRoad.SpawnTables, mPtrsOff, mTablesOff, le)

	writeFloors(buf, cfg.SoloRoad.Floors, sFloorOff, le)
	writeSpawnSection(buf, cfg.SoloRoad.SpawnTables, sPtrsOff, sTablesOff, le)

	return buf, nil
}

// RoadModeFields carries the computed field values for one RoadMode struct.
type RoadModeFields struct {
	FloorCount, SpawnCount, TablePtrCnt  uint32
	FloorPtr, TablePtrsPtr, CountPtrsPtr uint32
}

func writeRoadMode(buf []byte, offset int, le binary.ByteOrder, f RoadModeFields) {
	le.PutUint32(buf[offset:], f.FloorCount)
	le.PutUint32(buf[offset+4:], f.SpawnCount)
	le.PutUint32(buf[offset+8:], f.TablePtrCnt)
	le.PutUint32(buf[offset+12:], f.FloorPtr)
	le.PutUint32(buf[offset+16:], f.TablePtrsPtr)
	le.PutUint32(buf[offset+20:], f.CountPtrsPtr)
}

func writeFloors(buf []byte, floors []FloorConfig, base uint32, le binary.ByteOrder) {
	for i, f := range floors {
		off := base + uint32(i)*floorStatsByteSize
		le.PutUint32(buf[off:], f.FloorNumber)
		le.PutUint32(buf[off+4:], f.SpawnTableIndex)
		le.PutUint32(buf[off+8:], f.Unk0)
		le.PutUint32(buf[off+12:], math.Float32bits(f.PointMulti1))
		le.PutUint32(buf[off+16:], math.Float32bits(f.PointMulti2))
		le.PutUint32(buf[off+20:], f.FinalLoop)
	}
}

func writeSpawnSection(buf []byte, tables []SpawnTableConfig, ptrsBase, tablesBase uint32, le binary.ByteOrder) {
	for i, t := range tables {
		tableOff := tablesBase + uint32(i)*spawnTableByteSize
		// Pointer entry
		le.PutUint32(buf[ptrsBase+uint32(i)*spawnPtrEntrySize:], tableOff)
		// SpawnTable (32 bytes)
		le.PutUint32(buf[tableOff:], t.Monster1ID)
		le.PutUint32(buf[tableOff+4:], t.Monster1Variant)
		le.PutUint32(buf[tableOff+8:], t.Monster2ID)
		le.PutUint32(buf[tableOff+12:], t.Monster2Variant)
		le.PutUint32(buf[tableOff+16:], t.StatTable)
		le.PutUint32(buf[tableOff+20:], t.MapZoneOverride)
		le.PutUint32(buf[tableOff+24:], t.SpawnWeighting)
		le.PutUint32(buf[tableOff+28:], t.AdditionalFlag)
	}
}

// validateRengokuConfig checks that all spawn_table_index references are
// within range for both road modes.
func validateRengokuConfig(cfg RengokuConfig) error {
	for _, road := range []struct {
		name string
		r    RoadConfig
	}{{"multi_road", cfg.MultiRoad}, {"solo_road", cfg.SoloRoad}} {
		n := len(road.r.SpawnTables)
		for i, f := range road.r.Floors {
			if int(f.SpawnTableIndex) >= n {
				return fmt.Errorf("rengoku: %s floor %d: spawn_table_index %d out of range (have %d tables)",
					road.name, i, f.SpawnTableIndex, n)
			}
		}
	}
	return nil
}

// ─── Shared helper ───────────────────────────────────────────────────────────

// encodeRengokuECD wraps decryption.EncodeECD with error logging.
func encodeRengokuECD(raw []byte, logger *zap.Logger) ([]byte, error) {
	enc, err := decryption.EncodeECD(raw, decryption.DefaultECDKey)
	if err != nil {
		logger.Error("rengoku: ECD encryption failed", zap.Error(err))
	}
	return enc, err
}

// ─── JSON loader ─────────────────────────────────────────────────────────────

// loadRengokuFromJSON attempts to load rengoku configuration from
// rengoku_data.json in binPath. It returns the ECD-encrypted binary ready for
// caching, or nil if the file is absent or cannot be processed.
func loadRengokuFromJSON(binPath string, logger *zap.Logger) []byte {
	path := filepath.Join(binPath, "rengoku_data.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil // file absent — not an error
	}

	var cfg RengokuConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		logger.Error("rengoku_data.json: JSON parse error",
			zap.String("path", path), zap.Error(err))
		return nil
	}

	bin, err := BuildRengokuBinary(cfg)
	if err != nil {
		logger.Error("rengoku_data.json: binary build failed",
			zap.String("path", path), zap.Error(err))
		return nil
	}

	// Validate the freshly built binary (should always pass, but good to confirm).
	info, parseErr := parseRengokuBinary(bin)
	if parseErr != nil {
		logger.Error("rengoku_data.json: structural validation of built binary failed",
			zap.String("path", path), zap.Error(parseErr))
		return nil
	}

	enc, err := encodeRengokuECD(bin, logger)
	if err != nil {
		return nil
	}

	logger.Info("Hunting Road config (from JSON)",
		zap.Int("multi_floors", info.MultiFloors),
		zap.Int("multi_spawn_tables", info.MultiSpawnTables),
		zap.Int("solo_floors", info.SoloFloors),
		zap.Int("solo_spawn_tables", info.SoloSpawnTables),
		zap.Int("unique_monsters", info.UniqueMonsters),
	)
	logger.Info("Loaded rengoku_data.json", zap.Int("bytes", len(enc)))
	return enc
}
