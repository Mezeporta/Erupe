package channelserver

import (
	"bytes"
	"encoding/binary"
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// TestGetPointers tests the pointer map generation for different game versions
func TestGetPointers(t *testing.T) {
	tests := []struct {
		name       string
		clientMode cfg.Mode
		wantGender int
		wantHR     int
	}{
		{
			name:       "ZZ_version",
			clientMode: cfg.ZZ,
			wantGender: 81,
			wantHR:     130550,
		},
		{
			name:       "Z2_version",
			clientMode: cfg.Z2,
			wantGender: 81,
			wantHR:     94550,
		},
		{
			name:       "G10_version",
			clientMode: cfg.G10,
			wantGender: 81,
			wantHR:     94550,
		},
		{
			name:       "F5_version",
			clientMode: cfg.F5,
			wantGender: 81,
			wantHR:     62550,
		},
		{
			name:       "S6_version",
			clientMode: cfg.S6,
			wantGender: 81,
			wantHR:     14550,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pointers := getPointers(tt.clientMode)

			if pointers[pGender] != tt.wantGender {
				t.Errorf("pGender = %d, want %d", pointers[pGender], tt.wantGender)
			}

			if pointers[pHR] != tt.wantHR {
				t.Errorf("pHR = %d, want %d", pointers[pHR], tt.wantHR)
			}

			// Verify all required pointers exist
			requiredPointers := []SavePointer{pGender, pRP, pHouseTier, pHouseData, pBookshelfData,
				pGalleryData, pToreData, pGardenData, pPlaytime, pWeaponType, pWeaponID, pHR, lBookshelfData}

			for _, ptr := range requiredPointers {
				if _, exists := pointers[ptr]; !exists {
					t.Errorf("pointer %v not found in map", ptr)
				}
			}
		})
	}
}

// TestCharacterSaveData_Compress tests savedata compression
func TestCharacterSaveData_Compress(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid_small_data",
			data:    []byte{0x01, 0x02, 0x03, 0x04},
			wantErr: false,
		},
		{
			name:    "valid_large_data",
			data:    bytes.Repeat([]byte{0xAA}, 10000),
			wantErr: false,
		},
		{
			name:    "empty_data",
			data:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &CharacterSaveData{
				decompSave: tt.data,
			}

			err := save.Compress()
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(save.compSave) == 0 {
				t.Error("compressed save is empty")
			}
		})
	}
}

// TestCharacterSaveData_Decompress tests savedata decompression
func TestCharacterSaveData_Decompress(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() []byte
		wantErr bool
	}{
		{
			name: "valid_compressed_data",
			setup: func() []byte {
				data := []byte{0x01, 0x02, 0x03, 0x04}
				compressed, _ := nullcomp.Compress(data)
				return compressed
			},
			wantErr: false,
		},
		{
			name: "valid_large_compressed_data",
			setup: func() []byte {
				data := bytes.Repeat([]byte{0xBB}, 5000)
				compressed, _ := nullcomp.Compress(data)
				return compressed
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &CharacterSaveData{
				compSave: tt.setup(),
			}

			err := save.Decompress()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompress() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(save.decompSave) == 0 {
				t.Error("decompressed save is empty")
			}
		})
	}
}

// TestCharacterSaveData_RoundTrip tests compression and decompression
func TestCharacterSaveData_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "small_data",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		},
		{
			name: "repeating_pattern",
			data: bytes.Repeat([]byte{0xCC}, 1000),
		},
		{
			name: "mixed_data",
			data: []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD, 0x03, 0xFC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &CharacterSaveData{
				decompSave: tt.data,
			}

			// Compress
			if err := save.Compress(); err != nil {
				t.Fatalf("Compress() failed: %v", err)
			}

			// Clear decompressed data
			save.decompSave = nil

			// Decompress
			if err := save.Decompress(); err != nil {
				t.Fatalf("Decompress() failed: %v", err)
			}

			// Verify round trip
			if !bytes.Equal(save.decompSave, tt.data) {
				t.Errorf("round trip failed: got %v, want %v", save.decompSave, tt.data)
			}
		})
	}
}

