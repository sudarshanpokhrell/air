package main

import (
	"net/http"

	"github.com/sudarshanpokhrell/air/internal/store"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Middleware chains.
	user := func(h http.HandlerFunc) http.Handler {
		return app.RequireUser(h)
	}
	globalAdmin := func(h http.HandlerFunc) http.Handler {
		return app.RequireUser(app.RequireGlobalAdmin(h))
	}
	appDeveloper := func(h http.HandlerFunc) http.Handler {
		return app.RequireUser(app.LoadApp(app.RequireAppRole(store.RoleDeveloper)(h)))
	}
	appAdmin := func(h http.HandlerFunc) http.Handler {
		return app.RequireUser(app.LoadApp(app.RequireAppRole(store.RoleAdmin)(h)))
	}
	cli := func(h http.HandlerFunc) http.Handler {
		return app.RequireAPIKey(h)
	}

	mux.HandleFunc("GET /api/v1/health", app.healthcheckHandler)
	mux.HandleFunc("GET /api/manifest", app.manifestHandler)

	mux.HandleFunc("POST /api/v1/auth/login", app.loginHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", app.logoutHandler)

	mux.Handle("GET /api/v1/me", user(app.getMeHandler))
	mux.Handle("PATCH /api/v1/me", user(app.updateMeHandler))
	mux.Handle("POST /api/v1/me/password", user(app.changePasswordHandler))
	mux.Handle("GET /api/v1/apps", user(app.listAppsHandler))

	mux.Handle("GET /api/v1/users", globalAdmin(app.listUsersHandler))
	mux.Handle("POST /api/v1/users", globalAdmin(app.createUserHandler))
	mux.Handle("PATCH /api/v1/users/{userID}", globalAdmin(app.updateUserHandler))
	mux.Handle("POST /api/v1/apps", globalAdmin(app.createAppHandler))

	mux.Handle("GET /api/v1/apps/{slug}", appDeveloper(app.getAppHandler))
	mux.Handle("GET /api/v1/apps/{slug}/members", appDeveloper(app.listMembersHandler))
	mux.Handle("GET /api/v1/apps/{slug}/api-keys", appDeveloper(app.listAPIKeysHandler))
	mux.Handle("POST /api/v1/apps/{slug}/api-keys", appDeveloper(app.createAPIKeyHandler))
	mux.Handle("DELETE /api/v1/apps/{slug}/api-keys/{keyID}", appDeveloper(app.revokeAPIKeyHandler))
	mux.Handle("GET /api/v1/apps/{slug}/updates", appDeveloper(app.listUpdatesHandler))
	mux.Handle("POST /api/v1/apps/{slug}/updates/rollback", appDeveloper(app.rollbackHandler))
	mux.Handle("PATCH /api/v1/apps/{slug}/updates/{groupID}", appDeveloper(app.setRolloutHandler))

	mux.Handle("POST /api/v1/apps/{slug}/members", appAdmin(app.addMemberHandler))
	mux.Handle("PATCH /api/v1/apps/{slug}/members/{userID}", appAdmin(app.updateMemberHandler))
	mux.Handle("DELETE /api/v1/apps/{slug}/members/{userID}", appAdmin(app.removeMemberHandler))
	mux.Handle("POST /api/v1/apps/{slug}/platforms", appAdmin(app.addPlatformHandler))
	mux.Handle("PATCH /api/v1/apps/{slug}/platforms/{platform}", appAdmin(app.updatePlatformHandler))
	mux.Handle("PUT /api/v1/apps/{slug}/code-signing", appAdmin(app.setCodeSigningHandler))

	mux.Handle("GET /api/v1/config", cli(app.getPublishConfigHandler))
	mux.Handle("POST /api/v1/assets/check", cli(app.checkAssetsHandler))
	mux.Handle("PUT /api/v1/assets/{hash}", cli(app.uploadAssetHandler))
	mux.Handle("GET /api/v1/updates", cli(app.listUpdatesHandler))
	mux.Handle("POST /api/v1/updates", cli(app.createUpdateHandler))
	mux.Handle("POST /api/v1/updates/rollback", cli(app.rollbackHandler))
	mux.Handle("PATCH /api/v1/updates/{groupID}", cli(app.setRolloutHandler))

	return app.recoverPanic(app.logRequest(mux))
}
