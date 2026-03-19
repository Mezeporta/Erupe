package channelserver

import (
	"encoding/binary"
	"strings"
	"testing"
)

// buildRengokuData constructs a minimal but structurally valid rengoku binary
// for testing.  It contains one floor and one spawn table per road mode.
//
// Layout:
//
//	0x00–0x13  header (magic + version + padding)
//	0x14–0x2B  multiDef RoadMode
//	0x2C–0x43  soloDef  RoadMode
//	0x44–0x5B  multiDef FloorStats (24 bytes)
//	0x5C–0x63  multiDef spawnTablePtrs (1×u32 = 4 bytes)
//	0x64–0x67  multiDef spawnCountPtrs (1×u32 = 4 bytes)
//	0x68–0x87  multiDef SpawnTable (32 bytes)
//	0x88–0x9F  soloDef  FloorStats (24 bytes)
//	0xA0–0xA3  soloDef  spawnTablePtrs (1×u32)
//	0xA4–0xA7  soloDef  spawnCountPtrs (1×u32)
//	0xA8–0xC7  soloDef  SpawnTable (32 bytes)
func buildRengokuData(multiMonster1, multiMonster2, soloMonster1, soloMonster2 uint32) []byte {
	buf := make([]byte, 0xC8)

	// Header
	buf[0] = 'r'
	buf[1] = 'e'
	buf[2] = 'f'
	buf[3] = 0x1A
	buf[4] = 1 // version

	le := binary.LittleEndian

	// multiDef RoadMode at 0x14
	le.PutUint32(buf[0x14:], 1)    // floorStatsCount
	le.PutUint32(buf[0x18:], 1)    // spawnCountCount
	le.PutUint32(buf[0x1C:], 1)    // spawnTablePtrCount
	le.PutUint32(buf[0x20:], 0x44) // floorStatsPtr
	le.PutUint32(buf[0x24:], 0x5C) // spawnTablePtrsPtr
	le.PutUint32(buf[0x28:], 0x64) // spawnCountPtrsPtr

	// soloDef RoadMode at 0x2C
	le.PutUint32(buf[0x2C:], 1)    // floorStatsCount
	le.PutUint32(buf[0x30:], 1)    // spawnCountCount
	le.PutUint32(buf[0x34:], 1)    // spawnTablePtrCount
	le.PutUint32(buf[0x38:], 0x88) // floorStatsPtr
	le.PutUint32(buf[0x3C:], 0xA0) // spawnTablePtrsPtr
	le.PutUint32(buf[0x40:], 0xA4) // spawnCountPtrsPtr

	// multiDef FloorStats at 0x44 (24 bytes)
	le.PutUint32(buf[0x44:], 1) // floorNumber

	// multiDef spawnTablePtrs at 0x5C: points to SpawnTable at 0x68
	le.PutUint32(buf[0x5C:], 0x68)

	// multiDef SpawnTable at 0x68 (32 bytes)
	le.PutUint32(buf[0x68:], multiMonster1)
	le.PutUint32(buf[0x70:], multiMonster2)

	// soloDef FloorStats at 0x88 (24 bytes)
	le.PutUint32(buf[0x88:], 1) // floorNumber

	// soloDef spawnTablePtrs at 0xA0: points to SpawnTable at 0xA8
	le.PutUint32(buf[0xA0:], 0xA8)

	// soloDef SpawnTable at 0xA8 (32 bytes)
	le.PutUint32(buf[0xA8:], soloMonster1)
	le.PutUint32(buf[0xB0:], soloMonster2)

	return buf
}

func TestParseRengokuBinary_ValidMinimal(t *testing.T) {
	data := buildRengokuData(101, 102, 103, 101) // monster 101 appears in both roads

	info, err := parseRengokuBinary(data)
	if err != nil {
		t.Fatalf("parseRengokuBinary: %v", err)
	}
	if info.MultiFloors != 1 {
		t.Errorf("MultiFloors = %d, want 1", info.MultiFloors)
	}
	if info.MultiSpawnTables != 1 {
		t.Errorf("MultiSpawnTables = %d, want 1", info.MultiSpawnTables)
	}
	if info.SoloFloors != 1 {
		t.Errorf("SoloFloors = %d, want 1", info.SoloFloors)
	}
	if info.SoloSpawnTables != 1 {
		t.Errorf("SoloSpawnTables = %d, want 1", info.SoloSpawnTables)
	}
	// IDs present: 101, 102, 103 → 3 unique (101 shared between roads)
	if info.UniqueMonsters != 3 {
		t.Errorf("UniqueMonsters = %d, want 3", info.UniqueMonsters)
	}
}

func TestParseRengokuBinary_ZeroMonsterIDsExcluded(t *testing.T) {
	data := buildRengokuData(0, 55, 0, 0) // only monster 55 is non-zero

	info, err := parseRengokuBinary(data)
	if err != nil {
		t.Fatalf("parseRengokuBinary: %v", err)
	}
	if info.UniqueMonsters != 1 {
		t.Errorf("UniqueMonsters = %d, want 1 (zeros excluded)", info.UniqueMonsters)
	}
}

func TestParseRengokuBinary_Errors(t *testing.T) {
	validData := buildRengokuData(1, 2, 3, 4)

	cases := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{
			name:    "too_small",
			data:    make([]byte, 10),
			wantErr: "too small",
		},
		{
			name: "bad_magic",
			data: func() []byte {
				d := make([]byte, len(validData))
				copy(d, validData)
				d[0] = 0xFF
				return d
			}(),
			wantErr: "invalid magic",
		},
		{
			name: "wrong_version",
			data: func() []byte {
				d := make([]byte, len(validData))
				copy(d, validData)
				d[4] = 2
				return d
			}(),
			wantErr: "unexpected version",
		},
		{
			name: "floorStats_ptr_out_of_bounds",
			data: func() []byte {
				d := make([]byte, len(validData))
				copy(d, validData)
				// Set multiDef floorStatsPtr to beyond file end
				binary.LittleEndian.PutUint32(d[0x20:], uint32(len(d)+1))
				return d
			}(),
			wantErr: "out of bounds",
		},
		{
			name: "spawnTable_ptr_target_out_of_bounds",
			data: func() []byte {
				d := make([]byte, len(validData))
				copy(d, validData)
				// Point the spawn table pointer to just before the end so SpawnTable
				// (32 bytes) would extend beyond the file.
				binary.LittleEndian.PutUint32(d[0x5C:], uint32(len(d)-4))
				return d
			}(),
			wantErr: "out of bounds",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseRengokuBinary(tc.data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
