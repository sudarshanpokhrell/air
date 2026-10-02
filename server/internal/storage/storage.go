// Package storage stores asset files in an S3-compatible bucket (Cloudflare R2).
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
