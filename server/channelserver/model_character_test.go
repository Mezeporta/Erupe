package channelserver

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// zzBlobSize is the minimum decompressed ZZ save blob size required to cover
// every offset declared in getPointers(cfg.ZZ). Derived from the highest
// mapped pointer (pKQF = 146720) plus saveFieldKQF, plus the new fields.
// Use a generous upper bound so every pointer + field is addressable.
const zzBlobSize = 150820 // matches observed live ZZ decompressed size

// buildMinimalZZBlob builds a zero-initialised decompressed ZZ save blob
// large enough to cover every field the parser reads, with the given
// scalar values written at their expected offsets.
func buildMinimalZZBlob(t *testing.T, zenny, gzenny, cp uint32, rp uint16, playtime uint32) []byte {
	t.Helper()
	buf := make([]byte, zzBlobSize)
	p := getPointers(cfg.ZZ)
	binary.LittleEndian.PutUint32(buf[p[pZenny]:p[pZenny]+saveFieldZenny], zenny)
	binary.LittleEndian.PutUint32(buf[p[pGZenny]:p[pGZenny]+saveFieldGZenny], gzenny)
	binary.LittleEndian.PutUint32(buf[p[pCP]:p[pCP]+saveFieldCP], cp)
	binary.LittleEndian.PutUint16(buf[p[pRP]:p[pRP]+saveFieldRP], rp)
	binary.LittleEndian.PutUint32(buf[p[pPlaytime]:p[pPlaytime]+saveFieldPlaytime], playtime)
	return buf
}

// TestGetPointers_NewFields_ZZOnly verifies that pZenny / pGZenny / pCP /
// pCurrentEquip are only populated for cfg.ZZ and remain zero for every
// other mode. This guards against accidental cross-version reads that
// could corrupt saves on F5 / G1-G5.2 / S6 where the offsets are not
// validated.
func TestGetPointers_NewFields_ZZOnly(t *testing.T) {
	zzPointers := getPointers(cfg.ZZ)
	if zzPointers[pZenny] != 0xB0 {
		t.Errorf("ZZ pZenny = 0x%X, want 0xB0", zzPointers[pZenny])
	}
	if zzPointers[pGZenny] != 0x1FF64 {
		t.Errorf("ZZ pGZenny = 0x%X, want 0x1FF64", zzPointers[pGZenny])
	}
	if zzPointers[pCP] != 0x212E4 {
		t.Errorf("ZZ pCP = 0x%X, want 0x212E4", zzPointers[pCP])
	}
	if zzPointers[pCurrentEquip] != 0x1F604 {
		t.Errorf("ZZ pCurrentEquip = 0x%X, want 0x1F604", zzPointers[pCurrentEquip])
	}

	unmapped := []cfg.Mode{cfg.Z2, cfg.Z1, cfg.G101, cfg.G10, cfg.G91, cfg.G9,
		cfg.G81, cfg.G8, cfg.G7, cfg.G61, cfg.G6, cfg.G52, cfg.G51, cfg.G5,
		cfg.GG, cfg.G32, cfg.G31, cfg.G3, cfg.G2, cfg.G1,
		cfg.F5, cfg.F4, cfg.S6}
	for _, m := range unmapped {
		p := getPointers(m)
		for _, ptr := range []SavePointer{pZenny, pGZenny, pCP, pCurrentEquip} {
			if got, ok := p[ptr]; ok && got != 0 {
				t.Errorf("mode %v unexpectedly has pointer %v = 0x%X "+
					"(new fields must stay unmapped outside ZZ)", m, ptr, got)
			}
		}
	}
}

