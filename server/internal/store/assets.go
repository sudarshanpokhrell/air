package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Asset is one stored file, identified by the hash of its content.
type Asset struct {
	Hash        string
	Key         string
	ContentType string
	FileExt     string
	SizeBytes   int64
	CreatedAt   time.Time
}

type AssetStore struct {
	db *sql.DB
}

// InsertAsset records a file that was stored. Inserting the same hash twice
// is not an error: the same hash always means the same bytes.
func (s *AssetStore) InsertAsset(ctx context.Context, a *Asset) error {
	const q = `
		INSERT INTO assets (hash, key, content_type, file_ext, size_bytes)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		ON CONFLICT (hash) DO NOTHING`

	_, err := s.db.ExecContext(ctx, q, a.Hash, a.Key, a.ContentType, a.FileExt, a.SizeBytes)
	return err
}

func (s *AssetStore) GetAsset(ctx context.Context, hash string) (*Asset, error) {
	const q = `
		SELECT hash, key, content_type, COALESCE(file_ext, ''), size_bytes, created_at
		FROM assets
		WHERE hash = $1`

	var a Asset
	err := s.db.QueryRowContext(ctx, q, hash).
		Scan(&a.Hash, &a.Key, &a.ContentType, &a.FileExt, &a.SizeBytes, &a.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// MissingAssets returns the hashes (in input order, without duplicates)
// that are not stored yet. The CLI uploads only these.
func (s *AssetStore) MissingAssets(ctx context.Context, hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return []string{}, nil
	}

	const q = `
		SELECT h
		FROM (
			SELECT h, min(i) AS i
			FROM unnest($1::text[]) WITH ORDINALITY AS t(h, i)
			GROUP BY h
		) input
		WHERE NOT EXISTS (SELECT 1 FROM assets a WHERE a.hash = input.h)
		ORDER BY i`

	rows, err := s.db.QueryContext(ctx, q, hashes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	missing := []string{}
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		missing = append(missing, h)
	}
	return missing, rows.Err()
}
