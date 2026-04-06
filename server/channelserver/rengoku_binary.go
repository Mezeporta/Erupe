package channelserver

import (
	"encoding/binary"
	"fmt"
)

// rengoku binary layout (after ECD decryption + JKR decompression):
//
//	@0x00: magic bytes 'r','e','f',0x1A
//	@0x04: version (u8, expected 1)
//	@0x05: 15 bytes of header offsets (unused by this parser)
//	@0x14: RoadMode multiDef (24 bytes)
//	@0x2C: RoadMode soloDef  (24 bytes)
const (
	rengokuMinSize     = 0x44 // header (0x14) + two RoadModes (2×24)
	rengokuMultiOffset = 0x14
	rengokuSoloOffset  = 0x2C
	floorStatsByteSize = 24
	spawnTableByteSize = 32
	spawnPtrEntrySize  = 4 // each spawn-table pointer is a u32
)

// rengokuRoadMode holds a parsed RoadMode struct. All pointer fields are file
// offsets into the raw (decrypted + decompressed) byte slice.
type rengokuRoadMode struct {
	FloorStatsCount    uint32
	SpawnCountCount    uint32
	SpawnTablePtrCount uint32
	FloorStatsPtr      uint32 // → FloorStats[FloorStatsCount]
	SpawnTablePtrsPtr  uint32 // → u32[SpawnTablePtrCount] → SpawnTable[]
	SpawnCountPtrsPtr  uint32 // → u32[SpawnCountCount]
}

// RengokuBinaryInfo summarises the validated rengoku_data.bin contents for
// structured logging. It is populated by parseRengokuBinary.
type RengokuBinaryInfo struct {
	MultiFloors      int
	MultiSpawnTables int
	SoloFloors       int
	SoloSpawnTables  int
	UniqueMonsters   int
}

// parseRengokuBinary validates the structural integrity of a decrypted and
// decompressed rengoku_data.bin and returns a summary of its contents.
//
// It checks:
//   - magic bytes and version
//   - all pointer-derived ranges lie within the file
//   - individual spawn-table pointers fall within the file
func parseRengokuBinary(data []byte) (*RengokuBinaryInfo, error) {
	if len(data) < rengokuMinSize {
		return nil, fmt.Errorf("rengoku: file too small (%d bytes, need %d)", len(data), rengokuMinSize)
	}

	// Magic: 'r','e','f',0x1A
	if data[0] != 'r' || data[1] != 'e' || data[2] != 'f' || data[3] != 0x1A {
		return nil, fmt.Errorf("rengoku: invalid magic %02x %02x %02x %02x",
			data[0], data[1], data[2], data[3])
	}

	if data[4] != 1 {
		return nil, fmt.Errorf("rengoku: unexpected version %d (want 1)", data[4])
	}

	multi, err := readRoadMode(data, rengokuMultiOffset)
	if err != nil {
		return nil, fmt.Errorf("rengoku: multiDef: %w", err)
	}
	solo, err := readRoadMode(data, rengokuSoloOffset)
	if err != nil {
		return nil, fmt.Errorf("rengoku: soloDef: %w", err)
	}

	if err := validateRoadMode(data, multi, "multiDef"); err != nil {
		return nil, err
	}
	if err := validateRoadMode(data, solo, "soloDef"); err != nil {
		return nil, err
	}

	uniqueMonsters := countUniqueMonsters(data, multi)
	for id := range countUniqueMonsters(data, solo) {
		uniqueMonsters[id] = struct{}{}
	}

	return &RengokuBinaryInfo{
		MultiFloors:      int(multi.FloorStatsCount),
		MultiSpawnTables: int(multi.SpawnTablePtrCount),
		SoloFloors:       int(solo.FloorStatsCount),
		SoloSpawnTables:  int(solo.SpawnTablePtrCount),
		UniqueMonsters:   len(uniqueMonsters),
	}, nil
}

