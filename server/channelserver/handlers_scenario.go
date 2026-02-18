package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// Scenario represents scenario counter data.
type Scenario struct {
	MainID uint32
	// 0 = Basic
	// 1 = Veteran
	// 3 = Other
	// 6 = Pallone
	// 7 = Diva
	CategoryID uint8
}

func handleMsgMhfInfoScenarioCounter(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoScenarioCounter)
	var scenarios []Scenario
	var scenario Scenario
	scenarioData, err := s.server.db.Queryx("SELECT scenario_id, category_id FROM scenario_counter")
	if err != nil {
		_ = scenarioData.Close()
		s.logger.Error("Failed to get scenario counter info from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	for scenarioData.Next() {
		err = scenarioData.Scan(&scenario.MainID, &scenario.CategoryID)
		if err != nil {
			continue
		}
		scenarios = append(scenarios, scenario)
	}

	// Trim excess scenarios
	if len(scenarios) > 128 {
		scenarios = scenarios[:128]
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(scenarios)))
	for _, scenario := range scenarios {
		bf.WriteUint32(scenario.MainID)
		// If item exchange
		switch scenario.CategoryID {
		case 3, 6, 7:
			bf.WriteBool(true)
		default:
			bf.WriteBool(false)
		}
		bf.WriteUint8(scenario.CategoryID)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
