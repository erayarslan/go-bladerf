package bladerf

// #include "macro_wrapper.h"
import "C"
import "bladerf/channel"

func CHANNEL_RX(ch int) channel.Channel {
	return channel.Channel(C.ChannelRX(C.int(ch)))
}

func CHANNEL_TX(ch int) channel.Channel {
	return channel.Channel(C.ChannelTX(C.int(ch)))
}
