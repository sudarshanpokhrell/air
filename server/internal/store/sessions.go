package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

type Session struct {
	Plaintext string
	Hash      []byte
	UserID    string
	Expiry    time.Time
}

func NewSession(userID string, ttl time.Duration) (*Session, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	s := &Session{
		Plaintext: base64.RawURLEncoding.EncodeToString(b),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
	}
	s.Hash = hashSessionToken(s.Plaintext)
	return s, nil
}

func hashSessionToken(plaintext string) []byte {
	sum := sha256.Sum256([]byte(plaintext))
	return sum[:]
}

type SessionStore struct {
	db *sql.DB
}

func (s *SessionStore) CreateSession(ctx context.Context, session *Session) error {
	const q = `INSERT INTO sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, q, session.UserID, session.Hash, session.Expiry)
	return err
}

func (s *SessionStore) GetUserBySession(ctx context.Context, plaintext string) (*User, error) {
	const q = `
		SELECT u.id, u.email, u.name, u.password_hash, u.is_admin, u.is_active, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now() AND u.is_active`

	var u User

	err := s.db.QueryRowContext(ctx, q, hashSessionToken(plaintext)).
		Scan(&u.ID, &u.Email, &u.Name, &u.Password.hash, &u.IsAdmin, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *SessionStore) DeleteSession(ctx context.Context, plaintext string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hashSessionToken(plaintext))
	return err
}

func (s *SessionStore) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (s *SessionStore) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
