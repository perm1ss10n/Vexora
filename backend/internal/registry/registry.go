package registry

import "context"

// Store — минимальный интерфейс для Registry (на вырост).
type Store interface {
	Close() error

	// Touch — отметить, что устройство "видели" в момент tsMillis.
	// Должно гарантировать наличие записи устройства (upsert).
	Touch(ctx context.Context, deviceID string, tsMillis int64, source string) error
	UpdateState(ctx context.Context, deviceID string, status string, link *string, fw *string, tsMillis int64) error
	MarkOffline(ctx context.Context, deviceID string, tsMillis int64, reason string) error
}
