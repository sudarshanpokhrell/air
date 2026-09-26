package main

import "net/http"

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	if err := app.writeJSON(w, status, envelope{"error": message}, nil); err != nil {
		app.logger.Error("write error response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("internal error", "method", r.Method, "path", r.URL.Path, "err", err)
	app.errorResponse(w, r, http.StatusInternalServerError, "internal server error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, "the resource could not be found")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	app.errorResponse(w, r, http.StatusUnauthorized, "invalid or missing credentials")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusForbidden, "you don't have permission to perform this action")
}

func (app *application) invalidLoginResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusUnauthorized, "invalid email or password")
}

func (app *application) notImplementedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotImplemented, "not implemented yet")
}
