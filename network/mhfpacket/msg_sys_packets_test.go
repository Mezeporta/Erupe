package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// TestMsgSysCastBinaryParse tests parsing MsgSysCastBinary
func TestMsgSysCastBinaryParse(t *testing.T) {
	tests := []struct {
		name          string
		unk           uint32
		broadcastType uint8
		messageType   uint8
		payload       []byte
	}{
		{"empty payload", 0, 1, 2, []byte{}},
		{"small payload", 0x006400C8, 3, 4, []byte{0xAA, 0xBB, 0xCC}},
		{"large payload", 0xFFFFFFFF, 0xFF, 0xFF, make([]byte, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.unk)
			bf.WriteUint8(tt.broadcastType)
			bf.WriteUint8(tt.messageType)
			bf.WriteUint16(uint16(len(tt.payload)))
			bf.WriteBytes(tt.payload)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgSysCastBinary{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.Unk != tt.unk {
				t.Errorf("Unk = %d, want %d", pkt.Unk, tt.unk)
			}
			if pkt.BroadcastType != tt.broadcastType {
				t.Errorf("BroadcastType = %d, want %d", pkt.BroadcastType, tt.broadcastType)
			}
			if pkt.MessageType != tt.messageType {
				t.Errorf("MessageType = %d, want %d", pkt.MessageType, tt.messageType)
			}
			if len(pkt.RawDataPayload) != len(tt.payload) {
				t.Errorf("RawDataPayload len = %d, want %d", len(pkt.RawDataPayload), len(tt.payload))
			}
		})
	}
}

// TestMsgSysCastBinaryOpcode tests Opcode method
func TestMsgSysCastBinaryOpcode(t *testing.T) {
	pkt := &MsgSysCastBinary{}
	if pkt.Opcode() != network.MSG_SYS_CAST_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_CAST_BINARY", pkt.Opcode())
	}
}

// TestMsgSysCreateSemaphoreOpcode tests Opcode method
func TestMsgSysCreateSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysCreateSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_CREATE_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_CREATE_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysCastedBinaryOpcode tests Opcode method
func TestMsgSysCastedBinaryOpcode(t *testing.T) {
	pkt := &MsgSysCastedBinary{}
	if pkt.Opcode() != network.MSG_SYS_CASTED_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_CASTED_BINARY", pkt.Opcode())
	}
}

// TestMsgSysSetStageBinaryOpcode tests Opcode method
func TestMsgSysSetStageBinaryOpcode(t *testing.T) {
	pkt := &MsgSysSetStageBinary{}
	if pkt.Opcode() != network.MSG_SYS_SET_STAGE_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_SET_STAGE_BINARY", pkt.Opcode())
	}
}

// TestMsgSysGetStageBinaryOpcode tests Opcode method
func TestMsgSysGetStageBinaryOpcode(t *testing.T) {
	pkt := &MsgSysGetStageBinary{}
	if pkt.Opcode() != network.MSG_SYS_GET_STAGE_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_STAGE_BINARY", pkt.Opcode())
	}
}

// TestMsgSysWaitStageBinaryOpcode tests Opcode method
func TestMsgSysWaitStageBinaryOpcode(t *testing.T) {
	pkt := &MsgSysWaitStageBinary{}
	if pkt.Opcode() != network.MSG_SYS_WAIT_STAGE_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_WAIT_STAGE_BINARY", pkt.Opcode())
	}
}

// TestMsgSysEnumerateClientOpcode tests Opcode method
func TestMsgSysEnumerateClientOpcode(t *testing.T) {
	pkt := &MsgSysEnumerateClient{}
	if pkt.Opcode() != network.MSG_SYS_ENUMERATE_CLIENT {
		t.Errorf("Opcode() = %s, want MSG_SYS_ENUMERATE_CLIENT", pkt.Opcode())
	}
}

// TestMsgSysEnumerateStageOpcode tests Opcode method
func TestMsgSysEnumerateStageOpcode(t *testing.T) {
	pkt := &MsgSysEnumerateStage{}
	if pkt.Opcode() != network.MSG_SYS_ENUMERATE_STAGE {
		t.Errorf("Opcode() = %s, want MSG_SYS_ENUMERATE_STAGE", pkt.Opcode())
	}
}

