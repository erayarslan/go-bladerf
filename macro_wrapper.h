#include <libbladeRF.h>

extern uint64_t ReTuneNow;
extern uint8_t TriggerRegArm;
extern uint8_t TriggerRegFire;
extern uint8_t TriggerRegMaster;
extern uint8_t TriggerRegLine;

int ChannelRx(const int ch);
int ChannelTx(const int ch);
int ChannelIsTx(const int ch);