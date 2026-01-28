#include "state/state_machine.h"
#include "log/logger.h"

static State g_state = State::BOOT;

void StateMachine::init() {
    g_state = State::BOOT;
    LOGI("STATE", "init()");
}

void StateMachine::loop() {
    // 3.1: пусто, потом добавим переходы и таймеры
}

void StateMachine::set(State s) {
    g_state = s;
    LOGI("STATE", "set()");
}

State StateMachine::get() {
    return g_state;
}