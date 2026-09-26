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
	Kind           string  // KindUpdate or KindRollbackToEmbedded
	LaunchAsset    *string // bundle hash; nil for KindRollbackToEmbedded
	Message        string
	GitCommit      string
	RolloutPercent int
	CreatedAt      time.Time
}

// NewUpdate is the input for one publish (one or more platforms).
type NewUpdate struct {
	AppID          string
	Channel        string
	Kind           string // "" means KindUpdate
	Message        string
	GitCommit      string
	RolloutPercent int // 0 means 100 (everyone)
	Platforms      []PlatformBuild
}

// PlatformBuild is what one platform contributes to a publish.
type PlatformBuild struct {
	Platform       string
	RuntimeVersion string
	LaunchAsset    string   // bundle hash; "" for KindRollbackToEmbedded
	Assets         []string // other asset hashes (images, fonts)
}

type UpdateStore struct {
	db *sql.DB
}

const updateColumns = `
	id, group_id, app_id, channel, platform, runtime_version, kind,
	launch_asset, COALESCE(message, ''), COALESCE(git_commit, ''),
	rollout_percent, created_at`

type scanner interface {
	Scan(dest ...any) error
}

func scanUpdate(row scanner) (*Update, error) {
	var u Update
	err := row.Scan(
		&u.ID, &u.GroupID, &u.AppID, &u.Channel, &u.Platform, &u.RuntimeVersion, &u.Kind,
		&u.LaunchAsset, &u.Message, &u.GitCommit,
		&u.RolloutPercent, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func scanUpdates(rows *sql.Rows) ([]*Update, error) {
	defer rows.Close()
	updates := []*Update{}
	for rows.Next() {
		u, err := scanUpdate(rows)
		if err != nil {
			return nil, err
		}
		updates = append(updates, u)
	}
	return updates, rows.Err()
}

// CreateUpdate inserts one row per platform, all sharing one group_id and
// created_at, plus their update_assets rows. Everything happens in one
// transaction: if any platform fails, nothing is saved.
func (s *UpdateStore) CreateUpdate(ctx context.Context, nu NewUpdate) ([]*Update, error) {
	if len(nu.Platforms) == 0 {
		return nil, errors.New("no platforms")
	}
	kind := nu.Kind
	if kind == "" {
		kind = KindUpdate
	}
	rollout := nu.RolloutPercent
	if rollout == 0 {
		rollout = 100
	}

	const insertUpdate = `
		INSERT INTO updates (
			id, group_id, app_id, channel, platform, runtime_version, kind,
			launch_asset, message, git_commit, rollout_percent
		)
		VALUES (
			gen_random_uuid(), $1, $2, $3, $4::platform, $5, $6::update_kind,
			NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), $10
		)
		RETURNING ` + updateColumns

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
			u, err := scanUpdate(tx.QueryRowContext(ctx, insertUpdate,
				groupID, nu.AppID, nu.Channel, p.Platform, p.RuntimeVersion, kind,
				p.LaunchAsset, nu.Message, nu.GitCommit, rollout,
			))
			if err != nil {
				return fmt.Errorf("insert %s update: %w", p.Platform, err)
			}

			if len(p.Assets) > 0 {
				if _, err := tx.ExecContext(ctx, insertAssets, u.ID, p.Assets); err != nil {
					return fmt.Errorf("insert %s assets: %w", p.Platform, err)
				}
			}
			created = append(created, u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// LatestUpdates returns the newest updates a device could receive, newest first.
// More than one is returned so rollouts can fall back to an older update.
func (s *UpdateStore) LatestUpdates(ctx context.Context, appID, channel, platform, runtimeVersion string, limit int) ([]*Update, error) {
	q := `
		SELECT ` + updateColumns + `
		FROM updates
		WHERE app_id = $1 AND channel = $2 AND platform = $3::platform AND runtime_version = $4
		ORDER BY created_at DESC
		LIMIT $5`

	rows, err := s.db.QueryContext(ctx, q, appID, channel, platform, runtimeVersion, limit)
	if err != nil {
		return nil, err
	}
	return scanUpdates(rows)
}

// ListUpdates returns an app's update history, newest first.
func (s *UpdateStore) ListUpdates(ctx context.Context, appID string, limit int) ([]*Update, error) {
	q := `
		SELECT ` + updateColumns + `
		FROM updates
		WHERE app_id = $1
		ORDER BY created_at DESC, platform
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, q, appID, limit)
	if err != nil {
		return nil, err
	}
	return scanUpdates(rows)
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

// SetRolloutPercent changes the rollout of every platform in a publish.
// Returns ErrNotFound if the group doesn't belong to the app.
func (s *UpdateStore) SetRolloutPercent(ctx context.Context, appID, groupID string, percent int) error {
	const q = `
		UPDATE updates
		SET rollout_percent = $3
		WHERE app_id = $1 AND group_id = $2`

	res, err := s.db.ExecContext(ctx, q, appID, groupID, percent)
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