// readRoadMode reads a 24-byte RoadMode struct from data at offset.
func readRoadMode(data []byte, offset int) (rengokuRoadMode, error) {
	end := offset + 24
	if len(data) < end {
		return rengokuRoadMode{}, fmt.Errorf("RoadMode at 0x%X extends beyond file", offset)
	}
	d := data[offset:]
	return rengokuRoadMode{
		FloorStatsCount:    binary.LittleEndian.Uint32(d[0:]),
		SpawnCountCount:    binary.LittleEndian.Uint32(d[4:]),
		SpawnTablePtrCount: binary.LittleEndian.Uint32(d[8:]),
		FloorStatsPtr:      binary.LittleEndian.Uint32(d[12:]),
		SpawnTablePtrsPtr:  binary.LittleEndian.Uint32(d[16:]),
		SpawnCountPtrsPtr:  binary.LittleEndian.Uint32(d[20:]),
	}, nil
}

// ptrInBounds returns true if the region [ptr, ptr+size) fits within data.
// It guards against overflow when ptr+size wraps uint32.
func ptrInBounds(data []byte, ptr, size uint32) bool {
	end := ptr + size
	if end < ptr { // overflow
		return false
	}
	return int(end) <= len(data)
}

// validateRoadMode checks that all pointer-derived byte ranges for a RoadMode
// lie within data.
func validateRoadMode(data []byte, rm rengokuRoadMode, label string) error {
	fileLen := uint32(len(data))

	// Floor-stats array bounds.
	if !ptrInBounds(data, rm.FloorStatsPtr, rm.FloorStatsCount*floorStatsByteSize) {
		return fmt.Errorf("rengoku: %s: floorStats array [0x%X, +%d×%d] out of bounds (file %d B)",
			label, rm.FloorStatsPtr, rm.FloorStatsCount, floorStatsByteSize, fileLen)
	}

	// Spawn-table pointer array bounds.
	if !ptrInBounds(data, rm.SpawnTablePtrsPtr, rm.SpawnTablePtrCount*spawnPtrEntrySize) {
		return fmt.Errorf("rengoku: %s: spawnTablePtrs array [0x%X, +%d×4] out of bounds (file %d B)",
			label, rm.SpawnTablePtrsPtr, rm.SpawnTablePtrCount, fileLen)
	}

	// Spawn-count pointer array bounds.
	if !ptrInBounds(data, rm.SpawnCountPtrsPtr, rm.SpawnCountCount*spawnPtrEntrySize) {
		return fmt.Errorf("rengoku: %s: spawnCountPtrs array [0x%X, +%d×4] out of bounds (file %d B)",
			label, rm.SpawnCountPtrsPtr, rm.SpawnCountCount, fileLen)
	}

	// Individual spawn-table pointer targets.
	ptrBase := rm.SpawnTablePtrsPtr
	for i := uint32(0); i < rm.SpawnTablePtrCount; i++ {
		tablePtr := binary.LittleEndian.Uint32(data[ptrBase+i*4:])
		if !ptrInBounds(data, tablePtr, spawnTableByteSize) {
			return fmt.Errorf("rengoku: %s: spawnTable[%d] at 0x%X is out of bounds (file %d B)",
				label, i, tablePtr, fileLen)
		}
	}

	return nil
}

// countUniqueMonsters iterates all SpawnTables for a RoadMode and returns a
// set of unique non-zero monster IDs (from both monsterID1 and monsterID2).
func countUniqueMonsters(data []byte, rm rengokuRoadMode) map[uint32]struct{} {
	ids := make(map[uint32]struct{})
	ptrBase := rm.SpawnTablePtrsPtr
	for i := uint32(0); i < rm.SpawnTablePtrCount; i++ {
		tablePtr := binary.LittleEndian.Uint32(data[ptrBase+i*4:])
		if !ptrInBounds(data, tablePtr, spawnTableByteSize) {
			continue
		}
		t := data[tablePtr:]
		id1 := binary.LittleEndian.Uint32(t[0:])
		id2 := binary.LittleEndian.Uint32(t[8:])
		if id1 != 0 {
			ids[id1] = struct{}{}
		}
		if id2 != 0 {
			ids[id2] = struct{}{}
		}
	}
	return ids
}
