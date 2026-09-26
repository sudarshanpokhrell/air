// Package storage stores asset files: R2 in production, local disk in development.
package storage

import (
	"context"
	"io"
)

type Storage interface {
	// Put stores an asset under its hash.
	Put(ctx context.Context, hash, contentType string, size int64, body io.ReadSeeker) error
	// PublicURL is where devices download the asset from.
	PublicURL(hash string) string
}

//TODO: Implelment s3 compatible storage
