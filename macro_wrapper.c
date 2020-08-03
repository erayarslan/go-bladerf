#include "macro_wrapper.h"

int ChannelRX(const int ch)
{
 return BLADERF_CHANNEL_RX(ch);
}

int ChannelTX(const int ch)
{
 return BLADERF_CHANNEL_TX(ch);
}