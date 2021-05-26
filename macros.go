package bladerf

// #include "macro_wrapper.h"
import "C"
import "unsafe"

var ReTuneNow = Timestamp(C.ReTuneNow)
var MetaFlagTxBurstStart = uint32(C.MetaFlagTxBurstStart)
var MetaFlagTxBurstEnd = uint32(C.MetaFlagTxBurstEnd)
var MetaFlagTxNow = uint32(C.MetaFlagTxNow)
var MetaFlagTxUpdateTimestamp = uint32(C.MetaFlagTxUpdateTimestamp)
var MetaFlagRxNow = uint32(C.MetaFlagRxNow)
var MetaFlagRxHwUnderflow = uint32(C.MetaFlagRxHwUnderflow)
var MetaFlagRxHwMiniexp1 = uint32(C.MetaFlagRxHwMiniexp1)
var MetaFlagRxHwMiniexp2 = uint32(C.MetaFlagRxHwMiniexp2)
var TriggerRegArm = C.TriggerRegArm
var TriggerRegFire = C.TriggerRegFire
var TriggerRegMaster = C.TriggerRegMaster
var TriggerRegLine = C.TriggerRegLine
var StreamNoData = unsafe.Pointer(C.StreamNoData)
var StreamShutdown = unsafe.Pointer(C.StreamShutdown)

func ChannelRx(ch int) Channel {
	return Channel(C.ChannelRx(C.int(ch)))
}

func ChannelTx(ch int) Channel {
	return Channel(C.ChannelTx(C.int(ch)))
}

func ChannelIsTx(ch int) bool {
	return C.ChannelIsTx(C.int(ch)) == 1
}
