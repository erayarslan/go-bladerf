package bladerf

// #include <libbladeRF.h>
import "C"

type Backend int
type Channel int
type ChannelLayout int
type ClockSelect int
type Correction int
type Direction int
type DeviceSpeed int
type Format int
type FpgaSize int
type FpgaSource int
type GainMode int
type Module int
type Loopback int
type PmicRegister int
type PowerSource int
type RxMux int
type TriggerRole int
type TriggerSignal int
type ExpansionBoard int
type VctcxoTamerMode int

const (
	BackendAny     Backend = C.BLADERF_BACKEND_ANY
	BackendLinux   Backend = C.BLADERF_BACKEND_LINUX
	BackendLibUSB  Backend = C.BLADERF_BACKEND_LIBUSB
	BackendCypress Backend = C.BLADERF_BACKEND_CYPRESS
	BackendDummy   Backend = C.BLADERF_BACKEND_DUMMY
)

const (
	RxX1 ChannelLayout = C.BLADERF_RX_X1
	TxX1 ChannelLayout = C.BLADERF_TX_X1
	RxX2 ChannelLayout = C.BLADERF_RX_X2
	TxX2 ChannelLayout = C.BLADERF_TX_X2
)

const (
	ClockSelectOnboard  ClockSelect = C.CLOCK_SELECT_ONBOARD
	ClockSelectExternal ClockSelect = C.CLOCK_SELECT_EXTERNAL
)
const (
	CorrectionDcoffI Correction = C.BLADERF_CORR_DCOFF_I
	CorrectionDcoffQ Correction = C.BLADERF_CORR_DCOFF_Q
	CorrectionPhase  Correction = C.BLADERF_CORR_PHASE
	CorrectionGain   Correction = C.BLADERF_CORR_GAIN
)

const (
	Tx Direction = C.BLADERF_TX
	Rx Direction = C.BLADERF_RX
)

const (
	SpeedUnknown DeviceSpeed = C.BLADERF_DEVICE_SPEED_UNKNOWN
	SpeedHigh    DeviceSpeed = C.BLADERF_DEVICE_SPEED_HIGH
	SpeedSuper   DeviceSpeed = C.BLADERF_DEVICE_SPEED_SUPER
)

const (
	FormatSc16Q11     Format = C.BLADERF_FORMAT_SC16_Q11
	FormatSc16Q11Meta Format = C.BLADERF_FORMAT_SC16_Q11_META
)

const (
	FpgaSizeUnknown FpgaSize = C.BLADERF_FPGA_UNKNOWN
	FpgaSize40kle   FpgaSize = C.BLADERF_FPGA_40KLE
	FpgaSize115kle  FpgaSize = C.BLADERF_FPGA_115KLE
	FpgaSizeA4      FpgaSize = C.BLADERF_FPGA_A4
	FpgaSizeA9      FpgaSize = C.BLADERF_FPGA_A9
)

const (
	FpgaSourceUnknown FpgaSource = C.BLADERF_FPGA_SOURCE_UNKNOWN
	FpgaSourceFlash   FpgaSource = C.BLADERF_FPGA_SOURCE_FLASH
	FpgaSourceHost    FpgaSource = C.BLADERF_FPGA_SOURCE_HOST
)

const (
	GainModeDefault       GainMode = C.BLADERF_GAIN_DEFAULT
	GainModeManual        GainMode = C.BLADERF_GAIN_MGC
	GainModeFastAttackAgc GainMode = C.BLADERF_GAIN_FASTATTACK_AGC
	GainModeSlowAttackAgc GainMode = C.BLADERF_GAIN_SLOWATTACK_AGC
	GainModeHybridAgc     GainMode = C.BLADERF_GAIN_HYBRID_AGC
)

const (
	ModuleInvalid Module = C.BLADERF_MODULE_INVALID
	ModuleTx      Module = C.BLADERF_MODULE_TX
	ModuleRx      Module = C.BLADERF_MODULE_RX
)

const (
	LoopbackDisabled       Loopback = C.BLADERF_LB_NONE
	LoopbackFirmware       Loopback = C.BLADERF_LB_FIRMWARE
	LoopbackBbTxlpfRxvga2  Loopback = C.BLADERF_LB_BB_TXLPF_RXVGA2
	LoopbackBbTxvga1Rxvga2 Loopback = C.BLADERF_LB_BB_TXVGA1_RXVGA2
	LoopbackBbTxlpfRxlpf   Loopback = C.BLADERF_LB_BB_TXLPF_RXLPF
	LoopbackBbTxvga1Rxlpf  Loopback = C.BLADERF_LB_BB_TXVGA1_RXLPF
	LoopbackRfLna1         Loopback = C.BLADERF_LB_RF_LNA1
	LoopbackRfLna2         Loopback = C.BLADERF_LB_RF_LNA2
	LoopbackRfLna3         Loopback = C.BLADERF_LB_RF_LNA3
	LoopbackRficBist       Loopback = C.BLADERF_LB_RFIC_BIST
)

