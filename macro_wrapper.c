#include "macro_wrapper.h"

uint64_t ReTuneNow = BLADERF_RETUNE_NOW;

int ChannelRx(const int ch) {
  return BLADERF_CHANNEL_RX(ch);
}

int ChannelTx(const int ch) {
  return BLADERF_CHANNEL_TX(ch);
}

int ChannelIsTx(const int ch) {
  return BLADERF_CHANNEL_IS_TX(ch);
}