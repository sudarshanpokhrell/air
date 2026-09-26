package main

import (
	"context"
	"crypto/subtle"
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
	appCtx    contextKey = "app"
	apiKeyCtx contextKey = "api_key"
)

func bearerToken(r *http.Request) string {
	scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

// RequireAdmin protects admin routes with ADMIN_TOKEN.
func (app *application) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if app.config.AdminToken == "" ||
			subtle.ConstantTimeCompare([]byte(token), []byte(app.config.AdminToken)) != 1 {
			app.unauthorizedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAPIKey authenticates the CLI and puts the key (and so its app) in the context.
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
