package protocol

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// hash for verifying the content (verifying the download)
func AssetHash(b []byte) string {
	hash := sha256.Sum256(b)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// key for every file in manifest (which phonw uses it as file's name for cache)
func AssetKey(b []byte) string {
	key := md5.Sum(b)
	return hex.EncodeToString(key[:])
}
