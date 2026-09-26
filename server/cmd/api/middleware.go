package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sudarshanpokhrell/air/internal/auth"
	"github.com/sudarshanpokhrell/air/internal/store"
)

type contextKey string

const (
	userCtx    contextKey = "user"
	appCtx     contextKey = "app"
	appRoleCtx contextKey = "app_role"
	apiKeyCtx  contextKey = "api_key"
)

func bearerToken(r *http.Request) string {
	scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func (app *application) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)

		if err != nil || cookie.Value == "" {
			app.unauthorizedResponse(w, r)
			return
		}

		user, err := app.store.Sessions.GetUserBySession(r.Context(), cookie.Value)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.unauthorizedResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) RequireGlobalAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.contextUser(r).IsAdmin {
			app.forbiddenResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoadApp loads the app named by {slug} into the context.
func (app *application) LoadApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a, err := app.store.Apps.GetAppBySlug(r.Context(), r.PathValue("slug"))
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), appCtx, a)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAppRole allows users whose role in the app is at least min.
// Global admins act as app admins. Non-members get 404 so they can't tell
// the app exists. Use after RequireUser and LoadApp.
func (app *application) RequireAppRole(min string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := app.contextUser(r)

			role := store.RoleAdmin
			if !user.IsAdmin {
				var err error
				role, err = app.store.Members.GetMemberRole(r.Context(), app.contextApp(r).ID, user.ID)
				if err != nil {
					switch {
					case errors.Is(err, store.ErrNotFound):
						app.notFoundResponse(w, r)
					default:
						app.serverErrorResponse(w, r, err)
					}
					return
				}
			}

			if !store.RoleAtLeast(role, min) {
				app.forbiddenResponse(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), appRoleCtx, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAPIKey authenticates the CLI and puts the key (and so its app) in the
// context. The store only returns keys whose creator is active and still has
// access to the key's app.
func (app *application) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			app.unauthorizedResponse(w, r)
			return
		}

		key, err := app.store.APIKeys.GetAPIKeyByHash(r.Context(), auth.HashAPIKey(token))
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.unauthorizedResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), apiKeyCtx, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// recoverPanic turns a panic in a handler into a 500 instead of dropping the connection.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("panic: %v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// logRequest logs method, path, status and duration of every request.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		app.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
		)
	})
}
