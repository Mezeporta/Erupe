package channelserver

import (
	"testing"

	"erupe-ce/network"
)

func TestHandlerTableInitialized(t *testing.T) {
	if handlerTable == nil {
		t.Fatal("handlerTable should be initialized by init()")
	}
}

func TestHandlerTableHasEntries(t *testing.T) {
	if len(handlerTable) == 0 {
		t.Error("handlerTable should have entries")
	}

	// Should have many handlers
	if len(handlerTable) < 100 {
		t.Errorf("handlerTable has %d entries, expected 100+", len(handlerTable))
	}
}

func TestHandlerTableSystemPackets(t *testing.T) {
	// Test that key system packets have handlers
	systemPackets := []network.PacketID{
		network.MSG_HEAD,
		network.MSG_SYS_END,
		network.MSG_SYS_NOP,
		network.MSG_SYS_ACK,
		network.MSG_SYS_LOGIN,
		network.MSG_SYS_LOGOUT,
		network.MSG_SYS_PING,
		network.MSG_SYS_TIME,
	}

	for _, opcode := range systemPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for %s", opcode)
			}
		})
	}
}

func TestHandlerTableStagePackets(t *testing.T) {
	// Test stage-related packet handlers
	stagePackets := []network.PacketID{
		network.MSG_SYS_CREATE_STAGE,
		network.MSG_SYS_STAGE_DESTRUCT,
		network.MSG_SYS_ENTER_STAGE,
		network.MSG_SYS_BACK_STAGE,
		network.MSG_SYS_MOVE_STAGE,
		network.MSG_SYS_LEAVE_STAGE,
		network.MSG_SYS_LOCK_STAGE,
		network.MSG_SYS_UNLOCK_STAGE,
	}

	for _, opcode := range stagePackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for stage packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableBinaryPackets(t *testing.T) {
	// Test binary message handlers
	binaryPackets := []network.PacketID{
		network.MSG_SYS_CAST_BINARY,
		network.MSG_SYS_CASTED_BINARY,
		network.MSG_SYS_SET_STAGE_BINARY,
		network.MSG_SYS_GET_STAGE_BINARY,
	}

	for _, opcode := range binaryPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for binary packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableReservedPackets(t *testing.T) {
	// Reserved packets should still have handlers (usually no-ops)
	reservedPackets := []network.PacketID{
		network.MSG_SYS_reserve01,
		network.MSG_SYS_reserve02,
		network.MSG_SYS_reserve03,
		network.MSG_SYS_reserve04,
		network.MSG_SYS_reserve05,
		network.MSG_SYS_reserve06,
		network.MSG_SYS_reserve07,
	}

	for _, opcode := range reservedPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for reserved packet %s", opcode)
			}
		})
	}
}

func TestHandlerFuncType(t *testing.T) {
	// Verify all handlers are valid functions
	for opcode, handler := range handlerTable {
		if handler == nil {
			t.Errorf("handler for %s is nil", opcode)
		}
	}
}

func TestHandlerTableObjectPackets(t *testing.T) {
	objectPackets := []network.PacketID{
		network.MSG_SYS_ADD_OBJECT,
		network.MSG_SYS_DEL_OBJECT,
		network.MSG_SYS_DISP_OBJECT,
		network.MSG_SYS_HIDE_OBJECT,
	}

	for _, opcode := range objectPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for object packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableClientPackets(t *testing.T) {
	clientPackets := []network.PacketID{
		network.MSG_SYS_SET_STATUS,
		network.MSG_SYS_HIDE_CLIENT,
		network.MSG_SYS_ENUMERATE_CLIENT,
	}

	for _, opcode := range clientPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for client packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableSemaphorePackets(t *testing.T) {
	semaphorePackets := []network.PacketID{
		network.MSG_SYS_CREATE_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_RELEASE_SEMAPHORE,
	}

	for _, opcode := range semaphorePackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for semaphore packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableMHFPackets(t *testing.T) {
	// Test some core MHF packets have handlers
	mhfPackets := []network.PacketID{
		network.MSG_MHF_SAVEDATA,
		network.MSG_MHF_LOADDATA,
	}

	for _, opcode := range mhfPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for MHF packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableEnumeratePackets(t *testing.T) {
	enumPackets := []network.PacketID{
		network.MSG_SYS_ENUMERATE_CLIENT,
		network.MSG_SYS_ENUMERATE_STAGE,
	}

	for _, opcode := range enumPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for enumerate packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableLogPackets(t *testing.T) {
	logPackets := []network.PacketID{
		network.MSG_SYS_TERMINAL_LOG,
		network.MSG_SYS_ISSUE_LOGKEY,
		network.MSG_SYS_RECORD_LOG,
	}

	for _, opcode := range logPackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for log packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableFilePackets(t *testing.T) {
	filePackets := []network.PacketID{
		network.MSG_SYS_GET_FILE,
	}

	for _, opcode := range filePackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for file packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableEchoPacket(t *testing.T) {
	if _, ok := handlerTable[network.MSG_SYS_ECHO]; !ok {
		t.Error("handler missing for MSG_SYS_ECHO")
	}
}

func TestHandlerTableReserveStagePackets(t *testing.T) {
	reservePackets := []network.PacketID{
		network.MSG_SYS_RESERVE_STAGE,
		network.MSG_SYS_UNRESERVE_STAGE,
		network.MSG_SYS_SET_STAGE_PASS,
		network.MSG_SYS_WAIT_STAGE_BINARY,
	}

	for _, opcode := range reservePackets {
		t.Run(opcode.String(), func(t *testing.T) {
			if _, ok := handlerTable[opcode]; !ok {
				t.Errorf("handler missing for reserve stage packet %s", opcode)
			}
		})
	}
}

func TestHandlerTableThresholdPacket(t *testing.T) {
	if _, ok := handlerTable[network.MSG_SYS_EXTEND_THRESHOLD]; !ok {
		t.Error("handler missing for MSG_SYS_EXTEND_THRESHOLD")
	}
}

func TestHandlerTableNoNilValues(t *testing.T) {
	nilCount := 0
	for opcode, handler := range handlerTable {
		if handler == nil {
			nilCount++
			t.Errorf("nil handler for opcode %s", opcode)
		}
	}
	if nilCount > 0 {
		t.Errorf("found %d nil handlers in handlerTable", nilCount)
	}
}
