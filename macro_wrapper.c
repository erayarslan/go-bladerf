#include "macro_wrapper.h"

uint64_t ReTuneNow = BLADERF_RETUNE_NOW;
uint32_t MetaFlagTxBurstStart = BLADERF_META_FLAG_TX_BURST_START;
uint32_t MetaFlagTxBurstEnd = BLADERF_META_FLAG_TX_BURST_END;
uint32_t MetaFlagTxNow = BLADERF_META_FLAG_TX_NOW;
uint32_t MetaFlagTxUpdateTimestamp = BLADERF_META_FLAG_TX_UPDATE_TIMESTAMP;
uint32_t MetaFlagRxNow = BLADERF_META_FLAG_RX_NOW;
uint32_t MetaFlagRxHwUnderflow = BLADERF_META_FLAG_RX_HW_UNDERFLOW;
uint32_t MetaFlagRxHwMiniexp1 = BLADERF_META_FLAG_RX_HW_MINIEXP1;
uint32_t MetaFlagRxHwMiniexp2 = BLADERF_META_FLAG_RX_HW_MINIEXP2;
uint8_t TriggerRegArm = BLADERF_TRIGGER_REG_ARM;
uint8_t TriggerRegFire = BLADERF_TRIGGER_REG_FIRE;
uint8_t TriggerRegMaster = BLADERF_TRIGGER_REG_MASTER;
uint8_t TriggerRegLine = BLADERF_TRIGGER_REG_LINE;
void* StreamNoData = BLADERF_STREAM_NO_DATA;
void* StreamShutdown = BLADERF_STREAM_SHUTDOWN;

int ChannelRx(const int ch) {
  return BLADERF_CHANNEL_RX(ch);
}

int ChannelTx(const int ch) {
  return BLADERF_CHANNEL_TX(ch);
}

int ChannelIsTx(const int ch) {
  return BLADERF_CHANNEL_IS_TX(ch);
}