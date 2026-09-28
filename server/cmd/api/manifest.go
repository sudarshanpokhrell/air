package main

import "net/http"

// GET /api/manifest?app={slug}
// Called by expo-updates on the phone. Responds with multipart (internal/protocol),

func (app *application) manifestHandler(w http.ResponseWriter, r *http.Request) {
	// TODO:
	//  1. headers: expo-platform, expo-runtime-version, expo-channel-name,
	//     expo-current-update-id, eas-client-id
	//  2. slug → app (unknown → 404)
	//  3. platform registered + enabled? else no update
	//  4. store.Updates.LatestUpdates → first where rollout.InRollout
	//  5. none / same id → 204 No Content (no body, so nothing to sign for
	//     code-signed apps; fallback if clients reject it: CLI-signed
	//     noUpdateAvailable directive stored on the app)
	//     rollback_to_embedded → stored directive
	//     update → protocol.WriteManifest(stored manifest bytes, stored signature)
	app.notImplementedResponse(w, r)
}
