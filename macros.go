package bladerf

// #include "macro_wrapper.h"
import "C"

func ChannelRx(ch int) Channel {
	return Channel(C.ChannelRX(C.int(ch)))
}

func ChannelTx(ch int) Channel {
	return Channel(C.ChannelTX(C.int(ch)))
}