// TestMsgSysCreateMutexOpcode tests Opcode method
func TestMsgSysCreateMutexOpcode(t *testing.T) {
	pkt := &MsgSysCreateMutex{}
	if pkt.Opcode() != network.MSG_SYS_CREATE_MUTEX {
		t.Errorf("Opcode() = %s, want MSG_SYS_CREATE_MUTEX", pkt.Opcode())
	}
}

// TestMsgSysCreateOpenMutexOpcode tests Opcode method
func TestMsgSysCreateOpenMutexOpcode(t *testing.T) {
	pkt := &MsgSysCreateOpenMutex{}
	if pkt.Opcode() != network.MSG_SYS_CREATE_OPEN_MUTEX {
		t.Errorf("Opcode() = %s, want MSG_SYS_CREATE_OPEN_MUTEX", pkt.Opcode())
	}
}

// TestMsgSysDeleteMutexOpcode tests Opcode method
func TestMsgSysDeleteMutexOpcode(t *testing.T) {
	pkt := &MsgSysDeleteMutex{}
	if pkt.Opcode() != network.MSG_SYS_DELETE_MUTEX {
		t.Errorf("Opcode() = %s, want MSG_SYS_DELETE_MUTEX", pkt.Opcode())
	}
}

// TestMsgSysOpenMutexOpcode tests Opcode method
func TestMsgSysOpenMutexOpcode(t *testing.T) {
	pkt := &MsgSysOpenMutex{}
	if pkt.Opcode() != network.MSG_SYS_OPEN_MUTEX {
		t.Errorf("Opcode() = %s, want MSG_SYS_OPEN_MUTEX", pkt.Opcode())
	}
}

// TestMsgSysCloseMutexOpcode tests Opcode method
func TestMsgSysCloseMutexOpcode(t *testing.T) {
	pkt := &MsgSysCloseMutex{}
	if pkt.Opcode() != network.MSG_SYS_CLOSE_MUTEX {
		t.Errorf("Opcode() = %s, want MSG_SYS_CLOSE_MUTEX", pkt.Opcode())
	}
}

// TestMsgSysDeleteSemaphoreOpcode tests Opcode method
func TestMsgSysDeleteSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysDeleteSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_DELETE_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_DELETE_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysAcquireSemaphoreOpcode tests Opcode method
func TestMsgSysAcquireSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysAcquireSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_ACQUIRE_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_ACQUIRE_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysReleaseSemaphoreOpcode tests Opcode method
func TestMsgSysReleaseSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysReleaseSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_RELEASE_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_RELEASE_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysCheckSemaphoreOpcode tests Opcode method
func TestMsgSysCheckSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysCheckSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_CHECK_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_CHECK_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysCreateAcquireSemaphoreOpcode tests Opcode method
func TestMsgSysCreateAcquireSemaphoreOpcode(t *testing.T) {
	pkt := &MsgSysCreateAcquireSemaphore{}
	if pkt.Opcode() != network.MSG_SYS_CREATE_ACQUIRE_SEMAPHORE {
		t.Errorf("Opcode() = %s, want MSG_SYS_CREATE_ACQUIRE_SEMAPHORE", pkt.Opcode())
	}
}

// TestMsgSysOperateRegisterOpcode tests Opcode method
func TestMsgSysOperateRegisterOpcode(t *testing.T) {
	pkt := &MsgSysOperateRegister{}
	if pkt.Opcode() != network.MSG_SYS_OPERATE_REGISTER {
		t.Errorf("Opcode() = %s, want MSG_SYS_OPERATE_REGISTER", pkt.Opcode())
	}
}

// TestMsgSysLoadRegisterOpcode tests Opcode method
func TestMsgSysLoadRegisterOpcode(t *testing.T) {
	pkt := &MsgSysLoadRegister{}
	if pkt.Opcode() != network.MSG_SYS_LOAD_REGISTER {
		t.Errorf("Opcode() = %s, want MSG_SYS_LOAD_REGISTER", pkt.Opcode())
	}
}

// TestMsgSysNotifyRegisterOpcode tests Opcode method
func TestMsgSysNotifyRegisterOpcode(t *testing.T) {
	pkt := &MsgSysNotifyRegister{}
	if pkt.Opcode() != network.MSG_SYS_NOTIFY_REGISTER {
		t.Errorf("Opcode() = %s, want MSG_SYS_NOTIFY_REGISTER", pkt.Opcode())
	}
}

