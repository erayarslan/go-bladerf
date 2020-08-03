package bladerf

// #include "macro_wrapper.h"
import "C"

func CHANNEL_RX(ch int) Channel {
	return Channel(C.ChannelRX(C.int(ch)))
}

func CHANNEL_TX(ch int) Channel {
	return Channel(C.ChannelTX(C.int(ch)))
}