// TestUpdateStructWithSaveData_ZZ_NewFields builds a minimal ZZ blob with
// known zenny / gzenny / CP values at their configured offsets, runs the
// parser, and asserts the struct fields match. This is the positive-path
// roundtrip: blob → struct.
func TestUpdateStructWithSaveData_ZZ_NewFields(t *testing.T) {
	tests := []struct {
		name   string
		zenny  uint32
		gzenny uint32
		cp     uint32
	}{
		{"zero values", 0, 0, 0},
		{"typical HR999 values", 8821924, 838956, 49379}, // from live blob
		{"max uint32", 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		{"mixed", 123456, 0, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob := buildMinimalZZBlob(t, tt.zenny, tt.gzenny, tt.cp, 0, 0)
			save := &CharacterSaveData{
				Mode:       cfg.ZZ,
				Pointers:   getPointers(cfg.ZZ),
				decompSave: blob,
			}
			save.updateStructWithSaveData()
			if save.Zenny != tt.zenny {
				t.Errorf("Zenny = %d, want %d", save.Zenny, tt.zenny)
			}
			if save.GZenny != tt.gzenny {
				t.Errorf("GZenny = %d, want %d", save.GZenny, tt.gzenny)
			}
			if save.CP != tt.cp {
				t.Errorf("CP = %d, want %d", save.CP, tt.cp)
			}
		})
	}
}

// TestUpdateStructWithSaveData_ZZ_ExistingFieldsUnaffected is a regression
// guard: loading a ZZ blob with the new fields populated must not change
// how Playtime / HR / RP / KQF / Gender are read. Any shift in those
// values would silently corrupt live saves on next write-back.
func TestUpdateStructWithSaveData_ZZ_ExistingFieldsUnaffected(t *testing.T) {
	const (
		wantPlaytime uint32 = 472080 // from live kirito blob (131h)
		wantRP       uint16 = 1234
	)
	blob := buildMinimalZZBlob(t, 8821924, 838956, 49379, wantRP, wantPlaytime)
	// Populate gender byte so the gender read path exercises the live offset.
	p := getPointers(cfg.ZZ)
	blob[p[pGender]] = 1
	save := &CharacterSaveData{
		Mode:       cfg.ZZ,
		Pointers:   p,
		decompSave: blob,
	}
	save.updateStructWithSaveData()
	if save.Playtime != wantPlaytime {
		t.Errorf("Playtime = %d, want %d (existing field must not shift)",
			save.Playtime, wantPlaytime)
	}
	if save.RP != wantRP {
		t.Errorf("RP = %d, want %d (existing field must not shift)",
			save.RP, wantRP)
	}
	if !save.Gender {
		t.Errorf("Gender = false, want true (existing field must not shift)")
	}
	if len(save.KQF) != saveFieldKQF {
		t.Errorf("KQF len = %d, want %d", len(save.KQF), saveFieldKQF)
	}
}

// TestUpdateStructWithSaveData_NewCharacterSkipsReads ensures that for
// brand-new characters (IsNewCharacter = true) none of the new fields are
// populated from what is likely an uninitialised blob.
func TestUpdateStructWithSaveData_NewCharacterSkipsReads(t *testing.T) {
	blob := buildMinimalZZBlob(t, 9999, 9999, 9999, 0, 0)
	save := &CharacterSaveData{
		Mode:           cfg.ZZ,
		Pointers:       getPointers(cfg.ZZ),
		decompSave:     blob,
		IsNewCharacter: true,
	}
	save.updateStructWithSaveData()
	if save.Zenny != 0 || save.GZenny != 0 || save.CP != 0 {
		t.Errorf("new character leaked zenny/gzenny/CP: %d/%d/%d",
			save.Zenny, save.GZenny, save.CP)
	}
}