// TestMsgSysCreateObjectOpcode tests Opcode method
func TestMsgSysCreateObjectOpcode(t *testing.T) {
	pkt := &MsgSysCreateObject{}
	if pkt.Opcode() != network.MSG_SYS_CREATE_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_CREATE_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysDeleteObjectOpcode tests Opcode method
func TestMsgSysDeleteObjectOpcode(t *testing.T) {
	pkt := &MsgSysDeleteObject{}
	if pkt.Opcode() != network.MSG_SYS_DELETE_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_DELETE_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysPositionObjectOpcode tests Opcode method
func TestMsgSysPositionObjectOpcode(t *testing.T) {
	pkt := &MsgSysPositionObject{}
	if pkt.Opcode() != network.MSG_SYS_POSITION_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_POSITION_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysRotateObjectOpcode tests Opcode method
func TestMsgSysRotateObjectOpcode(t *testing.T) {
	pkt := &MsgSysRotateObject{}
	if pkt.Opcode() != network.MSG_SYS_ROTATE_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_ROTATE_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysDuplicateObjectOpcode tests Opcode method
func TestMsgSysDuplicateObjectOpcode(t *testing.T) {
	pkt := &MsgSysDuplicateObject{}
	if pkt.Opcode() != network.MSG_SYS_DUPLICATE_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_DUPLICATE_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysSetObjectBinaryOpcode tests Opcode method
func TestMsgSysSetObjectBinaryOpcode(t *testing.T) {
	pkt := &MsgSysSetObjectBinary{}
	if pkt.Opcode() != network.MSG_SYS_SET_OBJECT_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_SET_OBJECT_BINARY", pkt.Opcode())
	}
}

// TestMsgSysGetObjectBinaryOpcode tests Opcode method
func TestMsgSysGetObjectBinaryOpcode(t *testing.T) {
	pkt := &MsgSysGetObjectBinary{}
	if pkt.Opcode() != network.MSG_SYS_GET_OBJECT_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_OBJECT_BINARY", pkt.Opcode())
	}
}

// TestMsgSysGetObjectOwnerOpcode tests Opcode method
func TestMsgSysGetObjectOwnerOpcode(t *testing.T) {
	pkt := &MsgSysGetObjectOwner{}
	if pkt.Opcode() != network.MSG_SYS_GET_OBJECT_OWNER {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_OBJECT_OWNER", pkt.Opcode())
	}
}

// TestMsgSysUpdateObjectBinaryOpcode tests Opcode method
func TestMsgSysUpdateObjectBinaryOpcode(t *testing.T) {
	pkt := &MsgSysUpdateObjectBinary{}
	if pkt.Opcode() != network.MSG_SYS_UPDATE_OBJECT_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_UPDATE_OBJECT_BINARY", pkt.Opcode())
	}
}

// TestMsgSysCleanupObjectOpcode tests Opcode method
func TestMsgSysCleanupObjectOpcode(t *testing.T) {
	pkt := &MsgSysCleanupObject{}
	if pkt.Opcode() != network.MSG_SYS_CLEANUP_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_CLEANUP_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysInsertUserOpcode tests Opcode method
func TestMsgSysInsertUserOpcode(t *testing.T) {
	pkt := &MsgSysInsertUser{}
	if pkt.Opcode() != network.MSG_SYS_INSERT_USER {
		t.Errorf("Opcode() = %s, want MSG_SYS_INSERT_USER", pkt.Opcode())
	}
}

// TestMsgSysDeleteUserOpcode tests Opcode method
func TestMsgSysDeleteUserOpcode(t *testing.T) {
	pkt := &MsgSysDeleteUser{}
	if pkt.Opcode() != network.MSG_SYS_DELETE_USER {
		t.Errorf("Opcode() = %s, want MSG_SYS_DELETE_USER", pkt.Opcode())
	}
}

// TestMsgSysSetUserBinaryOpcode tests Opcode method
func TestMsgSysSetUserBinaryOpcode(t *testing.T) {
	pkt := &MsgSysSetUserBinary{}
	if pkt.Opcode() != network.MSG_SYS_SET_USER_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_SET_USER_BINARY", pkt.Opcode())
	}
}

// TestMsgSysGetUserBinaryOpcode tests Opcode method
func TestMsgSysGetUserBinaryOpcode(t *testing.T) {
	pkt := &MsgSysGetUserBinary{}
	if pkt.Opcode() != network.MSG_SYS_GET_USER_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_USER_BINARY", pkt.Opcode())
	}
}

// TestMsgSysNotifyUserBinaryOpcode tests Opcode method
func TestMsgSysNotifyUserBinaryOpcode(t *testing.T) {
	pkt := &MsgSysNotifyUserBinary{}
	if pkt.Opcode() != network.MSG_SYS_NOTIFY_USER_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_NOTIFY_USER_BINARY", pkt.Opcode())
	}
}