// TestCharacterSaveData_updateStructWithSaveData tests parsing save data
func TestCharacterSaveData_updateStructWithSaveData(t *testing.T) {
	tests := []struct {
		name           string
		isNewCharacter bool
		setupSaveData  func() []byte
		wantName       string
		wantGender     bool
	}{
		{
			name:           "male_character",
			isNewCharacter: false,
			setupSaveData: func() []byte {
				data := make([]byte, 150000)
				copy(data[88:], []byte("TestChar\x00"))
				data[81] = 0 // Male
				return data
			},
			wantName:   "TestChar",
			wantGender: false,
		},
		{
			name:           "female_character",
			isNewCharacter: false,
			setupSaveData: func() []byte {
				data := make([]byte, 150000)
				copy(data[88:], []byte("FemaleChar\x00"))
				data[81] = 1 // Female
				return data
			},
			wantName:   "FemaleChar",
			wantGender: true,
		},
		{
			name:           "new_character_skips_parsing",
			isNewCharacter: true,
			setupSaveData: func() []byte {
				data := make([]byte, 150000)
				copy(data[88:], []byte("NewChar\x00"))
				return data
			},
			wantName:   "NewChar",
			wantGender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &CharacterSaveData{
				Mode:           cfg.Z2,
				Pointers:       getPointers(cfg.Z2),
				decompSave:     tt.setupSaveData(),
				IsNewCharacter: tt.isNewCharacter,
			}

			save.updateStructWithSaveData()

			if save.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", save.Name, tt.wantName)
			}

			if save.Gender != tt.wantGender {
				t.Errorf("Gender = %v, want %v", save.Gender, tt.wantGender)
			}
		})
	}
}

// TestCharacterSaveData_updateSaveDataWithStruct tests writing struct to save data
func TestCharacterSaveData_updateSaveDataWithStruct(t *testing.T) {
	tests := []struct {
		name   string
		rp     uint16
		kqf    []byte
		wantRP uint16
	}{
		{
			name:   "update_rp_value",
			rp:     1234,
			kqf:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			wantRP: 1234,
		},
		{
			name:   "zero_rp_value",
			rp:     0,
			kqf:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantRP: 0,
		},
		{
			name:   "max_rp_value",
			rp:     65535,
			kqf:    []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			wantRP: 65535,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			save := &CharacterSaveData{
				Mode:       cfg.G10,
				Pointers:   getPointers(cfg.G10),
				decompSave: make([]byte, 150000),
				RP:         tt.rp,
				KQF:        tt.kqf,
			}

			save.updateSaveDataWithStruct()

			// Verify RP was written correctly
			rpOffset := save.Pointers[pRP]
			gotRP := binary.LittleEndian.Uint16(save.decompSave[rpOffset : rpOffset+2])
			if gotRP != tt.wantRP {
				t.Errorf("RP in save data = %d, want %d", gotRP, tt.wantRP)
			}

			// Verify KQF was written correctly
			kqfOffset := save.Pointers[pKQF]
			gotKQF := save.decompSave[kqfOffset : kqfOffset+8]
			if !bytes.Equal(gotKQF, tt.kqf) {
				t.Errorf("KQF in save data = %v, want %v", gotKQF, tt.kqf)
			}
		})
	}
}

// TestHandleMsgMhfSexChanger tests the sex changer handler
func TestHandleMsgMhfSexChanger(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
	}{
		{
			name:      "basic_sex_change",
			ackHandle: 1234,
		},
		{
			name:      "different_ack_handle",
			ackHandle: 9999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)

			pkt := &mhfpacket.MsgMhfSexChanger{
				AckHandle: tt.ackHandle,
			}

			handleMsgMhfSexChanger(s, pkt)

			// Verify ACK was sent
			if len(s.sendPackets) == 0 {
				t.Fatal("no ACK packet was sent")
			}

			// Drain the channel
			<-s.sendPackets
		})
	}
}

