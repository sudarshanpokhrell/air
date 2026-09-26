package main

import "net/http"

type SetCodeSigningPayload struct {
	Certificate       string `json:"certificate"`         // PEM, public
	KeyID             string `json:"key_id"`              // usually "main"
	NoUpdateDirective string `json:"no_update_directive"` // base64 of signed {"type":"noUpdateAvailable"}
	NoUpdateSignature string `json:"no_update_signature"` // expo-signature header value
}

// PUT /api/v1/apps/{slug}/code-signing  (app admin)
func (app *application) setCodeSigningHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: parse certificate → signing.Verify(noUpdate directive) → save on the app
	app.notImplementedResponse(w, r)
}
