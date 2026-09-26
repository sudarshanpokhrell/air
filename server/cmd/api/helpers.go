package main

import (
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"

	"github.com/sudarshanpokhrell/air/internal/store"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	maps.Copy(w.Header(), headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		//TODO: Friendlier messages
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}

func (app *application) contextApp(r *http.Request) *store.App {
	a, ok := r.Context().Value(appCtx).(*store.App)
	if !ok {
		panic("missing app in request context")
	}
	return a
}

func (app *application) contextUser(r *http.Request) *store.User {
	u, ok := r.Context().Value(userCtx).(*store.User)
	if !ok {
		panic("missing user in request context")
	}
	return u
}

// contextAppRole is the caller's effective role in the app from LoadApp.
func (app *application) contextAppRole(r *http.Request) string {
	role, ok := r.Context().Value(appRoleCtx).(string)
	if !ok {
		panic("missing app role in request context")
	}
	return role
}

// requestAppID is the app a request acts on: from the API key (CLI) or
// from {slug} (dashboard). Lets one handler serve both.
func (app *application) requestAppID(r *http.Request) string {
	if k, ok := r.Context().Value(apiKeyCtx).(*store.APIKey); ok {
		return k.AppID
	}
	return app.contextApp(r).ID
}

func (app *application) contextAPIKey(r *http.Request) *store.APIKey {
	k, ok := r.Context().Value(apiKeyCtx).(*store.APIKey)
	if !ok {
		panic("missing api key in request context")
	}
	return k
}
