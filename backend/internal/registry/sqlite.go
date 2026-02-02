package registry

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

type SQLiteConfig struct {
	Path string // путь к файлу БД
}

func LoadSQLiteConfigFromEnv() SQLiteConfig {
	// По умолчанию: backend/.data/vexora.db (чтобы не засорять корень)
	p := os.Getenv("REGISTRY_DB_PATH")
	if p == "" {
		p = "./.data/vexora.db"
	}
	return SQLiteConfig{Path: p}
}

func NewSQLite(cfg SQLiteConfig) (*SQLiteStore, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("registry db path is empty")
	}

	// создаём директорию под БД
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite настройки для нормальной работы
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma journal_mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma synchronous: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) DB() *sql.DB {
	if s == nil {
		return nil
	}
	return s.db
}

func (s *SQLiteStore) migrate() error {
	// Минимальная таблица с полями “на вырост” под 2.3.2/2.3.3/2.3.4
	// first_seen_ts / last_seen_ts — Unix millis
	ddl := `
CREATE TABLE IF NOT EXISTS devices (
  device_id TEXT PRIMARY KEY,

  first_seen_ts     INTEGER NOT NULL,
  last_seen_ts      INTEGER NOT NULL,

  status            TEXT DEFAULT NULL,   -- online/offline/degraded/error
  link              TEXT DEFAULT NULL,   -- wifi/gsm
  fw                TEXT DEFAULT NULL,

  last_state_ts     INTEGER DEFAULT NULL,
  last_telemetry_ts INTEGER DEFAULT NULL,

  updated_at_ts     INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(last_seen_ts);

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  refresh_hash TEXT NOT NULL,
  expires_at INTEGER NOT NULL,
  created_at INTEGER NOT NULL,
  last_used_at INTEGER NOT NULL,
  revoked_at INTEGER NULL,
  user_agent TEXT NULL,
  ip TEXT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_refresh_hash ON sessions(refresh_hash);
`
	_, err := s.db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("registry migrate: %w", err)
	}
	return nil
}

// Touch: upsert устройства + обновление last_seen_ts / updated_at_ts.
// first_seen_ts выставляем только при первом появлении.
func (s *SQLiteStore) Touch(ctx context.Context, deviceID string, tsMillis int64, source string) error {
	if deviceID == "" {
		return nil
	}

	// если прилетел нулевой/битый ts — подстрахуемся текущим временем
	if tsMillis <= 0 {
		tsMillis = time.Now().UnixMilli()
	}

	// SQLite upsert: при конфликте по PK — обновляем last_seen_ts/updated_at_ts.
	// first_seen_ts остаётся прежним.
	q := `
INSERT INTO devices(device_id, first_seen_ts, last_seen_ts, updated_at_ts)
VALUES (?, ?, ?, ?)
ON CONFLICT(device_id) DO UPDATE SET
  last_seen_ts = excluded.last_seen_ts,
  updated_at_ts = excluded.updated_at_ts;
`
	_, err := s.db.ExecContext(ctx, q, deviceID, tsMillis, tsMillis, tsMillis)
	if err != nil {
		return fmt.Errorf("registry touch deviceId=%s: %w", deviceID, err)
	}
	_ = source // пока не используем, пригодится для аудита/метрик
	return nil
}
func (s *SQLiteStore) UpdateState(
	ctx context.Context,
	deviceID string,
	status string,
	link *string,
	fw *string,
	tsMillis int64,
) error {
	if deviceID == "" {
		return nil
	}
	if tsMillis <= 0 {
		tsMillis = time.Now().UnixMilli()
	}

	q := `
UPDATE devices SET
  status = ?,
  link = COALESCE(?, link),
  fw = COALESCE(?, fw),
  last_state_ts = ?,
  last_seen_ts = ?,
  updated_at_ts = ?
WHERE device_id = ?;
`
	_, err := s.db.ExecContext(
		ctx,
		q,
		status,
		link,
		fw,
		tsMillis,
		tsMillis,
		tsMillis,
		deviceID,
	)
	if err != nil {
		return fmt.Errorf("registry update state deviceId=%s: %w", deviceID, err)
	}
	return nil
}

func (s *SQLiteStore) MarkOffline(
	ctx context.Context,
	deviceID string,
	tsMillis int64,
	reason string,
) error {
	if deviceID == "" {
		return nil
	}
	if tsMillis <= 0 {
		tsMillis = time.Now().UnixMilli()
	}

	q := `
UPDATE devices SET
  status = 'offline',
  last_seen_ts = ?,
  updated_at_ts = ?
WHERE device_id = ?;
`
	_, err := s.db.ExecContext(
		ctx,
		q,
		tsMillis,
		tsMillis,
		deviceID,
	)
	if err != nil {
		return fmt.Errorf("registry mark offline deviceId=%s: %w", deviceID, err)
	}
	_ = reason // оставим под аудит/логи позже
	return nil
}

type DeviceRecord struct {
	DeviceID        string
	Status          string
	LastSeenMillis  int64
	TelemetryMillis sql.NullInt64
	FW              sql.NullString
}

func (s *SQLiteStore) ListDevices(ctx context.Context) ([]DeviceRecord, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT device_id, status, last_seen_ts, last_telemetry_ts, fw
FROM devices
ORDER BY last_seen_ts DESC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("registry list devices: %w", err)
	}
	defer rows.Close()

	devices := []DeviceRecord{}
	for rows.Next() {
		var record DeviceRecord
		if err := rows.Scan(
			&record.DeviceID,
			&record.Status,
			&record.LastSeenMillis,
			&record.TelemetryMillis,
			&record.FW,
		); err != nil {
			return nil, fmt.Errorf("registry scan devices: %w", err)
		}
		devices = append(devices, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("registry list devices rows: %w", err)
	}
	return devices, nil
}
