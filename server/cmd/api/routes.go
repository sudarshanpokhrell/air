package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Middleware chains for each group of routes.
	admin := func(h http.HandlerFunc) http.Handler { return app.RequireAdmin(h) }
	adminApp := func(h http.HandlerFunc) http.Handler { return app.RequireAdmin(app.LoadApp(h)) }
	publish := func(h http.HandlerFunc) http.Handler { return app.RequireAPIKey(h) }

	// Phones (expo-updates). The URL is baked into app.json, so keep it stable.
	mux.HandleFunc("GET /api/manifest", app.manifestHandler)

	mux.HandleFunc("GET /api/v1/health", app.healthcheckHandler)

	// Admin: set up apps, platforms, API keys, code signing.
	mux.Handle("GET /api/v1/admin/apps", admin(app.listAppsHandler))
	mux.Handle("POST /api/v1/admin/apps", admin(app.createAppHandler))

	mux.Handle("POST /api/v1/admin/apps/{slug}/platforms", adminApp(app.addPlatformHandler))
	mux.Handle("PATCH /api/v1/admin/apps/{slug}/platforms/{platform}", adminApp(app.updatePlatformHandler))

	mux.Handle("GET /api/v1/admin/apps/{slug}/api-keys", adminApp(app.listAPIKeysHandler))
	mux.Handle("POST /api/v1/admin/apps/{slug}/api-keys", adminApp(app.createAPIKeyHandler))
	mux.Handle("DELETE /api/v1/admin/apps/{slug}/api-keys/{keyID}", adminApp(app.revokeAPIKeyHandler))

	mux.Handle("PUT /api/v1/admin/apps/{slug}/code-signing", adminApp(app.setCodeSigningHandler))

	// Publishing (CLI). The API key decides which app.
	mux.Handle("GET /api/v1/config", publish(app.getPublishConfigHandler))

	mux.Handle("POST /api/v1/assets/check", publish(app.checkAssetsHandler))
	mux.Handle("PUT /api/v1/assets/{hash}", publish(app.uploadAssetHandler))

	mux.Handle("GET /api/v1/updates", publish(app.listUpdatesHandler))
	mux.Handle("POST /api/v1/updates", publish(app.createUpdateHandler))
	mux.Handle("POST /api/v1/updates/rollback", publish(app.rollbackHandler))
	mux.Handle("PATCH /api/v1/updates/{groupID}", publish(app.setRolloutHandler))

	return app.recoverPanic(app.logRequest(mux))
}
