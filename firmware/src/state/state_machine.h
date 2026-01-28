#pragma once
#include <Arduino.h>

enum class State {
    BOOT,
    INIT,
    ONLINE,
    OFFLINE,
    DEGRADED,
    ERROR
};

class StateMachine {
public:
    static void init();
    static void loop();
    static void set(State s);
    static State get();
};