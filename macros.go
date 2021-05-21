package bladerf

// #include "macro_wrapper.h"
import "C"

var ReTuneNow = Timestamp(C.ReTuneNow)

var TriggerRegArm = C.TriggerRegArm
var TriggerRegFire = C.TriggerRegFire
var TriggerRegMaster = C.TriggerRegMaster
var TriggerRegLine = C.TriggerRegLine

func ChannelRx(ch int) Channel {
	return Channel(C.ChannelRx(C.int(ch)))
}

func ChannelTx(ch int) Channel {
	return Channel(C.ChannelTx(C.int(ch)))
}

func ChannelIsTx(ch int) bool {
	return C.ChannelIsTx(C.int(ch)) == 1
}
