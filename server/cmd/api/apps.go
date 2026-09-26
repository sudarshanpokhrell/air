package main

import "net/http"

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

// GET /api/v1/admin/apps
func (app *application) listAppsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: store.Apps.GetApps → envelope{"apps": apps}
	app.notImplementedResponse(w, r)
}

// POST /api/v1/admin/apps
func (app *application) createAppHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: readJSON → validate slug/name → store.Apps.CreateApp (ErrConflict → 409) → 201
	app.notImplementedResponse(w, r)
}

// POST /api/v1/admin/apps/{slug}/platforms
func (app *application) addPlatformHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: contextApp → readJSON → validate platform → store.Apps.AddPlatform → 201
	app.notImplementedResponse(w, r)
}

// PATCH /api/v1/admin/apps/{slug}/platforms/{platform}
func (app *application) updatePlatformHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: enable / disable OTA for one platform
	app.notImplementedResponse(w, r)
}
