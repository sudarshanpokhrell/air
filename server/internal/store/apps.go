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
	ID        string
	Slug      string
	Name      string
	CreatedAt time.Time
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

	var apps []*App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.Slug, &a.Name, &a.CreatedAt); err != nil {
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
	err := s.db.QueryRowContext(ctx, q, slug).Scan(&a.ID, &a.Slug, &a.Name, &a.CreatedAt)
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
	err := s.db.QueryRowContext(ctx, q, appID, platform).
		Scan(&p.AppID, &p.Platform, &p.BundleID, &p.Enabled, &p.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *AppStore) CreateApp(ctx context.Context, app *App) error {
	const q = `
		INSERT INTO apps (slug, name)
		VALUES ($1, $2)
		RETURNING id, created_at`

	err := s.db.QueryRowContext(ctx, q, app.Slug, app.Name).Scan(&app.ID, &app.CreatedAt)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func (s *AppStore) AddPlatform(ctx context.Context, p *AppPlatform) error {
	const q = `
		INSERT INTO app_platforms (app_id, platform, bundle_id)
		VALUES ($1, $2::platform, NULLIF($3, ''))
		RETURNING enabled, created_at`

	err := s.db.QueryRowContext(ctx, q, p.AppID, p.Platform, p.BundleID).Scan(&p.Enabled, &p.CreatedAt)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}
