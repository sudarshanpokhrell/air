package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/sudarshanpokhrell/air/web"
)

// dashboardHandler serves the embedded React dashboard (web/dist).
// Unknown paths without a file extension get index.html, so client-side
// routes like /apps/example-app survive a page refresh.
func (app *application) dashboardHandler() http.Handler {
	distFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		panic(err) // the embed pattern guarantees "dist" exists
	}

	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		app.logger.Warn("dashboard not built; run `make web-build` and restart")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				app.notFoundResponse(w, r)
				return
			}
			http.Error(w, "dashboard not built: run `make web-build`", http.StatusServiceUnavailable)
		})
	}

	fileServer := http.FileServerFS(distFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unknown API routes are JSON 404s, not the dashboard.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			app.notFoundResponse(w, r)
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}

		if _, err := fs.Stat(distFS, name); err != nil {
			// A missing file (e.g. /assets/old.js) is a real 404.
			if path.Ext(name) != "" {
				http.NotFound(w, r)
				return
			}
			// A client-side route: serve the app shell.
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			name = "index.html"
		}

		// Vite puts a content hash in asset file names, so they never change.
		// index.html must always be revalidated so new deploys are picked up.
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		fileServer.ServeHTTP(w, r)
	})
}
