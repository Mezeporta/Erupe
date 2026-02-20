package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	"erupe-ce/common/token"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

const warehouseNamesQuery = `
SELECT
COALESCE(item0name, ''),
COALESCE(item1name, ''),
COALESCE(item2name, ''),
COALESCE(item3name, ''),
COALESCE(item4name, ''),
COALESCE(item5name, ''),
COALESCE(item6name, ''),
COALESCE(item7name, ''),
COALESCE(item8name, ''),
COALESCE(item9name, ''),
COALESCE(equip0name, ''),
COALESCE(equip1name, ''),
COALESCE(equip2name, ''),
COALESCE(equip3name, ''),
COALESCE(equip4name, ''),
COALESCE(equip5name, ''),
COALESCE(equip6name, ''),
COALESCE(equip7name, ''),
COALESCE(equip8name, ''),
COALESCE(equip9name, '')
FROM warehouse
`

func handleMsgMhfUpdateInterior(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateInterior)
	if len(pkt.InteriorData) > 64 {
		s.logger.Warn("Interior payload too large", zap.Int("len", len(pkt.InteriorData)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if _, err := s.server.db.Exec(`UPDATE user_binary SET house_furniture=$1 WHERE id=$2`, pkt.InteriorData, s.charID); err != nil {
		s.logger.Error("Failed to update house furniture", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// HouseData represents player house/my house data.
type HouseData struct {
	CharID        uint32 `db:"id"`
	HR            uint16 `db:"hr"`
	GR            uint16 `db:"gr"`
	Name          string `db:"name"`
	HouseState    uint8  `db:"house_state"`
	HousePassword string `db:"house_password"`
}

func handleMsgMhfEnumerateHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateHouse)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	var houses []HouseData
	houseQuery := `SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
		FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE c.id=$1`
	switch pkt.Method {
	case 1:
		friendsList, _ := s.server.charRepo.ReadString(s.charID, "friends")
		cids := stringsupport.CSVElems(friendsList)
		for _, cid := range cids {
			house := HouseData{}
			row := s.server.db.QueryRowx(houseQuery, cid)
			err := row.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
			}
		}
	case 2:
		guild, err := GetGuildInfoByCharacterId(s, s.charID)
		if err != nil || guild == nil {
			break
		}
		guildMembers, err := GetGuildMembers(s, guild.ID, false)
		if err != nil {
			break
		}
		for _, member := range guildMembers {
			house := HouseData{}
			row := s.server.db.QueryRowx(houseQuery, member.CharID)
			err = row.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
			}
		}
	case 3:
		houseQuery = `SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
			FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE name ILIKE $1`
		house := HouseData{}
		rows, err := s.server.db.Queryx(houseQuery, fmt.Sprintf(`%%%s%%`, pkt.Name))
		if err != nil {
			s.logger.Error("Failed to query houses by name", zap.Error(err))
		} else {
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				if err := rows.StructScan(&house); err == nil {
					houses = append(houses, house)
				}
			}
		}
	case 4:
		house := HouseData{}
		row := s.server.db.QueryRowx(houseQuery, pkt.CharID)
		err := row.StructScan(&house)
		if err == nil {
			houses = append(houses, house)
		}
	case 5: // Recent visitors
		break
	}
	for _, house := range houses {
		bf.WriteUint32(house.CharID)
		bf.WriteUint8(house.HouseState)
		if len(house.HousePassword) > 0 {
			bf.WriteUint8(3)
		} else {
			bf.WriteUint8(0)
		}
		bf.WriteUint16(house.HR)
		if s.server.erupeConfig.RealClientMode >= _config.G10 {
			bf.WriteUint16(house.GR)
		}
		ps.Uint8(bf, house.Name, true)
	}
	_, _ = bf.Seek(0, 0)
	bf.WriteUint16(uint16(len(houses)))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateHouse)
	// 01 = closed
	// 02 = open anyone
	// 03 = open friends
	// 04 = open guild
	// 05 = open friends+guild
	if _, err := s.server.db.Exec(`UPDATE user_binary SET house_state=$1, house_password=$2 WHERE id=$3`, pkt.State, pkt.Password, s.charID); err != nil {
		s.logger.Error("Failed to update house state", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)
	bf := byteframe.NewByteFrame()

	state := uint8(2) // Default to password-protected if DB fails
	var password string
	if err := s.server.db.QueryRow(`SELECT COALESCE(house_state, 2) as house_state, COALESCE(house_password, '') as house_password FROM user_binary WHERE id=$1
	`, pkt.CharID).Scan(&state, &password); err != nil {
		s.logger.Error("Failed to read house state", zap.Error(err))
	}

	if pkt.Destination != 9 && len(pkt.Password) > 0 && pkt.CheckPass {
		if pkt.Password != password {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	if pkt.Destination != 9 && state > 2 {
		allowed := false

		// Friends list verification
		if state == 3 || state == 5 {
			friendsList, _ := s.server.charRepo.ReadString(pkt.CharID, "friends")
			cids := stringsupport.CSVElems(friendsList)
			for _, cid := range cids {
				if uint32(cid) == s.charID {
					allowed = true
					break
				}
			}
		}

		// Guild verification
		if state > 3 {
			ownGuild, err := GetGuildInfoByCharacterId(s, s.charID)
			isApplicant, _ := ownGuild.HasApplicationForCharID(s, s.charID)
			if err == nil && ownGuild != nil {
				othersGuild, err := GetGuildInfoByCharacterId(s, pkt.CharID)
				if err == nil && othersGuild != nil {
					if othersGuild.ID == ownGuild.ID && !isApplicant {
						allowed = true
					}
				}
			}
		}

		if !allowed {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	var houseTier, houseData, houseFurniture, bookshelf, gallery, tore, garden []byte
	_ = s.server.db.QueryRow(`SELECT house_tier, house_data, house_furniture, bookshelf, gallery, tore, garden FROM user_binary WHERE id=$1
	`, pkt.CharID).Scan(&houseTier, &houseData, &houseFurniture, &bookshelf, &gallery, &tore, &garden)
	if houseFurniture == nil {
		houseFurniture = make([]byte, 20)
	}

	switch pkt.Destination {
	case 3: // Others house
		bf.WriteBytes(houseTier)
		bf.WriteBytes(houseData)
		bf.WriteBytes(make([]byte, 19)) // Padding?
		bf.WriteBytes(houseFurniture)
	case 4: // Bookshelf
		bf.WriteBytes(bookshelf)
	case 5: // Gallery
		bf.WriteBytes(gallery)
	case 8: // Tore
		bf.WriteBytes(tore)
	case 9: // Own house
		bf.WriteBytes(houseFurniture)
	case 10: // Garden
		bf.WriteBytes(garden)
		goocoos := getGoocooData(s, pkt.CharID)
		bf.WriteUint16(uint16(len(goocoos)))
		bf.WriteUint16(0)
		for _, goocoo := range goocoos {
			bf.WriteBytes(goocoo)
		}
	}
	if len(bf.Data()) == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfGetMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetMyhouseInfo)
	var data []byte
	_ = s.server.db.QueryRow(`SELECT mission FROM user_binary WHERE id=$1`, s.charID).Scan(&data)
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 9))
	}
}

func handleMsgMhfUpdateMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateMyhouseInfo)
	if len(pkt.Data) > 512 {
		s.logger.Warn("MyhouseInfo payload too large", zap.Int("len", len(pkt.Data)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if _, err := s.server.db.Exec("UPDATE user_binary SET mission=$1 WHERE id=$2", pkt.Data, s.charID); err != nil {
		s.logger.Error("Failed to update myhouse mission", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadDecoMyset)
	defaultData := []byte{0x01, 0x00}
	if s.server.erupeConfig.RealClientMode < _config.G10 {
		defaultData = []byte{0x00, 0x00}
	}
	loadCharacterData(s, pkt.AckHandle, "decomyset", defaultData)
}

func handleMsgMhfSaveDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveDecoMyset)
	if len(pkt.RawDataPayload) < 3 {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	temp, err := s.server.charRepo.LoadColumn(s.charID, "decomyset")
	if err != nil {
		s.logger.Error("Failed to load decomyset", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	// Version handling
	bf := byteframe.NewByteFrame()
	var size uint
	if s.server.erupeConfig.RealClientMode >= _config.G10 {
		size = 76
		bf.WriteUint8(1)
	} else {
		size = 68
		bf.WriteUint8(0)
	}

	// Handle nil data
	if len(temp) == 0 {
		temp = append(bf.Data(), uint8(0))
	}

	// Build a map of set data
	sets := make(map[uint16][]byte)
	oldSets := byteframe.NewByteFrameFromBytes(temp[2:])
	for i := uint8(0); i < temp[1]; i++ {
		index := oldSets.ReadUint16()
		sets[index] = oldSets.ReadBytes(size)
	}

	// Overwrite existing sets
	newSets := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[2:])
	for i := uint8(0); i < pkt.RawDataPayload[1]; i++ {
		index := newSets.ReadUint16()
		sets[index] = newSets.ReadBytes(size)
	}

	// Serialise the set data
	bf.WriteUint8(uint8(len(sets)))
	for u, b := range sets {
		bf.WriteUint16(u)
		bf.WriteBytes(b)
	}

	dumpSaveData(s, bf.Data(), "decomyset")
	if err := s.server.charRepo.SaveColumn(s.charID, "decomyset", bf.Data()); err != nil {
		s.logger.Error("Failed to save decomyset", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// Title represents a hunter title entry.
type Title struct {
	ID       uint16    `db:"id"`
	Acquired time.Time `db:"unlocked_at"`
	Updated  time.Time `db:"updated_at"`
}

func handleMsgMhfEnumerateTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateTitle)
	var count uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0) // Unk
	rows, err := s.server.db.Queryx("SELECT id, unlocked_at, updated_at FROM titles WHERE char_id=$1", s.charID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	for rows.Next() {
		title := &Title{}
		err = rows.StructScan(&title)
		if err != nil {
			continue
		}
		count++
		bf.WriteUint16(title.ID)
		bf.WriteUint16(0) // Unk
		bf.WriteUint32(uint32(title.Acquired.Unix()))
		bf.WriteUint32(uint32(title.Updated.Unix()))
	}
	_, _ = bf.Seek(0, io.SeekStart)
	bf.WriteUint16(count)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTitle)
	for _, title := range pkt.TitleIDs {
		var exists int
		err := s.server.db.QueryRow(`SELECT count(*) FROM titles WHERE id=$1 AND char_id=$2`, title, s.charID).Scan(&exists)
		if err != nil || exists == 0 {
			if _, err := s.server.db.Exec(`INSERT INTO titles VALUES ($1, $2, now(), now())`, title, s.charID); err != nil {
				s.logger.Error("Failed to insert title", zap.Error(err))
			}
		} else {
			if _, err := s.server.db.Exec(`UPDATE titles SET updated_at=now() WHERE id=$1 AND char_id=$2`, title, s.charID); err != nil {
				s.logger.Error("Failed to update title", zap.Error(err))
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfResetTitle(s *Session, p mhfpacket.MHFPacket) {}

func initializeWarehouse(s *Session) {
	var t int
	err := s.server.db.QueryRow("SELECT character_id FROM warehouse WHERE character_id=$1", s.charID).Scan(&t)
	if err != nil {
		if _, err := s.server.db.Exec("INSERT INTO warehouse (character_id) VALUES ($1)", s.charID); err != nil {
			s.logger.Error("Failed to initialize warehouse", zap.Error(err))
		}
	}
}

func handleMsgMhfOperateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateWarehouse)
	initializeWarehouse(s)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Operation)
	switch pkt.Operation {
	case 0:
		var count uint8
		itemNames := make([]string, 10)
		equipNames := make([]string, 10)
		_ = s.server.db.QueryRow(fmt.Sprintf("%s WHERE character_id=$1", warehouseNamesQuery), s.charID).Scan(&itemNames[0],
			&itemNames[1], &itemNames[2], &itemNames[3], &itemNames[4], &itemNames[5], &itemNames[6], &itemNames[7], &itemNames[8], &itemNames[9], &equipNames[0],
			&equipNames[1], &equipNames[2], &equipNames[3], &equipNames[4], &equipNames[5], &equipNames[6], &equipNames[7], &equipNames[8], &equipNames[9])
		bf.WriteUint32(0)
		bf.WriteUint16(10000) // Usages
		temp := byteframe.NewByteFrame()
		for i, name := range itemNames {
			if len(name) > 0 {
				count++
				temp.WriteUint8(0)
				temp.WriteUint8(uint8(i))
				ps.Uint8(temp, name, true)
			}
		}
		for i, name := range equipNames {
			if len(name) > 0 {
				count++
				temp.WriteUint8(1)
				temp.WriteUint8(uint8(i))
				ps.Uint8(temp, name, true)
			}
		}
		bf.WriteUint8(count)
		bf.WriteBytes(temp.Data())
	case 1:
		bf.WriteUint8(0)
	case 2:
		if pkt.BoxIndex > 9 {
			break
		}
		switch pkt.BoxType {
		case 0:
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET item%dname=$1 WHERE character_id=$2", pkt.BoxIndex), pkt.Name, s.charID); err != nil {
				s.logger.Error("Failed to rename warehouse item box", zap.Error(err))
			}
		case 1:
			if _, err := s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET equip%dname=$1 WHERE character_id=$2", pkt.BoxIndex), pkt.Name, s.charID); err != nil {
				s.logger.Error("Failed to rename warehouse equip box", zap.Error(err))
			}
		}
	case 3:
		bf.WriteUint32(0)     // Usage renewal time, >1 = disabled
		bf.WriteUint16(10000) // Usages
	case 4:
		bf.WriteUint32(0)
		bf.WriteUint16(10000) // Usages
		bf.WriteUint8(0)
	}
	// Opcodes
	// 0 = Get box names
	// 1 = Commit usage
	// 2 = Rename
	// 3 = Get usage limit
	// 4 = Get gift box names (doesn't do anything?)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func addWarehouseItem(s *Session, item mhfitem.MHFItemStack) {
	giftBox := warehouseGetItems(s, 10)
	item.WarehouseID = token.RNG.Uint32()
	giftBox = append(giftBox, item)
	if _, err := s.server.db.Exec("UPDATE warehouse SET item10=$1 WHERE character_id=$2", mhfitem.SerializeWarehouseItems(giftBox), s.charID); err != nil {
		s.logger.Error("Failed to update warehouse gift box", zap.Error(err))
	}
}

func warehouseGetItems(s *Session, index uint8) []mhfitem.MHFItemStack {
	initializeWarehouse(s)
	var data []byte
	var items []mhfitem.MHFItemStack
	if index > 10 {
		return items
	}
	_ = s.server.db.QueryRow(fmt.Sprintf(`SELECT item%d FROM warehouse WHERE character_id=$1`, index), s.charID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}

func warehouseGetEquipment(s *Session, index uint8) []mhfitem.MHFEquipment {
	var data []byte
	var equipment []mhfitem.MHFEquipment
	if index > 10 {
		return equipment
	}
	_ = s.server.db.QueryRow(fmt.Sprintf(`SELECT equip%d FROM warehouse WHERE character_id=$1`, index), s.charID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			equipment = append(equipment, mhfitem.ReadWarehouseEquipment(box, s.server.erupeConfig.RealClientMode))
		}
	}
	return equipment
}

func handleMsgMhfEnumerateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateWarehouse)
	bf := byteframe.NewByteFrame()
	switch pkt.BoxType {
	case 0:
		items := warehouseGetItems(s, pkt.BoxIndex)
		bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	case 1:
		equipment := warehouseGetEquipment(s, pkt.BoxIndex)
		bf.WriteBytes(mhfitem.SerializeWarehouseEquipment(equipment, s.server.erupeConfig.RealClientMode))
	}
	if bf.Index() > 0 {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfUpdateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateWarehouse)
	if pkt.BoxIndex > 10 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	saveStart := time.Now()

	var err error
	var boxTypeName string
	var dataSize int

	switch pkt.BoxType {
	case 0:
		boxTypeName = "items"
		newStacks := mhfitem.DiffItemStacks(warehouseGetItems(s, pkt.BoxIndex), pkt.UpdatedItems)
		serialized := mhfitem.SerializeWarehouseItems(newStacks)
		dataSize = len(serialized)

		s.logger.Debug("Warehouse save request",
			zap.Uint32("charID", s.charID),
			zap.String("box_type", boxTypeName),
			zap.Uint8("box_index", pkt.BoxIndex),
			zap.Int("item_count", len(pkt.UpdatedItems)),
			zap.Int("data_size", dataSize),
		)

		_, err = s.server.db.Exec(fmt.Sprintf(`UPDATE warehouse SET item%d=$1 WHERE character_id=$2`, pkt.BoxIndex), serialized, s.charID)
		if err != nil {
			s.logger.Error("Failed to update warehouse items",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
				zap.Uint8("box_index", pkt.BoxIndex),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	case 1:
		boxTypeName = "equipment"
		var fEquip []mhfitem.MHFEquipment
		oEquips := warehouseGetEquipment(s, pkt.BoxIndex)
		for _, uEquip := range pkt.UpdatedEquipment {
			exists := false
			for i := range oEquips {
				if oEquips[i].WarehouseID == uEquip.WarehouseID {
					exists = true
					// Will set removed items to 0
					oEquips[i].ItemID = uEquip.ItemID
					break
				}
			}
			if !exists {
				uEquip.WarehouseID = token.RNG.Uint32()
				fEquip = append(fEquip, uEquip)
			}
		}
		for _, oEquip := range oEquips {
			if oEquip.ItemID > 0 {
				fEquip = append(fEquip, oEquip)
			}
		}

		serialized := mhfitem.SerializeWarehouseEquipment(fEquip, s.server.erupeConfig.RealClientMode)
		dataSize = len(serialized)

		s.logger.Debug("Warehouse save request",
			zap.Uint32("charID", s.charID),
			zap.String("box_type", boxTypeName),
			zap.Uint8("box_index", pkt.BoxIndex),
			zap.Int("equip_count", len(pkt.UpdatedEquipment)),
			zap.Int("data_size", dataSize),
		)

		_, err = s.server.db.Exec(fmt.Sprintf(`UPDATE warehouse SET equip%d=$1 WHERE character_id=$2`, pkt.BoxIndex), serialized, s.charID)
		if err != nil {
			s.logger.Error("Failed to update warehouse equipment",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
				zap.Uint8("box_index", pkt.BoxIndex),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	saveDuration := time.Since(saveStart)
	s.logger.Info("Warehouse saved successfully",
		zap.Uint32("charID", s.charID),
		zap.String("box_type", boxTypeName),
		zap.Uint8("box_index", pkt.BoxIndex),
		zap.Int("data_size", dataSize),
		zap.Duration("duration", saveDuration),
	)

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
