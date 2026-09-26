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
		// TODO: friendlier messages like trackforge (syntax, type, unknown field, too large)
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

func (app *application) contextAPIKey(r *http.Request) *store.APIKey {
	k, ok := r.Context().Value(apiKeyCtx).(*store.APIKey)
	if !ok {
		panic("missing api key in request context")
	}
	return k
}
