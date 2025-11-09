package mhfmon

import (
	"testing"
)

func TestMonsters_Length(t *testing.T) {
	// Verify that the Monsters slice has entries
	actualLen := len(Monsters)
	if actualLen == 0 {
		t.Fatal("Monsters slice is empty")
	}
	// The slice has 177 entries (some constants may not have entries)
	if actualLen < 170 {
		t.Errorf("Monsters length = %d, seems too small", actualLen)
	}
}

func TestMonsters_IndexMatchesConstant(t *testing.T) {
	// Test that the index in the slice matches the constant value
	tests := []struct {
		index int
		name  string
		large bool
	}{
		{Mon0, "Mon0", false},
		{Rathian, "Rathian", true},
		{Fatalis, "Fatalis", true},
		{Kelbi, "Kelbi", false},
		{Rathalos, "Rathalos", true},
		{Diablos, "Diablos", true},
		{Rajang, "Rajang", true},
		{Zinogre, "Zinogre", true},
		{Deviljho, "Deviljho", true},
		{KingShakalaka, "King Shakalaka", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.index >= len(Monsters) {
				t.Fatalf("Index %d out of bounds", tt.index)
			}
			monster := Monsters[tt.index]
			if monster.Name != tt.name {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, monster.Name, tt.name)
			}
			if monster.Large != tt.large {
				t.Errorf("Monsters[%d].Large = %v, want %v", tt.index, monster.Large, tt.large)
			}
		})
	}
}

func TestMonsters_AllLargeMonsters(t *testing.T) {
	// Verify some known large monsters
	largeMonsters := []int{
		Rathian,
		Fatalis,
		YianKutKu,
		LaoShanLung,
		Cephadrome,
		Rathalos,
		Diablos,
		Khezu,
		Gravios,
		Tigrex,
		Zinogre,
		Deviljho,
		Brachydios,
	}

	for _, idx := range largeMonsters {
		if !Monsters[idx].Large {
			t.Errorf("Monsters[%d] (%s) should be marked as large", idx, Monsters[idx].Name)
		}
	}
}

func TestMonsters_AllSmallMonsters(t *testing.T) {
	// Verify some known small monsters
	smallMonsters := []int{
		Kelbi,
		Mosswine,
		Bullfango,
		Felyne,
		Aptonoth,
		Genprey,
		Velociprey,
		Melynx,
		Hornetaur,
		Apceros,
		Ioprey,
		Giaprey,
		Cephalos,
		Blango,
		Conga,
		Remobra,
		GreatThunderbug,
		Shakalaka,
	}

	for _, idx := range smallMonsters {
		if Monsters[idx].Large {
			t.Errorf("Monsters[%d] (%s) should be marked as small", idx, Monsters[idx].Name)
		}
	}
}

func TestMonsters_Constants(t *testing.T) {
	// Test that constants have expected values
	tests := []struct {
		constant int
		expected int
	}{
		{Mon0, 0},
		{Rathian, 1},
		{Fatalis, 2},
		{Kelbi, 3},
		{Rathalos, 11},
		{Diablos, 14},
		{Rajang, 53},
		{Zinogre, 146},
		{Deviljho, 147},
		{Brachydios, 148},
		{KingShakalaka, 176},
	}

	for _, tt := range tests {
		if tt.constant != tt.expected {
			t.Errorf("Constant = %d, want %d", tt.constant, tt.expected)
		}
	}
}

func TestMonsters_NameConsistency(t *testing.T) {
	// Test that specific monsters have correct names
	tests := []struct {
		index        int
		expectedName string
	}{
		{Rathian, "Rathian"},
		{Rathalos, "Rathalos"},
		{YianKutKu, "Yian Kut-Ku"},
		{LaoShanLung, "Lao-Shan Lung"},
		{KushalaDaora, "Kushala Daora"},
		{Tigrex, "Tigrex"},
		{Rajang, "Rajang"},
		{Zinogre, "Zinogre"},
		{Deviljho, "Deviljho"},
		{Brachydios, "Brachydios"},
		{Nargacuga, "Nargacuga"},
		{GoreMagala, "Gore Magala"},
		{ShagaruMagala, "Shagaru Magala"},
		{KingShakalaka, "King Shakalaka"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			if Monsters[tt.index].Name != tt.expectedName {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.expectedName)
			}
		})
	}
}

func TestMonsters_SubspeciesNames(t *testing.T) {
	// Test subspecies have appropriate names
	tests := []struct {
		index        int
		expectedName string
	}{
		{PinkRathian, "Pink Rathian"},
		{AzureRathalos, "Azure Rathalos"},
		{SilverRathalos, "Silver Rathalos"},
		{GoldRathian, "Gold Rathian"},
		{BlackDiablos, "Black Diablos"},
		{WhiteMonoblos, "White Monoblos"},
		{RedKhezu, "Red Khezu"},
		{CrimsonFatalis, "Crimson Fatalis"},
		{WhiteFatalis, "White Fatalis"},
		{StygianZinogre, "Stygian Zinogre"},
		{SavageDeviljho, "Savage Deviljho"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			if Monsters[tt.index].Name != tt.expectedName {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.expectedName)
			}
		})
	}
}