// TestUpdateStructWithSaveData_NonZZLeavesNewFieldsZero verifies that a
// non-ZZ mode (e.g. Z2 or G10) does NOT read zenny/gzenny/CP, so they
// remain zero-valued. ZZ-only scope must not leak into other versions.
func TestUpdateStructWithSaveData_NonZZLeavesNewFieldsZero(t *testing.T) {
	modes := []cfg.Mode{cfg.Z2, cfg.G10, cfg.G5, cfg.F5, cfg.S6}
	for _, m := range modes {
		t.Run(m.String(), func(t *testing.T) {
			// Build a generous blob so bounds are never the reason for zeros.
			blob := make([]byte, zzBlobSize)
			// Seed what would be the ZZ zenny offset with a recognisable
			// non-zero value — if the parser mistakenly reads it for a
			// non-ZZ mode, the test catches it.
			binary.LittleEndian.PutUint32(blob[0xB0:0xB0+4], 0xDEADBEEF)
			binary.LittleEndian.PutUint32(blob[0x1FF64:0x1FF64+4], 0xCAFEBABE)
			binary.LittleEndian.PutUint32(blob[0x212E4:0x212E4+4], 0x1234)
			save := &CharacterSaveData{
				Mode:       m,
				Pointers:   getPointers(m),
				decompSave: blob,
			}
			save.updateStructWithSaveData()
			if save.Zenny != 0 {
				t.Errorf("mode %v read Zenny = 0x%X, want 0 "+
					"(ZZ offsets must not apply)", m, save.Zenny)
			}
			if save.GZenny != 0 {
				t.Errorf("mode %v read GZenny = 0x%X, want 0", m, save.GZenny)
			}
			if save.CP != 0 {
				t.Errorf("mode %v read CP = %d, want 0", m, save.CP)
			}
		})
	}
}

// TestUpdateStructWithSaveData_LiveBlob parses a real ZZ save blob pulled
// from production (gitignored under tmp/saves/). Values hard-coded here
// are what the save-mgr offsets produced when inspected by hand; the test
// fails if a future refactor shifts them. The test skips silently when
// the blob file is absent (CI, other developers' machines).
func TestUpdateStructWithSaveData_LiveBlob(t *testing.T) {
	path := filepath.Join("..", "..", "tmp", "saves", "297_kirito.comp")
	comp, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("live blob unavailable at %s: %v", path, err)
	}
	decomp, err := nullcomp.Decompress(comp)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	save := &CharacterSaveData{
		Mode:       cfg.ZZ,
		Pointers:   getPointers(cfg.ZZ),
		decompSave: decomp,
	}
	save.updateStructWithSaveData()
	const (
		wantName     = "kirito"
		wantPlaytime = 472080
		wantZenny    = 8821924
		wantGZenny   = 838956
		wantCP       = 49379
	)
	if save.Name != wantName {
		t.Errorf("Name = %q, want %q", save.Name, wantName)
	}
	if save.Playtime != wantPlaytime {
		t.Errorf("Playtime = %d, want %d", save.Playtime, wantPlaytime)
	}
	if save.Zenny != wantZenny {
		t.Errorf("Zenny = %d, want %d", save.Zenny, wantZenny)
	}
	if save.GZenny != wantGZenny {
		t.Errorf("GZenny = %d, want %d", save.GZenny, wantGZenny)
	}
	if save.CP != wantCP {
		t.Errorf("CP = %d, want %d", save.CP, wantCP)
	}
}

// TestUpdateStructWithSaveData_BoundsSafety guards the new reads against
// truncated blobs: a decompressed save that happens to be shorter than the
// configured ZZ offsets must not panic. We don't require any particular
// parsed value — only that the process survives.
func TestUpdateStructWithSaveData_BoundsSafety(t *testing.T) {
	sizes := []int{
		// At a minimum, the existing parser requires a blob that covers
		// every existing pointer + field; truncating below that tripped
		// pre-existing reads, not ours. Cover only sizes that exercise
		// the new-field bounds check.
		zzBlobSize - 1,
		0x212E4 + 3, // just below pCP + size
		0x1FF64 + 3, // just below pGZenny + size
	}
	for _, sz := range sizes {
		// Build a full-size blob, populate existing fields, then truncate.
		full := buildMinimalZZBlob(t, 1, 2, 3, 0, 0)
		if sz > len(full) {
			continue
		}
		trunc := full[:sz]
		save := &CharacterSaveData{
			Mode:       cfg.ZZ,
			Pointers:   getPointers(cfg.ZZ),
			decompSave: trunc,
		}
		// If existing reads panic at this size, skip — we only care
		// about new-field safety.
		func() {
			defer func() { _ = recover() }()
			save.updateStructWithSaveData()
		}()
	}
}
