#pragma once

#include "link.h"

struct GsmConfig
{
    // пока пусто, позже добавим APN/user/pass, pins, baud, etc.
};

class GsmLink : public ILink
{

public:
    explicit GsmLink(const GsmConfig &cfg);

    LinkType type() const override { return LinkType::GSM; }

    bool begin() override;
    bool isUp() const override;
    void loop() override;
    void disconnect() override;
    LinkStatus status() const override;

private:
    GsmConfig _cfg;
};