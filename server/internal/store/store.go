package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

// isUniqueViolation reports whether err is a Postgres unique constraint error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type Store struct {
	Apps interface {
		GetApps(ctx context.Context) ([]*App, error)
		GetAppBySlug(ctx context.Context, slug string) (*App, error)
		GetAppPlatform(ctx context.Context, appID, platform string) (*AppPlatform, error)
		CreateApp(ctx context.Context, app *App) error
		AddPlatform(ctx context.Context, p *AppPlatform) error
	}
	APIKeys interface {
		CreateAPIKey(ctx context.Context, k *APIKey, keyHash []byte) error
		GetAPIKeyByHash(ctx context.Context, keyHash []byte) (*APIKey, error)
		ListAPIKeys(ctx context.Context, appID string) ([]*APIKey, error)
		RevokeAPIKey(ctx context.Context, id string) error
	}
	Assets interface {
		InsertAsset(ctx context.Context, a *Asset) error
		GetAsset(ctx context.Context, hash string) (*Asset, error)
		MissingAssets(ctx context.Context, hashes []string) ([]string, error)
	}
	Updates interface {
		CreateUpdate(ctx context.Context, nu NewUpdate) ([]*Update, error)
		LatestUpdates(ctx context.Context, appID, channel, platform, runtimeVersion string, limit int) ([]*Update, error)
		ListUpdates(ctx context.Context, appID string, limit int) ([]*Update, error)
		GetUpdateAssets(ctx context.Context, updateID string) ([]*Asset, error)
		SetRolloutPercent(ctx context.Context, appID, groupID string, percent int) error
	}
}

func NewStore(db *sql.DB) Store {
	return Store{
		Apps:    &AppStore{db: db},
		APIKeys: &APIKeyStore{db: db},
		Assets:  &AssetStore{db: db},
		Updates: &UpdateStore{db: db},
	}
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
