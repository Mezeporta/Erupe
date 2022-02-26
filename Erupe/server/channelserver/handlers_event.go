package channelserver

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network/mhfpacket"
	timeServerFix "github.com/Solenataris/Erupe/server/channelserver/timeserver"
)

func handleMsgMhfRegisterEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegisterEvent)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Unk2)
	bf.WriteUint8(pkt.Unk4)
	bf.WriteUint16(0x1142)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReleaseEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReleaseEvent)

	// Do this ack manually because it uses a non-(0|1) error code
	/*
		_ACK_SUCCESS = 0
		_ACK_ERROR = 1

		_ACK_EINPROGRESS = 16
		_ACK_ENOENT = 17
		_ACK_ENOSPC = 18
		_ACK_ETIMEOUT = 19

		_ACK_EINVALID = 64
		_ACK_EFAILED = 65
		_ACK_ENOMEM = 66
		_ACK_ENOTEXIT = 67
		_ACK_ENOTREADY = 68
		_ACK_EALREADY = 69
		_ACK_DISABLE_WORK = 71
	*/
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        pkt.AckHandle,
		IsBufferResponse: false,
		ErrorCode:        0x41,
		AckData:          []byte{0x00, 0x00, 0x00, 0x00},
	})
}

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

var persistentEventSchedule []activeFeature

type activeFeature struct {
	StartTime      time.Time
	ActiveFeatures uint32
	Unk1           uint16
}

func handleMsgMhfGetWeeklySchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySchedule)
	// ActiveFeatures is a bit field, 0x3FFF is all 14 active features.
	// Long term it should probably be made persistent and simply cycle a couple daily
	// Times seem to need to be timeServerFix.midnight which is likely why matching timezone was required originally
	if len(persistentEventSchedule) == 0 {
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.OpcodeMessages {
			s.logger.Info("\nGenerating active feature...")
		}
		persistentEventSchedule = make([]activeFeature, 8)
		//weapons := generateRandomNumber(1, 14, 8)
		for x := -1; x < 7; x++ {
			var feat uint32
			feat |= 65535
			persistentEventSchedule[x+1] = activeFeature{
				StartTime:      Time_Current_Midnight().Add(time.Duration(24*x) * time.Hour),
				ActiveFeatures: feat,
				Unk1:           0,
			}
			fmt.Println(feat)
		}
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(persistentEventSchedule)))                           // Entry count, client only parses the first 7 or 8.
	resp.WriteUint32(uint32(Time_Current_Adjusted().Add(-5 * time.Minute).Unix())) // 5 minutes ago server time

	for _, es := range persistentEventSchedule {
		resp.WriteUint32(uint32(es.StartTime.Unix()))
		resp.WriteUint32(es.ActiveFeatures)
		resp.WriteUint16(es.Unk1)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func generateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}
	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

type loginBoost struct {
	WeekReq, WeekCount uint8
	Available          bool
	Expiration         uint32
}

func handleMsgMhfGetKeepLoginBoostStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKeepLoginBoostStatus)

	var loginBoostStatus []loginBoost
	insert := false
	boostState, err := s.server.db.Query("SELECT week_req, week_count, available, end_time FROM login_boost_state WHERE char_id=$1 ORDER BY week_req ASC", s.charID)
	if err != nil {
		panic(err)
	}
	for boostState.Next() {
		var boost loginBoost
		err = boostState.Scan(&boost.WeekReq, &boost.WeekCount, &boost.Available, &boost.Expiration)
		if err != nil {
			panic(err)
		}
		loginBoostStatus = append(loginBoostStatus, boost)
	}
	if len(loginBoostStatus) == 0 {
		// create default Entries (should only been week 1 with )
		insert = true
		loginBoostStatus = []loginBoost{
			{
				WeekReq:    1,    // weeks needed
				WeekCount:  0,    // weeks passed
				Available:  true, // available
				Expiration: 0,    //uint32(t.Add(120 * time.Minute).Unix()), // uncomment to enable permanently
			},
			{
				WeekReq:    2,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    3,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    4,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    5,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
		}
	}
	resp := byteframe.NewByteFrame()
	CurrentWeek := Time_Current_Week_uint8()
	for d := range loginBoostStatus {
		if CurrentWeek == 1 && loginBoostStatus[d].WeekCount <= 5 {
			loginBoostStatus[d].WeekCount = 0
		}
		if loginBoostStatus[d].WeekReq == CurrentWeek || loginBoostStatus[d].WeekCount != 0 {
			loginBoostStatus[d].WeekCount = CurrentWeek
		}
		if !loginBoostStatus[d].Available && loginBoostStatus[d].WeekCount >= loginBoostStatus[d].WeekReq && uint32(time.Now().In(time.FixedZone("UTC+1", 1*60*60)).Unix()) >= loginBoostStatus[d].Expiration {
			loginBoostStatus[d].Expiration = 1
		}
		if !insert {
			_, err := s.server.db.Exec(`UPDATE login_boost_state SET week_count=$1, end_time=$2 WHERE char_id=$3 AND week_req=$4`, loginBoostStatus[d].WeekCount, loginBoostStatus[d].Expiration, s.charID, loginBoostStatus[d].WeekReq)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, v := range loginBoostStatus {
		if insert {
			_, err := s.server.db.Exec(`INSERT INTO login_boost_state (char_id, week_req, week_count, available, end_time) VALUES ($1,$2,$3,$4,$5)`, s.charID, v.WeekReq, v.WeekCount, v.Available, v.Expiration)
			if err != nil {
				panic(err)
			}
		}
		resp.WriteUint8(v.WeekReq)
		resp.WriteUint8(v.WeekCount)
		resp.WriteBool(v.Available)
		resp.WriteUint32(v.Expiration)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseKeepLoginBoost(s *Session, p mhfpacket.MHFPacket) {
	// Directly interacts with MhfGetKeepLoginBoostStatus
	// TODO: make these states persistent on a per character basis
	pkt := p.(*mhfpacket.MsgMhfUseKeepLoginBoost)
	var t = time.Now().In(time.FixedZone("UTC+1", 1*60*60))
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(0)

	// response is end timestamp based on input
	switch pkt.BoostWeekUsed {
	case 1:
		t = t.Add(120 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 2:
		t = t.Add(240 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 3:
		t = t.Add(120 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 4:
		t = t.Add(180 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 5:
		t = t.Add(240 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	}
	_, err := s.server.db.Exec(`UPDATE login_boost_state SET available='false', end_time=$1 WHERE char_id=$2 AND week_req=$3`, uint32(t.Unix()), s.charID, pkt.BoostWeekUsed)
	if err != nil {
		panic(err)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdSchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSchedule)
	var t = timeServerFix.Tstatic_midnight()
	var event int = s.server.erupeConfig.DevModeOptions.Event

	year, month, day := t.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	// Events with time limits are Festival with Sign up, Soul Week and Winners Weeks
	// Diva Defense with Prayer, Interception and Song weeks
	// Mezeporta Festival with simply 'available' being a weekend thing
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0x1d5fda5c) // Unk (1d5fda5c, 0b5397df)

	if event == 1 {
		resp.WriteUint32(uint32(midnight.Add(24 * 21 * time.Hour).Unix())) // Week 1 Timestamp, Festi start?
	} else {
		resp.WriteUint32(uint32(midnight.Add(-24 * 21 * time.Hour).Unix())) // Week 1 Timestamp, Festi start?
	}

	if event == 2 {
		resp.WriteUint32(uint32(midnight.Add(24 * 14 * time.Hour).Unix())) // Week 2 Timestamp
		resp.WriteUint32(uint32(midnight.Add(24 * 14 * time.Hour).Unix())) // Week 2 Timestamp
	} else {
		resp.WriteUint32(uint32(midnight.Add(-24 * 14 * time.Hour).Unix())) // Week 2 Timestamp
		resp.WriteUint32(uint32(midnight.Add(-24 * 14 * time.Hour).Unix())) // Week 2 Timestamp
	}

	if event == 3 {
		resp.WriteUint32(uint32(midnight.Add((24) * 7 * time.Hour).Unix()))  // Diva Defense Interception
		resp.WriteUint32(uint32(midnight.Add((24) * 14 * time.Hour).Unix())) // Diva Defense Greeting Song
	} else {
		resp.WriteUint32(uint32(midnight.Add((-24) * 7 * time.Hour).Unix()))  // Diva Defense Interception
		resp.WriteUint32(uint32(midnight.Add((-24) * 14 * time.Hour).Unix())) // Diva Defense Greeting Song
	}

	resp.WriteUint16(0x19) // Unk 00011001
	resp.WriteUint16(0x2d) // Unk 00101101
	resp.WriteUint16(0x02) // Unk 00000010
	resp.WriteUint16(0x02) // Unk 00000010

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

/*
func handleMsgMhfGetUdSchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSchedule)
	resp := byteframe.NewByteFrame()

	resp.WriteUint32(0x1d5fda5c)                  // Unk (1d5fda5c, 0b5397df)
	resp.WriteUint32(uint32(ScheduleEvent(s, 1))) // Week 1 Timestamp, Festi start?
	resp.WriteUint32(uint32(ScheduleEvent(s, 2))) // Diva Defense Interception 1
	resp.WriteUint32(uint32(ScheduleEvent(s, 3))) // Week 2 Timestamp
	resp.WriteUint32(uint32(ScheduleEvent(s, 4))) // Diva Defense Interception 2
	resp.WriteUint32(uint32(ScheduleEvent(s, 5))) // Week 3 Timestamp
	resp.WriteUint32(uint32(ScheduleEvent(s, 6))) // Diva Defense Greeting Song 3
	resp.WriteUint16(0x19)                        // Unk 00011001
	resp.WriteUint16(0x2d)                        // Unk 00101101
	resp.WriteUint16(0x02)                        // Unk 00000010
	resp.WriteUint16(0x02)                        // Unk 00000010

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

var timedb int64
var countEvent int = 1
var BlockSchedulEvent bool
var t_Next_SchedulEvent = Time_Current_Adjusted()
var t_curr_SchedulEvent = Time_Current_Adjusted().Unix()

func ScheduleEvent(s *Session, fixWeek int) uint32 {
	if !BlockSchedulEvent {
		if s.server.erupeConfig.DevModeOptions.ServerName != "" { // IF (SERVERNAME == NAME)
			err := s.server.db.QueryRow("SELECT event_id FROM servers WHERE server_name=$1", s.server.erupeConfig.DevModeOptions.ServerName).Scan(&countEvent)
			if err != nil {
				panic(err)
			}
			s.server.db.QueryRow("SELECT date_expiration FROM servers server_name=$1", s.server.erupeConfig.DevModeOptions.ServerName).Scan(&timedb)
			if t_curr_SchedulEvent >= timedb {
				countEvent += 1
				if countEvent == 7 {
					countEvent = 1
				}
				var t_Add_Next_SchedulEvent = t_Next_SchedulEvent.Add(7 * 24 * time.Hour).Unix()
				_, err := s.server.db.Exec("UPDATE servers SET event_id=$1, event_expiration=$2 WHERE server_name=$3", countEvent, t_Add_Next_SchedulEvent, s.server.erupeConfig.DevModeOptions.ServerName)
				if err == nil {
					s.server.db.QueryRow("SELECT event_id FROM servers WHERE id=$1").Scan(&countEvent)
				}
			}
			BlockSchedulEvent = fixWeek == countEvent
		} else { // ELSE (SERVERNAME == NULL)
			err := s.server.db.QueryRow("SELECT event_id FROM event_week WHERE id=1").Scan(&countEvent)
			if err != nil {
				var t_Add_Next_SchedulEvent = t_Next_SchedulEvent.Add(7 * 24 * time.Hour).Unix()
				s.server.db.Exec("INSERT INTO event_week (id, event_id, date_expiration) VALUES (1, $1, $2)", countEvent, t_Add_Next_SchedulEvent)
				s.server.db.QueryRow("SELECT event_id FROM event_week WHERE id=1").Scan(&countEvent)
			}
			s.server.db.QueryRow("SELECT date_expiration FROM event_week WHERE id=1").Scan(&timedb)
			if t_curr_SchedulEvent >= timedb {
				countEvent += 1
				if countEvent == 7 {
					countEvent = 1
				}
				var t_Add_Next_SchedulEvent = t_Next_SchedulEvent.Add(7 * 24 * time.Hour).Unix()
				_, err := s.server.db.Exec("UPDATE event_week SET event_id=$1, date_expiration=$2 WHERE id=1", countEvent, t_Add_Next_SchedulEvent)
				if err == nil {
					s.server.db.QueryRow("SELECT event_id FROM event_week WHERE id=1").Scan(&countEvent)
				}
			}
			BlockSchedulEvent = fixWeek == countEvent
		}
	}
	if fixWeek == countEvent {
		return uint32(Time_Current_Midnight().Add(7 * 24 * time.Hour).Unix())
	} else {
		return uint32(Time_Current_Midnight().Add(-24 * 21 * time.Hour).Unix())
	}
}
*/

func handleMsgMhfGetUdInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdInfo)
	// Message that appears on the Diva Defense NPC and triggers the green exclamation mark
	udInfos := []struct {
		Text      string
		StartTime time.Time
		EndTime   time.Time
	}{
		/*{
			Text:      " ~C17【Erupe】 is dead event!\n\n■Features\n~C18 Dont bother walking around!\n~C17 Take down your DB by doing \n~C17 nearly anything!",
			StartTime: Time_static().Add(time.Duration(-5) * time.Minute), // Event started 5 minutes ago,
			EndTime:   Time_static().Add(time.Duration(24) * time.Hour),   // Event ends in 5 minutes,
		}, */
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udInfos)))
	for _, udInfo := range udInfos {
		resp.WriteBytes(fixedSizeShiftJIS(udInfo.Text, 1024))
		resp.WriteUint32(uint32(udInfo.StartTime.Unix()))
		resp.WriteUint32(uint32(udInfo.EndTime.Unix()))
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTime)

	doAckBufSucceed(s, pkt.AckHandle, []byte{})
	updateRights(s)
}

func handleMsgMhfGetBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTimeLimit)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetBoostRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostRight)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfPostBoostTimeQuestReturn(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeQuestReturn)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfStartBoostTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostBoostTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}