const (
	PmicConfiguration PmicRegister = C.BLADERF_PMIC_CONFIGURATION
	PmicVoltageShunt  PmicRegister = C.BLADERF_PMIC_VOLTAGE_SHUNT
	PmicVoltageBus    PmicRegister = C.BLADERF_PMIC_VOLTAGE_BUS
	PmicPower         PmicRegister = C.BLADERF_PMIC_POWER
	PmicCurrent       PmicRegister = C.BLADERF_PMIC_CURRENT
	PmicCalibration   PmicRegister = C.BLADERF_PMIC_CALIBRATION
)

const (
	PsUnknown PowerSource = C.BLADERF_UNKNOWN
	PsDc      PowerSource = C.BLADERF_PS_DC
	PsUsbVbus PowerSource = C.BLADERF_PS_USB_VBUS
)

const (
	RxMuxInvalid         RxMux = C.BLADERF_RX_MUX_INVALID
	RxMuxBaseband        RxMux = C.BLADERF_RX_MUX_BASEBAND
	RxMux12BitCounter    RxMux = C.BLADERF_RX_MUX_12BIT_COUNTER
	RxMux32BitCounter    RxMux = C.BLADERF_RX_MUX_32BIT_COUNTER
	RxMuxDigitalLoopback RxMux = C.BLADERF_RX_MUX_DIGITAL_LOOPBACK
)

const (
	TriggerRoleInvalid  TriggerRole = C.BLADERF_TRIGGER_ROLE_INVALID
	TriggerRoleDisabled TriggerRole = C.BLADERF_TRIGGER_ROLE_DISABLED
	TriggerRoleMaster   TriggerRole = C.BLADERF_TRIGGER_ROLE_MASTER
	TriggerRoleSlave    TriggerRole = C.BLADERF_TRIGGER_ROLE_SLAVE
)

const (
	TriggerSignalInvalid  TriggerSignal = C.BLADERF_TRIGGER_INVALID
	TriggerSignalJ714     TriggerSignal = C.BLADERF_TRIGGER_J71_4
	TriggerSignalJ511     TriggerSignal = C.BLADERF_TRIGGER_J51_1
	TriggerSignalMiniExp1 TriggerSignal = C.BLADERF_TRIGGER_MINI_EXP_1
	TriggerSignalUser0    TriggerSignal = C.BLADERF_TRIGGER_USER_0
	TriggerSignalUser1    TriggerSignal = C.BLADERF_TRIGGER_USER_1
	TriggerSignalUser2    TriggerSignal = C.BLADERF_TRIGGER_USER_2
	TriggerSignalUser3    TriggerSignal = C.BLADERF_TRIGGER_USER_3
	TriggerSignalUser4    TriggerSignal = C.BLADERF_TRIGGER_USER_4
	TriggerSignalUser5    TriggerSignal = C.BLADERF_TRIGGER_USER_5
	TriggerSignalUser6    TriggerSignal = C.BLADERF_TRIGGER_USER_6
	TriggerSignalUser7    TriggerSignal = C.BLADERF_TRIGGER_USER_7
)

const (
	ExpansionBoardNone ExpansionBoard = C.BLADERF_XB_NONE
	ExpansionBoard100  ExpansionBoard = C.BLADERF_XB_100
	ExpansionBoard200  ExpansionBoard = C.BLADERF_XB_200
	ExpansionBoard300  ExpansionBoard = C.BLADERF_XB_300
)

const (
	VctcxoTamerModeInvalid  VctcxoTamerMode = C.BLADERF_VCTCXO_TAMER_INVALID
	VctcxoTamerModeDisabled VctcxoTamerMode = C.BLADERF_VCTCXO_TAMER_DISABLED
	VctcxoTamerMode1Pps     VctcxoTamerMode = C.BLADERF_VCTCXO_TAMER_1_PPS
	VctcxoTamerMode10Mhz    VctcxoTamerMode = C.BLADERF_VCTCXO_TAMER_10_MHZ
)