// TestGetCharacterSaveData_Integration tests retrieving character save data from database
func TestGetCharacterSaveData_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tests := []struct {
		name           string
		charName       string
		isNewCharacter bool
		wantError      bool
	}{
		{
			name:           "existing_character",
			charName:       "TestChar",
			isNewCharacter: false,
			wantError:      false,
		},
		{
			name:           "new_character",
			charName:       "NewChar",
			isNewCharacter: true,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user and character
			userID := CreateTestUser(t, db, "testuser_"+tt.name)
			charID := CreateTestCharacter(t, db, userID, tt.charName)

			// Update is_new_character flag
			_, err := db.Exec("UPDATE characters SET is_new_character = $1 WHERE id = $2", tt.isNewCharacter, charID)
			if err != nil {
				t.Fatalf("Failed to update character: %v", err)
			}

			// Create test session
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = charID
			SetTestDB(s.server, db)
			s.server.erupeConfig.RealClientMode = cfg.Z2

			// Get character save data
			saveData, err := GetCharacterSaveData(s, charID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetCharacterSaveData() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if saveData == nil {
					t.Fatal("saveData is nil")
				}

				if saveData.CharID != charID {
					t.Errorf("CharID = %d, want %d", saveData.CharID, charID)
				}

				if saveData.Name != tt.charName {
					t.Errorf("Name = %q, want %q", saveData.Name, tt.charName)
				}

				if saveData.IsNewCharacter != tt.isNewCharacter {
					t.Errorf("IsNewCharacter = %v, want %v", saveData.IsNewCharacter, tt.isNewCharacter)
				}
			}
		})
	}
}

// TestCharacterSaveData_Save_Integration tests saving character data to database
func TestCharacterSaveData_Save_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// Create test user and character
	userID := CreateTestUser(t, db, "savetest")
	charID := CreateTestCharacter(t, db, userID, "SaveChar")

	// Create test session
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = charID
	SetTestDB(s.server, db)
	s.server.erupeConfig.RealClientMode = cfg.Z2

	// Load character save data
	saveData, err := GetCharacterSaveData(s, charID)
	if err != nil {
		t.Fatalf("Failed to get save data: %v", err)
	}

	// Modify save data
	saveData.HR = 999
	saveData.GR = 100
	saveData.Gender = true
	saveData.WeaponType = 5
	saveData.WeaponID = 1234

	// Save it
	saveData.Save(s)

	// Reload and verify
	var hr, gr uint16
	var gender bool
	var weaponType uint8
	var weaponID uint16

	err = db.QueryRow("SELECT hr, gr, is_female, weapon_type, weapon_id FROM characters WHERE id = $1",
		charID).Scan(&hr, &gr, &gender, &weaponType, &weaponID)
	if err != nil {
		t.Fatalf("Failed to query updated character: %v", err)
	}

	if hr != 999 {
		t.Errorf("HR = %d, want 999", hr)
	}
	if gr != 100 {
		t.Errorf("GR = %d, want 100", gr)
	}
	if !gender {
		t.Error("Gender should be true (female)")
	}
	if weaponType != 5 {
		t.Errorf("WeaponType = %d, want 5", weaponType)
	}
	if weaponID != 1234 {
		t.Errorf("WeaponID = %d, want 1234", weaponID)
	}
}

// TestGRPtoGR tests the GRP to GR conversion function
func TestGRPtoGR(t *testing.T) {
	tests := []struct {
		name   string
		grp    int
		wantGR uint16
	}{
		{
			name:   "zero_grp",
			grp:    0,
			wantGR: 1, // Function returns 1 for 0 GRP
		},
		{
			name:   "low_grp",
			grp:    10000,
			wantGR: 10, // Function returns 10 for 10000 GRP
		},
		{
			name:   "mid_grp",
			grp:    500000,
			wantGR: 88, // Function returns 88 for 500000 GRP
		},
		{
			name:   "high_grp",
			grp:    2000000,
			wantGR: 265, // Function returns 265 for 2000000 GRP
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGR := grpToGR(tt.grp)
			if gotGR != tt.wantGR {
				t.Errorf("grpToGR(%d) = %d, want %d", tt.grp, gotGR, tt.wantGR)
			}
		})
	}
}

// BenchmarkCompress benchmarks savedata compression
func BenchmarkCompress(b *testing.B) {
	data := bytes.Repeat([]byte{0xAA, 0xBB, 0xCC, 0xDD}, 25000) // 100KB
	save := &CharacterSaveData{
		decompSave: data,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = save.Compress()
	}
}

// BenchmarkDecompress benchmarks savedata decompression
func BenchmarkDecompress(b *testing.B) {
	data := bytes.Repeat([]byte{0xAA, 0xBB, 0xCC, 0xDD}, 25000)
	compressed, _ := nullcomp.Compress(data)

	save := &CharacterSaveData{
		compSave: compressed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = save.Decompress()
	}
}