// TestMsgSysUpdateRightOpcode tests Opcode method
func TestMsgSysUpdateRightOpcode(t *testing.T) {
	pkt := &MsgSysUpdateRight{}
	if pkt.Opcode() != network.MSG_SYS_UPDATE_RIGHT {
		t.Errorf("Opcode() = %s, want MSG_SYS_UPDATE_RIGHT", pkt.Opcode())
	}
}

// TestMsgSysAuthQueryOpcode tests Opcode method
func TestMsgSysAuthQueryOpcode(t *testing.T) {
	pkt := &MsgSysAuthQuery{}
	if pkt.Opcode() != network.MSG_SYS_AUTH_QUERY {
		t.Errorf("Opcode() = %s, want MSG_SYS_AUTH_QUERY", pkt.Opcode())
	}
}

// TestMsgSysAuthDataOpcode tests Opcode method
func TestMsgSysAuthDataOpcode(t *testing.T) {
	pkt := &MsgSysAuthData{}
	if pkt.Opcode() != network.MSG_SYS_AUTH_DATA {
		t.Errorf("Opcode() = %s, want MSG_SYS_AUTH_DATA", pkt.Opcode())
	}
}

// TestMsgSysAuthTerminalOpcode tests Opcode method
func TestMsgSysAuthTerminalOpcode(t *testing.T) {
	pkt := &MsgSysAuthTerminal{}
	if pkt.Opcode() != network.MSG_SYS_AUTH_TERMINAL {
		t.Errorf("Opcode() = %s, want MSG_SYS_AUTH_TERMINAL", pkt.Opcode())
	}
}

// TestMsgSysRightsReloadOpcode tests Opcode method
func TestMsgSysRightsReloadOpcode(t *testing.T) {
	pkt := &MsgSysRightsReload{}
	if pkt.Opcode() != network.MSG_SYS_RIGHTS_RELOAD {
		t.Errorf("Opcode() = %s, want MSG_SYS_RIGHTS_RELOAD", pkt.Opcode())
	}
}

// TestMsgSysTerminalLogOpcode tests Opcode method
func TestMsgSysTerminalLogOpcode(t *testing.T) {
	pkt := &MsgSysTerminalLog{}
	if pkt.Opcode() != network.MSG_SYS_TERMINAL_LOG {
		t.Errorf("Opcode() = %s, want MSG_SYS_TERMINAL_LOG", pkt.Opcode())
	}
}

// TestMsgSysIssueLogkeyOpcode tests Opcode method
func TestMsgSysIssueLogkeyOpcode(t *testing.T) {
	pkt := &MsgSysIssueLogkey{}
	if pkt.Opcode() != network.MSG_SYS_ISSUE_LOGKEY {
		t.Errorf("Opcode() = %s, want MSG_SYS_ISSUE_LOGKEY", pkt.Opcode())
	}
}

// TestMsgSysRecordLogOpcode tests Opcode method
func TestMsgSysRecordLogOpcode(t *testing.T) {
	pkt := &MsgSysRecordLog{}
	if pkt.Opcode() != network.MSG_SYS_RECORD_LOG {
		t.Errorf("Opcode() = %s, want MSG_SYS_RECORD_LOG", pkt.Opcode())
	}
}

