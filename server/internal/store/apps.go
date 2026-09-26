package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
)

type App struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AppPlatform struct {
	AppID     string
	Platform  string
	BundleID  string
	Enabled   bool
	CreatedAt time.Time
}

type AppStore struct {
	db *sql.DB
}

func (s *AppStore) GetApps(ctx context.Context) ([]*App, error) {
	const q = `
		SELECT id, slug, name, created_at
		FROM apps
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*App{}
	for rows.Next() {
		var a App
		err := rows.Scan(
			&a.ID,
			&a.Slug,
			&a.Name,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, rows.Err()
}

func (s *AppStore) ListAppsForUser(ctx context.Context, userID string) ([]*App, error) {
	const q = `
		SELECT a.id, a.slug, a.name, a.created_at
		FROM apps a
		JOIN app_members m ON m.app_id = a.id
		WHERE m.user_id = $1
		ORDER BY a.created_at`

	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*App{}
	for rows.Next() {
		var a App
		err := rows.Scan(
			&a.ID,
			&a.Slug,
			&a.Name,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, rows.Err()
}

func (s *AppStore) GetAppBySlug(ctx context.Context, slug string) (*App, error) {
	const q = `
		SELECT id, slug, name, created_at
		FROM apps
		WHERE slug = $1`

	var a App
	err := s.db.QueryRowContext(ctx, q, slug).Scan(
		&a.ID,
		&a.Slug,
		&a.Name,
		&a.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AppStore) GetAppPlatform(ctx context.Context, appID, platform string) (*AppPlatform, error) {
	const q = `
		SELECT app_id, platform, COALESCE(bundle_id, ''), enabled, created_at
		FROM app_platforms
		WHERE app_id = $1 AND platform = $2::platform`

	var p AppPlatform
	err := s.db.QueryRowContext(ctx, q, appID, platform).Scan(
		&p.AppID,
		&p.Platform,
		&p.BundleID,
		&p.Enabled,
		&p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateApp inserts the app and makes creatorID its app admin, in one
// transaction. Returns ErrConflict if the slug is taken.
func (s *AppStore) CreateApp(ctx context.Context, app *App, creatorID string) error {
	const insertApp = `
		INSERT INTO apps (slug, name)
		VALUES ($1, $2)
		RETURNING id, created_at`

	const insertMember = `
		INSERT INTO app_members (app_id, user_id, role)
		VALUES ($1, $2, 'admin')`

	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, insertApp, app.Slug, app.Name).Scan(
			&app.ID,
			&app.CreatedAt,
		)
		if isUniqueViolation(err) {
			return ErrConflict
		}
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, insertMember, app.ID, creatorID)
		return err
	})
}

func (s *AppStore) AddPlatform(ctx context.Context, p *AppPlatform) error {
	const q = `
		INSERT INTO app_platforms (app_id, platform, bundle_id)
		VALUES ($1, $2::platform, NULLIF($3, ''))
		RETURNING enabled, created_at`

	err := s.db.QueryRowContext(ctx, q,
		p.AppID,
		p.Platform,
		p.BundleID,
	).Scan(
		&p.Enabled,
		&p.CreatedAt,
	)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}
