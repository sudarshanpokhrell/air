package main

import "net/http"

type PublishPlatformPayload struct {
	Platform  string `json:"platform"`
	Manifest  string `json:"manifest"`            // base64 of the exact manifest bytes (built by the CLI)
	Signature string `json:"signature,omitempty"` // only for signed apps
}

type CreateUpdatePayload struct {
	Channel   string                   `json:"channel"`
	Message   string                   `json:"message"`
	GitCommit string                   `json:"git_commit"`
	Platforms []PublishPlatformPayload `json:"platforms"`
}

type SetRolloutPayload struct {
	RolloutPercent *int `json:"rollout_percent"`
}

// GET /api/v1/config
func (app *application) getPublishConfigHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: envelope{"asset_base_url": ..., "code_signing": {"enabled": ..., "key_id": ...}}
	app.notImplementedResponse(w, r)
}

// GET /api/v1/updates
func (app *application) listUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: store.Updates.ListUpdates(contextAPIKey.AppID) → envelope{"updates": updates}
	app.notImplementedResponse(w, r)
}

// POST /api/v1/updates
func (app *application) createUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: readJSON → decode each manifest → validate (assets exist, urls, createdAt)
	//       signed app: signing.Verify; unsigned app: no signature allowed
	//       store.Updates.CreateUpdate → 201
	app.notImplementedResponse(w, r)
}

// POST /api/v1/updates/rollback
func (app *application) rollbackHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: to previous (republish old manifest's assets) or to embedded (directive from CLI)
	app.notImplementedResponse(w, r)
}

// PATCH /api/v1/updates/{groupID}
func (app *application) setRolloutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: validate UUID + 0-100 → store.Updates.SetRolloutPercent
	app.notImplementedResponse(w, r)
}
