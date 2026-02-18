package mhfmon

const (
	Mon0 = iota
	Rathian
	Fatalis
	Kelbi
	Mosswine
	Bullfango
	YianKutKu
	LaoShanLung
	Cephadrome
	Felyne
	VeggieElder
	Rathalos
	Aptonoth
	Genprey
	Diablos
	Khezu
	Velociprey
	Gravios
	Mon18
	Vespoid
	Gypceros
	Plesioth
	Basarios
	Melynx
	Hornetaur
	Apceros
	Monoblos
	Velocidrome
	Gendrome
	Mon29
	Ioprey
	Iodrome
	Mon32
	Kirin
	Cephalos
	Giaprey
	CrimsonFatalis
	PinkRathian
	BlueYianKutKu
	PurpleGypceros
	YianGaruga
	SilverRathalos
	GoldRathian
	BlackDiablos
	WhiteMonoblos
	RedKhezu
	GreenPlesioth
	BlackGravios
	DaimyoHermitaur
	AzureRathalos
	AshenLaoShanLung
	Blangonga
	Congalala
	Rajang
	KushalaDaora
	ShenGaoren
	GreatThunderbug
	Shakalaka
	YamaTsukami
	Chameleos
	RustedKushalaDaora
	Blango
	Conga
	Remobra
	Lunastra
	Teostra
	Hermitaur
	ShogunCeanataur
	Bulldrome
	Anteka
	Popo
	WhiteFatalis
	Mon72
	Ceanataur
	Hypnocatrice
	Lavasioth
	Tigrex
	Akantor
	BrightHypnoc
	RedLavasioth
	Espinas
	BurningEspinas
	WhiteHypnoc
	AqraVashimu
	AqraJebia
	Berukyurosu
	Mon86
	Mon87
	Mon88
	Pariapuria
	PearlEspinas
	KamuOrugaron
	NonoOrugaron
	Raviente
	Dyuragaua
	Doragyurosu
	Gurenzeburu
	Burukku
	Erupe
	Rukodiora
	Unknown
	Gogomoa
	Kokomoa
	TaikunZamuza
	Abiorugu
	Kuarusepusu
	Odibatorasu
	Disufiroa
	Rebidiora
	Anorupatisu
	Hyujikiki
	Midogaron
	Giaorugu
	MiRu
	Farunokku
	Pokaradon
	Shantien
	Pokara
	Mon118
	Goruganosu
	Aruganosu
	Baruragaru
	Zerureusu
	Gougarf
	Uruki
	Forokururu
	Meraginasu
	Diorex
	GarubaDaora
	Inagami
	Varusaburosu
	Poborubarumu
	Block1Duremudira
	Mon133
	Mon134
	Mon135
	Mon136
	Mon137
	Mon138
	Gureadomosu
	Harudomerugu
	Toridcless
	Gasurabazura
	Kusubami
	YamaKurai
	Block2Duremudira
	Zinogre
	Deviljho
	Brachydios
	BerserkRaviente
	ToaTesukatora
	Barioth
	Uragaan
	StygianZinogre
	Guanzorumu
	SavageDeviljho
	Mon156
	Egyurasu
	Voljang
	Nargacuga
	Keoaruboru
	Zenaserisu
	GoreMagala
	BlinkingNargacuga
	ShagaruMagala
	Amatsu
	Eruzerion
	MusouDuremudira
	Mon168
	Seregios
	Bogabadorumu
	Mon171
	MusouBogabadorumu
	CostumedUruki
	MusouZerureusu
	Rappy
	KingShakalaka
)

// Monster holds display metadata for a single monster species.
type Monster struct {
	Name  string
	Large bool
}

