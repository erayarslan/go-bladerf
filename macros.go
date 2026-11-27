package bladerf

// #include "macro_wrapper.h"
import "C"
import "unsafe"

var (
	ReTuneNow                 = Timestamp(C.ReTuneNow)
	MetaFlagTxBurstStart      = uint32(C.MetaFlagTxBurstStart)
	MetaFlagTxBurstEnd        = uint32(C.MetaFlagTxBurstEnd)
	MetaFlagTxNow             = uint32(C.MetaFlagTxNow)
	MetaFlagTxUpdateTimestamp = uint32(C.MetaFlagTxUpdateTimestamp)
	MetaFlagRxNow             = uint32(C.MetaFlagRxNow)
	MetaFlagRxHwUnderflow     = uint32(C.MetaFlagRxHwUnderflow)
	MetaFlagRxHwMiniexp1      = uint32(C.MetaFlagRxHwMiniexp1)
	MetaFlagRxHwMiniexp2      = uint32(C.MetaFlagRxHwMiniexp2)
	TriggerRegArm             = uint8(C.TriggerRegArm)
	TriggerRegFire            = uint8(C.TriggerRegFire)
	TriggerRegMaster          = uint8(C.TriggerRegMaster)
	TriggerRegLine            = uint8(C.TriggerRegLine)
	StreamNoData              = unsafe.Pointer(C.StreamNoData)
	StreamShutdown            = unsafe.Pointer(C.StreamShutdown)
)

func ChannelRx(ch int) Channel {
	return Channel(C.ChannelRx(C.int(ch)))
}

func ChannelTx(ch int) Channel {
	return Channel(C.ChannelTx(C.int(ch)))
}

func ChannelIsTx(ch Channel) bool {
	return C.ChannelIsTx(C.int(ch)) != 0
}

func (ch Channel) IsTx() bool {
	return ChannelIsTx(ch)
}
