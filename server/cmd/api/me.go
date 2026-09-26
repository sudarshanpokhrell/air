package main

import (
	"net/http"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

type UpdateMePayload struct {
	Name *string `json:"name"`
}

type ChangePasswordPayload struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (app *application) getMeHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.writeJSON(w, http.StatusOK, envelope{"user": app.contextUser(r)}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMeHandler(w http.ResponseWriter, r *http.Request) {
	var payload UpdateMePayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextUser(r)
	if payload.Name != nil {
		user.Name = strings.TrimSpace(*payload.Name)
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

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/me/password
// Changes the password and logs out every other session.
func (app *application) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload ChangePasswordPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(payload.CurrentPassword != "", "current_password", "must be provided")
	store.ValidatePasswordField(v, "new_password", payload.NewPassword)
	v.Check(payload.NewPassword != payload.CurrentPassword, "new_password", "must be different from the current password")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.contextUser(r)

	ok, err := user.Password.Compare(payload.CurrentPassword)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !ok {
		app.failedValidationResponse(w, r, map[string]string{"current_password": "is incorrect"})
		return
	}

	if err := user.Password.Set(payload.NewPassword); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.store.Users.UpdatePassword(r.Context(), user); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//log out everywhere but keep this browser open
	if err := app.store.Sessions.DeleteUserSessions(r.Context(), user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.startSession(w, r, user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
