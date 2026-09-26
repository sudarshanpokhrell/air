package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type APIKey struct {
	ID        string
	AppID     string
	Name      string // label
	CreatedAt time.Time
	RevokedAt *time.Time
}

type APIKeyStore struct {
	db *sql.DB
}

func (s *APIKeyStore) CreateAPIKey(ctx context.Context, k *APIKey, keyHash []byte) error {
	const q = `
		INSERT INTO api_keys (app_id, name, key_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return s.db.QueryRowContext(ctx, q, k.AppID, k.Name, keyHash).Scan(&k.ID, &k.CreatedAt)
}

func (s *APIKeyStore) GetAPIKeyByHash(ctx context.Context, keyHash []byte) (*APIKey, error) {
	const q = `
		SELECT id, app_id, name, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1 AND revoked_at IS NULL`

	var k APIKey
	err := s.db.QueryRowContext(ctx, q, keyHash).
		Scan(&k.ID, &k.AppID, &k.Name, &k.CreatedAt, &k.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *APIKeyStore) ListAPIKeys(ctx context.Context, appID string) ([]*APIKey, error) {
	const q = `
		SELECT id, app_id, name, created_at, revoked_at
		FROM api_keys
		WHERE app_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, q, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.AppID, &k.Name, &k.CreatedAt, &k.RevokedAt); err != nil {
			return nil, err
		}
		keys = append(keys, &k)
	}
	return keys, rows.Err()
}

func (s *APIKeyStore) RevokeAPIKey(ctx context.Context, id string) error {
	const q = `
		UPDATE api_keys
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL`

	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
