package channelserver

import (
	"encoding/json"
	"testing"
)

func TestGuildIconScan_Bytes(t *testing.T) {
	jsonData := []byte(`{"Parts":[{"Index":1,"ID":100,"Page":2,"Size":3,"Rotation":4,"Red":255,"Green":128,"Blue":0,"PosX":50,"PosY":60}]}`)

	gi := &GuildIcon{}
	err := gi.Scan(jsonData)
	if err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}

	if len(gi.Parts) != 1 {
		t.Fatalf("Parts length = %d, want 1", len(gi.Parts))
	}

	part := gi.Parts[0]
	if part.Index != 1 {
		t.Errorf("Index = %d, want 1", part.Index)
	}
	if part.ID != 100 {
		t.Errorf("ID = %d, want 100", part.ID)
	}
	if part.Page != 2 {
		t.Errorf("Page = %d, want 2", part.Page)
	}
	if part.Size != 3 {
		t.Errorf("Size = %d, want 3", part.Size)
	}
	if part.Rotation != 4 {
		t.Errorf("Rotation = %d, want 4", part.Rotation)
	}
	if part.Red != 255 {
		t.Errorf("Red = %d, want 255", part.Red)
	}
	if part.Green != 128 {
		t.Errorf("Green = %d, want 128", part.Green)
	}
	if part.Blue != 0 {
		t.Errorf("Blue = %d, want 0", part.Blue)
	}
	if part.PosX != 50 {
		t.Errorf("PosX = %d, want 50", part.PosX)
	}
	if part.PosY != 60 {
		t.Errorf("PosY = %d, want 60", part.PosY)
	}
}

func TestGuildIconScan_String(t *testing.T) {
	jsonStr := `{"Parts":[{"Index":5,"ID":200,"Page":1,"Size":2,"Rotation":0,"Red":100,"Green":50,"Blue":25,"PosX":300,"PosY":400}]}`

	gi := &GuildIcon{}
	err := gi.Scan(jsonStr)
	if err != nil {
		t.Fatalf("Scan(string) error = %v", err)
	}

	if len(gi.Parts) != 1 {
		t.Fatalf("Parts length = %d, want 1", len(gi.Parts))
	}
	if gi.Parts[0].ID != 200 {
		t.Errorf("ID = %d, want 200", gi.Parts[0].ID)
	}
	if gi.Parts[0].PosX != 300 {
		t.Errorf("PosX = %d, want 300", gi.Parts[0].PosX)
	}
}

func TestGuildIconScan_MultipleParts(t *testing.T) {
	jsonData := []byte(`{"Parts":[{"Index":0,"ID":1,"Page":0,"Size":0,"Rotation":0,"Red":0,"Green":0,"Blue":0,"PosX":0,"PosY":0},{"Index":1,"ID":2,"Page":0,"Size":0,"Rotation":0,"Red":0,"Green":0,"Blue":0,"PosX":0,"PosY":0},{"Index":2,"ID":3,"Page":0,"Size":0,"Rotation":0,"Red":0,"Green":0,"Blue":0,"PosX":0,"PosY":0}]}`)

	gi := &GuildIcon{}
	err := gi.Scan(jsonData)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(gi.Parts) != 3 {
		t.Fatalf("Parts length = %d, want 3", len(gi.Parts))
	}
	for i, part := range gi.Parts {
		if part.Index != uint16(i) {
			t.Errorf("Parts[%d].Index = %d, want %d", i, part.Index, i)
		}
	}
}

func TestGuildIconScan_EmptyParts(t *testing.T) {
	gi := &GuildIcon{}
	err := gi.Scan([]byte(`{"Parts":[]}`))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(gi.Parts) != 0 {
		t.Errorf("Parts length = %d, want 0", len(gi.Parts))
	}
}

func TestGuildIconScan_InvalidJSON(t *testing.T) {
	gi := &GuildIcon{}
	err := gi.Scan([]byte(`{invalid`))
	if err == nil {
		t.Error("Scan() with invalid JSON should return error")
	}
}

