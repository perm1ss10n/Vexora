package validate

import (
    "errors"
    "strings"

    "github.com/perm1ss10n/vexora/backend/internal/model"
)

func EnvelopeBasic(e model.Envelope) error {
    if e.V <= 0 {
        return errors.New("invalid v")
    }
    if strings.TrimSpace(e.DeviceID) == "" {
        return errors.New("missing deviceId")
    }
    if e.Ts <= 0 {
        return errors.New("invalid ts")
    }
    return nil
}