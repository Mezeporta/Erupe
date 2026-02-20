package channelserver

// Raviente register type IDs (used in MsgSysLoadRegister / MsgSysNotifyRegister)
const (
	raviRegisterState   = uint32(0x40000)
	raviRegisterSupport = uint32(0x50000)
	raviRegisterGeneral = uint32(0x60000)
)

// Raviente semaphore constants
const (
	raviSemaphoreStride = 0x10000    // ID spacing between hs_l0* semaphores
	raviSemaphoreMax    = uint16(127) // max players per Raviente semaphore
)
