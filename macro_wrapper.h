#ifndef GO_BLADERF_MACRO_WRAPPER_H
#define GO_BLADERF_MACRO_WRAPPER_H

#include <libbladeRF.h>

extern uint64_t ReTuneNow;
extern uint32_t MetaFlagTxBurstStart;
extern uint32_t MetaFlagTxBurstEnd;
extern uint32_t MetaFlagTxNow;
extern uint32_t MetaFlagTxUpdateTimestamp;
extern uint32_t MetaFlagRxNow;
extern uint32_t MetaFlagRxHwUnderflow;
extern uint32_t MetaFlagRxHwMiniexp1;
extern uint32_t MetaFlagRxHwMiniexp2;
extern uint8_t TriggerRegArm;
extern uint8_t TriggerRegFire;
extern uint8_t TriggerRegMaster;
extern uint8_t TriggerRegLine;
extern void* StreamNoData;
extern void* StreamShutdown;

extern void* StreamCallback(struct bladerf *dev, struct bladerf_stream *stream, struct bladerf_metadata *md, void *samples, size_t num_samples, void *user_data);

int ChannelRx(const int ch);
int ChannelTx(const int ch);
int ChannelIsTx(const int ch);

#endif
