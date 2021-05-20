#include <libbladeRF.h>

extern uint64_t ReTuneNow;

int ChannelRx(const int ch);
int ChannelTx(const int ch);
int ChannelIsTx(const int ch);