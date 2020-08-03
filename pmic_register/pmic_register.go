package pmic_register

// #include <libbladeRF.h>
import "C"

type PMICRegister int

const (
	Configuration PMICRegister = C.BLADERF_PMIC_CONFIGURATION
	Voltage_shunt PMICRegister = C.BLADERF_PMIC_VOLTAGE_SHUNT
	Voltage_bus   PMICRegister = C.BLADERF_PMIC_VOLTAGE_BUS
	Power         PMICRegister = C.BLADERF_PMIC_POWER
	Current       PMICRegister = C.BLADERF_PMIC_CURRENT
	Calibration   PMICRegister = C.BLADERF_PMIC_CALIBRATION
)
