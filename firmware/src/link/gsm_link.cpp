#include "gsm_link.h"

GsmLink::GsmLink(const GsmConfig &cfg) : _cfg(cfg) {}

bool GsmLink::begin()
{
    // TODO: TinyGSM attach
    return false;
}

bool GsmLink::isUp() const
{
    return false;
}

void GsmLink::loop()
{
    // TODO
}

void GsmLink::disconnect()
{
    // TODO
}

LinkStatus GsmLink::status() const
{
    LinkStatus st;
    st.type = LinkType::GSM;
    st.connected = false;
    st.rssi = 0;
    st.ip = "";
    return st;
}