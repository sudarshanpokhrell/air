package main

import "net/http"

type CheckAssetsPayload struct {
	Hashes []string `json:"hashes"`
}

// POST /api/v1/assets/check
func (app *application) checkAssetsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: validate hashes → store.Assets.MissingAssets → envelope{"missing": missing}
	app.notImplementedResponse(w, r)
}

// PUT /api/v1/assets/{hash}  (raw file body, not JSON)
func (app *application) uploadAssetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: already stored? drain body → 200
	//       stream to temp file while hashing → hash must match {hash}
	//       storage.Put → store.Assets.InsertAsset → 201
	app.notImplementedResponse(w, r)
}
