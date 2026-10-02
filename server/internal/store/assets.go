package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Asset struct {
	Hash        string    `json:"hash"`
	Key         string    `json:"key"`
	ContentType string    `json:"content_type"`
	FileExt     string    `json:"file_ext"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

type AssetStore struct {
	db *sql.DB
}

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
