package fpga_source

// #include <libbladeRF.h>
import "C"

type FPGASource int

const (
	SourceUnknown FPGASource = C.BLADERF_FPGA_SOURCE_UNKNOWN
	SourceFlash   FPGASource = C.BLADERF_FPGA_SOURCE_FLASH
	SourceHost    FPGASource = C.BLADERF_FPGA_SOURCE_HOST
)
