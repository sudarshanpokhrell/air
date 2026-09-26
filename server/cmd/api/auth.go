package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sudarshanpokhrell/air/internal/store"
)

const (
	sessionCookieName = "air_session"
	sessionTTL        = 7 * 24 * time.Hour
)

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/v1/auth/login
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//TODO: rate-limit

	user, err := app.store.Users.GetUserByEmail(r.Context(), strings.TrimSpace(payload.Email))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.invalidLoginResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	ok, err := user.Password.Compare(payload.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !ok || !user.IsActive {
		app.invalidLoginResponse(w, r)
		return
	}

	if err := app.startSession(w, r, user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /api/v1/auth/logout
func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		if err := app.store.Sessions.DeleteSession(r.Context(), cookie.Value); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) startSession(w http.ResponseWriter, r *http.Request, userID string) error {
	session, err := store.NewSession(userID, sessionTTL)
	if err != nil {
		return err
	}

	if err := app.store.Sessions.CreateSession(r.Context(), session); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Plaintext,
		Path:     "/",
		Expires:  session.Expiry,
		HttpOnly: true,
		Secure:   app.config.Env != "development",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (app *application) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.config.Env != "development",
		SameSite: http.SameSiteLaxMode,
	})
}
