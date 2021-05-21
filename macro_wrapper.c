#include "macro_wrapper.h"

uint64_t ReTuneNow = BLADERF_RETUNE_NOW;
uint8_t TriggerRegArm = BLADERF_TRIGGER_REG_ARM;
uint8_t TriggerRegFire = BLADERF_TRIGGER_REG_FIRE;
uint8_t TriggerRegMaster = BLADERF_TRIGGER_REG_MASTER;
uint8_t TriggerRegLine = BLADERF_TRIGGER_REG_LINE;

int ChannelRx(const int ch) {
  return BLADERF_CHANNEL_RX(ch);
}

int ChannelTx(const int ch) {
  return BLADERF_CHANNEL_TX(ch);
}

int ChannelIsTx(const int ch) {
  return BLADERF_CHANNEL_IS_TX(ch);
}