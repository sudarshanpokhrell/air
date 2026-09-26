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
	env := envelope{"app": app.contextApp(r), "my_role": app.contextAppRole(r)}
	if err := app.writeJSON(w, http.StatusOK, env, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/apps/{slug}/platforms  (app admin)
func (app *application) addPlatformHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: readJSON → validate platform → store.Apps.AddPlatform (ErrConflict → 409) → 201
	app.notImplementedResponse(w, r)
}

// PATCH /api/v1/apps/{slug}/platforms/{platform}  (app admin)
func (app *application) updatePlatformHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: enable / disable OTA for one platform
	app.notImplementedResponse(w, r)
}
