package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

const (
	maxAssetBytes  = 50 << 20
	maxCheckHashes = 5000
)

type CheckAssetsPayload struct {
	Hashes []string `json:"hashes"`
}

// POST /api/v1/assets/check
// The CLI sends every hash of a publish and uploads only the missing ones.
func (app *application) checkAssetsHandler(w http.ResponseWriter, r *http.Request) {
	var payload CheckAssetsPayload

	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(len(payload.Hashes) >= 1, "hashes", "must contain at least 1 hash")
	v.Check(len(payload.Hashes) <= maxCheckHashes, "hashes", fmt.Sprintf("must not contain more than %d hashes", maxCheckHashes))

	for i, h := range payload.Hashes {
		v.Check(validator.HashRX.MatchString(h), fmt.Sprintf("hashes[%d]", i), "must be a base64url sha256 hash")
	}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	missing, err := app.store.Assets.MissingAssets(r.Context(), payload.Hashes)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"missing": missing}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// PUT /api/v1/assets/{hash}?ext=.png
// The body is re-hashed here: phones verify the hash, so the client's is never trusted.

func (app *application) uploadAssetHandler(w http.ResponseWriter, r *http.Request) {

	hash := r.PathValue("hash")
	ext := r.URL.Query().Get("ext")
	contentType := r.Header.Get("Content-Type")

	v := validator.New()
	v.Check(validator.HashRX.MatchString(hash), "hash", "must be a base64url sha256 hash")
	v.Check(ext == "" || validator.ExtRX.MatchString(ext), "ext", "must look like .png")
	v.Check(contentType != "", "content_type", "Content-Type header must be provided")
	v.Check(len(contentType) <= 255, "content_type", "must not be more than 255 bytes long")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAssetBytes)

	asset, err := app.store.Assets.GetAsset(r.Context(), hash)
	if err == nil {
		io.Copy(io.Discard, r.Body)
		if err := app.writeJSON(w, http.StatusOK, envelope{"asset": asset}, nil); err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		app.serverErrorResponse(w, r, err)
		return
	}

	tmp, err := os.CreateTemp("", "air-asset-*")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	sha, sum := sha256.New(), md5.New()
	size, err := io.Copy(io.MultiWriter(tmp, sha, sum), r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			app.payloadTooLargeResponse(w, r)
			return
		}
		app.badRequestResponse(w, r, errors.New("could not read the request body"))
		return
	}

	if got := base64.RawURLEncoding.EncodeToString(sha.Sum(nil)); got != hash {
		app.badRequestResponse(w, r, fmt.Errorf("hash mismatch: the body hashes to %s", got))
		return
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.storage.Put(r.Context(), hash, contentType, size, tmp); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	asset = &store.Asset{
		Hash:        hash,
		Key:         hex.EncodeToString(sum.Sum(nil)),
		ContentType: contentType,
		FileExt:     ext,
		SizeBytes:   size,
	}
	if err := app.store.Assets.InsertAsset(r.Context(), asset); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelope{"asset": asset}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
