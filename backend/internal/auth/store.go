package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

const RefreshTokenTTL = 30 * 24 * time.Hour

type User struct {
	ID        string
	Email     string
	CreatedAt int64
}

type Session struct {
	ID          string
	UserID      string
	RefreshHash string
	ExpiresAt   int64
	CreatedAt   int64
	LastUsedAt  int64
	RevokedAt   sql.NullInt64
	UserAgent   sql.NullString
	IP          sql.NullString
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string, now time.Time) (User, error) {
	user := User{ID: uuid.NewString(), Email: email, CreatedAt: now.UnixMilli()}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO users(id, email, password_hash, created_at) VALUES (?, ?, ?, ?);`,
		user.ID,
		user.Email,
		passwordHash,
		user.CreatedAt,
	)
	return user, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var passwordHash string
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ?;`,
		email,
	)
	if err := row.Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt); err != nil {
		return user, "", err
	}
	return user, passwordHash, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
	var user User
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, created_at FROM users WHERE id = ?;`,
		id,
	)
	if err := row.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
		return user, err
	}
	return user, nil
}

func (s *Store) CreateSession(
	ctx context.Context,
	userID string,
	refreshHash string,
	expiresAt time.Time,
	userAgent string,
	ip string,
	now time.Time,
) (Session, error) {
	session := Session{
		ID:          uuid.NewString(),
		UserID:      userID,
		RefreshHash: refreshHash,
		ExpiresAt:   expiresAt.UnixMilli(),
		CreatedAt:   now.UnixMilli(),
		LastUsedAt:  now.UnixMilli(),
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
		IP:          sql.NullString{String: ip, Valid: ip != ""},
	}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sessions(id, user_id, refresh_hash, expires_at, created_at, last_used_at, revoked_at, user_agent, ip)
      VALUES (?, ?, ?, ?, ?, ?, NULL, ?, ?);`,
		session.ID,
		session.UserID,
		session.RefreshHash,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastUsedAt,
		session.UserAgent,
		session.IP,
	)
	return session, err
}

func (s *Store) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (Session, error) {
	var session Session
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, refresh_hash, expires_at, created_at, last_used_at, revoked_at, user_agent, ip
      FROM sessions WHERE refresh_hash = ?;`,
		refreshHash,
	)
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastUsedAt,
		&session.RevokedAt,
		&session.UserAgent,
		&session.IP,
	); err != nil {
		return session, err
	}
	return session, nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionID string, now time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at = ? WHERE id = ?;`,
		now.UnixMilli(),
		sessionID,
	)
	return err
}

func (s *Store) UpdateSessionLastUsed(ctx context.Context, sessionID string, now time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET last_used_at = ? WHERE id = ?;`,
		now.UnixMilli(),
		sessionID,
	)
	return err
}

func (s *Store) RefreshSession(ctx context.Context, sessionID string, now time.Time) error {
	res, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET last_used_at = ? WHERE id = ?;`,
		now.UnixMilli(),
		sessionID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("session not found")
	}
	return err
}
