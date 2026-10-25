package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// The SPA build is produced by Vite (`make ui-build`) and embedded into the
// binary, so the program ships the frontend without touching the filesystem at
// run time. Every path inside ui/dist is served unchanged, which is why asset
// names carrying a content hash never appear in Go code.
//
// The pattern must stay inside this package's directory, so ui/dist is spelled
// out even though only the dist part is served.
//
//go:embed all:ui/dist
var spaDist embed.FS

const (
	// spaDistDir is the directory Vite writes to; it is the root of the served FS.
	spaDistDir = "ui/dist"
	// spaIndexFile answers the root path and is never cached, because its
	// content names the hashed assets and changes with every build.
	spaIndexFile = "index.html"
	// spaAssetsDir holds content-hashed files, so they can be cached forever.
	spaAssetsDir = "assets/"
)

// spaContentTypes is a fixed table instead of mime.TypeByExtension: the latter
// reads /etc/mime.types, so the same binary would answer differently on
// different machines.
var spaContentTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".html":  "text/html; charset=utf-8",
	".ico":   "image/x-icon",
	".js":    "text/javascript; charset=utf-8",
	".json":  "application/json",
	".map":   "application/json",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".woff2": "font/woff2",
}

// newSPAHandler returns the handler backed by the embedded build.
func newSPAHandler() (http.Handler, error) {
	fsys, err := fs.Sub(spaDist, spaDistDir)
	if err != nil {
		return nil, err
	}

	return newFileHandler(fsys), nil
}

// newFileHandler serves files from fsys. It exists as a seam: tests pass an
// fstest.MapFS, so they exercise the serving rules without the real build.
func newFileHandler(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = spaIndexFile
		}

		serveFile(w, r, fsys, name)
	})
}

// serveFile writes name from fsys, or 404 when it is absent.
//
// There is deliberately no SPA fallback: the app uses HashRouter, so the route
// lives after '#' and the server only ever sees "/". A missing path is a bug in
// the request, not a client-side route.
func serveFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	// ValidPath rejects the empty name, "." and ".." components, so a request
	// cannot escape the embedded tree. net/http hands over an already decoded
	// path, which means %2e%2e arrives here as "..".
	if !fs.ValidPath(name) {
		http.NotFound(w, r)

		return
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		// Unreadable covers both a missing file and a directory, since dist is
		// flat: directories are never servable entries on their own.
		http.NotFound(w, r)

		return
	}

	header := w.Header()
	header.Set("Content-Type", spaContentType(name))
	header.Set("Content-Length", strconv.Itoa(len(data)))
	header.Set("Cache-Control", spaCacheControl(name))

	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodHead {
		return
	}

	_, _ = w.Write(data)
}

// spaContentType guesses from the extension and falls back to octet-stream,
// which browsers download rather than render — the safe default for unknown
// files.
func spaContentType(name string) string {
	if ct, ok := spaContentTypes[strings.ToLower(path.Ext(name))]; ok {
		return ct
	}

	return "application/octet-stream"
}

// spaCacheControl allows long caching only for hashed assets: any other file
// (notably index.html) keeps its name across builds, so it must be revalidated.
func spaCacheControl(name string) string {
	if strings.HasPrefix(name, spaAssetsDir) {
		return "public, max-age=31536000, immutable"
	}

	return "no-cache"
}
