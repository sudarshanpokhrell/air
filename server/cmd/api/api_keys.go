package main

import "net/http"

type CreateAPIKeyPayload struct {
	Name string `json:"name"`
}

// GET /api/v1/admin/apps/{slug}/api-keys
func (app *application) listAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: store.APIKeys.ListAPIKeys(contextApp.ID) → envelope{"api_keys": keys}
	app.notImplementedResponse(w, r)
}

// POST /api/v1/admin/apps/{slug}/api-keys
func (app *application) createAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: auth.GenerateAPIKey → store.APIKeys.CreateAPIKey(hash) → return the plain key ONCE
	app.notImplementedResponse(w, r)
}

// DELETE /api/v1/admin/apps/{slug}/api-keys/{keyID}
func (app *application) revokeAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: validate UUID → store.APIKeys.RevokeAPIKey → 204
	app.notImplementedResponse(w, r)
}