func TestMonsters_PlaceholderMonsters(t *testing.T) {
	// Test that placeholder monsters exist
	placeholders := []int{Mon0, Mon18, Mon29, Mon32, Mon72, Mon86, Mon87, Mon88, Mon118, Mon133, Mon134, Mon135, Mon136, Mon137, Mon138, Mon156, Mon168, Mon171}

	for _, idx := range placeholders {
		if idx >= len(Monsters) {
			t.Errorf("Placeholder monster index %d out of bounds", idx)
			continue
		}
		// Placeholder monsters should be marked as small (non-large)
		if Monsters[idx].Large {
			t.Errorf("Placeholder Monsters[%d] (%s) should not be marked as large", idx, Monsters[idx].Name)
		}
	}
}

func TestMonsters_FrontierMonsters(t *testing.T) {
	// Test some MH Frontier-specific monsters
	frontierMonsters := []struct {
		index int
		name  string
	}{
		{Espinas, "Espinas"},
		{Berukyurosu, "Berukyurosu"},
		{Pariapuria, "Pariapuria"},
		{Raviente, "Raviente"},
		{Dyuragaua, "Dyuragaua"},
		{Doragyurosu, "Doragyurosu"},
		{Gurenzeburu, "Gurenzeburu"},
		{Rukodiora, "Rukodiora"},
		{Gogomoa, "Gogomoa"},
		{Disufiroa, "Disufiroa"},
		{Rebidiora, "Rebidiora"},
		{MiRu, "Mi-Ru"},
		{Shantien, "Shantien"},
		{Zerureusu, "Zerureusu"},
		{GarubaDaora, "Garuba Daora"},
		{Harudomerugu, "Harudomerugu"},
		{Toridcless, "Toridcless"},
		{Guanzorumu, "Guanzorumu"},
		{Egyurasu, "Egyurasu"},
		{Bogabadorumu, "Bogabadorumu"},
	}

	for _, tt := range frontierMonsters {
		t.Run(tt.name, func(t *testing.T) {
			if tt.index >= len(Monsters) {
				t.Fatalf("Index %d out of bounds", tt.index)
			}
			if Monsters[tt.index].Name != tt.name {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.name)
			}
			// Most Frontier monsters should be large
			if !Monsters[tt.index].Large {
				t.Logf("Frontier monster %s is marked as small", tt.name)
			}
		})
	}
}

func TestMonsters_DuremudiraVariants(t *testing.T) {
	// Test Duremudira variants
	tests := []struct {
		index int
		name  string
	}{
		{Block1Duremudira, "1st Block Duremudira"},
		{Block2Duremudira, "2nd Block Duremudira"},
		{MusouDuremudira, "Musou Duremudira"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Monsters[tt.index].Name != tt.name {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.name)
			}
			if !Monsters[tt.index].Large {
				t.Errorf("Duremudira variant should be marked as large")
			}
		})
	}
}

func TestMonsters_RalienteVariants(t *testing.T) {
	// Test Raviente variants
	tests := []struct {
		index int
		name  string
	}{
		{Raviente, "Raviente"},
		{BerserkRaviente, "Berserk Raviente"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Monsters[tt.index].Name != tt.name {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.name)
			}
			if !Monsters[tt.index].Large {
				t.Errorf("Raviente variant should be marked as large")
			}
		})
	}
}

func TestMonsters_NoHoles(t *testing.T) {
	// Verify that there are no nil entries or empty names (except for placeholder "MonXX" entries)
	for i, monster := range Monsters {
		if monster.Name == "" {
			t.Errorf("Monsters[%d] has empty name", i)
		}
	}
}

func TestMonster_Struct(t *testing.T) {
	// Test that Monster struct is properly defined
	m := Monster{
		Name:  "Test Monster",
		Large: true,
	}

	if m.Name != "Test Monster" {
		t.Errorf("Name = %q, want %q", m.Name, "Test Monster")
	}
	if !m.Large {
		t.Error("Large should be true")
	}
}

func BenchmarkAccessMonster(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Monsters[Rathalos]
	}
}

func BenchmarkAccessMonsterName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Monsters[Zinogre].Name
	}
}

func BenchmarkAccessMonsterLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Monsters[Deviljho].Large
	}
}

func TestMonsters_CrossoverMonsters(t *testing.T) {
	// Test crossover monsters (from other games)
	tests := []struct {
		index int
		name  string
	}{
		{Zinogre, "Zinogre"},        // From MH Portable 3rd
		{Deviljho, "Deviljho"},      // From MH3
		{Brachydios, "Brachydios"},  // From MH3G
		{Barioth, "Barioth"},        // From MH3
		{Uragaan, "Uragaan"},        // From MH3
		{Nargacuga, "Nargacuga"},    // From MH Freedom Unite
		{GoreMagala, "Gore Magala"}, // From MH4
		{Amatsu, "Amatsu"},          // From MH Portable 3rd
		{Seregios, "Seregios"},      // From MH4G
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Monsters[tt.index].Name != tt.name {
				t.Errorf("Monsters[%d].Name = %q, want %q", tt.index, Monsters[tt.index].Name, tt.name)
			}
			if !Monsters[tt.index].Large {
				t.Errorf("Crossover large monster %s should be marked as large", tt.name)
			}
		})
	}
}