// TestMsgSysEchoOpcode tests Opcode method
func TestMsgSysEchoOpcode(t *testing.T) {
	pkt := &MsgSysEcho{}
	if pkt.Opcode() != network.MSG_SYS_ECHO {
		t.Errorf("Opcode() = %s, want MSG_SYS_ECHO", pkt.Opcode())
	}
}

// TestMsgSysGetFileOpcode tests Opcode method
func TestMsgSysGetFileOpcode(t *testing.T) {
	pkt := &MsgSysGetFile{}
	if pkt.Opcode() != network.MSG_SYS_GET_FILE {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_FILE", pkt.Opcode())
	}
}

// TestMsgSysHideClientOpcode tests Opcode method
func TestMsgSysHideClientOpcode(t *testing.T) {
	pkt := &MsgSysHideClient{}
	if pkt.Opcode() != network.MSG_SYS_HIDE_CLIENT {
		t.Errorf("Opcode() = %s, want MSG_SYS_HIDE_CLIENT", pkt.Opcode())
	}
}

// TestMsgSysSetStatusOpcode tests Opcode method
func TestMsgSysSetStatusOpcode(t *testing.T) {
	pkt := &MsgSysSetStatus{}
	if pkt.Opcode() != network.MSG_SYS_SET_STATUS {
		t.Errorf("Opcode() = %s, want MSG_SYS_SET_STATUS", pkt.Opcode())
	}
}

// TestMsgSysStageDestructOpcode tests Opcode method
func TestMsgSysStageDestructOpcode(t *testing.T) {
	pkt := &MsgSysStageDestruct{}
	if pkt.Opcode() != network.MSG_SYS_STAGE_DESTRUCT {
		t.Errorf("Opcode() = %s, want MSG_SYS_STAGE_DESTRUCT", pkt.Opcode())
	}
}

// TestMsgSysLeaveStageOpcode tests Opcode method
func TestMsgSysLeaveStageOpcode(t *testing.T) {
	pkt := &MsgSysLeaveStage{}
	if pkt.Opcode() != network.MSG_SYS_LEAVE_STAGE {
		t.Errorf("Opcode() = %s, want MSG_SYS_LEAVE_STAGE", pkt.Opcode())
	}
}

// TestMsgSysReserveStageOpcode tests Opcode method
func TestMsgSysReserveStageOpcode(t *testing.T) {
	pkt := &MsgSysReserveStage{}
	if pkt.Opcode() != network.MSG_SYS_RESERVE_STAGE {
		t.Errorf("Opcode() = %s, want MSG_SYS_RESERVE_STAGE", pkt.Opcode())
	}
}

// TestMsgSysUnreserveStageOpcode tests Opcode method
func TestMsgSysUnreserveStageOpcode(t *testing.T) {
	pkt := &MsgSysUnreserveStage{}
	if pkt.Opcode() != network.MSG_SYS_UNRESERVE_STAGE {
		t.Errorf("Opcode() = %s, want MSG_SYS_UNRESERVE_STAGE", pkt.Opcode())
	}
}

// TestMsgSysSetStagePassOpcode tests Opcode method
func TestMsgSysSetStagePassOpcode(t *testing.T) {
	pkt := &MsgSysSetStagePass{}
	if pkt.Opcode() != network.MSG_SYS_SET_STAGE_PASS {
		t.Errorf("Opcode() = %s, want MSG_SYS_SET_STAGE_PASS", pkt.Opcode())
	}
}

// TestMsgSysLockGlobalSemaOpcode tests Opcode method
func TestMsgSysLockGlobalSemaOpcode(t *testing.T) {
	pkt := &MsgSysLockGlobalSema{}
	if pkt.Opcode() != network.MSG_SYS_LOCK_GLOBAL_SEMA {
		t.Errorf("Opcode() = %s, want MSG_SYS_LOCK_GLOBAL_SEMA", pkt.Opcode())
	}
}

// TestMsgSysUnlockGlobalSemaOpcode tests Opcode method
func TestMsgSysUnlockGlobalSemaOpcode(t *testing.T) {
	pkt := &MsgSysUnlockGlobalSema{}
	if pkt.Opcode() != network.MSG_SYS_UNLOCK_GLOBAL_SEMA {
		t.Errorf("Opcode() = %s, want MSG_SYS_UNLOCK_GLOBAL_SEMA", pkt.Opcode())
	}
}

