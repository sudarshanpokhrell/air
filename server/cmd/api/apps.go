package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

type CreateAppPayload struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type AddPlatformPayload struct {
	Platform string `json:"platform"`
	BundleID string `json:"bundle_id"`
}

type UpdatePlatformPayload struct {
	Enabled *bool `json:"enabled"`
}

func (app *application) listAppsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextUser(r)

	var apps []*store.App
	var err error

	if user.IsAdmin {
		apps, err = app.store.Apps.GetApps(r.Context())
	} else {
		apps, err = app.store.Apps.ListAppsForUser(r.Context(), user.ID)
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"apps": apps}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/apps  (global admin)
func (app *application) createAppHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateAppPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	a := &store.App{Slug: strings.TrimSpace(payload.Slug), Name: strings.TrimSpace(payload.Name)}

	v := validator.New()
	v.Check(validator.SlugRX.MatchString(a.Slug), "slug", "must be 2-63 characters: lowercase letters, digits and dashes")
	v.Check(a.Name != "", "name", "must be provided")
	v.Check(len(a.Name) <= 100, "name", "must not be more than 100 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.store.Apps.CreateApp(r.Context(), a, app.contextUser(r).ID); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, "an app with this slug already exists")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelope{"app": a}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GET /api/v1/apps/{slug}  (developer+)
func (app *application) getAppHandler(w http.ResponseWriter, r *http.Request) {
	a := app.contextApp(r)

	// Only whether signing is on and the key id; the certificate stays out.
	codeSigning := envelope{"enabled": a.CodeSigning != nil, "key_id": ""}
	if a.CodeSigning != nil {
		codeSigning["key_id"] = a.CodeSigning.KeyID
	}

	env := envelope{
		"app":          a,
		"my_role":      app.contextAppRole(r),
		"code_signing": codeSigning,
	}
	if err := app.writeJSON(w, http.StatusOK, env, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addPlatformHandler(w http.ResponseWriter, r *http.Request) {
	var payload AddPlatformPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	p := &store.AppPlatform{
		AppID:    app.contextApp(r).ID,
		Platform: strings.TrimSpace(payload.Platform),
		BundleID: strings.TrimSpace(payload.BundleID),
	}

	v := validator.New()
	v.Check(validator.In(p.Platform, store.PlatformIOS, store.PlatformAndroid), "platform", "must be ios or android")
	v.Check(len(p.BundleID) <= 255, "bundle_id", "must not be more than 255 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.store.Apps.AddPlatform(r.Context(), p); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, "this platform is already added")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelope{"platform": p}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// PATCH /api/v1/apps/{slug}/platforms/{platform}  (app admin)
// enabled: false is the kill switch: devices on that platform stop getting updates.
func (app *application) updatePlatformHandler(w http.ResponseWriter, r *http.Request) {
	platform := r.PathValue("platform")
	if !validator.In(platform, store.PlatformIOS, store.PlatformAndroid) {
		app.notFoundResponse(w, r)
		return
	}

	var payload UpdatePlatformPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(payload.Enabled != nil, "enabled", "must be provided")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	p, err := app.store.Apps.SetPlatformEnabled(r.Context(), app.contextApp(r).ID, platform, *payload.Enabled)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"platform": p}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
