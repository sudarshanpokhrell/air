package main

import "net/http"

type SetCodeSigningPayload struct {
	Certificate string `json:"certificate"` // PEM, public
	KeyID       string `json:"key_id"`      // usually "main"
}

func (app *application) setCodeSigningHandler(w http.ResponseWriter, r *http.Request) {
	app.notImplementedResponse(w, r)
}
