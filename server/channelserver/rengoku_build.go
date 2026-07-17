package channelserver

/*
	JSON-based rengoku_data.bin builder.

	Operators can place rengoku_data.json in the bin/ directory instead of
	(or alongside) rengoku_data.bin. When the JSON file is found it takes
	precedence: it is parsed, validated, assembled into the raw binary layout,
	and ECD-encrypted before being cached. The .bin file is used as a fallback.

	Each spawn slot (referenced by a floor's SpawnTableIndex) holds a *pool* of
	candidate spawn tables — the real client rolls one candidate per slot using
	each candidate's spawn_weighting, so a slot's count must be >= 1. A pool of
	size 1 is a valid degenerate case (fixed spawn, no rolling).

	Binary layout produced by BuildRengokuBinary:
	  0x00–0x13  header  (20 bytes: magic + version + zeros)
	  0x14–0x2B  multiDef RoadMode  (24 bytes)
	  0x2C–0x43  soloDef  RoadMode  (24 bytes)
	  -- multi road data --
	  floorStats[]          (floorStatsCount × 24 bytes)
	  spawnTablePtrs[]      (slotCount × 4 bytes — ptr to each slot's first candidate)
	  spawnCountPtrs[]      (slotCount × 4 bytes — candidate count per slot, >= 1)
	  spawnTables[]         (Σ pool sizes × 32 bytes, pools stored back to back)
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
// spawn pools. Floors reference spawn pools by zero-based index.
type RoadConfig struct {
	Floors     []FloorConfig        `json:"floors"`
	SpawnPools [][]SpawnTableConfig `json:"spawn_pools"`
}

// FloorConfig describes one floor within a road mode.
//
//   - SpawnTableIndex: zero-based index into this road's SpawnPools slice,
//     selecting which pool of candidate monster configurations is active on
//     this floor. The client rolls one candidate from the pool per spawn,
//     weighted by each candidate's SpawnWeighting.
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

	mSlotCount := uint32(len(cfg.MultiRoad.SpawnPools))
	mCandidateCount := countCandidates(cfg.MultiRoad.SpawnPools)
	sSlotCount := uint32(len(cfg.SoloRoad.SpawnPools))
	sCandidateCount := countCandidates(cfg.SoloRoad.SpawnPools)

	// Multi road sections
	mFloorOff := dataStart
	mFloorSz := uint32(len(cfg.MultiRoad.Floors)) * floorStatsByteSize
	mPtrsOff := mFloorOff + mFloorSz
	mPtrsSz := mSlotCount * spawnPtrEntrySize
	mCntOff := mPtrsOff + mPtrsSz
	mCntSz := mSlotCount * spawnPtrEntrySize
	mTablesOff := mCntOff + mCntSz
	mTablesSz := mCandidateCount * spawnTableByteSize

	// Solo road sections (appended directly after multi)
	sFloorOff := mTablesOff + mTablesSz
	sFloorSz := uint32(len(cfg.SoloRoad.Floors)) * floorStatsByteSize
	sPtrsOff := sFloorOff + sFloorSz
	sPtrsSz := sSlotCount * spawnPtrEntrySize
	sCntOff := sPtrsOff + sPtrsSz
	sCntSz := sSlotCount * spawnPtrEntrySize
	sTablesOff := sCntOff + sCntSz
	sTablesSz := sCandidateCount * spawnTableByteSize

	totalSize := sTablesOff + sTablesSz
	buf := make([]byte, totalSize)

	// ── Header ───────────────────────────────────────────────────────────────
	buf[0], buf[1], buf[2], buf[3] = 'r', 'e', 'f', 0x1A
	buf[4] = 1 // version

	le := binary.LittleEndian

	// ── RoadMode structs ─────────────────────────────────────────────────────
	writeRoadMode(buf, 0x14, le, RoadModeFields{
		FloorCount:   uint32(len(cfg.MultiRoad.Floors)),
		SpawnCount:   mSlotCount,
		TablePtrCnt:  mSlotCount,
		FloorPtr:     mFloorOff,
		TablePtrsPtr: mPtrsOff,
		CountPtrsPtr: mCntOff,
	})
	writeRoadMode(buf, 0x2C, le, RoadModeFields{
		FloorCount:   uint32(len(cfg.SoloRoad.Floors)),
		SpawnCount:   sSlotCount,
		TablePtrCnt:  sSlotCount,
		FloorPtr:     sFloorOff,
		TablePtrsPtr: sPtrsOff,
		CountPtrsPtr: sCntOff,
	})

	// ── Data sections ────────────────────────────────────────────────────────
	writeFloors(buf, cfg.MultiRoad.Floors, mFloorOff, le)
	writeSpawnSection(buf, cfg.MultiRoad.SpawnPools, mPtrsOff, mCntOff, mTablesOff, le)

	writeFloors(buf, cfg.SoloRoad.Floors, sFloorOff, le)
	writeSpawnSection(buf, cfg.SoloRoad.SpawnPools, sPtrsOff, sCntOff, sTablesOff, le)

	return buf, nil
}

// countCandidates returns the total number of candidate spawn tables across
// all pools (i.e. the sum of each pool's length).
func countCandidates(pools [][]SpawnTableConfig) uint32 {
	var n uint32
	for _, pool := range pools {
		n += uint32(len(pool))
	}
	return n
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

// writeSpawnSection writes, for each pool (slot): a pointer to the pool's
// first candidate table, the candidate count (len(pool)), and the candidate
// tables themselves, packed back to back across all pools.
func writeSpawnSection(buf []byte, pools [][]SpawnTableConfig, ptrsBase, cntsBase, tablesBase uint32, le binary.ByteOrder) {
	tableOff := tablesBase
	for i, pool := range pools {
		le.PutUint32(buf[ptrsBase+uint32(i)*spawnPtrEntrySize:], tableOff)
		le.PutUint32(buf[cntsBase+uint32(i)*spawnPtrEntrySize:], uint32(len(pool)))
		for _, t := range pool {
			writeSpawnTable(buf, tableOff, t, le)
			tableOff += spawnTableByteSize
		}
	}
}

// writeSpawnTable writes one 32-byte SpawnTable entry at off.
func writeSpawnTable(buf []byte, off uint32, t SpawnTableConfig, le binary.ByteOrder) {
	le.PutUint32(buf[off:], t.Monster1ID)
	le.PutUint32(buf[off+4:], t.Monster1Variant)
	le.PutUint32(buf[off+8:], t.Monster2ID)
	le.PutUint32(buf[off+12:], t.Monster2Variant)
	le.PutUint32(buf[off+16:], t.StatTable)
	le.PutUint32(buf[off+20:], t.MapZoneOverride)
	le.PutUint32(buf[off+24:], t.SpawnWeighting)
	le.PutUint32(buf[off+28:], t.AdditionalFlag)
}

// validateRengokuConfig checks that all spawn_table_index references are
// within range, and that every spawn pool has at least one candidate table,
// for both road modes. A pool with zero candidates would be written with a
// candidate count of 0, which crashes the real client on Hunting Road entry.
func validateRengokuConfig(cfg RengokuConfig) error {
	for _, road := range []struct {
		name string
		r    RoadConfig
	}{{"multi_road", cfg.MultiRoad}, {"solo_road", cfg.SoloRoad}} {
		n := len(road.r.SpawnPools)
		for i, f := range road.r.Floors {
			if int(f.SpawnTableIndex) >= n {
				return fmt.Errorf("rengoku: %s floor %d: spawn_table_index %d out of range (have %d spawn pools)",
					road.name, i, f.SpawnTableIndex, n)
			}
		}
		for i, pool := range road.r.SpawnPools {
			if len(pool) == 0 {
				return fmt.Errorf("rengoku: %s spawn_pools[%d]: pool is empty (need at least 1 candidate)",
					road.name, i)
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
