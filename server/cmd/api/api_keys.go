package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/auth"
	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

type CreateAPIKeyPayload struct {
	Name string `json:"name"`
}

// GET /api/v1/apps/{slug}/api-keys  (developer+)
// App admins see every key; developers see their own.
func (app *application) listAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := app.store.APIKeys.ListAPIKeys(r.Context(), app.contextApp(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if app.contextAppRole(r) != store.RoleAdmin {
		userID := app.contextUser(r).ID
		own := keys[:0]
		for _, k := range keys {
			if k.CreatedBy == userID {
				own = append(own, k)
			}
		}
		keys = own
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"api_keys": keys}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/apps/{slug}/api-keys  (developer+)
// The plain key is returned only in this response; only its hash is stored.
func (app *application) createAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateAPIKeyPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	k := &store.APIKey{
		AppID:     app.contextApp(r).ID,
		Name:      strings.TrimSpace(payload.Name),
		CreatedBy: app.contextUser(r).ID,
	}

	v := validator.New()
	v.Check(k.Name != "", "name", "must be provided")
	v.Check(len(k.Name) <= 100, "name", "must not be more than 100 bytes long")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	plain, err := auth.GenerateAPIKey()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.store.APIKeys.CreateAPIKey(r.Context(), k, auth.HashAPIKey(plain)); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"api_key": k, "key": plain}
	if err := app.writeJSON(w, http.StatusCreated, env, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// DELETE /api/v1/apps/{slug}/api-keys/{keyID}  (developer+)
// Developers can revoke their own keys; app admins can revoke any.
func (app *application) revokeAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyID := r.PathValue("keyID")
	if !validator.UUIDRX.MatchString(keyID) {
		app.notFoundResponse(w, r)
		return
	}

	k, err := app.store.APIKeys.GetAPIKeyByID(r.Context(), app.contextApp(r).ID, keyID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if app.contextAppRole(r) != store.RoleAdmin && k.CreatedBy != app.contextUser(r).ID {
		app.forbiddenResponse(w, r)
		return
	}

	if err := app.store.APIKeys.RevokeAPIKey(r.Context(), k.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.conflictResponse(w, r, "key is already revoked")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