// TestMsgSysTransBinaryOpcode tests Opcode method
func TestMsgSysTransBinaryOpcode(t *testing.T) {
	pkt := &MsgSysTransBinary{}
	if pkt.Opcode() != network.MSG_SYS_TRANS_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_TRANS_BINARY", pkt.Opcode())
	}
}

// TestMsgSysCollectBinaryOpcode tests Opcode method
func TestMsgSysCollectBinaryOpcode(t *testing.T) {
	pkt := &MsgSysCollectBinary{}
	if pkt.Opcode() != network.MSG_SYS_COLLECT_BINARY {
		t.Errorf("Opcode() = %s, want MSG_SYS_COLLECT_BINARY", pkt.Opcode())
	}
}

// TestMsgSysGetStateOpcode tests Opcode method
func TestMsgSysGetStateOpcode(t *testing.T) {
	pkt := &MsgSysGetState{}
	if pkt.Opcode() != network.MSG_SYS_GET_STATE {
		t.Errorf("Opcode() = %s, want MSG_SYS_GET_STATE", pkt.Opcode())
	}
}

// TestMsgSysSerializeOpcode tests Opcode method
func TestMsgSysSerializeOpcode(t *testing.T) {
	pkt := &MsgSysSerialize{}
	if pkt.Opcode() != network.MSG_SYS_SERIALIZE {
		t.Errorf("Opcode() = %s, want MSG_SYS_SERIALIZE", pkt.Opcode())
	}
}

// TestMsgSysEnumlobbyOpcode tests Opcode method
func TestMsgSysEnumlobbyOpcode(t *testing.T) {
	pkt := &MsgSysEnumlobby{}
	if pkt.Opcode() != network.MSG_SYS_ENUMLOBBY {
		t.Errorf("Opcode() = %s, want MSG_SYS_ENUMLOBBY", pkt.Opcode())
	}
}

// TestMsgSysEnumuserOpcode tests Opcode method
func TestMsgSysEnumuserOpcode(t *testing.T) {
	pkt := &MsgSysEnumuser{}
	if pkt.Opcode() != network.MSG_SYS_ENUMUSER {
		t.Errorf("Opcode() = %s, want MSG_SYS_ENUMUSER", pkt.Opcode())
	}
}

// TestMsgSysInfokyserverOpcode tests Opcode method
func TestMsgSysInfokyserverOpcode(t *testing.T) {
	pkt := &MsgSysInfokyserver{}
	if pkt.Opcode() != network.MSG_SYS_INFOKYSERVER {
		t.Errorf("Opcode() = %s, want MSG_SYS_INFOKYSERVER", pkt.Opcode())
	}
}

// TestMsgSysExtendThresholdOpcode tests Opcode method
func TestMsgSysExtendThresholdOpcode(t *testing.T) {
	pkt := &MsgSysExtendThreshold{}
	if pkt.Opcode() != network.MSG_SYS_EXTEND_THRESHOLD {
		t.Errorf("Opcode() = %s, want MSG_SYS_EXTEND_THRESHOLD", pkt.Opcode())
	}
}

// TestMsgSysAddObjectOpcode tests Opcode method
func TestMsgSysAddObjectOpcode(t *testing.T) {
	pkt := &MsgSysAddObject{}
	if pkt.Opcode() != network.MSG_SYS_ADD_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_ADD_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysDelObjectOpcode tests Opcode method
func TestMsgSysDelObjectOpcode(t *testing.T) {
	pkt := &MsgSysDelObject{}
	if pkt.Opcode() != network.MSG_SYS_DEL_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_DEL_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysDispObjectOpcode tests Opcode method
func TestMsgSysDispObjectOpcode(t *testing.T) {
	pkt := &MsgSysDispObject{}
	if pkt.Opcode() != network.MSG_SYS_DISP_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_DISP_OBJECT", pkt.Opcode())
	}
}

// TestMsgSysHideObjectOpcode tests Opcode method
func TestMsgSysHideObjectOpcode(t *testing.T) {
	pkt := &MsgSysHideObject{}
	if pkt.Opcode() != network.MSG_SYS_HIDE_OBJECT {
		t.Errorf("Opcode() = %s, want MSG_SYS_HIDE_OBJECT", pkt.Opcode())
	}
}
