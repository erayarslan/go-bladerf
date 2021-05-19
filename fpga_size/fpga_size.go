package fpga_size

// #include <libbladeRF.h>
import "C"

type FPGASize int

const (
	SizeUnknown FPGASize = C.BLADERF_FPGA_UNKNOWN
	Size40KLE   FPGASize = C.BLADERF_FPGA_40KLE
	Size115KLE  FPGASize = C.BLADERF_FPGA_115KLE
	SizeA4      FPGASize = C.BLADERF_FPGA_A4
	SizeA9      FPGASize = C.BLADERF_FPGA_A9
)
