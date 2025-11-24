package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
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
	s.server.db.Exec(`UPDATE user_binary SET house_furniture=$1 WHERE id=$2`, pkt.InteriorData, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type HouseData struct {
	CharID        uint32 `db:"id"`
	HRP           uint16 `db:"hrp"`
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
	houseQuery := `SELECT c.id, hrp, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
		FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE c.id=$1`
	switch pkt.Method {
	case 1:
		var friendsList string
		s.server.db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.charID).Scan(&friendsList)
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
		houseQuery = `SELECT c.id, hrp, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
			FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE name ILIKE $1`
		house := HouseData{}
		rows, _ := s.server.db.Queryx(houseQuery, fmt.Sprintf(`%%%s%%`, pkt.Name))
		for rows.Next() {
			err := rows.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
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
		bf.WriteUint16(house.HRP)
		bf.WriteUint16(house.GR)
		ps.Uint8(bf, house.Name, true)
	}
	bf.Seek(0, 0)
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
	s.server.db.Exec(`UPDATE user_binary SET house_state=$1, house_password=$2 WHERE id=$3`, pkt.State, pkt.Password, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)
	bf := byteframe.NewByteFrame()

	var state uint8
	var password string
	s.server.db.QueryRow(`SELECT COALESCE(house_state, 2) as house_state, COALESCE(house_password, '') as house_password FROM user_binary WHERE id=$1
	`, pkt.CharID).Scan(&state, &password)

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
			var friendsList string
			s.server.db.QueryRow(`SELECT friends FROM characters WHERE id=$1`, pkt.CharID).Scan(&friendsList)
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
	s.server.db.QueryRow(`SELECT house_tier, house_data, house_furniture, bookshelf, gallery, tore, garden FROM user_binary WHERE id=$1
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
		c, d := getGookData(s, pkt.CharID)
		bf.WriteUint16(c)
		bf.WriteUint16(0)
		bf.WriteBytes(d)
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
	s.server.db.QueryRow(`SELECT mission FROM user_binary WHERE id=$1`, s.charID).Scan(&data)
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 9))
	}
}

func handleMsgMhfUpdateMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateMyhouseInfo)
	s.server.db.Exec("UPDATE user_binary SET mission=$1 WHERE id=$2", pkt.Unk0, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

// maxDecoMysets is the maximum number of decoration presets allowed per character.
// The client has a fixed buffer for 40 presets and will crash if more are sent.
const maxDecoMysets = 40

// decoMysetSize is the size in bytes of each decoration preset entry (2 byte index + 76 byte data).
const decoMysetSize = 78

func handleMsgMhfLoadDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadDecoMyset)
	var data []byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load decomyset", zap.Error(err))
	}
	if len(data) == 0 {
		data = []byte{0x01, 0x00}
	}
	// Validate and fix corrupted data with more than maxDecoMysets presets
	if len(data) >= 2 && data[1] > maxDecoMysets {
		s.logger.Warn("Decoration myset data exceeds limit, truncating",
			zap.Uint32("charID", s.charID),
			zap.Uint8("savedSets", data[1]),
			zap.Uint8("maxSets", maxDecoMysets))
		// Truncate to maxDecoMysets: 2 byte header + (maxDecoMysets * decoMysetSize)
		maxLen := 2 + (maxDecoMysets * decoMysetSize)
		if len(data) > maxLen {
			data = data[:maxLen]
		}
		data[1] = maxDecoMysets
		// Update the database with the fixed data
		_, err := s.server.db.Exec("UPDATE characters SET decomyset=$1 WHERE id=$2", data, s.charID)
		if err != nil {
			s.logger.Error("Failed to save truncated decomyset", zap.Error(err))
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSaveDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveDecoMyset)
	// https://gist.github.com/Andoryuuta/9c524da7285e4b5ca7e52e0fc1ca1daf
	var loadData []byte
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[1:]) // skip first unk byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&loadData)
	if err != nil {
		s.logger.Error("Failed to load decomyset", zap.Error(err))
	} else {
		numSets := bf.ReadUint8() // sets being written
		// empty save
		if len(loadData) == 0 {
			loadData = []byte{0x01, 0x00}
		}

		savedSets := loadData[1] // existing saved sets
		// no sets, new slice with just first 2 bytes for appends later
		if savedSets == 0 {
			loadData = []byte{0x01, 0x00}
		}
		for i := 0; i < int(numSets); i++ {
			writeSet := bf.ReadUint16()
			dataChunk := bf.ReadBytes(76)
			setBytes := append([]byte{uint8(writeSet >> 8), uint8(writeSet & 0xff)}, dataChunk...)
			for x := 0; true; x++ {
				if x == int(savedSets) {
					// appending set - check limit before adding
					if savedSets >= maxDecoMysets {
						s.logger.Warn("Decoration myset limit reached, ignoring new preset",
							zap.Uint32("charID", s.charID),
							zap.Uint8("currentSets", savedSets),
							zap.Uint8("maxSets", maxDecoMysets))
						break
					}
					if loadData[len(loadData)-1] == 0x10 {
						// sanity check for if there was a messy manual import
						loadData = append(loadData[:len(loadData)-2], setBytes...)
					} else {
						loadData = append(loadData, setBytes...)
					}
					savedSets++
					break
				}
				currentSet := loadData[3+(x*78)]
				if int(currentSet) == int(writeSet) {
					// replacing a set
					loadData = append(loadData[:2+(x*78)], append(setBytes, loadData[2+((x+1)*78):]...)...)
					break
				} else if int(currentSet) > int(writeSet) {
					// inserting before current set - check limit before adding
					if savedSets >= maxDecoMysets {
						s.logger.Warn("Decoration myset limit reached, ignoring new preset",
							zap.Uint32("charID", s.charID),
							zap.Uint8("currentSets", savedSets),
							zap.Uint8("maxSets", maxDecoMysets))
						break
					}
					loadData = append(loadData[:2+((x)*78)], append(setBytes, loadData[2+((x)*78):]...)...)
					savedSets++
					break
				}
			}
			loadData[1] = savedSets // update set count
		}
		dumpSaveData(s, loadData, "decomyset")
		_, err := s.server.db.Exec("UPDATE characters SET decomyset=$1 WHERE id=$2", loadData, s.charID)
		if err != nil {
			s.logger.Error("Failed to save decomyset", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

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
	bf.Seek(0, io.SeekStart)
	bf.WriteUint16(count)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTitle)
	var exists int
	err := s.server.db.QueryRow("SELECT count(*) FROM titles WHERE id=$1 AND char_id=$2", pkt.TitleID, s.charID).Scan(&exists)
	if err != nil || exists == 0 {
		s.server.db.Exec("INSERT INTO titles VALUES ($1, $2, now(), now())", pkt.TitleID, s.charID)
	} else {
		s.server.db.Exec("UPDATE titles SET updated_at=now()")
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfResetTitle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateWarehouse)
	var t int
	err := s.server.db.QueryRow("SELECT character_id FROM warehouse WHERE character_id=$1", s.charID).Scan(&t)
	if err != nil {
		s.server.db.Exec("INSERT INTO warehouse (character_id) VALUES ($1)", s.charID)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Operation)
	switch pkt.Operation {
	case 0:
		var count uint8
		itemNames := make([]string, 10)
		equipNames := make([]string, 10)
		s.server.db.QueryRow(fmt.Sprintf("%s WHERE character_id=$1", warehouseNamesQuery), s.charID).Scan(&itemNames[0],
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
		s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET %s%dname=$1 WHERE character_id=$2", pkt.BoxType, pkt.BoxIndex), pkt.Name, s.charID)
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

func addWarehouseGift(s *Session, boxType string, giftStack mhfpacket.WarehouseStack) {
	giftBox := getWarehouseBox(s, boxType, 10)
	if boxType == "item" {
		exists := false
		for i, stack := range giftBox {
			if stack.ItemID == giftStack.ItemID {
				exists = true
				giftBox[i].Quantity += giftStack.Quantity
				break
			}
		}
		if exists == false {
			giftBox = append(giftBox, giftStack)
		}
	} else {
		giftBox = append(giftBox, giftStack)
	}
	s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET %s10=$1 WHERE character_id=$2", boxType), boxToBytes(giftBox, boxType), s.charID)
}

func getWarehouseBox(s *Session, boxType string, boxIndex uint8) []mhfpacket.WarehouseStack {
	var data []byte
	s.server.db.QueryRow(fmt.Sprintf("SELECT %s%d FROM warehouse WHERE character_id=$1", boxType, boxIndex), s.charID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		stacks := make([]mhfpacket.WarehouseStack, numStacks)
		for i := 0; i < int(numStacks); i++ {
			if boxType == "item" {
				stacks[i].ID = box.ReadUint32()
				stacks[i].Index = box.ReadUint16()
				stacks[i].ItemID = box.ReadUint16()
				stacks[i].Quantity = box.ReadUint16()
				box.ReadUint16()
			} else {
				stacks[i].ID = box.ReadUint32()
				stacks[i].Index = box.ReadUint16()
				stacks[i].EquipType = box.ReadUint16()
				stacks[i].ItemID = box.ReadUint16()
				stacks[i].Data = box.ReadBytes(56)
			}
		}
		return stacks
	} else {
		return make([]mhfpacket.WarehouseStack, 0)
	}
}

func boxToBytes(stacks []mhfpacket.WarehouseStack, boxType string) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(stacks)))
	for i, stack := range stacks {
		if boxType == "item" {
			bf.WriteUint32(stack.ID)
			bf.WriteUint16(uint16(i + 1))
			bf.WriteUint16(stack.ItemID)
			bf.WriteUint16(stack.Quantity)
			bf.WriteUint16(0)
		} else {
			bf.WriteUint32(stack.ID)
			bf.WriteUint16(uint16(i + 1))
			bf.WriteUint16(stack.EquipType)
			bf.WriteUint16(stack.ItemID)
			bf.WriteBytes(stack.Data)
		}
	}
	bf.WriteUint16(0)
	return bf.Data()
}

func handleMsgMhfEnumerateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateWarehouse)
	box := getWarehouseBox(s, pkt.BoxType, pkt.BoxIndex)
	if len(box) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, boxToBytes(box, pkt.BoxType))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfUpdateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateWarehouse)
	box := getWarehouseBox(s, pkt.BoxType, pkt.BoxIndex)
	// Update existing stacks
	var newStacks []mhfpacket.WarehouseStack
	for _, update := range pkt.Updates {
		exists := false
		if pkt.BoxType == "item" {
			for i, stack := range box {
				if stack.Index == update.Index {
					exists = true
					box[i].Quantity = update.Quantity
					break
				}
			}
		} else {
			for i, stack := range box {
				if stack.Index == update.Index {
					exists = true
					box[i].ItemID = update.ItemID
					break
				}
			}
		}
		if exists == false {
			newStacks = append(newStacks, update)
		}
	}
	// Append new stacks
	for _, stack := range newStacks {
		box = append(box, stack)
	}
	// Slice empty stacks
	var cleanedBox []mhfpacket.WarehouseStack
	for _, stack := range box {
		if pkt.BoxType == "item" {
			if stack.Quantity > 0 {
				cleanedBox = append(cleanedBox, stack)
			}
		} else {
			if stack.ItemID != 0 {
				cleanedBox = append(cleanedBox, stack)
			}
		}
	}
	s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET %s%d=$1 WHERE character_id=$2", pkt.BoxType, pkt.BoxIndex), boxToBytes(cleanedBox, pkt.BoxType), s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
