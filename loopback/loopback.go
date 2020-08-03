package loopback

// #include <libbladeRF.h>
import "C"

type Loopback int

const (
	Disabled         Loopback = C.BLADERF_LB_NONE
	Firmware         Loopback = C.BLADERF_LB_FIRMWARE
	BB_TXLPF_RXVGA2  Loopback = C.BLADERF_LB_BB_TXLPF_RXVGA2
	BB_TXVGA1_RXVGA2 Loopback = C.BLADERF_LB_BB_TXVGA1_RXVGA2
	BB_TXLPF_RXLPF   Loopback = C.BLADERF_LB_BB_TXLPF_RXLPF
	BB_TXVGA1_RXLPF  Loopback = C.BLADERF_LB_BB_TXVGA1_RXLPF
	RF_LNA1          Loopback = C.BLADERF_LB_RF_LNA1
	RF_LNA2          Loopback = C.BLADERF_LB_RF_LNA2
	RF_LNA3          Loopback = C.BLADERF_LB_RF_LNA3
	RFIC_BIST        Loopback = C.BLADERF_LB_RFIC_BIST
)
