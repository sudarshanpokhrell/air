package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type APIKey struct {
	ID        string     `json:"id"`
	AppID     string     `json:"-"`
	Name      string     `json:"name"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

const apiKeyColumns = `k.id, k.app_id, k.name, k.created_by, k.created_at, k.revoked_at`

func scanAPIKey(row scanner, k *APIKey) error {
	return row.Scan(&k.ID, &k.AppID, &k.Name, &k.CreatedBy, &k.CreatedAt, &k.RevokedAt)
}

type APIKeyStore struct {
	db *sql.DB
}

func (s *APIKeyStore) CreateAPIKey(ctx context.Context, k *APIKey, keyHash []byte) error {
	const q = `
		INSERT INTO api_keys (app_id, name, key_hash, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return s.db.QueryRowContext(ctx, q, k.AppID, k.Name, keyHash, k.CreatedBy).Scan(&k.ID, &k.CreatedAt)
}

// GetAPIKeyByHash returns a usable key: not revoked, its creator is active
// and still has access to the key's app (global admin or app member).
func (s *APIKeyStore) GetAPIKeyByHash(ctx context.Context, keyHash []byte) (*APIKey, error) {
	q := `
		SELECT ` + apiKeyColumns + `
		FROM api_keys k
		JOIN users u ON u.id = k.created_by
		WHERE k.key_hash = $1
		  AND k.revoked_at IS NULL
		  AND u.is_active
		  AND (u.is_admin OR EXISTS (
		      SELECT 1 FROM app_members m WHERE m.app_id = k.app_id AND m.user_id = u.id
		  ))`

	var k APIKey
	err := scanAPIKey(s.db.QueryRowContext(ctx, q, keyHash), &k)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// GetAPIKeyByID returns a key of an app (for revoke ownership checks).
func (s *APIKeyStore) GetAPIKeyByID(ctx context.Context, appID, id string) (*APIKey, error) {
	q := `SELECT ` + apiKeyColumns + ` FROM api_keys k WHERE k.app_id = $1 AND k.id = $2`

	var k APIKey
	err := scanAPIKey(s.db.QueryRowContext(ctx, q, appID, id), &k)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *APIKeyStore) ListAPIKeys(ctx context.Context, appID string) ([]*APIKey, error) {
	q := `
		SELECT ` + apiKeyColumns + `
		FROM api_keys k
		WHERE k.app_id = $1
		ORDER BY k.created_at`

	rows, err := s.db.QueryContext(ctx, q, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []*APIKey{}
	for rows.Next() {
		var k APIKey
		if err := scanAPIKey(rows, &k); err != nil {
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
