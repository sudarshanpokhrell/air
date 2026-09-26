package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

type CreateUserPayload struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

// Pointers: an omitted field is left unchanged.
type UpdateUserPayload struct {
	Name     *string `json:"name"`
	IsAdmin  *bool   `json:"is_admin"`
	IsActive *bool   `json:"is_active"`
}

// GET /api/v1/users  (global admin)
func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.store.Users.ListUsers(r.Context())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/users  (global admin)
func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Email:   strings.TrimSpace(payload.Email),
		Name:    strings.TrimSpace(payload.Name),
		IsAdmin: payload.IsAdmin,
	}

	v := validator.New()
	if store.ValidateUser(v, user, payload.Password); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.store.Users.CreateUser(r.Context(), user); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictResponse(w, r, "a user with this email already exists")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// PATCH /api/v1/users/{userID}  (global admin)
// Rename, make/remove global admin, deactivate/reactivate.
func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if !validator.UUIDRX.MatchString(userID) {
		app.notFoundResponse(w, r)
		return
	}

	var payload UpdateUserPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Admins can't remove their own admin role or deactivate themselves,
	// so the server can't end up with no admin by accident.
	self := app.contextUser(r).ID == userID
	if self && ((payload.IsAdmin != nil && !*payload.IsAdmin) || (payload.IsActive != nil && !*payload.IsActive)) {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, "you can't remove your own admin role or deactivate yourself")
		return
	}

	user, err := app.store.Users.GetUserByID(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if payload.Name != nil {
		user.Name = strings.TrimSpace(*payload.Name)
	}
	if payload.IsAdmin != nil {
		user.IsAdmin = *payload.IsAdmin
	}
	if payload.IsActive != nil {
		user.IsActive = *payload.IsActive
	}

	v := validator.New()
	if store.ValidateName(v, user.Name); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.store.Users.UpdateUser(r.Context(), user); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// A deactivated user is logged out everywhere immediately.
	// (Their API keys stop working on their own: the key lookup checks is_active.)
	if !user.IsActive {
		if err := app.store.Sessions.DeleteUserSessions(r.Context(), user.ID); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
