package web

import (
	"bytes"
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
//
// authRequired is baked into the served index.html rather than exposed through
// an endpoint: the UI needs the answer before it renders anything, and the page
// that boots it is the one place the server can hand it over without a request.
func newSPAHandler(authRequired bool, logger logger) (http.Handler, error) {
	fsys, err := fs.Sub(spaDist, spaDistDir)
	if err != nil {
		logger.Error("open embedded SPA directory failed", "dir", spaDistDir, "error", err)

		return nil, err
	}

	return newFileHandler(fsys, authRequired, logger), nil
}

// newFileHandler serves files from fsys. It exists as a seam: tests pass an
// fstest.MapFS, so they exercise the serving rules without the real build.
func newFileHandler(fsys fs.FS, authRequired bool, logger logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			logger.Warn("SPA request method not allowed", "method", r.Method, "path", r.URL.Path, "status", http.StatusMethodNotAllowed)

			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = spaIndexFile
		}

		serveFile(w, r, fsys, name, authRequired, logger)
	})
}

// serveFile writes name from fsys, or 404 when it is absent.
//
// There is deliberately no SPA fallback: the app uses HashRouter, so the route
// lives after '#' and the server only ever sees "/". A missing path is a bug in
// the request, not a client-side route.
func serveFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string, authRequired bool, logger logger) {
	// ValidPath rejects the empty name, "." and ".." components, so a request
	// cannot escape the embedded tree. net/http hands over an already decoded
	// path, which means %2e%2e arrives here as "..".
	if !fs.ValidPath(name) {
		logger.Warn("SPA path is invalid", "path", r.URL.Path, "status", http.StatusNotFound)

		http.NotFound(w, r)

		return
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		// Unreadable covers both a missing file and a directory, since dist is
		// flat: directories are never servable entries on their own.
		logger.Warn("SPA file not found", "path", name, "error", err)

		http.NotFound(w, r)

		return
	}

	// Only the entry document is rewritten: the flag is meaningless in an asset,
	// and injecting into a hashed file would break the cache promise made for it.
	if name == spaIndexFile {
		data = injectAuthFlag(data, authRequired)
	}

	header := w.Header()
	header.Set("Content-Type", spaContentType(name))
	header.Set("Content-Length", strconv.Itoa(len(data)))
	header.Set("Cache-Control", spaCacheControl(name))

	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodHead {
		return
	}

	if _, err := w.Write(data); err != nil {
		logger.Error("write SPA file failed", "path", name, "error", err)
	}
}

// spaAuthFlagScript sets the global the UI reads on boot.
//
// It is a plain assignment rather than JSON-in-attribute: the value is a bool we
// generate ourselves, so there is nothing to escape, and the UI only has to read
// one property. The name matches what `main.tsx` looks up.
const spaAuthFlagScript = `<script>window.requireAuth=%s</script>`

// injectAuthFlag inserts the flag into the document's head.
//
// The head is the target because the entry module is deferred to the body:
// anything in head runs first, so the UI never reads an undefined global. A
// document without a head is treated as malformed rather than patched, since a
// build that lost its head would not boot anyway.
func injectAuthFlag(html []byte, authRequired bool) []byte {
	value := "false"
	if authRequired {
		value = "true"
	}

	script := []byte(strings.Replace(spaAuthFlagScript, "%s", value, 1))

	const headClose = "</head>"

	at := bytes.Index(html, []byte(headClose))
	if at < 0 {
		// No head: put the flag first anyway, so a trimmed build still carries the
		// answer instead of silently defaulting to "no auth".
		return append(script, html...)
	}

	out := make([]byte, 0, len(html)+len(script))
	out = append(out, html[:at]...)
	out = append(out, script...)
	out = append(out, html[at:]...)

	return out
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
