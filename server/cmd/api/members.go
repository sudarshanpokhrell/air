package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

type AddMemberPayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberPayload struct {
	Role string `json:"role"`
}

func validateAppRole(v *validator.Validator, role string) {
	v.Check(validator.In(role, store.RoleAdmin, store.RoleDeveloper), "role", `must be "admin" or "developer"`)
}

// GET /api/v1/apps/{slug}/members  (developer+)
func (app *application) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	members, err := app.store.Members.ListMembers(r.Context(), app.contextApp(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"members": members}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/apps/{slug}/members  (app admin)
// Adds an existing user (by email) to the app.
func (app *application) addMemberHandler(w http.ResponseWriter, r *http.Request) {
	var payload AddMemberPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	store.ValidateEmail(v, strings.TrimSpace(payload.Email))
	validateAppRole(v, payload.Role)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.store.Users.GetUserByEmail(r.Context(), strings.TrimSpace(payload.Email))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.failedValidationResponse(w, r, map[string]string{"email": "no user with this email"})
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.store.Members.AddMember(r.Context(), app.contextApp(r).ID, user.ID, payload.Role); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, "user is already a member of this app")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// PATCH /api/v1/apps/{slug}/members/{userID}  (app admin)
func (app *application) updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if !validator.UUIDRX.MatchString(userID) {
		app.notFoundResponse(w, r)
		return
	}

	var payload UpdateMemberPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	if validateAppRole(v, payload.Role); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err := app.store.Members.UpdateMemberRole(r.Context(), app.contextApp(r).ID, userID, payload.Role)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/apps/{slug}/members/{userID}  (app admin)
// The user's API keys for this app stop working immediately.
func (app *application) removeMemberHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if !validator.UUIDRX.MatchString(userID) {
		app.notFoundResponse(w, r)
		return
	}

	if err := app.store.Members.RemoveMember(r.Context(), app.contextApp(r).ID, userID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