// Monsters is an ordered table of all MHF monsters, indexed by enemy ID.
var Monsters = []Monster{
	{"Mon0", false},
	{"Rathian", true},
	{"Fatalis", true},
	{"Kelbi", false},
	{"Mosswine", false},
	{"Bullfango", false},
	{"Yian Kut-Ku", true},
	{"Lao-Shan Lung", true},
	{"Cephadrome", true},
	{"Felyne", false},
	{"Veggie Elder", false},
	{"Rathalos", true},
	{"Aptonoth", false},
	{"Genprey", false},
	{"Diablos", true},
	{"Khezu", true},
	{"Velociprey", false},
	{"Gravios", true},
	{"Mon18", false},
	{"Vespoid", false},
	{"Gypceros", true},
	{"Plesioth", true},
	{"Basarios", true},
	{"Melynx", false},
	{"Hornetaur", false},
	{"Apceros", false},
	{"Monoblos", true},
	{"Velocidrome", true},
	{"Gendrome", true},
	{"Mon29", false},
	{"Ioprey", false},
	{"Iodrome", true},
	{"Mon32", false},
	{"Kirin", true},
	{"Cephalos", false},
	{"Giaprey", false},
	{"Crimson Fatalis", true},
	{"Pink Rathian", true},
	{"Blue Yian Kut-Ku", true},
	{"Purple Gypceros", true},
	{"Yian Garuga", true},
	{"Silver Rathalos", true},
	{"Gold Rathian", true},
	{"Black Diablos", true},
	{"White Monoblos", true},
	{"Red Khezu", true},
	{"Green Plesioth", true},
	{"Black Gravios", true},
	{"Daimyo Hermitaur", true},
	{"Azure Rathalos", true},
	{"Ashen Lao-Shan Lung", true},
	{"Blangonga", true},
	{"Congalala", true},
	{"Rajang", true},
	{"Kushala Daora", true},
	{"Shen Gaoren", true},
	{"Great Thunderbug", false},
	{"Shakalaka", false},
	{"Yama Tsukami", true},
	{"Chameleos", true},
	{"Rusted Kushala Daora", true},
	{"Blango", false},
	{"Conga", false},
	{"Remobra", false},
	{"Lunastra", true},
	{"Teostra", true},
	{"Hermitaur", false},
	{"Shogun Ceanataur", true},
	{"Bulldrome", true},
	{"Anteka", false},
	{"Popo", false},
	{"White Fatalis", true},
	{"Mon72", false},
	{"Ceanataur", false},
	{"Hypnocatrice", true},
	{"Lavasioth", true},
	{"Tigrex", true},
	{"Akantor", true},
	{"Bright Hypnocatrice", true},
	{"Red Lavasioth", true},
	{"Espinas", true},
	{"Burning Espinas", true},
	{"White Hypnocatrice", true},
	{"Aqra Vashimu", true},
	{"Aqra Jebia", true},
	{"Berukyurosu", true},
	{"Mon86", false},
	{"Mon87", false},
	{"Mon88", false},
	{"Pariapuria", true},
	{"Pearl Espinas", true},
	{"Kamu Orugaron", true},
	{"Nono Orugaron", true},
	{"Raviente", true}, // + Violent
	{"Dyuragaua", true},
	{"Doragyurosu", true},
	{"Gurenzeburu", true},
	{"Burukku", false},
	{"Erupe", false},
	{"Rukodiora", true},
	{"Unknown", true},
	{"Gogomoa", true},
	{"Kokomoa", false},
	{"Taikun Zamuza", true},
	{"Abiorugu", true},
	{"Kuarusepusu", true},
	{"Odibatorasu", true},
	{"Disufiroa", true},
	{"Rebidiora", true},
	{"Anorupatisu", true},
	{"Hyujikiki", true},
	{"Midogaron", true},
	{"Giaorugu", true},
	{"Mi-Ru", true}, // + Musou
	{"Farunokku", true},
	{"Pokaradon", true},
	{"Shantien", true},
	{"Pokara", false},
	{"Mon118", false},
	{"Goruganosu", true},
	{"Aruganosu", true},
	{"Baruragaru", true},
	{"Zerureusu", true},
	{"Gougarf", true}, // Both
	{"Uruki", false},
	{"Forokururu", true},
	{"Meraginasu", true},
	{"Diorex", true},
	{"Garuba Daora", true},
	{"Inagami", true},
	{"Varusablos", true},
	{"Poborubarumu", true},
	{"1st Block Duremudira", true},
	{"Mon133", false},
	{"Mon134", false},
	{"Mon135", false},
	{"Mon136", false},
	{"Mon137", false},
	{"Mon138", false},
	{"Gureadomosu", true},
	{"Harudomerugu", true},
	{"Toridcless", true},
	{"Gasurabazura", true},
	{"Kusubami", false},
	{"Yama Kurai", true},
	{"2nd Block Duremudira", true},
	{"Zinogre", true},
	{"Deviljho", true},
	{"Brachydios", true},
	{"Berserk Raviente", true},
	{"Toa Tesukatora", true},
	{"Barioth", true},
	{"Uragaan", true},
	{"Stygian Zinogre", true},
	{"Guanzorumu", true},
	{"Savage Deviljho", true}, // + Starving/Heavenly
	{"Mon156", false},
	{"Egyurasu", false},
	{"Voljang", true},
	{"Nargacuga", true},
	{"Keoaruboru", true},
	{"Zenaserisu", true},
	{"Gore Magala", true},
	{"Blinking Nargacuga", true},
	{"Shagaru Magala", true},
	{"Amatsu", true},
	{"Eruzerion", true}, // + Musou
	{"Musou Duremudira", true},
	{"Mon168", false},
	{"Seregios", true},
	{"Bogabadorumu", true},
	{"Mon171", false},
	{"Musou Bogabadorumu", true},
	{"Costumed Uruki", false},
	{"Musou Zerureusu", true},
	{"Rappy", false},
	{"King Shakalaka", false},
}