func TestGuildIconScan_InvalidJSONString(t *testing.T) {
	gi := &GuildIcon{}
	err := gi.Scan("{invalid")
	if err == nil {
		t.Error("Scan() with invalid JSON string should return error")
	}
}

func TestGuildIconScan_UnsupportedType(t *testing.T) {
	gi := &GuildIcon{}
	// Passing an unsupported type should not error (just no-op)
	err := gi.Scan(12345)
	if err != nil {
		t.Errorf("Scan(int) unexpected error = %v", err)
	}
}

func TestGuildIconValue(t *testing.T) {
	gi := &GuildIcon{
		Parts: []GuildIconPart{
			{Index: 1, ID: 100, Page: 2, Size: 3, Rotation: 4, Red: 255, Green: 128, Blue: 0, PosX: 50, PosY: 60},
		},
	}

	val, err := gi.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	jsonBytes, ok := val.([]byte)
	if !ok {
		t.Fatalf("Value() returned %T, want []byte", val)
	}

	// Verify round-trip
	gi2 := &GuildIcon{}
	err = json.Unmarshal(jsonBytes, gi2)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if len(gi2.Parts) != 1 {
		t.Fatalf("round-trip Parts length = %d, want 1", len(gi2.Parts))
	}
	if gi2.Parts[0].ID != 100 {
		t.Errorf("round-trip ID = %d, want 100", gi2.Parts[0].ID)
	}
	if gi2.Parts[0].Red != 255 {
		t.Errorf("round-trip Red = %d, want 255", gi2.Parts[0].Red)
	}
}

func TestGuildIconValue_Empty(t *testing.T) {
	gi := &GuildIcon{}
	val, err := gi.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	if val == nil {
		t.Error("Value() should not return nil")
	}
}

func TestGuildIconScanValueRoundTrip(t *testing.T) {
	original := &GuildIcon{
		Parts: []GuildIconPart{
			{Index: 0, ID: 10, Page: 1, Size: 2, Rotation: 45, Red: 200, Green: 150, Blue: 100, PosX: 500, PosY: 600},
			{Index: 1, ID: 20, Page: 3, Size: 4, Rotation: 90, Red: 50, Green: 75, Blue: 255, PosX: 100, PosY: 200},
		},
	}

	// Value -> Scan round trip
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	restored := &GuildIcon{}
	err = restored.Scan(val)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(restored.Parts) != len(original.Parts) {
		t.Fatalf("Parts length = %d, want %d", len(restored.Parts), len(original.Parts))
	}

	for i := range original.Parts {
		if restored.Parts[i] != original.Parts[i] {
			t.Errorf("Parts[%d] mismatch: got %+v, want %+v", i, restored.Parts[i], original.Parts[i])
		}
	}
}

func TestFestivalColorCodes(t *testing.T) {
	tests := []struct {
		colour FestivalColor
		code   int16
	}{
		{FestivalColorBlue, 0},
		{FestivalColorRed, 1},
		{FestivalColorNone, -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.colour), func(t *testing.T) {
			code, ok := FestivalColorCodes[tt.colour]
			if !ok {
				t.Fatalf("FestivalColorCodes missing key %s", tt.colour)
			}
			if code != tt.code {
				t.Errorf("FestivalColorCodes[%s] = %d, want %d", tt.colour, code, tt.code)
			}
		})
	}
}

func TestFestivalColorConstants(t *testing.T) {
	if FestivalColorNone != "none" {
		t.Errorf("FestivalColorNone = %s, want none", FestivalColorNone)
	}
	if FestivalColorRed != "red" {
		t.Errorf("FestivalColorRed = %s, want red", FestivalColorRed)
	}
	if FestivalColorBlue != "blue" {
		t.Errorf("FestivalColorBlue = %s, want blue", FestivalColorBlue)
	}
}

func TestGuildApplicationTypeConstants(t *testing.T) {
	if GuildApplicationTypeApplied != "applied" {
		t.Errorf("GuildApplicationTypeApplied = %s, want applied", GuildApplicationTypeApplied)
	}
	if GuildApplicationTypeInvited != "invited" {
		t.Errorf("GuildApplicationTypeInvited = %s, want invited", GuildApplicationTypeInvited)
	}
}
