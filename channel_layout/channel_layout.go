package channel_layout

// #include <libbladeRF.h>
import "C"

type ChannelLayout int

const (
	RX_X1 ChannelLayout = C.BLADERF_RX_X1
	TX_X1 ChannelLayout = C.BLADERF_TX_X1
	RX_X2 ChannelLayout = C.BLADERF_RX_X2
	TX_X2 ChannelLayout = C.BLADERF_TX_X2
)
