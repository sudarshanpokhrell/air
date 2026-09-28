package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	KindUpdate             = "update"
	KindRollbackToEmbedded = "rollback_to_embedded"
)

// Update is one row of the updates table: one platform of one publish.
type Update struct {
	ID             string
	GroupID        string
	AppID          string
	Channel        string
	Platform       string
	RuntimeVersion string
	Kind           string
	LaunchAsset    *string // bundle hash; nil for KindRollbackToEmbedded
	Message        string
	GitCommit      string
	Manifest       []byte     // exact bytes from the CLI, served unchanged
	Signature      string     // expo-signature header value; "" when unsigned
	RolledBackAt   *time.Time // set by rollback; such rows are never served
	CreatedAt      time.Time
}

// Update is for the new update (for 1 or 2 platform)
type NewUpdate struct {
	AppID     string
	Channel   string
	Kind      string // "" means KindUpdate
	Message   string
	GitCommit string
	Platforms []PlatformBuild
}

// every update has a platfrom related info
type PlatformBuild struct {
	ID             string
	Platform       string
	RuntimeVersion string
	LaunchAsset    string   // bundle hash; "" for KindRollbackToEmbedded
	Assets         []string // other asset hashes (images, fonts)
	Manifest       []byte   // manifest (or directive) JSON exactly as signed
	Signature      string   // "" when unsigned
}

type UpdateStore struct {
	db *sql.DB
}

// CreateUpdate inserts one row per platform
func (s *UpdateStore) CreateUpdate(ctx context.Context, nu NewUpdate) ([]*Update, error) {
	if len(nu.Platforms) == 0 {
		return nil, errors.New("no platforms")
	}
	kind := nu.Kind
	if kind == "" {
		kind = KindUpdate
	}

	const insertUpdate = `
		INSERT INTO updates (
			id, group_id, app_id, channel, platform, runtime_version, kind,
			launch_asset, message, git_commit, manifest, signature
		)
		VALUES (
			$1, $2, $3, $4, $5::platform, $6, $7::update_kind,
			NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, NULLIF($12, '')
		)
		RETURNING
			id, group_id, app_id, channel, platform, runtime_version, kind,
			launch_asset, COALESCE(message, ''), COALESCE(git_commit, ''),
			manifest, COALESCE(signature, ''), rolled_back_at, created_at`

	const insertAssets = `
		INSERT INTO update_assets (update_id, asset_hash)
		SELECT $1, unnest($2::text[])
		ON CONFLICT DO NOTHING`

	var created []*Update
	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		var groupID string
		if err := tx.QueryRowContext(ctx, `SELECT gen_random_uuid()`).Scan(&groupID); err != nil {
			return err
		}

		for _, p := range nu.Platforms {
			var u Update
			err := tx.QueryRowContext(ctx, insertUpdate,
				p.ID, groupID, nu.AppID, nu.Channel, p.Platform, p.RuntimeVersion, kind,
				p.LaunchAsset, nu.Message, nu.GitCommit, p.Manifest, p.Signature,
			).Scan(
				&u.ID, &u.GroupID, &u.AppID, &u.Channel, &u.Platform, &u.RuntimeVersion, &u.Kind,
				&u.LaunchAsset, &u.Message, &u.GitCommit,
				&u.Manifest, &u.Signature, &u.RolledBackAt, &u.CreatedAt,
			)
			if isUniqueViolation(err) {
				return ErrConflict
			}
			if err != nil {
				return fmt.Errorf("insert %s update: %w", p.Platform, err)
			}

			if len(p.Assets) > 0 {
				if _, err := tx.ExecContext(ctx, insertAssets, u.ID, p.Assets); err != nil {
					return fmt.Errorf("insert %s assets: %w", p.Platform, err)
				}
			}
			created = append(created, &u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// LatestUpdates returns the newest updates a device could receive, newest first.
// Rolled-back updates are skipped.
func (s *UpdateStore) LatestUpdates(ctx context.Context, appID, channel, platform, runtimeVersion string, limit int) ([]*Update, error) {
	const q = `
		SELECT
			id, group_id, app_id, channel, platform, runtime_version, kind,
			launch_asset, COALESCE(message, ''), COALESCE(git_commit, ''),
			manifest, COALESCE(signature, ''), rolled_back_at, created_at
		FROM updates
		WHERE app_id = $1 AND channel = $2 AND platform = $3::platform AND runtime_version = $4
		  AND rolled_back_at IS NULL
		ORDER BY created_at DESC
		LIMIT $5`

	rows, err := s.db.QueryContext(ctx, q, appID, channel, platform, runtimeVersion, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updates := []*Update{}
	for rows.Next() {
		var u Update
		err := rows.Scan(
			&u.ID, &u.GroupID, &u.AppID, &u.Channel, &u.Platform, &u.RuntimeVersion, &u.Kind,
			&u.LaunchAsset, &u.Message, &u.GitCommit,
			&u.Manifest, &u.Signature, &u.RolledBackAt, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		updates = append(updates, &u)
	}
	return updates, rows.Err()
}

// ListUpdates returns an app's update history, newest first.
func (s *UpdateStore) ListUpdates(ctx context.Context, appID string, limit int) ([]*Update, error) {
	const q = `
		SELECT
			id, group_id, app_id, channel, platform, runtime_version, kind,
			launch_asset, COALESCE(message, ''), COALESCE(git_commit, ''),
			manifest, COALESCE(signature, ''), rolled_back_at, created_at
		FROM updates
		WHERE app_id = $1
		ORDER BY created_at DESC, platform
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, q, appID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updates := []*Update{}
	for rows.Next() {
		var u Update
		err := rows.Scan(
			&u.ID, &u.GroupID, &u.AppID, &u.Channel, &u.Platform, &u.RuntimeVersion, &u.Kind,
			&u.LaunchAsset, &u.Message, &u.GitCommit,
			&u.Manifest, &u.Signature, &u.RolledBackAt, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		updates = append(updates, &u)
	}
	return updates, rows.Err()
}

// GetUpdateAssets returns the assets of an update (not including the launch asset).
func (s *UpdateStore) GetUpdateAssets(ctx context.Context, updateID string) ([]*Asset, error) {
	const q = `
		SELECT a.hash, a.key, a.content_type, COALESCE(a.file_ext, ''), a.size_bytes, a.created_at
		FROM update_assets ua
		JOIN assets a ON a.hash = ua.asset_hash
		WHERE ua.update_id = $1
		ORDER BY a.hash`

	rows, err := s.db.QueryContext(ctx, q, updateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := []*Asset{}
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.Hash, &a.Key, &a.ContentType, &a.FileExt, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, &a)
	}
	return assets, rows.Err()
}
